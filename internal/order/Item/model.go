package item

import (
	"lazeez-core/internal/common"

	"github.com/lib/pq"
)

type OrderItem struct {
	common.Base
	OrderID         string         `db:"order_id"`
	MenuItemID      string         `db:"menu_item_id"`
	ModifierOptions pq.StringArray `db:"modifier_options"`
	Quantity        int            `db:"quantity"`
	Price           float64        `db:"price"`
	Total           float64        `db:"total"`
	// Station routes this line to a preparation station ("kitchen" or "bar").
	Station string `db:"station"`
	// ItemStatus tracks fulfillment: sent -> accepted -> preparing -> ready (or cancelled).
	ItemStatus string `db:"item_status"`
}

func (o *OrderItem) Table() string {
	return "order_items"
}

func (o *OrderItem) Columns() []string {
	return []string{"id", "order_id", "menu_item_id", "modifier_options", "quantity", "price", "total", "station", "item_status", "deleted_at", "is_deleted"}
}

func (o *OrderItem) Values() []any {
	return []any{o.ID, o.OrderID, o.MenuItemID, o.ModifierOptions, o.Quantity, o.Price, o.Total, o.Station, o.ItemStatus, o.DeletedAt, o.IsDeleted}
}

func (o *OrderItem) Addr() []any {
	return []any{&o.ID, &o.OrderID, &o.MenuItemID, &o.ModifierOptions, &o.Quantity, &o.Price, &o.Total, &o.Station, &o.ItemStatus, &o.DeletedAt, &o.IsDeleted, &o.CreatedAt, &o.UpdatedAt}
}

func (o *OrderItem) ToDTO() OrderItemDTO {
	return OrderItemDTO{
		ID:              o.ID,
		MenuItemID:      o.MenuItemID,
		ModifierOptions: []string(o.ModifierOptions),
		Quantity:        o.Quantity,
		Price:           o.Price,
		Total:           o.Total,
		Station:         o.Station,
		ItemStatus:      o.ItemStatus,
	}
}
