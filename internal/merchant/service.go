package merchant

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MerchantService interface {
	Create(ctx context.Context, merchant Merchant) (Merchant, error)
	Get(ctx context.Context, id string) (Merchant, error)
	Update(ctx context.Context, merchant Merchant) (Merchant, error)
	Delete(ctx context.Context, id string) error
}

type merchantService struct {
	merchantRepository MerchantRepository
	logger             config.Logger
}

func NewMerchantService(merchantRepository MerchantRepository, logger config.Logger) MerchantService {
	return &merchantService{
		merchantRepository: merchantRepository,
		logger:             logger,
	}
}

func (s *merchantService) Create(ctx context.Context, merchant Merchant) (Merchant, error) {
	s.logger.Info("Creating merchant", "merchant", merchant)
	merchant.ID = common.GenerateUUID()
	return s.merchantRepository.Create(ctx, merchant)
}

func (s *merchantService) Get(ctx context.Context, id string) (Merchant, error) {
	s.logger.Info("Getting merchant by ID", "id", id)
	return s.merchantRepository.Get(ctx, id)
}

func (s *merchantService) Update(ctx context.Context, merchant Merchant) (Merchant, error) {
	s.logger.Info("Updating merchant", "merchant", merchant)
	existingMerchant, err := s.Get(ctx, merchant.ID)
	if err != nil {
		s.logger.Error("Failed to get merchant by ID", "error", err)
		return existingMerchant, err
	}
	existingMerchant.Name = merchant.Name
	_, err = s.merchantRepository.Update(ctx, existingMerchant)
	if err != nil {
		s.logger.Error("Failed to update merchant", "error", err)
		return existingMerchant, err
	}
	return existingMerchant, nil
}

func (s *merchantService) Delete(ctx context.Context, id string) error {
	s.logger.Info("Deleting merchant", "id", id)
	existingMerchant, err := s.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get merchant by ID", "error", err)
		return err
	}
	err = s.merchantRepository.Delete(ctx, existingMerchant.ID)
	if err != nil {
		s.logger.Error("Failed to delete merchant", "error", err)
		return err
	}
	return nil
}
