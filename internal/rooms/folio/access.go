package folio

import "lazeez-core/internal/users"

// canSettleBills reports roles allowed to settle a room bill: front desk agents
// and room service staff. Managers and admins are read-only.
func canSettleBills(role string) bool {
	switch users.Role(role) {
	case users.RoleFrontDeskAgent, users.RoleRoomServiceStaff:
		return true
	default:
		return false
	}
}
