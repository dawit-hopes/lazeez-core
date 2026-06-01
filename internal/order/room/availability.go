package roomorder

import (
	"context"
	"errors"
	item "lazeez-core/internal/order/Item"

	"lazeez-core/config"
	"lazeez-core/internal/common"
)

const (
	unavailableReasonNotFound    = "not_found"
	unavailableReasonUnavailable = "unavailable"
	unavailableReasonExcluded    = "excluded"
)

// UnavailableMenuItem describes a cart line that cannot be ordered at the branch.
type UnavailableMenuItem struct {
	MenuItemID string `json:"menu_item_id"`
	Name       string `json:"name,omitempty"`
	Reason     string `json:"reason"`
}

// MenuItemsUnavailableError is returned when one or more order items are not orderable.
type MenuItemsUnavailableError struct {
	*common.Errors
	Items []UnavailableMenuItem
}

func (e *MenuItemsUnavailableError) ErrorData() any {
	return map[string]any{
		"unavailable_items": e.Items,
	}
}

func newMenuItemsUnavailableError(items []UnavailableMenuItem) *MenuItemsUnavailableError {
	return &MenuItemsUnavailableError{
		Errors: common.ErrMenuItemsUnavailable,
		Items:  dedupeUnavailableMenuItems(items),
	}
}

func dedupeUnavailableMenuItems(items []UnavailableMenuItem) []UnavailableMenuItem {
	if len(items) <= 1 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]UnavailableMenuItem, 0, len(items))
	for _, it := range items {
		if _, ok := seen[it.MenuItemID]; ok {
			continue
		}
		seen[it.MenuItemID] = struct{}{}
		out = append(out, it)
	}
	return out
}

func uniqueMenuItemIDs(orderItems []item.OrderItemRequest) []string {
	seen := make(map[string]struct{}, len(orderItems))
	ids := make([]string, 0, len(orderItems))
	for _, oi := range orderItems {
		if oi.MenuItemID == "" {
			continue
		}
		if _, ok := seen[oi.MenuItemID]; ok {
			continue
		}
		seen[oi.MenuItemID] = struct{}{}
		ids = append(ids, oi.MenuItemID)
	}
	return ids
}

func uniqueModifierOptionIDs(orderItems []item.OrderItemRequest) []string {
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for _, oi := range orderItems {
		for _, modifierOptionID := range oi.ModifierOptions {
			if modifierOptionID == "" {
				continue
			}
			if _, ok := seen[modifierOptionID]; ok {
				continue
			}
			seen[modifierOptionID] = struct{}{}
			ids = append(ids, modifierOptionID)
		}
	}
	return ids
}

func (s *roomOrderService) validateOrderBeforeCreate(ctx context.Context, branchID string, orderItems []item.OrderItemRequest) error {
	menuIDs := uniqueMenuItemIDs(orderItems)
	modifierOptionIDs := uniqueModifierOptionIDs(orderItems)

	if err := s.validateOrderMenuItems(ctx, branchID, orderItems, menuIDs); err != nil {
		logOrderValidationFailure(s.logger, branchID, err)
		return err
	}
	if err := s.validateOrderModifierOptions(ctx, modifierOptionIDs); err != nil {
		logOrderValidationFailure(s.logger, branchID, err)
		return err
	}
	return nil
}

func logOrderValidationFailure(logger config.Logger, branchID string, err error) {
	var unavailable *MenuItemsUnavailableError
	switch {
	case errors.As(err, &unavailable):
		logger.Warn(
			"room order rejected: menu items unavailable at branch",
			"branch_id", branchID,
			"unavailable_count", len(unavailable.Items),
			"unavailable_items", unavailable.Items,
		)
	case errors.Is(err, common.ErrMenuPriceMismatch):
		logger.Warn("room order rejected: menu item price mismatch", "branch_id", branchID, "error", err)
	case errors.Is(err, common.ErrModifierOptionNotFound):
		logger.Warn("room order rejected: modifier option not found", "branch_id", branchID, "error", err)
	default:
		logger.Error("room order availability validation failed", "branch_id", branchID, "error", err)
	}
}

func (s *roomOrderService) validateOrderModifierOptions(ctx context.Context, modifierOptionIDs []string) error {
	if len(modifierOptionIDs) == 0 {
		return nil
	}

	found, err := s.modifierOptionService.GetByIDs(ctx, modifierOptionIDs)
	if err != nil {
		s.logger.Error("failed to load modifier options for room order validation", "error", err)
		return err
	}

	for _, id := range modifierOptionIDs {
		if _, ok := found[id]; !ok {
			s.logger.Warn("modifier option missing during room order validation", "modifier_option_id", id)
			return common.ErrModifierOptionNotFound
		}
	}
	return nil
}

func (s *roomOrderService) validateOrderMenuItems(ctx context.Context, branchID string, orderItems []item.OrderItemRequest, menuIDs []string) error {
	if len(menuIDs) == 0 {
		return common.ErrOrderItemsRequired
	}

	snapshots, err := s.menuService.GetBranchMenuSnapshots(ctx, branchID, menuIDs)
	if err != nil {
		s.logger.Error("failed to load branch menu snapshots for room order validation", "branch_id", branchID, "error", err)
		return err
	}

	unavailable := make([]UnavailableMenuItem, 0)
	for _, oi := range orderItems {
		snap, ok := snapshots[oi.MenuItemID]
		if !ok {
			unavailable = append(unavailable, UnavailableMenuItem{MenuItemID: oi.MenuItemID, Reason: unavailableReasonNotFound})
			continue
		}
		if snap.IsExcluded {
			unavailable = append(unavailable, UnavailableMenuItem{MenuItemID: oi.MenuItemID, Name: snap.Name, Reason: unavailableReasonExcluded})
			continue
		}
		if !snap.IsAvailable {
			unavailable = append(unavailable, UnavailableMenuItem{MenuItemID: oi.MenuItemID, Name: snap.Name, Reason: unavailableReasonUnavailable})
			continue
		}
		const epsilon = 0.01
		if oi.Price < snap.Price-epsilon || oi.Price > snap.Price+epsilon {
			s.logger.Warn(
				"menu item price mismatch during room order validation",
				"branch_id", branchID,
				"menu_item_id", oi.MenuItemID,
				"expected_price", snap.Price,
				"requested_price", oi.Price,
			)
			return common.ErrMenuPriceMismatch
		}
	}

	if len(unavailable) > 0 {
		return newMenuItemsUnavailableError(dedupeUnavailableMenuItems(unavailable))
	}
	return nil
}
