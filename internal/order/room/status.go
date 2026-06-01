package roomorder

import (
	"context"

	"lazeez-core/internal/common"
)

// validateRoomOrderStatusTransition enforces the room-order lifecycle:
// pending -> processing -> ready -> completed, with cancelled as a terminal branch.
// There is no payment gate: room orders are charged to the room bill.
func validateRoomOrderStatusTransition(currentStatus, nextStatus string) error {
	if currentStatus == nextStatus {
		return nil
	}

	switch currentStatus {
	case string(StatusPending):
		switch nextStatus {
		case string(StatusProcessing), string(StatusCancelled):
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

func (s *roomOrderService) updateOrderStatus(ctx context.Context, order *RoomOrderDTO, input RoomOrderUpdateInput) error {
	if input.OrderStatus == "" {
		return nil
	}

	if err := validateRoomOrderStatusTransition(order.OrderStatus, input.OrderStatus); err != nil {
		s.logger.Warn(
			"invalid room order status transition",
			"order_id", order.ID,
			"current_status", order.OrderStatus,
			"requested_status", input.OrderStatus,
			"error", err,
		)
		return err
	}

	if err := s.orderRepository.UpdateStatus(ctx, order.ID, input.OrderStatus, input.CancellationReason); err != nil {
		return err
	}

	// Cancelling an order removes it from the room bill total.
	if input.OrderStatus == string(StatusCancelled) {
		if err := s.recomputeFolio(ctx, order.BookingID); err != nil {
			s.logger.Error("failed to recompute folio after cancel", "booking_id", order.BookingID, "error", err)
		}
	}
	return nil
}
