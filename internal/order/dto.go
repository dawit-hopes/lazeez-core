package order

import (
	item "lazeez-core/internal/order/Item"
	"time"
)

type Status string

// Order lifecycle values (see scripts/init.sql COMMENT ON orders.order_status).
const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

const (
	PaymentStatusPending = "pending"
	PaymentStatusSuccess = "success"
	PaymentStatusFailed  = "failed"
)

type OrderInput struct {
	TableNumber   int                     `json:"table_number"`
	BranchID      string                  `json:"branch_id"`
	OrderItems    []item.OrderItemRequest `json:"order_items"`
	PaymentMethod string                  `json:"payment_method"`
	Total         float64                 `json:"total"`
	SessionKey    string                  `json:"session_key"`
}

type OrderUpdateInput struct {
	OrderStatus        string `json:"order_status"`
	CancellationReason string `json:"cancellation_reason,omitempty"`
}

type OrderDTO struct {
	ID                   string              `json:"id"`
	OrderNumber          int                 `json:"order_number"`
	TableNumber          int                 `json:"table_number"`
	BranchID             string              `json:"branch_id"`
	SessionKey           string              `json:"session_key"`
	OrderStatus          string              `json:"order_status"`
	CancellationReason   string              `json:"cancellation_reason,omitempty"`
	Total                float64             `json:"total"`
	PaymentMethod        string              `json:"payment_method"`
	PaymentStatus        string              `json:"payment_status"`
	PaymentDate          time.Time           `json:"payment_date"`
	PaymentAmount        float64             `json:"payment_amount"`
	PaymentCurrency      string              `json:"payment_currency"`
	PaymentTransactionID string              `json:"payment_transaction_id"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
	OrderItems           []item.OrderItemDTO `json:"order_items"`
}

type CreateOrderResponse struct {
	ID            string  `json:"id"`
	OrderNumber   int     `json:"order_number"`
	Total         float64 `json:"total"`
	OrderStatus   string  `json:"order_status"`
	PaymentStatus string  `json:"payment_status"`
	CheckoutURL   string  `json:"checkout_url"`
}

type PaymentWebHookPayload struct {
	Status    string `json:"status"`
	TRXRef    string `json:"trx_ref"`
	TxRef     string `json:"tx_ref"`
	RefID     string `json:"ref_id"`
	Reference string `json:"reference"`
}

func (p PaymentWebHookPayload) TransactionRef() string {
	if p.TxRef != "" {
		return p.TxRef
	}
	return p.TRXRef
}

func (p PaymentWebHookPayload) ChapaReference() string {
	if p.Reference != "" {
		return p.Reference
	}
	return p.RefID
}
