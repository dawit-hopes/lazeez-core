package merchant

import (
	"database/sql"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/users"
	"mime/multipart"
)

type BranchType string

const (
	BranchTypeRestaurant BranchType = "restaurant"
	BranchTypeHotel      BranchType = "hotel"
)

type MerchantRequest struct {
	Name       string               `json:"name"`
	Logo       multipart.File       `json:"logo"`
	LogoHeader multipart.FileHeader `json:"-"`
	BranchType BranchType           `json:"branch_type"`
}

type MerchantDTO struct {
	common.BaseDTO
	Name          string                   `json:"name"`
	Logo          string                   `json:"logo"`
	Branches      []*branch.BranchResponse `json:"branches,omitempty"`
	Users         []*users.UserDTO         `json:"users,omitempty"`
	TotalBranches int                      `json:"total_branches"`
	TotalUsers    int                      `json:"total_users"`
	BranchType    BranchType               `json:"branch_type"`
}

func (m *MerchantDTO) ToModel() Merchant {
	var deletedAt sql.NullTime
	if m.DeletedAt != nil {
		deletedAt = common.ToNullTime(*m.DeletedAt)
	}
	return Merchant{
		Base: common.Base{
			ID:        m.ID,
			IsDeleted: m.IsDeleted,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: deletedAt,
		},
		Name:       m.Name,
		BranchType: m.BranchType,
		Logo:       m.Logo,
	}
}

type MerchantResponseSimplified struct {
	Name       string     `json:"name"`
	BranchType BranchType `json:"branch_type"`
	Logo       string     `json:"logo,omitempty"`
}
