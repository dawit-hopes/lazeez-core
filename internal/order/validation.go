package order

import (
	item "lazeez-core/internal/order/Item"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *OrderInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.TableNumber,
			validation.Required.Error("table number is required"),
			validation.Min(1).Error("table number must be at least 1"),
			validation.Max(999).Error("table number must be at most 999"),
		),
		validation.Field(&r.BranchID,
			validation.Required.Error("branch id is required"),
		),
		validation.Field(&r.SessionKey,
			validation.Required.Error("session key is required for client orders"),
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
		validation.Field(&r.PaymentMethod,
			validation.Required.Error("payment method is required"),
			validation.In("cash", "card", "mobile", "other").Error("payment method must be cash, card, mobile, or other"),
		),
		validation.Field(&r.Total,
			validation.Required.Error("total is required"),
			validation.Min(0.01).Error("total must be greater than 0"),
		),
		validation.Field(&r.SessionKey,
			validation.Required.Error("session key is required"),
		),
		validation.Field(&r.PaymentMethod,
			validation.Required.Error("payment method is required"),
			validation.In("cash", "card", "mobile", "other").Error("payment method must be cash, card, mobile, or other"),
		),
		validation.Field(&r.Total,
			validation.Required.Error("total is required"),
			validation.Min(0.01).Error("total must be greater than 0"),
		),
	)
}

func (r *OrderUpdateInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.OrderStatus,
			validation.When(r.OrderStatus != "",
				validation.In(string(OrderStatusPending), string(OrderStatusProcessing), string(OrderStatusReady), string(OrderStatusCompleted), string(OrderStatusCancelled)).Error("invalid order status"),
			),
		),
	)
}
