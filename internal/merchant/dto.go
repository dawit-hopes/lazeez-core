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
	Name             string               `json:"name"`
	Logo             multipart.File       `json:"logo"`
	LogoHeader       multipart.FileHeader `json:"-"`
	BranchType       BranchType           `json:"branch_type"`
	SubscriptionPlan SubscriptionPlan     `json:"subscription_plan"`
}

type MerchantDTO struct {
	common.BaseDTO
	Name          string                   `json:"name"`
	Logo          string                   `json:"logo"`
	Branches      []*branch.BranchResponse `json:"branches,omitempty"`
	Users         []*users.UserDTO         `json:"users,omitempty"`
	TotalBranches int                      `json:"total_branches"`
	TotalUsers    int                      `json:"total_users"`
	BranchType       BranchType       `json:"branch_type"`
	TaxCharges       *TaxCharges      `json:"tax_charges,omitempty"`
	SubscriptionPlan SubscriptionPlan `json:"subscription_plan"`
}

func (m *MerchantDTO) ToModel() Merchant {
	var deletedAt sql.NullTime
	if m.DeletedAt != nil {
		deletedAt = common.ToNullTime(*m.DeletedAt)
	}
	merchant := Merchant{
		Base: common.Base{
			ID:        m.ID,
			IsDeleted: m.IsDeleted,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: deletedAt,
		},
		Name:             m.Name,
		BranchType:       m.BranchType,
		Logo:             m.Logo,
		SubscriptionPlan: m.SubscriptionPlan,
	}
	if m.TaxCharges != nil {
		vatPercent := m.TaxCharges.VatPercent
		var serviceCharge sql.NullFloat64
		if m.TaxCharges.ServiceChargePercent != nil {
			serviceCharge = sql.NullFloat64{Float64: *m.TaxCharges.ServiceChargePercent, Valid: true}
		}
		applyTaxChargesToModel(&merchant, vatPercent, serviceCharge)
	}
	return merchant
}

type MerchantResponseSimplified struct {
	Name             string           `json:"name"`
	BranchType       BranchType       `json:"branch_type,omitempty"`
	Logo             string           `json:"logo,omitempty"`
	TaxCharges       *TaxCharges      `json:"tax_charges,omitempty"`
	SubscriptionPlan SubscriptionPlan `json:"subscription_plan"`
}
