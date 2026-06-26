package item

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"

	"github.com/lib/pq"
)

type OrderItemService interface {
	Create(ctx context.Context, orderItem OrderItemRequest) error
	UpdateStatus(ctx context.Context, id, itemStatus string) error
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

	station := orderItem.Station
	if station == "" {
		station = "kitchen"
	}
	itemStatus := orderItem.ItemStatus
	if itemStatus == "" {
		itemStatus = "sent"
	}
	orderItemModel := OrderItem{
		OrderID:         orderItem.OrderID,
		MenuItemID:      orderItem.MenuItemID,
		ModifierOptions: mods,
		Quantity:        orderItem.Quantity,
		Price:           orderItem.Price,
		Total:           orderItem.Total,
		Station:         station,
		ItemStatus:      itemStatus,
	}
	orderItemModel.ID = common.GenerateUUID()

	if err := s.orderItemRepository.Create(ctx, orderItemModel); err != nil {
		s.logger.Error("failed to create order item", "error", err)
		return err
	}

	return nil
}

func (s *orderItemService) UpdateStatus(ctx context.Context, id, itemStatus string) error {
	if err := s.orderItemRepository.UpdateStatus(ctx, id, itemStatus); err != nil {
		s.logger.Error("failed to update order item status", "error", err)
		return err
	}
	return nil
}
