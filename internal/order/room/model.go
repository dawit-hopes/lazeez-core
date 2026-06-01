package roomorder

import (
	"lazeez-core/internal/common"

	"github.com/lib/pq"
)

type RoomOrder struct {
	common.Base
	OrderNumber        int     `json:"order_number" db:"order_number"`
	RoomID             string  `json:"room_id" db:"room_id"`
	BookingID          string  `json:"booking_id" db:"booking_id"`
	BranchID           string  `json:"branch_id" db:"branch_id"`
	SessionKey         string  `json:"session_key" db:"session_key"`
	OrderStatus        string  `json:"order_status" db:"order_status"`
	CancellationReason string  `json:"cancellation_reason" db:"cancellation_reason"`
	Total              float64 `json:"total" db:"total"`
	BillID             string  `json:"bill_id" db:"bill_id"`
}

func (o *RoomOrder) Table() string {
	return "room_orders"
}

func (o *RoomOrder) Columns() []string {
	return []string{"id", "order_number", "room_id", "booking_id", "branch_id", "session_key", "order_status", "cancellation_reason", "total", "bill_id", "deleted_at", "is_deleted"}
}

func (o *RoomOrder) Values() []any {
	return []any{o.ID, o.OrderNumber, o.RoomID, o.BookingID, o.BranchID, o.SessionKey, o.OrderStatus, o.CancellationReason, o.Total, o.BillID, o.DeletedAt, o.IsDeleted}
}

func (o *RoomOrder) Addr() []any {
	return []any{&o.ID, &o.OrderNumber, &o.RoomID, &o.BookingID, &o.BranchID, &o.SessionKey, &o.OrderStatus, &o.CancellationReason, &o.Total, &o.BillID, &o.DeletedAt, &o.IsDeleted, &o.CreatedAt, &o.UpdatedAt}
}

type RoomOrderItem struct {
	common.Base
	RoomOrderID     string         `db:"room_order_id"`
	MenuItemID      string         `db:"menu_item_id"`
	ModifierOptions pq.StringArray `db:"modifier_options"`
	Quantity        int            `db:"quantity"`
	Price           float64        `db:"price"`
	Total           float64        `db:"total"`
}

func (o *RoomOrderItem) Table() string {
	return "room_order_items"
}

func (o *RoomOrderItem) Columns() []string {
	return []string{"id", "room_order_id", "menu_item_id", "modifier_options", "quantity", "price", "total", "deleted_at", "is_deleted"}
}

func (o *RoomOrderItem) Values() []any {
	return []any{o.ID, o.RoomOrderID, o.MenuItemID, o.ModifierOptions, o.Quantity, o.Price, o.Total, o.DeletedAt, o.IsDeleted}
}

func (o *RoomOrderItem) Addr() []any {
	return []any{&o.ID, &o.RoomOrderID, &o.MenuItemID, &o.ModifierOptions, &o.Quantity, &o.Price, &o.Total, &o.DeletedAt, &o.IsDeleted, &o.CreatedAt, &o.UpdatedAt}
}
