package roomorder

import "lazeez-core/internal/users"

// canMutateRoomOrders reports roles allowed to update room-order status / cancel:
// front desk agents and room service staff. Branch managers and admins are read-only.
func canMutateRoomOrders(role string) bool {
	switch users.Role(role) {
	case users.RoleFrontDeskAgent, users.RoleRoomServiceStaff:
		return true
	default:
		return false
	}
}
