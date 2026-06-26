package order

import (
	"context"
	"time"

	"lazeez-core/internal/common"
	item "lazeez-core/internal/order/Item"
)

// CreateWaiter places a PIN-attributed dine-in order from the shared waiter tablet.
// branchID comes from the device token; price/total/station are resolved server-side
// and no payment is initialized (settlement happens later at the cashier check).
func (s *orderService) CreateWaiter(ctx context.Context, branchID string, input WaiterOrderInput) (*OrderDTO, error) {
	if branchID == "" {
		return nil, common.ErrUnAuthorized
	}

	waiterID, err := s.userService.ResolveWaiterByPIN(ctx, branchID, input.WaiterPIN)
	if err != nil {
		return nil, err
	}

	itemRequests, total, err := s.buildWaiterItems(ctx, branchID, input.OrderItems)
	if err != nil {
		return nil, err
	}

	checkID, err := s.checkService.GetOrCreateOpenCheck(ctx, branchID, input.TableNumber)
	if err != nil {
		s.logger.Error("failed to get or create open check", "branch_id", branchID, "table_number", input.TableNumber, "error", err)
		return nil, err
	}

	orderNumber, err := s.assignOrderNumber(ctx, branchID)
	if err != nil {
		s.logger.Error("failed to assign order number", "error", err)
		return nil, common.ErrInternalServerError
	}

	orderModel := Order{
		OrderNumber:   orderNumber,
		TableNumber:   input.TableNumber,
		BranchID:      branchID,
		Total:         total,
		OrderStatus:   string(StatusPlaced),
		PaymentStatus: PaymentStatusPending,
		PaymentMethod: "cash",
		PaymentDate:   time.Now(),
		PaymentAmount: total,
		PaymentCurrency: ETB,
		OrderSource:   OrderSourceWaiter,
	}
	orderModel.ID = common.GenerateUUID()
	orderModel.SetWaiterID(waiterID)
	orderModel.SetCheckID(checkID)

	created, err := s.orderRepository.Create(ctx, orderModel)
	if err != nil {
		s.logger.Error("failed to create waiter order", "error", err)
		return nil, common.ErrInternalServerError
	}

	for _, oi := range itemRequests {
		oi.OrderID = created.ID
		oi.BranchID = branchID
		if createErr := s.orderItemService.Create(ctx, oi); createErr != nil {
			s.logger.Error("failed to create waiter order item", "order_id", created.ID, "error", createErr)
			if delErr := s.orderRepository.Delete(ctx, created.ID); delErr != nil {
				s.logger.Error("failed to rollback waiter order after item create failure", "order_id", created.ID, "error", delErr)
			}
			return nil, createErr
		}
	}

	if err := s.checkService.Recompute(ctx, checkID); err != nil {
		s.logger.Error("failed to recompute check after waiter order", "check_id", checkID, "error", err)
	}

	return s.orderRepository.Get(ctx, created.ID)
}

// buildWaiterItems validates the cart against the branch menu and computes each line's
// price/total/station server-side, returning the item requests and the order total.
func (s *orderService) buildWaiterItems(ctx context.Context, branchID string, lines []WaiterOrderItemInput) ([]item.OrderItemRequest, float64, error) {
	if len(lines) == 0 {
		return nil, 0, common.ErrOrderItemsRequired
	}

	menuIDs := make([]string, 0, len(lines))
	modifierIDs := make([]string, 0)
	seenMenu := make(map[string]struct{})
	seenMod := make(map[string]struct{})
	for _, l := range lines {
		if l.MenuItemID != "" {
			if _, ok := seenMenu[l.MenuItemID]; !ok {
				seenMenu[l.MenuItemID] = struct{}{}
				menuIDs = append(menuIDs, l.MenuItemID)
			}
		}
		for _, m := range l.ModifierOptions {
			if m == "" {
				continue
			}
			if _, ok := seenMod[m]; !ok {
				seenMod[m] = struct{}{}
				modifierIDs = append(modifierIDs, m)
			}
		}
	}

	snapshots, err := s.menuService.GetBranchMenuSnapshots(ctx, branchID, menuIDs)
	if err != nil {
		s.logger.Error("failed to load branch menu snapshots for waiter order", "branch_id", branchID, "error", err)
		return nil, 0, err
	}

	modifiers, err := s.modifierOptionService.GetByIDs(ctx, modifierIDs)
	if err != nil {
		s.logger.Error("failed to load modifier options for waiter order", "branch_id", branchID, "error", err)
		return nil, 0, err
	}

	unavailable := make([]UnavailableMenuItem, 0)
	requests := make([]item.OrderItemRequest, 0, len(lines))
	orderTotal := 0.0

	for _, l := range lines {
		snap, ok := snapshots[l.MenuItemID]
		if !ok {
			unavailable = append(unavailable, UnavailableMenuItem{MenuItemID: l.MenuItemID, Reason: unavailableReasonNotFound})
			continue
		}
		if snap.IsExcluded {
			unavailable = append(unavailable, UnavailableMenuItem{MenuItemID: l.MenuItemID, Name: snap.Name, Reason: unavailableReasonExcluded})
			continue
		}
		if !snap.IsAvailable {
			unavailable = append(unavailable, UnavailableMenuItem{MenuItemID: l.MenuItemID, Name: snap.Name, Reason: unavailableReasonUnavailable})
			continue
		}

		unitPrice := snap.Price
		for _, modID := range l.ModifierOptions {
			mod, ok := modifiers[modID]
			if !ok {
				return nil, 0, common.ErrModifierOptionNotFound
			}
			unitPrice += mod.PriceAdjustment
		}
		lineTotal := unitPrice * float64(l.Quantity)
		orderTotal += lineTotal

		requests = append(requests, item.OrderItemRequest{
			MenuItemID:      l.MenuItemID,
			ModifierOptions: l.ModifierOptions,
			Quantity:        l.Quantity,
			Price:           unitPrice,
			Total:           lineTotal,
			Station:         resolveStation(snap.Station),
			ItemStatus:      ItemStatusSent,
		})
	}

	if len(unavailable) > 0 {
		return nil, 0, newMenuItemsUnavailableError(unavailable)
	}

	return requests, orderTotal, nil
}

// resolveStation defaults an unset effective station to the kitchen.
func resolveStation(station string) string {
	if station == StationBar {
		return StationBar
	}
	return StationKitchen
}

func (s *orderService) GetWaiter(ctx context.Context, id string, branchID string) (*OrderDTO, error) {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.BranchID != branchID || order.OrderSource != OrderSourceWaiter {
		return nil, common.ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) ListWaiter(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return s.orderRepository.ListWaiterByBranch(ctx, filter, branchID)
}

func (s *orderService) UpdateWaiter(ctx context.Context, id string, input WaiterUpdateInput, branchID string) error {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	if order.BranchID != branchID || order.OrderSource != OrderSourceWaiter {
		return common.ErrOrderNotFound
	}

	if err := validateWaiterOrderStatusTransition(order.OrderStatus, input.OrderStatus); err != nil {
		s.logger.Warn(
			"invalid waiter order status transition",
			"order_id", id,
			"current_status", order.OrderStatus,
			"requested_status", input.OrderStatus,
		)
		return err
	}

	if err := s.orderRepository.UpdateStatus(ctx, id, input.OrderStatus, input.CancellationReason, ""); err != nil {
		return err
	}

	if order.CheckID != "" {
		if err := s.checkService.Recompute(ctx, order.CheckID); err != nil {
			s.logger.Error("failed to recompute check after waiter order update", "check_id", order.CheckID, "error", err)
		}
	}
	return nil
}
