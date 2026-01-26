package branch

import "lazeez-core/internal/common"

type Branch struct {
	common.Base
	MerchantID  string `json:"merchant_id" db:"merchant_id"`
	BranchName  string `json:"branch_name" db:"branch_name"`
	Address     string `json:"address" db:"address"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
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
	return []string{"id", "merchant_id", "branch_name", "address", "phone_number", "created_at", "updated_at", "deleted_at", "is_deleted"}
}

func (b *Branch) Values() []any {
	return []any{b.ID, b.MerchantID, b.BranchName, b.Address, b.PhoneNumber, b.CreatedAt, b.UpdatedAt, b.DeletedAt, b.IsDeleted}
}

func (b *Branch) Addr() []any {
	return []any{&b.ID, &b.MerchantID, &b.BranchName, &b.Address, &b.PhoneNumber, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt, &b.IsDeleted}
}
