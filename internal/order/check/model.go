package check

import (
	"database/sql"
	"lazeez-core/internal/common"
)

// TableCheck is the per-seating tab that aggregates all waiter orders for a table.
// It mirrors the hotel folio pattern but is keyed on (branch_id, table_number) and
// owns the cashier bill math (subtotal + VAT + service charge).
type TableCheck struct {
	common.Base
	BranchID            string         `db:"branch_id"`
	TableNumber         int            `db:"table_number"`
	Status              string         `db:"status"`
	Subtotal            float64        `db:"subtotal"`
	VatAmount           float64        `db:"vat_amount"`
	ServiceChargeAmount float64        `db:"service_charge_amount"`
	Total               float64        `db:"total"`
	PaymentMethod       string         `db:"payment_method"`
	SettledBy           sql.NullString `db:"settled_by"`
	SettledAt           sql.NullTime   `db:"settled_at"`
}

func (c *TableCheck) Table() string {
	return "table_checks"
}

func (c *TableCheck) Columns() []string {
	return []string{
		"id", "branch_id", "table_number", "status", "subtotal", "vat_amount",
		"service_charge_amount", "total", "payment_method", "settled_by", "settled_at",
		"deleted_at", "is_deleted",
	}
}

func (c *TableCheck) Values() []any {
	return []any{
		c.ID, c.BranchID, c.TableNumber, c.Status, c.Subtotal, c.VatAmount,
		c.ServiceChargeAmount, c.Total, c.PaymentMethod, c.SettledBy, c.SettledAt,
		c.DeletedAt, c.IsDeleted,
	}
}

func (c *TableCheck) Addr() []any {
	return []any{
		&c.ID, &c.BranchID, &c.TableNumber, &c.Status, &c.Subtotal, &c.VatAmount,
		&c.ServiceChargeAmount, &c.Total, &c.PaymentMethod, &c.SettledBy, &c.SettledAt,
		&c.DeletedAt, &c.IsDeleted, &c.CreatedAt, &c.UpdatedAt,
	}
}

func (c *TableCheck) ToDTO() CheckDTO {
	dto := CheckDTO{
		BaseDTO:             c.Base.ToDTO(),
		BranchID:            c.BranchID,
		TableNumber:         c.TableNumber,
		Status:              CheckStatus(c.Status),
		Subtotal:            c.Subtotal,
		VatAmount:           c.VatAmount,
		ServiceChargeAmount: c.ServiceChargeAmount,
		Total:               c.Total,
		PaymentMethod:       c.PaymentMethod,
	}
	if c.SettledBy.Valid {
		dto.SettledBy = c.SettledBy.String
	}
	if c.SettledAt.Valid {
		t := c.SettledAt.Time
		dto.SettledAt = &t
	}
	return dto
}
