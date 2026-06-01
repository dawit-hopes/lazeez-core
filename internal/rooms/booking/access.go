package booking

import (
	"context"
	"lazeez-core/internal/common"
	"lazeez-core/internal/users"
)

// canMutateBookings reports roles allowed to check guests in/out, cancel, update or
// delete bookings. Only front desk agents may mutate bookings.
func canMutateBookings(role string) bool {
	return users.Role(role) == users.RoleFrontDeskAgent
}

// isBranchScopedReader reports branch-bound roles that read bookings within their own branch.
func isBranchScopedReader(role string) bool {
	return users.IsBranchManagerRoleString(role) || users.Role(role) == users.RoleFrontDeskAgent
}

// authorizeRead enforces who may read a specific booking.
func (s *bookingService) authorizeRead(ctx context.Context, b *Booking, role, branchID, merchantID string) error {
	if users.IsSuperAdminRoleString(role) {
		return nil
	}
	if isBranchScopedReader(role) {
		if branchID != "" && b.BranchID == branchID {
			return nil
		}
		return common.ErrUnAuthorized
	}
	if users.IsSuperBranchAdminRoleString(role) {
		if merchantID == "" {
			return common.ErrUnAuthorized
		}
		br, err := s.branchService.Get(ctx, b.BranchID)
		if err != nil {
			return err
		}
		if br.MerchantID == merchantID {
			return nil
		}
	}
	return common.ErrUnAuthorized
}

// resolveListScope returns the branch and merchant filters to apply for a list request.
func (s *bookingService) resolveListScope(role, branchID, merchantID string, filter common.Filter) (string, string, error) {
	if isBranchScopedReader(role) {
		if branchID == "" {
			s.logger.Error("branch id is required", "role", role)
			return "", "", common.ErrUnAuthorized
		}
		return branchID, "", nil
	}
	if users.IsSuperBranchAdminRoleString(role) {
		if merchantID == "" {
			s.logger.Error("merchant id is required", "role", role)
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
