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
			validation.When(r.BranchID == "" && r.SessionKey == "",
				validation.Required.Error("branch id or session key is required"),
			),
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
			validation.When(r.PaymentMethod != "",
				validation.In("cash", "card", "mobile", "other").Error("payment method must be cash, card, mobile, or other"),
			),
		),
		validation.Field(&r.Total,
			validation.Required.Error("total is required"),
			validation.Min(0.01).Error("total must be greater than 0"),
		),
	)
}

func (r *WaiterOrderInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.TableNumber,
			validation.Required.Error("table number is required"),
			validation.Min(1).Error("table number must be at least 1"),
			validation.Max(999).Error("table number must be at most 999"),
		),
		validation.Field(&r.WaiterPIN,
			validation.Required.Error("waiter pin is required"),
		),
		validation.Field(&r.OrderItems,
			validation.Required.Error("order items are required"),
			validation.Length(1, 100).Error("at least one order item is required"),
			validation.Each(validation.By(func(value interface{}) error {
				oi, ok := value.(WaiterOrderItemInput)
				if !ok {
					return nil
				}
				return validation.ValidateStruct(&oi,
					validation.Field(&oi.MenuItemID, validation.Required.Error("menu item id is required")),
					validation.Field(&oi.Quantity,
						validation.Required.Error("quantity is required"),
						validation.Min(1).Error("quantity must be at least 1"),
					),
				)
			})),
		),
	)
}

// WaiterUpdateInput is the waiter's serve/cancel request for an order they placed.
type WaiterUpdateInput struct {
	OrderStatus        string `json:"order_status"`
	CancellationReason string `json:"cancellation_reason,omitempty"`
}

func (r *WaiterUpdateInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.OrderStatus,
			validation.Required.Error("order status is required"),
			validation.In(
				string(StatusServed),
				string(StatusCancelled),
			).Error("waiter can only mark an order served or cancelled"),
		),
	)
}

func (r *ItemStatusUpdateInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.ItemStatus,
			validation.Required.Error("item status is required"),
			validation.In(
				ItemStatusAccepted,
				ItemStatusPreparing,
				ItemStatusReady,
				ItemStatusCancelled,
			).Error("invalid item status"),
		),
	)
}

func (r *OrderUpdateInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.OrderStatus,
			validation.When(r.OrderStatus != "",
				validation.In(
					string(StatusProcessing),
					string(StatusCompleted),
					string(StatusCancelled),
				).Error("invalid order status"),
			),
		),
	)
}
