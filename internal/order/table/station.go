package order

import (
	"context"

	"lazeez-core/internal/common"
)

// ListStationItems returns the open KDS feed (sent/accepted/preparing lines) for a
// station within the token's branch.
func (s *orderService) ListStationItems(ctx context.Context, branchID, station string) ([]*StationItemDTO, error) {
	if station != StationKitchen && station != StationBar {
		return nil, common.ErrInvalidRequest
	}
	return s.orderRepository.ListStationItems(ctx, branchID, station)
}

// UpdateItemStatus advances a single order item along the station lifecycle, then rolls
// the parent order status up from its items and recomputes the table check.
func (s *orderService) UpdateItemStatus(ctx context.Context, itemID, branchID, station, nextStatus string) error {
	itemCtx, err := s.orderRepository.GetItemContext(ctx, itemID)
	if err != nil {
		return err
	}
	if itemCtx.BranchID != branchID || itemCtx.OrderSource != OrderSourceWaiter {
		return common.ErrOrderItemNotFound
	}
	// A station may only touch its own lines.
	if station != "" && itemCtx.Station != station {
		return common.ErrOrderItemNotFound
	}

	if err := validateItemStatusTransition(itemCtx.ItemStatus, nextStatus); err != nil {
		s.logger.Warn(
			"invalid item status transition",
			"item_id", itemID,
			"current_status", itemCtx.ItemStatus,
			"requested_status", nextStatus,
		)
		return err
	}

	if err := s.orderItemService.UpdateStatus(ctx, itemID, nextStatus); err != nil {
		return err
	}

	if err := s.recomputeOrderRollup(ctx, itemCtx.OrderID); err != nil {
		s.logger.Error("failed to roll up order status after item update", "order_id", itemCtx.OrderID, "error", err)
	}

	if itemCtx.CheckID != "" {
		if err := s.checkService.Recompute(ctx, itemCtx.CheckID); err != nil {
			s.logger.Error("failed to recompute check after item update", "check_id", itemCtx.CheckID, "error", err)
		}
	}
	return nil
}

// recomputeOrderRollup re-derives a waiter order's status from its current item statuses.
// Served orders are terminal and left untouched.
func (s *orderService) recomputeOrderRollup(ctx context.Context, orderID string) error {
	order, err := s.orderRepository.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if order.OrderStatus == string(StatusServed) || order.OrderStatus == string(StatusCancelled) {
		return nil
	}
	rolled := rollupOrderStatus(order.OrderItems)
	if rolled == order.OrderStatus {
		return nil
	}
	return s.orderRepository.UpdateStatus(ctx, orderID, rolled, "", "")
}
