package auth

type UserRequest struct {
	PhoneNumber string `json:"phone_number"`
	BranchID    string `json:"branch_id"`
	MerchantID  string `json:"merchant_id"`
}

type CreateUserResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	BranchID    string `json:"branch_id"`
	MerchantID  string `json:"merchant_id"`
	Role        Role   `json:"role"`
}

type SetPasswordRequest struct {
	Password string `json:"password"`
}

type SetPasswordResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	BranchID    string `json:"branch_id"`
	MerchantID  string `json:"merchant_id"`
	Role        Role   `json:"role"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
