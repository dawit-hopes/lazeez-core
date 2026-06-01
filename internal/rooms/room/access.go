package room

import (
	"context"
	"lazeez-core/internal/common"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/users"
)

// canManageRooms reports roles allowed to create/update/delete/regenerate physical rooms.
func canManageRooms(role string) bool {
	return users.IsSuperAdminRoleString(role) || isBranchScopedManager(role)
}

// isBranchScopedManager reports branch-bound roles that manage rooms within their own branch.
func isBranchScopedManager(role string) bool {
	return users.IsBranchManagerRoleString(role) || users.Role(role) == users.RoleFrontDeskAgent
}

// ensureHotelBranch verifies the branch exists and its merchant is a hotel; returns the merchant id.
func (s *roomService) ensureHotelBranch(ctx context.Context, branchID string) (string, error) {
	br, err := s.branchService.Get(ctx, branchID)
	if err != nil {
		return "", err
	}
	m, err := s.merchantService.Get(ctx, br.MerchantID)
	if err != nil {
		return "", err
	}
	if m.BranchType != merchant.BranchTypeHotel {
		return "", common.ErrRoomsHotelOnly
	}
	return br.MerchantID, nil
}

// resolveManageBranch determines the target branch (and its merchant) for a write,
// enforcing role rules. Branch-scoped managers act on their JWT branch; super_admin
// must pass a branch id in the request.
func (s *roomService) resolveManageBranch(ctx context.Context, role, branchID, reqBranchID string) (string, string, error) {
	if !canManageRooms(role) {
		return "", "", common.ErrUnAuthorized
	}
	target := reqBranchID
	if isBranchScopedManager(role) {
		if branchID == "" {
			return "", "", common.ErrUnAuthorized
		}
		target = branchID
	}
	if target == "" {
		return "", "", common.ErrInvalidRequest
	}
	merchantID, err := s.ensureHotelBranch(ctx, target)
	if err != nil {
		return "", "", err
	}
	return target, merchantID, nil
}

// authorizeRead enforces who may read a specific physical room.
func (s *roomService) authorizeRead(ctx context.Context, rm *Room, role, branchID, merchantID string) error {
	if users.IsSuperAdminRoleString(role) {
		return nil
	}
	if isBranchScopedManager(role) {
		if branchID != "" && rm.BranchID == branchID {
			return nil
		}
		return common.ErrUnAuthorized
	}
	if users.IsSuperBranchAdminRoleString(role) {
		if merchantID == "" {
			return common.ErrUnAuthorized
		}
		br, err := s.branchService.Get(ctx, rm.BranchID)
		if err != nil {
			return err
		}
		if br.MerchantID == merchantID {
			return nil
		}
	}
	return common.ErrUnAuthorized
}

// authorizeManage enforces who may mutate a specific physical room.
func (s *roomService) authorizeManage(rm *Room, role, branchID string) error {
	if users.IsSuperAdminRoleString(role) {
		return nil
	}
	if isBranchScopedManager(role) {
		if branchID != "" && rm.BranchID == branchID {
			return nil
		}
		return common.ErrUnAuthorized
	}
	return common.ErrUnAuthorized
}
