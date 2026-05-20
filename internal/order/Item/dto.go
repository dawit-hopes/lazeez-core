package item

type OrderItemDTO struct {
	ID              string   `json:"id"`
	MenuItemID      string   `json:"menu_item_id"`
	Name            string   `json:"name,omitempty"`
	ModifierOptions []string `json:"modifier_options"`
	Quantity        int      `json:"quantity"`
	Price           float64  `json:"price"`
	Total           float64  `json:"total"`
}

type OrderItemRequest struct {
	OrderID         string   `json:"order_id,omitempty"` // Set by order service when creating
	BranchID        string   `json:"branch_id"`
	MenuItemID      string   `json:"menu_item_id"`
	ModifierOptions []string `json:"modifier_options"`
	Quantity        int      `json:"quantity"`
	Price           float64  `json:"price"`
	Total           float64  `json:"total"`
}
