package order

import (
	"lazeez-core/internal/common"
	"time"
)

type Order struct {
	common.Base
	OrderNumber          int       `json:"order_number" db:"order_number"`
	TableNumber          int       `json:"table_number" db:"table_number"`
	BranchID             string    `json:"branch_id" db:"branch_id"`
	SessionKey           string    `json:"session_key" db:"session_key"`
	OrderStatus          string    `json:"order_status" db:"order_status"`
	CancellationReason   string    `json:"cancellation_reason" db:"cancellation_reason"`
	Total                float64   `json:"total" db:"total"`
	PaymentMethod        string    `json:"payment_method" db:"payment_method"`
	PaymentStatus        string    `json:"payment_status" db:"payment_status"`
	PaymentDate          time.Time `json:"payment_date" db:"payment_date"`
	PaymentAmount        float64   `json:"payment_amount" db:"payment_amount"`
	PaymentCurrency      string    `json:"payment_currency" db:"payment_currency"`
	PaymentTransactionID string    `json:"payment_transaction_id" db:"payment_transaction_id"`
	OrderItems           []string  `json:"order_items" db:"order_items"`
}

func (o *Order) Table() string {
	return "orders"
}

func (o *Order) Columns() []string {
	return []string{"id", "order_number", "table_number", "branch_id", "session_key", "order_status", "cancellation_reason", "total", "payment_method", "payment_status", "payment_date", "payment_amount", "payment_currency", "payment_transaction_id", "deleted_at", "is_deleted"}
}

func (o *Order) Values() []any {
	return []any{o.ID, o.OrderNumber, o.TableNumber, o.BranchID, o.SessionKey, o.OrderStatus, o.CancellationReason, o.Total, o.PaymentMethod, o.PaymentStatus, o.PaymentDate, o.PaymentAmount, o.PaymentCurrency, o.PaymentTransactionID, o.DeletedAt, o.IsDeleted}
}

func (o *Order) Addr() []any {
	return []any{&o.ID, &o.OrderNumber, &o.TableNumber, &o.BranchID, &o.SessionKey, &o.OrderStatus, &o.CancellationReason, &o.Total, &o.PaymentMethod, &o.PaymentStatus, &o.PaymentDate, &o.PaymentAmount, &o.PaymentCurrency, &o.PaymentTransactionID, &o.DeletedAt, &o.IsDeleted, &o.CreatedAt, &o.UpdatedAt}
}

func (o *Order) ToDTO() OrderDTO {
	return OrderDTO{
		ID:                   o.ID,
		OrderNumber:          o.OrderNumber,
		TableNumber:          o.TableNumber,
		BranchID:             o.BranchID,
		SessionKey:           o.SessionKey,
		OrderStatus:          o.OrderStatus,
		CancellationReason:   o.CancellationReason,
		Total:                o.Total,
		PaymentMethod:        o.PaymentMethod,
		PaymentStatus:        o.PaymentStatus,
		PaymentDate:          o.PaymentDate,
		PaymentAmount:        o.PaymentAmount,
		PaymentCurrency:      o.PaymentCurrency,
		PaymentTransactionID: o.PaymentTransactionID,
	}
}
