package roomorder

import (
	item "lazeez-core/internal/order/Item"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *RoomOrderInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.SessionKey,
			validation.Required.Error("session key is required for room orders"),
		),
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
