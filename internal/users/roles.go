package users

// ValidRoles lists every role allowed in the database (users.role CHECK constraint).
var ValidRoles = []Role{
	RoleAdmin,
	RoleBranchManager,
	RoleSuperBranchManager,
	RoleFrontDeskAgent,
	RoleRoomServiceStaff,
	RoleWaiter,
	RoleKitchenStaff,
	RoleBarista,
	RoleCashier,
}

// IsBranchStaffRole reports roles scoped to a single branch (require branch_id and merchant at login).
func IsBranchStaffRole(r Role) bool {
	switch r {
	case RoleBranchManager, RoleFrontDeskAgent, RoleRoomServiceStaff,
		RoleWaiter, RoleKitchenStaff, RoleBarista, RoleCashier:
		return true
	default:
		return false
	}
}

// IsBranchStaffRoleString is a convenience wrapper for JWT/handler role strings.
func IsBranchStaffRoleString(role string) bool {
	return IsBranchStaffRole(Role(role))
}

// IsSuperAdminRoleString reports platform super_admin from a JWT role string.
func IsSuperAdminRoleString(role string) bool {
	return Role(role) == RoleAdmin
}

// IsSuperBranchAdminRoleString reports super_branch_admin from a JWT role string.
func IsSuperBranchAdminRoleString(role string) bool {
	return Role(role) == RoleSuperBranchManager
}

// IsBranchManagerRoleString reports branch_manager only (not other branch staff roles).
func IsBranchManagerRoleString(role string) bool {
	return Role(role) == RoleBranchManager
}

// IsBranchUserRoleString reports super_branch_admin or any branch-scoped staff role.
func IsBranchUserRoleString(role string) bool {
	return IsSuperBranchAdminRoleString(role) || IsBranchStaffRoleString(role)
}

// CanManageMerchantMaster reports whether role may mutate merchant-level master resources
// (e.g. master menu items, master room types) for the given merchant.
func CanManageMerchantMaster(role, userMerchantID, resourceMerchantID string) bool {
	if IsSuperAdminRoleString(role) {
		return true
	}
	return IsSuperBranchAdminRoleString(role) && userMerchantID != "" && userMerchantID == resourceMerchantID
}

// CreatableRoles returns roles the creator may assign when creating a user.
// branchType is only relevant when creator is branch_manager (hotel vs restaurant).
func CreatableRoles(creator Role, branchType BranchType) []Role {
	switch creator {
	case RoleSuperBranchManager:
		return []Role{RoleBranchManager}
	case RoleBranchManager:
		if branchType == BranchTypeHotel {
			return []Role{RoleFrontDeskAgent, RoleRoomServiceStaff}
		}
		return []Role{RoleWaiter, RoleKitchenStaff, RoleBarista, RoleCashier}
	default:
		return nil
	}
}

// CanCreatorAssignRole reports whether creator may create a user with targetRole.
func CanCreatorAssignRole(creator, target Role, branchType BranchType) bool {
	for _, allowed := range CreatableRoles(creator, branchType) {
		if allowed == target {
			return true
		}
	}
	return false
}

// ViewableRoles returns roles the viewer may list or fetch (direct reports they could create).
func ViewableRoles(viewer Role, branchType BranchType) []Role {
	switch viewer {
	case RoleAdmin:
		return []Role{RoleSuperBranchManager}
	default:
		return CreatableRoles(viewer, branchType)
	}
}

// CanViewerSeeUser reports whether viewer may access a user with targetRole.
func CanViewerSeeUser(viewer, target Role, branchType BranchType) bool {
	for _, allowed := range ViewableRoles(viewer, branchType) {
		if allowed == target {
			return true
		}
	}
	return false
}

func rolesToStrings(roles []Role) []string {
	out := make([]string, len(roles))
	for i, r := range roles {
		out[i] = string(r)
	}
	return out
}
