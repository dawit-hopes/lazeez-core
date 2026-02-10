package auth

import (
	"lazeez-core/internal/common"
)

type UserRequest struct {
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	BranchID    string `json:"branch_id"`
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

type SetPasswordRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type SetPasswordResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	BranchID    string `json:"branch_id"`
	MerchantID  string `json:"merchant_id"`
	Role        Role   `json:"role"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

type UserLookUpRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type UserDTO struct {
	common.BaseDTO
	PhoneNumber     string `json:"phone_number" `
	FullName        string `json:"full_name" `
	Password        string `json:"-" `
	Role            Role   `json:"role" `
	BranchID        string `json:"branch_id" `
	IsLocked        bool   `json:"is_locked" `
	IsFirstLogin    bool   `json:"is_first_login" `
	LoggingAttempts int    `json:"logging_attempts" `
}
