package item

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/menu"
	option "lazeez-core/internal/modifiers/option"

	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"
)

type OrderItemService interface {
	Create(ctx context.Context, orderItem OrderItemRequest) error
}

type orderItemService struct {
	orderItemRepository   OrderItemRepository
	menuService           menu.MenuService
	modifierOptionService option.ModifierOptionService
	branchService         branch.BranchService
	logger                config.Logger
}

func NewOrderItemService(orderItemRepository OrderItemRepository, menuService menu.MenuService, modifierOptionService option.ModifierOptionService, branchService branch.BranchService, logger config.Logger) OrderItemService {
	return &orderItemService{
		orderItemRepository:   orderItemRepository,
		menuService:           menuService,
		modifierOptionService: modifierOptionService,
		branchService:         branchService,
		logger:                logger,
	}
}

func (s *orderItemService) validateOrderItem(ctx context.Context, orderItem OrderItemRequest) error {
	g, ctx := errgroup.WithContext(ctx)

	// menu item
	g.Go(func() error {
		menuItem, err := s.menuService.Get(ctx, orderItem.MenuItemID, orderItem.BranchID, "")
		if err != nil {
			s.logger.Error("failed to get menu item", "error", err)
			return err
		}
		if menuItem.IsExcluded {
			s.logger.Error("menu item is excluded from branch", "menu_item_id", orderItem.MenuItemID, "branch_id", orderItem.BranchID)
			return common.ErrMenuNotFound
		}
		if !menuItem.IsAvailable {
			s.logger.Error("menu item is unavailable", "menu_item_id", orderItem.MenuItemID, "branch_id", orderItem.BranchID)
			return common.ErrMenuNotFound
		}
		if menuItem.Price != orderItem.Price {
			s.logger.Error("menu item price does not match order item price", "menu item price", menuItem.Price, "order item price", orderItem.Price)
			return common.ErrMenuPriceMismatch
		}
		return nil
	})

	// modifier options
	g.Go(func() error {
		modifierGroupGroup, modifierGroupCtx := errgroup.WithContext(ctx)
		for _, modifierOptionID := range orderItem.ModifierOptions {
			modifierOptionID := modifierOptionID
			if modifierOptionID == "" {
				continue
			}
			modifierGroupGroup.Go(func() error {
				_, err := s.modifierOptionService.Get(modifierGroupCtx, modifierOptionID)
				if err != nil {
					s.logger.Error("failed to get modifier option", "error", err)
					return err
				}
				return nil
			})
		}
		return modifierGroupGroup.Wait()
	})

	//branch
	g.Go(func() error {
		_, err := s.branchService.Get(ctx, orderItem.BranchID)
		if err != nil {
			s.logger.Error("failed to get branch", "error", err)
			return err
		}
		return nil
	})

	return g.Wait()
}

func (s *orderItemService) Create(ctx context.Context, orderItem OrderItemRequest) error {
	if err := s.validateOrderItem(ctx, orderItem); err != nil {
		s.logger.Error("failed to validate order item", "error", err)
		return err
	}

	mods := pq.StringArray(orderItem.ModifierOptions)
	if mods == nil {
		mods = pq.StringArray{}
	}

	orderItemModel := OrderItem{
		OrderID:         orderItem.OrderID,
		MenuItemID:      orderItem.MenuItemID,
		ModifierOptions: mods,
		Quantity:        orderItem.Quantity,
		Price:           orderItem.Price,
		Total:           orderItem.Total,
	}
	orderItemModel.ID = common.GenerateUUID()

	if err := s.orderItemRepository.Create(ctx, orderItemModel); err != nil {
		s.logger.Error("failed to create order item", "error", err)
		return err
	}

	return nil
}
