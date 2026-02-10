package merchant

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MerchantRepository interface {
	Create(ctx context.Context, merchant Merchant) (Merchant, error)
	Get(ctx context.Context, id string) (Merchant, error)
	Update(ctx context.Context, merchant Merchant) (Merchant, error)
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*Merchant, error)
	CheckExists(ctx context.Context, name string) error
}

type merchantRepository struct {
	dal    *common.DAL[*Merchant]
	logger config.Logger
}

func NewMerchantRepository(dal *common.DAL[*Merchant], logger config.Logger) MerchantRepository {
	return &merchantRepository{dal: dal, logger: logger}
}

func (r *merchantRepository) Create(ctx context.Context, merchant Merchant) (Merchant, error) {
	result, err := r.dal.Create(ctx, &merchant)
	if err != nil {
		r.logger.Error("failed to create merchant", "error", err)
		return merchant, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *merchantRepository) Get(ctx context.Context, id string) (Merchant, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("merchant not found", "error", err)
			return Merchant{}, common.ErrMerchantNotFound
		}
		r.logger.Error("failed to get merchant", "error", err)
		return Merchant{}, err
	}
	return *result, nil
}

func (r *merchantRepository) Update(ctx context.Context, merchant Merchant) (Merchant, error) {
	result, err := r.dal.Update(ctx, merchant.ID, &merchant)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("merchant not found", "error", err)
			return Merchant{}, common.ErrMerchantNotFound
		}
		r.logger.Error("failed to update merchant", "error", err)
		return Merchant{}, err
	}
	return *result, nil
}

func (r *merchantRepository) Delete(ctx context.Context, id string) error {
	err := r.dal.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("merchant not found", "error", err)
			return common.ErrMerchantNotFound
		}
		r.logger.Error("failed to delete merchant", "error", err)
		return err
	}
	return nil
}

func (r *merchantRepository) GetAll(ctx context.Context) ([]*Merchant, error) {
	// Get all non-deleted merchants (limit=0 means get all, handled by DAL)
	results, err := r.dal.List(ctx, map[string]any{}, 0, 0)
	if err != nil {
		r.logger.Error("failed to get all merchants", "error", err)
		return nil, err
	}
	return results, nil
}

func (r *merchantRepository) CheckExists(ctx context.Context, name string) error {
	filter := map[string]any{"name": name}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		r.logger.Error("failed to check if merchant exists", "error", err)
		return err
	}

	// If merchant exists, return error
	if result.ID != "" {
		return common.ErrMerchantAlreadyExists
	}
	return nil
}
