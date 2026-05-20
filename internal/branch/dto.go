package branch

import (
	"lazeez-core/internal/common"
)

type BranchResponse struct {
	common.BaseDTO
	MerchantID  string `json:"merchant_id"`
	BranchName  string `json:"branch_name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}


type BranchResponseSimplified struct {
	ID          string `json:"id"`
	BranchName  string `json:"branch_name"`
	Address     string `json:"address,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}
