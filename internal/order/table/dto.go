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

// Order source distinguishes guest QR orders from waiter-tablet orders.
const (
	OrderSourceClient = "client"
	OrderSourceWaiter = "waiter"
)

// Waiter order lifecycle (order_source = 'waiter'); rolled up from item statuses.
const (
	StatusPlaced        Status = "placed"
	StatusInPreparation Status = "in_preparation"
	StatusServed        Status = "served"
)

// Preparation stations an order item can route to.
const (
	StationKitchen = "kitchen"
	StationBar     = "bar"
)

// Order item fulfillment states (order_items.item_status).
const (
	ItemStatusSent      = "sent"
	ItemStatusAccepted  = "accepted"
	ItemStatusPreparing = "preparing"
	ItemStatusReady     = "ready"
	ItemStatusCancelled = "cancelled"
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

// WaiterOrderItemInput is a single waiter cart line. Price, total, and station are
// resolved server-side from the branch menu; only the selection is trusted.
type WaiterOrderItemInput struct {
	MenuItemID      string   `json:"menu_item_id"`
	ModifierOptions []string `json:"modifier_options"`
	Quantity        int      `json:"quantity"`
}

// WaiterOrderInput is a waiter-placed order from the shared tablet (PIN-attributed, no payment).
type WaiterOrderInput struct {
	TableNumber int                    `json:"table_number"`
	WaiterPIN   string                 `json:"waiter_pin"`
	OrderItems  []WaiterOrderItemInput `json:"order_items"`
}

// ItemStatusUpdateInput advances a single order item along the station lifecycle.
type ItemStatusUpdateInput struct {
	ItemStatus string `json:"item_status"`
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
	OrderSource          string              `json:"order_source,omitempty"`
	WaiterID             string              `json:"waiter_id,omitempty"`
	CheckID              string              `json:"check_id,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
	OrderItems           []item.OrderItemDTO `json:"order_items"`
}

// StationItemDTO is one open line on a station (kitchen/bar) KDS feed.
type StationItemDTO struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"order_id"`
	OrderNumber     int       `json:"order_number"`
	TableNumber     int       `json:"table_number"`
	MenuItemID      string    `json:"menu_item_id"`
	Name            string    `json:"name,omitempty"`
	Quantity        int       `json:"quantity"`
	ModifierOptions []string  `json:"modifier_options"`
	Station         string    `json:"station"`
	ItemStatus      string    `json:"item_status"`
	CreatedAt       time.Time `json:"created_at"`
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
