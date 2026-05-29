package users

// ValidRoles lists every role allowed in the database (users.role CHECK constraint).
var ValidRoles = []Role{
	RoleAdmin,
	RoleBranchManager,
	RoleSuperBranchManager,
	RoleFrontDeskAgent,
	RoleRoomServiceStaff,
}

// IsBranchStaffRole reports roles scoped to a single branch (require branch_id and merchant at login).
func IsBranchStaffRole(r Role) bool {
	switch r {
	case RoleBranchManager, RoleFrontDeskAgent, RoleRoomServiceStaff:
		return true
	default:
		return false
	}
}

// IsBranchStaffRoleString is a convenience wrapper for JWT/handler role strings.
func IsBranchStaffRoleString(role string) bool {
	return IsBranchStaffRole(Role(role))
}
