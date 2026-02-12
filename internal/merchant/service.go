package merchant

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MerchantService interface {
	Create(ctx context.Context, req MerchantRequest) (MerchantDTO, error)
	Get(ctx context.Context, id string) (MerchantDTO, error)
	Update(ctx context.Context, id string, req MerchantRequest) error
	GetAll(ctx context.Context) ([]*MerchantDTO, error)
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

func (s *merchantService) Create(ctx context.Context, req MerchantRequest) (MerchantDTO, error) {
	merchant := Merchant{
		Name: req.Name,
	}
	s.logger.Info("Creating merchant", "merchant", merchant)
	merchant.ID = common.GenerateUUID()
	err := s.merchantRepository.CheckExists(ctx, merchant.Name)
	if err != nil {
		s.logger.Error("Failed to check if merchant exists", "error", err)
		return MerchantDTO{}, err
	}

	if req.Logo != nil {
		// we will upload the image to the cloud storage and update the image url in the database
	}

	newMerchant, err := s.merchantRepository.Create(ctx, merchant)
	if err != nil {
		s.logger.Error("Failed to create merchant", "error", err)
		return MerchantDTO{}, err
	}
	return newMerchant.ToDTO(), nil
}

func (s *merchantService) Get(ctx context.Context, id string) (MerchantDTO, error) {
	s.logger.Info("Getting merchant by ID", "id", id)
	merchant, err := s.merchantRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get merchant by ID", "error", err)
		return MerchantDTO{}, err
	}
	return merchant.ToDTO(), nil
}

func (s *merchantService) Update(ctx context.Context, id string, req MerchantRequest) error {
	s.logger.Info("Updating merchant", "id", id, "name", req.Name)

	// Load existing merchant from repository
	existingMerchant, err := s.merchantRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get merchant by ID", "error", err)
		return err
	}

	// If name is provided and changed, ensure it's not taken by another merchant
	if req.Name != "" && req.Name != existingMerchant.Name {
		if err := s.merchantRepository.CheckExists(ctx, req.Name); err != nil {
			s.logger.Error("Failed to check if merchant exists", "error", err)
			return err
		}
		existingMerchant.Name = req.Name
	}

	if req.Logo != nil {
		// we will upload the image to the cloud storage and update the image url in the database
	}

	_, err = s.merchantRepository.Update(ctx, existingMerchant)
	if err != nil {
		s.logger.Error("Failed to update merchant", "error", err)
		return err
	}
	return nil
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

func (s *merchantService) GetAll(ctx context.Context) ([]*MerchantDTO, error) {
	s.logger.Info("Getting all merchants")
	merchants, err := s.merchantRepository.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all merchants", "error", err)
		return nil, err
	}
	merchantDTOs := make([]*MerchantDTO, len(merchants))
	for i, merchant := range merchants {
		result := merchant.ToDTO()
		merchantDTOs[i] = &result
	}
	return merchantDTOs, nil
}
