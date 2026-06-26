package users

import "lazeez-core/internal/common"

type BranchType string

const (
	BranchTypeRestaurant BranchType = "restaurant"
	BranchTypeHotel      BranchType = "hotel"
)

type UserRequest struct {
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	BranchID    string `json:"branch_id"`
	MerchantID  string `json:"merchant_id"`
	Role        Role   `json:"role"`
	// Pin is the waiter's PIN, required only when Role is waiter. Stored hashed; never returned.
	Pin string `json:"pin,omitempty"`
}

type SuperAdminUserRequest struct {
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	MerchantID  string `json:"merchant_id"`
}

type CreateUserResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	BranchID    string `json:"branch_id"`
	MerchantID  string `json:"merchant_id"`
	Role        Role   `json:"role"`
}

type SetPasswordResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	BranchID    string `json:"branch_id"`
	MerchantID  string `json:"merchant_id"`
	Role        Role   `json:"role"`
}

type UserLookUpRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type UserDTO struct {
	common.BaseDTO
	PhoneNumber     string     `json:"phone_number"`
	FullName        string     `json:"full_name"`
	Password        string     `json:"-"`
	Role            Role       `json:"role"`
	BranchID        string     `json:"branch_id"`
	MerchantID      string     `json:"merchant_id"` // Populated for branch-scoped roles and super_branch_admin from branch or users.merchant_id
	IsLocked        bool       `json:"is_locked"`
	IsFirstLogin    bool       `json:"is_first_login"`
	LoggingAttempts int        `json:"logging_attempts"`
	BranchType      BranchType `json:"branch_type,omitempty"` // Derived from merchants.branch_type at read time, not stored on users
}
