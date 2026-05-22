package item

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"

	"github.com/lib/pq"
)

type OrderItemService interface {
	Create(ctx context.Context, orderItem OrderItemRequest) error
}

type orderItemService struct {
	orderItemRepository OrderItemRepository
	logger              config.Logger
}

func NewOrderItemService(orderItemRepository OrderItemRepository, logger config.Logger) OrderItemService {
	return &orderItemService{
		orderItemRepository: orderItemRepository,
		logger:              logger,
	}
}

func (s *orderItemService) Create(ctx context.Context, orderItem OrderItemRequest) error {
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
