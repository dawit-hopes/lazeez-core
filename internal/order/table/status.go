package order

import (
	"context"

	"lazeez-core/internal/common"
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
