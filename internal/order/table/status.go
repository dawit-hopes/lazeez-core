package order

import (
	"context"

	"lazeez-core/internal/common"
	item "lazeez-core/internal/order/Item"
)

// validateOrderStatusTransition enforces the lifecycle documented in init.sql:
// pending -> processing -> ready -> completed, with cancelled as a terminal branch.
func validateOrderStatusTransition(currentStatus, nextStatus string, paymentStatus string) error {
	if currentStatus == nextStatus {
		return nil
	}

	switch currentStatus {
	case string(StatusPending):
		switch nextStatus {
		case string(StatusProcessing):
			if paymentStatus != PaymentStatusSuccess {
				return common.ErrOrderPaymentRequired
			}
			return nil
		case string(StatusCancelled):
			return nil
		}
	case string(StatusProcessing):
		switch nextStatus {
		case string(StatusReady), string(StatusCompleted), string(StatusCancelled):
			return nil
		}
	case string(StatusReady):
		switch nextStatus {
		case string(StatusCompleted), string(StatusCancelled):
			return nil
		}
	}

	return common.ErrInvalidOrderStatusTransition
}

func (s *orderService) updateOrderStatus(ctx context.Context, id string, input OrderUpdateInput) error {
	if input.OrderStatus == "" {
		return nil
	}

	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := validateOrderStatusTransition(order.OrderStatus, input.OrderStatus, order.PaymentStatus); err != nil {
		s.logger.Warn(
			"invalid order status transition",
			"order_id", id,
			"current_status", order.OrderStatus,
			"requested_status", input.OrderStatus,
			"payment_status", order.PaymentStatus,
			"error", err,
		)
		return err
	}

	return s.orderRepository.UpdateStatus(ctx, id, input.OrderStatus, input.CancellationReason, "")
}

// validateWaiterOrderStatusTransition enforces the waiter (order_source = 'waiter')
// lifecycle: placed -> in_preparation -> ready -> served, with cancelled allowed from
// any pre-served state. served is terminal.
func validateWaiterOrderStatusTransition(currentStatus, nextStatus string) error {
	if currentStatus == nextStatus {
		return nil
	}

	switch currentStatus {
	case string(StatusPlaced):
		switch nextStatus {
		case string(StatusInPreparation), string(StatusReady), string(StatusServed), string(StatusCancelled):
			return nil
		}
	case string(StatusInPreparation):
		switch nextStatus {
		case string(StatusReady), string(StatusServed), string(StatusCancelled):
			return nil
		}
	case string(StatusReady):
		switch nextStatus {
		case string(StatusServed), string(StatusCancelled):
			return nil
		}
	}

	return common.ErrInvalidOrderStatusTransition
}

// validateItemStatusTransition enforces the station lifecycle:
// sent -> accepted -> preparing -> ready, with cancelled allowed before ready.
func validateItemStatusTransition(currentStatus, nextStatus string) error {
	if currentStatus == nextStatus {
		return nil
	}

	switch currentStatus {
	case ItemStatusSent:
		switch nextStatus {
		case ItemStatusAccepted, ItemStatusPreparing, ItemStatusReady, ItemStatusCancelled:
			return nil
		}
	case ItemStatusAccepted:
		switch nextStatus {
		case ItemStatusPreparing, ItemStatusReady, ItemStatusCancelled:
			return nil
		}
	case ItemStatusPreparing:
		switch nextStatus {
		case ItemStatusReady, ItemStatusCancelled:
			return nil
		}
	}

	return common.ErrInvalidItemStatusTransition
}

// rollupOrderStatus derives an order's status from its item statuses:
// all non-cancelled items ready -> ready; any started (accepted/preparing/ready) -> in_preparation;
// otherwise placed. An all-cancelled order stays placed (it is effectively empty).
func rollupOrderStatus(items []item.OrderItemDTO) string {
	active := 0
	ready := 0
	started := 0
	for _, it := range items {
		if it.ItemStatus == ItemStatusCancelled {
			continue
		}
		active++
		if it.ItemStatus == ItemStatusReady {
			ready++
		}
		if it.ItemStatus == ItemStatusAccepted || it.ItemStatus == ItemStatusPreparing || it.ItemStatus == ItemStatusReady {
			started++
		}
	}
	if active == 0 {
		return string(StatusPlaced)
	}
	if ready == active {
		return string(StatusReady)
	}
	if started > 0 {
		return string(StatusInPreparation)
	}
	return string(StatusPlaced)
}
