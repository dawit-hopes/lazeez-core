package ingredient

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type IngredientService interface {
	Create(ctx context.Context, req IngredientRequest) error
	Get(ctx context.Context, id string) (*IngredientDTO, error)
	Update(ctx context.Context, id string, req IngredientRequest) error
	List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*IngredientDTO], error)
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	CheckExists(ctx context.Context, name string) error
}

type ingredientService struct {
	ingredientRepository IngredientRepository
	logger               config.Logger
}

func NewIngredientService(ingredientRepository IngredientRepository, logger config.Logger) IngredientService {
	return &ingredientService{
		ingredientRepository: ingredientRepository,
		logger:               logger,
	}
}

func (s *ingredientService) Create(ctx context.Context, req IngredientRequest) error {
	name := common.FormatText(req.Name)
	ingredient := Ingredient{
		Name: name,
		Icon: req.Icon,
	}
	ingredient.ID = common.GenerateUUID()

	if err := s.ingredientRepository.HardDeleteSoftDeletedByName(ctx, name); err != nil {
		s.logger.Error("Failed to purge soft-deleted ingredient", "error", err)
		return err
	}

	err := s.ingredientRepository.CheckExists(ctx, req.Name)
	if err != nil {
		s.logger.Error("Failed to check if ingredient exists", "error", err)
		return err
	}

	err = s.ingredientRepository.Create(ctx, ingredient)
	if err != nil {
		s.logger.Error("Failed to create ingredient", "error", err)
		return err
	}
	return nil
}

func (s *ingredientService) Get(ctx context.Context, id string) (*IngredientDTO, error) {
	ingredient, err := s.ingredientRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get ingredient", "error", err)
		return nil, err
	}
	ingredientDTO := ingredient.ToDTO()
	return &ingredientDTO, nil
}

func (s *ingredientService) Update(ctx context.Context, id string, req IngredientRequest) error {
	existing, err := s.ingredientRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get ingredient", "error", err)
		return err
	}
	if req.Name != "" {
		err = s.ingredientRepository.CheckExists(ctx, req.Name)
		if err != nil {
			s.logger.Error("Failed to check if ingredient exists", "error", err)
			return err
		}
		existing.Name = common.FormatText(req.Name)
	}

	if req.Icon != "" {
		existing.Icon = req.Icon
	}

	err = s.ingredientRepository.Update(ctx, existing)
	if err != nil {
		s.logger.Error("Failed to update ingredient", "error", err)
		return err
	}
	return nil
}

func (s *ingredientService) List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*IngredientDTO], error) {
	result, err := s.ingredientRepository.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list ingredients", "error", err)
		return nil, err
	}
	ingredientDTOs := make([]*IngredientDTO, len(result.Data))
	for i, ing := range result.Data {
		dto := ing.ToDTO()
		ingredientDTOs[i] = &dto
	}
	return &common.PaginatedResponse[[]*IngredientDTO]{
		Data: ingredientDTOs,
		Meta: result.Meta,
	}, nil
}

func (s *ingredientService) Delete(ctx context.Context, id string) error {
	if err := s.ingredientRepository.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete ingredient", "error", err)
		return err
	}
	return nil
}

func (s *ingredientService) UnDelete(ctx context.Context, id string) error {
	err := s.ingredientRepository.UnDelete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to undelete ingredient", "error", err)
		return err
	}
	return nil
}

func (s *ingredientService) CheckExists(ctx context.Context, name string) error {
	err := s.ingredientRepository.CheckExists(ctx, name)
	if err != nil {
		s.logger.Error("Failed to check if ingredient exists", "error", err)
		return err
	}
	return nil
}
