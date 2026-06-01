package folio

import (
	"database/sql"
	"lazeez-core/internal/common"
)

type Bill struct {
	common.Base
	BookingID     string         `json:"booking_id" db:"booking_id"`
	BranchID      string         `json:"branch_id" db:"branch_id"`
	Status        string         `json:"status" db:"status"`
	Total         float64        `json:"total" db:"total"`
	PaymentMethod string         `json:"payment_method" db:"payment_method"`
	SettledBy     sql.NullString `json:"-" db:"settled_by"`
	SettledAt     sql.NullTime   `json:"-" db:"settled_at"`
}

func (b *Bill) Table() string {
	return "room_bills"
}

func (b *Bill) Columns() []string {
	return []string{"id", "booking_id", "branch_id", "status", "total", "payment_method", "settled_by", "settled_at", "deleted_at", "is_deleted"}
}

func (b *Bill) Values() []any {
	return []any{b.ID, b.BookingID, b.BranchID, b.Status, b.Total, b.PaymentMethod, b.SettledBy, b.SettledAt, b.DeletedAt, b.IsDeleted}
}

func (b *Bill) Addr() []any {
	return []any{&b.ID, &b.BookingID, &b.BranchID, &b.Status, &b.Total, &b.PaymentMethod, &b.SettledBy, &b.SettledAt, &b.DeletedAt, &b.IsDeleted, &b.CreatedAt, &b.UpdatedAt}
}

func (b *Bill) ToDTO() BillDTO {
	dto := BillDTO{
		BaseDTO:       b.Base.ToDTO(),
		BookingID:     b.BookingID,
		BranchID:      b.BranchID,
		Status:        BillStatus(b.Status),
		Total:         b.Total,
		PaymentMethod: b.PaymentMethod,
	}
	if b.SettledBy.Valid {
		dto.SettledBy = b.SettledBy.String
	}
	if b.SettledAt.Valid {
		t := b.SettledAt.Time
		dto.SettledAt = &t
	}
	return dto
}
