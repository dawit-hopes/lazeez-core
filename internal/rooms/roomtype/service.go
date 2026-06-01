package rooms

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/users"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error)
	GetRoom(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error)
	ListRooms(ctx context.Context, filter common.Filter, role, branchID, merchantID string, scope ListScope) (*common.PaginatedResponse[[]*RoomDTO], error)
	UpdateRoom(ctx context.Context, id string, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error)
	DeleteRoom(ctx context.Context, id, role, branchID, merchantID string) error
	CloneRoom(ctx context.Context, masterID string, req CloneRoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error)
}

type roomService struct {
	roomRepository  RoomRepository
	branchService   branch.BranchService
	merchantService merchant.MerchantService
	fileService     files.FileService
	logger          config.Logger
}

func NewRoomService(
	roomRepository RoomRepository,
	branchService branch.BranchService,
	merchantService merchant.MerchantService,
	fileService files.FileService,
	logger config.Logger,
) RoomService {
	return &roomService{
		roomRepository:  roomRepository,
		branchService:   branchService,
		merchantService: merchantService,
		fileService:     fileService,
		logger:          logger,
	}
}

func (s *roomService) CreateRoom(ctx context.Context, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error) {
	if err := s.ensureHotelAccess(ctx, role, branchID, merchantID); err != nil {
		return nil, err
	}

	isMaster := req.BranchID == "" && (users.IsSuperBranchAdminRoleString(role) || users.IsSuperAdminRoleString(role))

	if users.IsBranchManagerRoleString(role) {
		if branchID == "" {
			return nil, common.ErrUnAuthorized
		}
		req.BranchID = branchID
		isMaster = false
	} else if !users.IsSuperAdminRoleString(role) && !users.IsSuperBranchAdminRoleString(role) {
		return nil, common.ErrUnAuthorized
	}

	if isMaster {
		if merchantID == "" {
			return nil, common.ErrBranchAdminMissingMerchant
		}
		req.MerchantID = merchantID
		if err := s.roomRepository.CheckMasterExists(ctx, common.FormatText(req.Name), req.MerchantID); err != nil {
			return nil, err
		}
	} else {
		targetBranchID := req.BranchID
		if targetBranchID == "" {
			return nil, common.ErrUnAuthorized
		}
		br, err := s.branchService.Get(ctx, targetBranchID)
		if err != nil {
			return nil, err
		}
		if users.IsBranchManagerRoleString(role) && targetBranchID != branchID {
			return nil, common.ErrUnAuthorized
		}
		if users.IsSuperBranchAdminRoleString(role) && merchantID != "" && br.MerchantID != merchantID {
			return nil, common.ErrUnAuthorized
		}
		req.MerchantID = br.MerchantID
		if err := s.roomRepository.CheckBranchExists(ctx, common.FormatText(req.Name), targetBranchID); err != nil {
			return nil, err
		}
	}

	room := req.ToModel()
	room.ID = common.GenerateUUID()
	room.Name = common.FormatText(req.Name)
	if isMaster {
		room.BranchID = common.ToNUllString("")
		room.ParentID = common.ToNUllString("")
	}

	dto, err := s.roomRepository.CreateRoom(ctx, room)
	if err != nil {
		s.logger.Error("failed to create room", "error", err)
		return nil, err
	}
	enriched := enrichRoomDTO(*dto, &room, role, branchID, merchantID)
	return &enriched, nil
}

func (s *roomService) GetRoom(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error) {
	if err := s.ensureHotelAccess(ctx, role, branchID, merchantID); err != nil {
		return nil, err
	}

	room, err := s.roomRepository.GetRoom(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeRead(ctx, room, role, branchID, merchantID); err != nil {
		return nil, err
	}
	dto := enrichRoomDTO(room.ToDTO(), room, role, branchID, merchantID)
	return &dto, nil
}

func (s *roomService) ListRooms(ctx context.Context, filter common.Filter, role, branchID, merchantID string, requestedScope ListScope) (*common.PaginatedResponse[[]*RoomDTO], error) {
	if err := s.ensureHotelAccess(ctx, role, branchID, merchantID); err != nil {
		return nil, err
	}

	if !users.IsBranchManagerRoleString(role) && !users.IsSuperBranchAdminRoleString(role) && !users.IsSuperAdminRoleString(role) {
		return nil, common.ErrUnAuthorized
	}

	scope, listBranchID, err := s.resolveListScope(ctx, role, branchID, merchantID, requestedScope, filter)
	if err != nil {
		return nil, err
	}

	result, err := s.roomRepository.ListScoped(ctx, filter, scope, listBranchID, merchantID)
	if err != nil {
		return nil, err
	}

	for i, dto := range result.Data {
		if dto == nil {
			continue
		}
		room := roomFromDTO(dto)
		enriched := enrichRoomDTO(*dto, room, role, branchID, merchantID)
		result.Data[i] = &enriched
	}
	return result, nil
}

func (s *roomService) resolveListScope(ctx context.Context, role, branchID, merchantID string, requestedScope ListScope, filter common.Filter) (ListScope, string, error) {
	if users.IsBranchManagerRoleString(role) {
		if branchID == "" {
			return "", "", common.ErrUnAuthorized
		}
		return ScopeBranchManage, branchID, nil
	}

	listBranchID := branchID
	if listBranchID == "" {
		listBranchID = filterBranchID(filter)
	}

	if users.IsSuperBranchAdminRoleString(role) {
		if requestedScope == ScopeBranchManage || listBranchID != "" {
			if listBranchID == "" {
				return "", "", common.ErrUnAuthorized
			}
			if err := s.ensureBranchInMerchant(ctx, listBranchID, merchantID); err != nil {
				return "", "", err
			}
			return ScopeBranchManage, listBranchID, nil
		}
		if requestedScope != "" && requestedScope != ScopeMaster {
			return "", "", common.ErrUnAuthorized
		}
		if merchantID == "" {
			return "", "", common.ErrBranchAdminMissingMerchant
		}
		return ScopeMaster, "", nil
	}

	if users.IsSuperAdminRoleString(role) {
		if requestedScope == ScopeBranchManage || listBranchID != "" {
			if listBranchID == "" {
				return "", "", common.ErrUnAuthorized
			}
			return ScopeBranchManage, listBranchID, nil
		}
		if requestedScope != "" && requestedScope != ScopeMaster {
			return "", "", common.ErrUnAuthorized
		}
		return ScopeMaster, "", nil
	}

	return "", "", common.ErrUnAuthorized
}

func roomFromDTO(dto *RoomDTO) *Room {
	room := &Room{
		Base:          common.Base{ID: dto.ID},
		Name:          dto.Name,
		Description:   dto.Description,
		MerchantID:    dto.MerchantID,
		PricePerNight: dto.PricePerNight,
	}
	if dto.BranchID != "" {
		room.BranchID = common.ToNUllString(dto.BranchID)
	}
	if dto.ParentID != "" {
		room.ParentID = common.ToNUllString(dto.ParentID)
	}
	return room
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

func (s *roomService) UpdateRoom(ctx context.Context, id string, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error) {
	if err := s.ensureHotelAccess(ctx, role, branchID, merchantID); err != nil {
		return nil, err
	}

	existing, err := s.roomRepository.GetRoom(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeMutate(existing, role, branchID, merchantID); err != nil {
		return nil, err
	}

	if existing.IsClone() {
		if req.Name != "" || req.Description != "" {
			return nil, common.ErrRoomClonePriceOnly
		}
		if req.PricePerNight <= 0 {
			return nil, common.ErrInvalidRequest
		}
		existing.PricePerNight = req.PricePerNight
		if err := s.roomRepository.UpdateRoom(ctx, id, *existing); err != nil {
			return nil, err
		}
		return s.GetRoom(ctx, id, role, branchID, merchantID)
	}

	if req.Name != "" && req.Name != existing.Name {
		name := common.FormatText(req.Name)
		if existing.IsMaster() {
			if err := s.roomRepository.CheckMasterExists(ctx, name, existing.MerchantID); err != nil {
				return nil, err
			}
		} else if err := s.roomRepository.CheckBranchExists(ctx, name, existing.BranchIDString()); err != nil {
			return nil, err
		}
		existing.Name = name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.PricePerNight > 0 {
		existing.PricePerNight = req.PricePerNight
	}

	if err := s.roomRepository.UpdateRoom(ctx, id, *existing); err != nil {
		return nil, err
	}
	return s.GetRoom(ctx, id, role, branchID, merchantID)
}

func (s *roomService) DeleteRoom(ctx context.Context, id, role, branchID, merchantID string) error {
	if err := s.ensureHotelAccess(ctx, role, branchID, merchantID); err != nil {
		return err
	}

	existing, err := s.roomRepository.GetRoom(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorizeMutate(existing, role, branchID, merchantID); err != nil {
		return err
	}
	return s.roomRepository.DeleteRoom(ctx, id)
}

func (s *roomService) CloneRoom(ctx context.Context, masterID string, req CloneRoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error) {
	if err := s.ensureHotelAccess(ctx, role, branchID, merchantID); err != nil {
		return nil, err
	}

	if !users.IsBranchManagerRoleString(role) || branchID == "" {
		return nil, common.ErrUnAuthorized
	}

	master, err := s.roomRepository.GetMaster(ctx, masterID, "")
	if err != nil {
		return nil, err
	}
	if merchantID != "" && master.MerchantID != merchantID {
		return nil, common.ErrUnAuthorized
	}
	br, err := s.branchService.Get(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if br.MerchantID != master.MerchantID {
		return nil, common.ErrUnAuthorized
	}
	if err := s.roomRepository.CheckCloneExists(ctx, branchID, masterID); err != nil {
		return nil, err
	}

	clone := Room{
		Name:          master.Name,
		Description:   master.Description,
		MerchantID:    master.MerchantID,
		BranchID:      common.ToNUllString(branchID),
		ParentID:      common.ToNUllString(masterID),
		PricePerNight: master.PricePerNight,
	}
	if req.PricePerNight > 0 {
		clone.PricePerNight = req.PricePerNight
	}
	clone.ID = common.GenerateUUID()

	dto, err := s.roomRepository.CreateRoom(ctx, clone)
	if err != nil {
		s.logger.Error("failed to clone room", "error", err)
		return nil, err
	}
	enriched := enrichRoomDTO(*dto, &clone, role, branchID, merchantID)
	return &enriched, nil
}

func (s *roomService) authorizeRead(ctx context.Context, room *Room, role, branchID, merchantID string) error {
	if users.IsSuperAdminRoleString(role) {
		s.logger.Info("super admin access to room", "role", role, "branch_id", branchID, "merchant_id", merchantID)
		return nil
	}
	if room.IsMaster() {
		if users.CanManageMerchantMaster(role, merchantID, room.MerchantID) {
			return nil
		}
		if users.IsBranchManagerRoleString(role) && branchID != "" {
			br, err := s.branchService.Get(ctx, branchID)
			if err != nil {
				s.logger.Error("failed to get branch", "error", err)
				return err
			}
			if br.MerchantID == room.MerchantID {
				return nil
			}
		}
		s.logger.Error("unauthorized access to master room", "role", role, "branch_id", branchID, "merchant_id", merchantID)
		return common.ErrUnAuthorized
	}
	if users.IsBranchManagerRoleString(role) && room.BranchIDString() == branchID {
		return nil
	}
	if users.CanManageMerchantMaster(role, merchantID, room.MerchantID) {
		return nil
	}
	return common.ErrUnAuthorized
}

func (s *roomService) authorizeMutate(room *Room, role, branchID, merchantID string) error {
	if room.IsMaster() {
		if users.IsBranchManagerRoleString(role) {
			s.logger.Error("unauthorized access to master room", "role", role, "branch_id", branchID, "merchant_id", merchantID)
			return common.ErrUnAuthorized
		}
		if !users.CanManageMerchantMaster(role, merchantID, room.MerchantID) {
			s.logger.Error("unauthorized access to master room", "role", role, "branch_id", branchID, "merchant_id", merchantID)
			return common.ErrUnAuthorized
		}
		return nil
	}
	if users.IsBranchManagerRoleString(role) {
		if branchID == "" || room.BranchIDString() != branchID {
			s.logger.Error("unauthorized access to branch room", "role", role, "branch_id", branchID, "merchant_id", merchantID)
			return common.ErrUnAuthorized
		}
		return nil
	}
	if users.CanManageMerchantMaster(role, merchantID, room.MerchantID) {
		return nil
	}
	return common.ErrUnAuthorized
}
