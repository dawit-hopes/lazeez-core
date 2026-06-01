package room

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/merchant"
	roomtype "lazeez-core/internal/rooms/roomtype"
	"lazeez-core/internal/users"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error)
	GetRoom(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error)
	ListRooms(ctx context.Context, filter common.Filter, role, branchID, merchantID string) (*common.PaginatedResponse[[]*RoomDTO], error)
	UpdateRoom(ctx context.Context, id string, req RoomUpdateRequestDTO, role, branchID, merchantID string) (*RoomDTO, error)
	DeleteRoom(ctx context.Context, id, role, branchID, merchantID string) error
	RegenerateQRCode(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error)
	GetRoomByReference(ctx context.Context, reference string) (*RoomDTO, error)
	UpdateRoomStatus(ctx context.Context, id, status string) error
	GetRoomById(ctx context.Context, id string) (*RoomDTO, error)
}

type roomService struct {
	repository      RoomRepository
	roomTypeRepo    roomtype.RoomRepository
	branchService   branch.BranchService
	merchantService merchant.MerchantService
	fileService     files.FileService
	logger          config.Logger
}

func NewRoomService(
	repository RoomRepository,
	roomTypeRepo roomtype.RoomRepository,
	branchService branch.BranchService,
	merchantService merchant.MerchantService,
	fileService files.FileService,
	logger config.Logger,
) RoomService {
	return &roomService{
		repository:      repository,
		roomTypeRepo:    roomTypeRepo,
		branchService:   branchService,
		merchantService: merchantService,
		fileService:     fileService,
		logger:          logger,
	}
}

func (s *roomService) generateQRCode(ctx context.Context, reference string) (string, error) {
	file, err := common.GenerateQRCodeHeader(reference, "room")
	if err != nil {
		s.logger.Error("failed to generate room QR code", "error", err)
		return "", err
	}
	return s.fileService.UploadFile(ctx, file)
}

// validateRoomType ensures the room type exists and is usable by the target branch/merchant.
func (s *roomService) validateRoomType(ctx context.Context, roomTypeID, branchID, merchantID string) error {
	rt, err := s.roomTypeRepo.GetRoom(ctx, roomTypeID)
	if err != nil {
		return err
	}
	if rt.IsMaster() {
		if rt.MerchantID == merchantID {
			return nil
		}
		return common.ErrUnAuthorized
	}
	if rt.BranchIDString() == branchID {
		return nil
	}
	return common.ErrUnAuthorized
}

func (s *roomService) CreateRoom(ctx context.Context, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error) {
	targetBranch, branchMerchantID, err := s.resolveManageBranch(ctx, role, branchID, req.BranchID)
	if err != nil {
		return nil, err
	}

	if err := s.validateRoomType(ctx, req.RoomTypeID, targetBranch, branchMerchantID); err != nil {
		s.logger.Error("invalid room type for branch", "error", err)
		return nil, err
	}

	roomNumber := common.FormatText(req.RoomNumber)
	if err := s.repository.CheckRoomNumberExists(ctx, targetBranch, roomNumber); err != nil {
		return nil, err
	}

	rm := req.ToModel()
	rm.ID = common.GenerateUUID()
	rm.RoomNumber = roomNumber
	rm.BranchID = targetBranch
	rm.Reference = common.GenerateReference()
	rm.QRVersion = 1

	qrURL, err := s.generateQRCode(ctx, rm.Reference)
	if err != nil {
		return nil, err
	}
	rm.QRCode = qrURL

	if err := s.repository.Create(ctx, &rm); err != nil {
		s.logger.Error("failed to create single room", "error", err)
		return nil, err
	}
	return s.roomDTO(ctx, rm.ID)
}

func (s *roomService) GetRoom(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error) {
	rm, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeRead(ctx, &rm.Room, role, branchID, merchantID); err != nil {
		return nil, err
	}
	dto := rm.ToDTO()
	return &dto, nil
}

func (s *roomService) ListRooms(ctx context.Context, filter common.Filter, role, branchID, merchantID string) (*common.PaginatedResponse[[]*RoomDTO], error) {
	listBranchID, listMerchantID, err := s.resolveListScope(role, branchID, merchantID, filter)
	if err != nil {
		return nil, err
	}

	result, err := s.repository.List(ctx, filter, listBranchID, listMerchantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*RoomDTO, len(result.Data))
	for i, rm := range result.Data {
		dto := rm.ToDTO()
		dtos[i] = &dto
	}
	return &common.PaginatedResponse[[]*RoomDTO]{
		Data: dtos,
		Meta: result.Meta,
	}, nil
}

func (s *roomService) roomDTO(ctx context.Context, id string) (*RoomDTO, error) {
	rm, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := rm.ToDTO()
	return &dto, nil
}

// resolveListScope returns the branch and merchant filters to apply for a list request.
func (s *roomService) resolveListScope(role, branchID, merchantID string, filter common.Filter) (string, string, error) {
	if isBranchScopedManager(role) {
		if branchID == "" {
			s.logger.Error("branch id is required", "role", role)
			return "", "", common.ErrUnAuthorized
		}
		return branchID, "", nil
	}
	if users.IsSuperBranchAdminRoleString(role) {
		if merchantID == "" {
			return "", "", common.ErrBranchAdminMissingMerchant
		}
		if bid := filterBranchID(filter); bid != "" {
			return bid, merchantID, nil
		}
		return "", merchantID, nil
	}
	if users.IsSuperAdminRoleString(role) {
		return filterBranchID(filter), "", nil
	}
	
	return "", "", common.ErrUnAuthorized
}

func filterBranchID(filter common.Filter) string {
	if filter.Filter == nil {
		return ""
	}
	if v, ok := filter.Filter["branch_id"].(string); ok {
		return v
	}
	return ""
}

func (s *roomService) UpdateRoom(ctx context.Context, id string, req RoomUpdateRequestDTO, role, branchID, merchantID string) (*RoomDTO, error) {
	if !canManageRooms(role) {
		return nil, common.ErrUnAuthorized
	}

	rm, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeManage(&rm.Room, role, branchID); err != nil {
		return nil, err
	}

	if req.RoomNumber != "" {
		roomNumber := common.FormatText(req.RoomNumber)
		if roomNumber != rm.Room.RoomNumber {
			if err := s.repository.CheckRoomNumberExists(ctx, rm.BranchID, roomNumber); err != nil {
				return nil, err
			}
			rm.RoomNumber = roomNumber
		}
	}
	if req.RoomTypeID != "" && req.RoomTypeID != rm.RoomTypeID {
		branchMerchantID, err := s.ensureHotelBranch(ctx, rm.BranchID)
		if err != nil {
			return nil, err
		}
		if err := s.validateRoomType(ctx, req.RoomTypeID, rm.BranchID, branchMerchantID); err != nil {
			return nil, err
		}
		rm.RoomTypeID = req.RoomTypeID
	}
	if req.Floor != nil {
		rm.Floor = *req.Floor
	}

	if err := s.repository.Update(ctx, rm.Room); err != nil {
		return nil, err
	}
	return s.roomDTO(ctx, id)
}

func (s *roomService) DeleteRoom(ctx context.Context, id, role, branchID, merchantID string) error {
	if !canManageRooms(role) {
		return common.ErrUnAuthorized
	}
	rm, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorizeManage(&rm.Room, role, branchID); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func (s *roomService) RegenerateQRCode(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error) {
	if !canManageRooms(role) {
		return nil, common.ErrUnAuthorized
	}
	rm, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeManage(&rm.Room, role, branchID); err != nil {
		return nil, err
	}
	if rm.Reference == "" {
		rm.Reference = common.GenerateReference()
	}
	qrURL, err := s.generateQRCode(ctx, rm.Reference)
	if err != nil {
		return nil, err
	}
	rm.QRCode = qrURL
	rm.QRVersion++
	if err := s.repository.Update(ctx, rm.Room); err != nil {
		return nil, err
	}
	return s.roomDTO(ctx, id)
}

func (s *roomService) GetRoomByReference(ctx context.Context, reference string) (*RoomDTO, error) {
	rm, err := s.repository.GetByReference(ctx, reference)
	if err != nil {
		return nil, err
	}
	return s.roomDTO(ctx, rm.ID)
}

func (s *roomService) UpdateRoomStatus(ctx context.Context, id, status string) error {
	return s.repository.UpdateStatus(ctx, id, status)
}

func (s *roomService) GetRoomById(ctx context.Context, id string) (*RoomDTO, error) {
	rm, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := rm.ToDTO()
	return &dto, nil
}