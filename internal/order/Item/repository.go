package item

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type OrderItemRepository interface {
	Create(ctx context.Context, orderItem OrderItem) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, itemStatus string) error
}

type orderItemRepository struct {
	dal    *common.DAL[*OrderItem]
	logger config.Logger
}

func NewOrderItemRepository(dal *common.DAL[*OrderItem], logger config.Logger) OrderItemRepository {
	return &orderItemRepository{dal: dal, logger: logger}
}

func (r *orderItemRepository) Create(ctx context.Context, orderItem OrderItem) error {
	_, err := r.dal.Create(ctx, &orderItem)
	if err != nil {
		r.logger.Error("failed to create order item", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *orderItemRepository) Delete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": true,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrOrderItemNotFound
		}
		r.logger.Error("failed to delete order item", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *orderItemRepository) UpdateStatus(ctx context.Context, id, itemStatus string) error {
	filter := map[string]any{"id": id, "is_deleted": false}
	updates := map[string]any{"item_status": itemStatus}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrOrderItemNotFound
		}
		r.logger.Error("failed to update order item status", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *orderItemRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrOrderItemNotFound
		}
		r.logger.Error("failed to undelete order item", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
