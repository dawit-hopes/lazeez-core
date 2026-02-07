package merchant

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MerchantService interface {
	Create(ctx context.Context, req MerchantRequest) (Merchant, error)
	Get(ctx context.Context, id string) (Merchant, error)
	Update(ctx context.Context, id string, req MerchantRequest) (Merchant, error)
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*Merchant, error)
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

func (s *merchantService) Create(ctx context.Context, req MerchantRequest) (Merchant, error) {
	merchant := Merchant{
		Name: req.Name,
	}
	s.logger.Info("Creating merchant", "merchant", merchant)
	merchant.ID = common.GenerateUUID()
	err := s.merchantRepository.CheckExists(ctx, merchant.Name)
	if err != nil {
		s.logger.Error("Failed to check if merchant exists", "error", err)
		return merchant, err
	}

	if req.Logo != nil {
		// we will upload the image to the cloud storage and update the image url in the database
	}

	newMerchant, err := s.merchantRepository.Create(ctx, merchant)
	if err != nil {
		s.logger.Error("Failed to create merchant", "error", err)
		return merchant, err
	}
	return newMerchant, nil
}

func (s *merchantService) Get(ctx context.Context, id string) (Merchant, error) {
	s.logger.Info("Getting merchant by ID", "id", id)
	return s.merchantRepository.Get(ctx, id)
}

func (s *merchantService) Update(ctx context.Context, id string, req MerchantRequest) (Merchant, error) {
	merchant := Merchant{
		Name: req.Name,
	}
	s.logger.Info("Updating merchant", "merchant", merchant)
	// check if merchant exists
	existingMerchant, err := s.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get merchant by ID", "error", err)
		return existingMerchant, err
	}

	// check if merchant name is already taken
	err = s.merchantRepository.CheckExists(ctx, merchant.Name)
	if err != nil {
		s.logger.Error("Failed to check if merchant exists", "error", err)
		return existingMerchant, err
	}

	// update merchant name
	existingMerchant.Name = merchant.Name
	if req.Logo != nil {
		// we will upload the image to the cloud storage and update the image url in the database
	}
	return s.merchantRepository.Update(ctx, existingMerchant)
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

func (s *merchantService) GetAll(ctx context.Context) ([]*Merchant, error) {
	s.logger.Info("Getting all merchants")
	merchants, err := s.merchantRepository.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all merchants", "error", err)
		return nil, err
	}
	return merchants, nil
}
