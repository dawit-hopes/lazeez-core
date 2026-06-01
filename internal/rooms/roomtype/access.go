package rooms

import (
	"context"
	"lazeez-core/internal/common"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/users"
)

func computeCanEdit(room *Room, role, branchID, merchantID string) bool {
	if room.IsMaster() {
		return users.CanManageMerchantMaster(role, merchantID, room.MerchantID)
	}
	if users.IsBranchManagerRoleString(role) {
		return branchID != "" && room.BranchIDString() == branchID
	}
	return users.CanManageMerchantMaster(role, merchantID, room.MerchantID)
}

func enrichRoomDTO(dto RoomDTO, room *Room, role, branchID, merchantID string) RoomDTO {
	dto.IsMaster = room.IsMaster()
	dto.IsClone = room.IsClone()
	dto.CanEdit = computeCanEdit(room, role, branchID, merchantID)
	return dto
}

func (s *roomService) ensureHotelAccess(ctx context.Context, role, branchID, merchantID string) error {
	if users.IsSuperAdminRoleString(role) {
		return nil
	}

	if users.IsBranchManagerRoleString(role) {
		if branchID == "" {
			return common.ErrUnAuthorized
		}
		br, err := s.branchService.Get(ctx, branchID)
		if err != nil {
			return err
		}
		targetMerchantID := merchantID
		if targetMerchantID == "" {
			targetMerchantID = br.MerchantID
		} else if targetMerchantID != br.MerchantID {
			return common.ErrUnAuthorized
		}
		m, err := s.merchantService.Get(ctx, targetMerchantID)
		if err != nil {
			return err
		}
		if m.BranchType != merchant.BranchTypeHotel {
			return common.ErrRoomsHotelOnly
		}
		return nil
	}

	if merchantID == "" {
		return common.ErrBranchAdminMissingMerchant
	}

	m, err := s.merchantService.Get(ctx, merchantID)
	if err != nil {
		return err
	}
	if m.BranchType != merchant.BranchTypeHotel {
		return common.ErrRoomsHotelOnly
	}
	return nil
}

func (s *roomService) ensureBranchInMerchant(ctx context.Context, branchID, merchantID string) error {
	br, err := s.branchService.Get(ctx, branchID)
	if err != nil {
		return err
	}
	if merchantID != "" && br.MerchantID != merchantID {
		return common.ErrUnAuthorized
	}
	return nil
}
