package item

import (
	"lazeez-core/internal/common"
)

type OrderItem struct {
	common.Base
	OrderID         string   `db:"order_id"`
	MenuItemID      string   `db:"menu_item_id"`
	ModifierOptions []string `db:"modifier_options"`
	Quantity        int      `db:"quantity"`
	Price           float64  `db:"price"`
	Total           float64  `db:"total"`
}

func (o *OrderItem) Table() string {
	return "order_items"
}

func (o *OrderItem) Columns() []string {
	return []string{"id", "order_id", "menu_item_id", "modifier_options", "quantity", "price", "total", "deleted_at", "is_deleted"}
}

func (o *OrderItem) Values() []any {
	return []any{o.ID, o.OrderID, o.MenuItemID, o.ModifierOptions, o.Quantity, o.Price, o.Total, o.DeletedAt, o.IsDeleted}
}

func (o *OrderItem) Addr() []any {
	return []any{&o.ID, &o.OrderID, &o.MenuItemID, &o.ModifierOptions, &o.Quantity, &o.Price, &o.Total, &o.DeletedAt, &o.IsDeleted, &o.CreatedAt, &o.UpdatedAt}
}

func (o *OrderItem) ToDTO() OrderItemDTO {
	return OrderItemDTO{
		ID:              o.ID,
		MenuItemID:      o.MenuItemID,
		ModifierOptions: o.ModifierOptions,
		Quantity:        o.Quantity,
		Price:           o.Price,
		Total:           o.Total,
	}
}
