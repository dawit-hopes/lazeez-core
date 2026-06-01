package roomorder

import (
	item "lazeez-core/internal/order/Item"
	"time"
)

type Status string

// Room order lifecycle. Unlike table orders there is no payment gate; orders are
// charged to the room bill, so pending can advance to processing directly.
const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type RoomOrderInput struct {
	OrderItems []item.OrderItemRequest `json:"order_items"`
	Total      float64                 `json:"total"`
	PassCode   string                  `json:"pass_code"`
	Reference  string                  `json:"reference"`
}

type RoomOrderUpdateInput struct {
	OrderStatus        string `json:"order_status"`
	CancellationReason string `json:"cancellation_reason,omitempty"`
}

type RoomOrderDTO struct {
	ID                 string              `json:"id"`
	OrderNumber        int                 `json:"order_number"`
	RoomID             string              `json:"room_id"`
	BookingID          string              `json:"booking_id"`
	BranchID           string              `json:"branch_id"`
	SessionKey         string              `json:"session_key"`
	OrderStatus        string              `json:"order_status"`
	CancellationReason string              `json:"cancellation_reason,omitempty"`
	Total              float64             `json:"total"`
	BillID             string              `json:"bill_id"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	OrderItems         []item.OrderItemDTO `json:"order_items"`
}

type CreateRoomOrderResponse struct {
	ID          string  `json:"id"`
	OrderNumber int     `json:"order_number"`
	Total       float64 `json:"total"`
	OrderStatus string  `json:"order_status"`
	BillID      string  `json:"bill_id"`
}
