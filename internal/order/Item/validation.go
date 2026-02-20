package item

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *OrderItemRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.BranchID,
			validation.Required.Error("branch id is required"),
		),
		validation.Field(&r.MenuItemID,
			validation.Required.Error("menu item id is required"),
		),
		validation.Field(&r.Quantity,
			validation.Required.Error("quantity is required"),
			validation.Min(1).Error("quantity must be at least 1"),
			validation.Max(99).Error("quantity must be at most 99"),
		),
		validation.Field(&r.Price,
			validation.Min(0.0).Error("price must be greater than or equal to 0"),
		),
		validation.Field(&r.Total,
			validation.Min(0.0).Error("total must be greater than or equal to 0"),
		),
	)
}
