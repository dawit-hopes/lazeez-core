package folio

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type FolioRepository interface {
	Create(ctx context.Context, bill *Bill) error
	GetByBooking(ctx context.Context, bookingID string) (*Bill, error)
	Update(ctx context.Context, bill Bill) error
}

type folioRepository struct {
	dal    *common.DAL[*Bill]
	logger config.Logger
}

func NewFolioRepository(dal *common.DAL[*Bill], logger config.Logger) FolioRepository {
	return &folioRepository{dal: dal, logger: logger}
}

func (r *folioRepository) Create(ctx context.Context, bill *Bill) error {
	if _, err := r.dal.Create(ctx, bill); err != nil {
		r.logger.Error("failed to create room bill", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *folioRepository) GetByBooking(ctx context.Context, bookingID string) (*Bill, error) {
	filter := map[string]any{"booking_id": bookingID, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrBillNotFound
		}
		r.logger.Error("failed to get room bill", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *folioRepository) Update(ctx context.Context, bill Bill) error {
	filter := map[string]any{"id": bill.ID, "is_deleted": false}
	updates := map[string]any{
		"status":         bill.Status,
		"total":          bill.Total,
		"payment_method": bill.PaymentMethod,
		"settled_by":     bill.SettledBy,
		"settled_at":     bill.SettledAt,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrBillNotFound
		}
		r.logger.Error("failed to update room bill", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
