package merchant

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MerchantRepository interface {
	Create(ctx context.Context, merchant Merchant) (Merchant, error)
	Get(ctx context.Context, id string) (Merchant, error)
	Update(ctx context.Context, merchant Merchant) (Merchant, error)
	Delete(ctx context.Context, id string) error
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
		return merchant, err
	}
	return *result, nil
}

func (r *merchantRepository) Get(ctx context.Context, id string) (Merchant, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		r.logger.Error("failed to get merchant", "error", err)
		return Merchant{}, err
	}
	return *result, nil
}

func (r *merchantRepository) Update(ctx context.Context, merchant Merchant) (Merchant, error) {
	result, err := r.dal.Update(ctx, merchant.ID, &merchant)
	if err != nil {
		r.logger.Error("failed to update merchant", "error", err)
		return Merchant{}, err
	}
	return *result, nil
}

func (r *merchantRepository) Delete(ctx context.Context, id string) error {
	err := r.dal.Delete(ctx, id)
	if err != nil {
		r.logger.Error("failed to delete merchant", "error", err)
		return err
	}
	return nil
}
