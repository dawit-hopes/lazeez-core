package branch

import "lazeez-core/internal/common"

type Branch struct {
	common.Base
	MerchantID  string `json:"merchant_id" db:"merchant_id"`
	BranchName  string `json:"branch_name" db:"branch_name"`
	Address     string `json:"address" db:"address"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
}

type CreateBranchRequest struct {
	MerchantID  string `json:"merchant_id"`
	BranchName  string `json:"branch_name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type UpdateBranchRequest struct {
	BranchName  string `json:"branch_name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type BranchStats struct {
	TotalOrdersToday int     `json:"total_orders_today"`
	ActiveTables     int     `json:"active_tables"`
	RevenueToday     float64 `json:"revenue_today"`
}

func (b *Branch) Table() string {
	return "branches"
}

func (b *Branch) Columns() []string {
	return []string{"id", "merchant_id", "branch_name", "address", "phone_number", "deleted_at", "is_deleted"}
}

func (b *Branch) Values() []any {
	return []any{b.ID, b.MerchantID, b.BranchName, b.Address, b.PhoneNumber, b.DeletedAt, b.IsDeleted}
}

func (b *Branch) Addr() []any {
	return []any{&b.ID, &b.MerchantID, &b.BranchName, &b.Address, &b.PhoneNumber, &b.DeletedAt, &b.IsDeleted, &b.CreatedAt, &b.UpdatedAt}
}

func (b *Branch) ToDTO() BranchResponse {
	return BranchResponse{
		BaseDTO: common.BaseDTO{
			ID:        b.ID,
			IsDeleted: b.IsDeleted,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(b.DeletedAt),
		},
		MerchantID:  b.MerchantID,
		BranchName:  b.BranchName,
		Address:     b.Address,
		PhoneNumber: b.PhoneNumber,
	}
}
