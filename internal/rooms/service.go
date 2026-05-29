package rooms

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/users"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error)
	GetRoom(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error)
	ListRooms(ctx context.Context, filter common.Filter, role, branchID, merchantID string, scope ListScope) (*common.PaginatedResponse[[]*RoomDTO], error)
	UpdateRoom(ctx context.Context, id string, req RoomRequestDTO, role, branchID, merchantID string) error
	DeleteRoom(ctx context.Context, id, role, branchID, merchantID string) error
	CloneRoom(ctx context.Context, masterID string, req CloneRoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error)
}

type roomService struct {
	roomRepository RoomRepository
	branchService  branch.BranchService
	logger         config.Logger
}

func NewRoomService(roomRepository RoomRepository, branchService branch.BranchService, logger config.Logger) RoomService {
	return &roomService{
		roomRepository: roomRepository,
		branchService:  branchService,
		logger:         logger,
	}
}

func isSuperAdmin(role string) bool {
	return role == string(users.RoleAdmin)
}

func isSuperBranchAdmin(role string) bool {
	return role == string(users.RoleSuperBranchManager)
}

func isBranchManager(role string) bool {
	return users.IsBranchStaffRoleString(role) && role == string(users.RoleBranchManager)
}

func canManageMaster(role, userMerchantID, roomMerchantID string) bool {
	if isSuperAdmin(role) {
		return true
	}
	return isSuperBranchAdmin(role) && userMerchantID != "" && userMerchantID == roomMerchantID
}

func (s *roomService) CreateRoom(ctx context.Context, req RoomRequestDTO, role, branchID, merchantID string) (*RoomDTO, error) {
	isMaster := req.BranchID == "" && (isSuperBranchAdmin(role) || isSuperAdmin(role))

	if isBranchManager(role) {
		if branchID == "" {
			return nil, common.ErrUnAuthorized
		}
		req.BranchID = branchID
		isMaster = false
	} else if !isSuperAdmin(role) && !isSuperBranchAdmin(role) {
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
		if isBranchManager(role) && targetBranchID != branchID {
			return nil, common.ErrUnAuthorized
		}
		if isSuperBranchAdmin(role) && merchantID != "" && br.MerchantID != merchantID {
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
	return dto, nil
}

func (s *roomService) GetRoom(ctx context.Context, id, role, branchID, merchantID string) (*RoomDTO, error) {
	room, err := s.roomRepository.GetRoom(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeRead(ctx, room, role, branchID, merchantID); err != nil {
		return nil, err
	}
	dto := room.ToDTO()
	return &dto, nil
}

func (s *roomService) ListRooms(ctx context.Context, filter common.Filter, role, branchID, merchantID string, scope ListScope) (*common.PaginatedResponse[[]*RoomDTO], error) {
	if scope == "" {
		switch {
		case isBranchManager(role):
			scope = ScopeBranchManage
		case isSuperBranchAdmin(role) || isSuperAdmin(role):
			if branchID != "" || filterBranchID(filter) != "" {
				scope = ScopeBranchManage
			} else {
				scope = ScopeMaster
			}
		default:
			return nil, common.ErrUnAuthorized
		}
	}

	listBranchID := branchID
	if listBranchID == "" {
		listBranchID = filterBranchID(filter)
	}

	if scope == ScopeMaster && merchantID == "" {
		return nil, common.ErrBranchAdminMissingMerchant
	}
	if scope == ScopeBranchManage && listBranchID == "" {
		return nil, common.ErrUnAuthorized
	}
	if !isBranchManager(role) && !isSuperBranchAdmin(role) && !isSuperAdmin(role) {
		return nil, common.ErrUnAuthorized
	}

	return s.roomRepository.ListScoped(ctx, filter, scope, listBranchID, merchantID)
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

func (s *roomService) UpdateRoom(ctx context.Context, id string, req RoomRequestDTO, role, branchID, merchantID string) error {
	existing, err := s.roomRepository.GetRoom(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorizeMutate(existing, role, branchID, merchantID); err != nil {
		return err
	}

	if req.Name != "" && req.Name != existing.Name {
		name := common.FormatText(req.Name)
		if existing.IsMaster() {
			if err := s.roomRepository.CheckMasterExists(ctx, name, existing.MerchantID); err != nil {
				return err
			}
		} else if err := s.roomRepository.CheckBranchExists(ctx, name, existing.BranchIDString()); err != nil {
			return err
		}
		existing.Name = name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.PricePerNight > 0 {
		existing.PricePerNight = req.PricePerNight
	}

	return s.roomRepository.UpdateRoom(ctx, id, *existing)
}

func (s *roomService) DeleteRoom(ctx context.Context, id, role, branchID, merchantID string) error {
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
	if !isBranchManager(role) || branchID == "" {
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
	if req.Name != "" {
		clone.Name = common.FormatText(req.Name)
	}
	if req.Description != "" {
		clone.Description = req.Description
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
	dto.IsMaster = false
	dto.IsClone = true
	dto.CanEdit = true
	return dto, nil
}

func (s *roomService) authorizeRead(ctx context.Context, room *Room, role, branchID, merchantID string) error {
	if isSuperAdmin(role) {
		return nil
	}
	if room.IsMaster() {
		if canManageMaster(role, merchantID, room.MerchantID) {
			return nil
		}
		if isBranchManager(role) && branchID != "" {
			br, err := s.branchService.Get(ctx, branchID)
			if err != nil {
				return err
			}
			if br.MerchantID == room.MerchantID {
				return nil
			}
		}
		return common.ErrUnAuthorized
	}
	if isBranchManager(role) && room.BranchIDString() == branchID {
		return nil
	}
	if canManageMaster(role, merchantID, room.MerchantID) {
		return nil
	}
	return common.ErrUnAuthorized
}

func (s *roomService) authorizeMutate(room *Room, role, branchID, merchantID string) error {
	if room.IsMaster() {
		if isBranchManager(role) {
			return common.ErrUnAuthorized
		}
		if !canManageMaster(role, merchantID, room.MerchantID) {
			return common.ErrUnAuthorized
		}
		return nil
	}
	if isBranchManager(role) {
		if branchID == "" || room.BranchIDString() != branchID {
			return common.ErrUnAuthorized
		}
		return nil
	}
	if canManageMaster(role, merchantID, room.MerchantID) {
		return nil
	}
	return common.ErrUnAuthorized
}
