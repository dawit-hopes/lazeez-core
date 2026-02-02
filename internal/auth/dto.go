package auth

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
