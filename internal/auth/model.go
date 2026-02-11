package auth

import "lazeez-core/internal/users"

type AuthPayload struct {
	UserID   string     `json:"uid"`
	BranchID string     `json:"bid,omitempty"`
	Role     users.Role `json:"rol"`
}
