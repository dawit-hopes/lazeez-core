package roomorder

import (
	item "lazeez-core/internal/order/Item"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *RoomOrderInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.OrderItems,
			validation.Required.Error("order items are required"),
			validation.Length(1, 100).Error("at least one order item is required"),
			validation.Each(validation.By(func(value interface{}) error {
				if oi, ok := value.(item.OrderItemRequest); ok {
					return oi.Validate()
				}
				return nil
			})),
		),
		validation.Field(&r.Total,
			validation.Required.Error("total is required"),
			validation.Min(0.01).Error("total must be greater than 0"),
		),
		validation.Field(&r.Reference,
			validation.Required.Error("reference is required"),
			validation.By(func(value any) error {
				ref, _ := value.(string)
				if strings.TrimSpace(ref) == "" {
					return validation.NewError("validation_reference", "reference is required")
				}
				return nil
			}),
		),
		validation.Field(&r.PassCode,
			validation.Required.Error("pass code is required"),
			validation.Length(4, 12).Error("pass code must be between 4 and 12 digits"),
		),
	)
}

func (r *RoomOrderUpdateInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.OrderStatus,
			validation.When(r.OrderStatus != "",
				validation.In(
					string(StatusProcessing),
					string(StatusReady),
					string(StatusCompleted),
					string(StatusCancelled),
				).Error("invalid order status"),
			),
		),
	)
}
