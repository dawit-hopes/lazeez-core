package category

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type CategoryService interface {
	Create(ctx context.Context, req CategoryRequest) error
	Get(ctx context.Context, id string) (*CategoryDTO, error)
	Update(ctx context.Context, id string, req CategoryRequest) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	List(ctx context.Context, filter common.Filter) ([]*CategoryDTO, error)
	CheckExists(ctx context.Context, name string) error
}

type categoryService struct {
	categoryRepository CategoryRepository
	logger             config.Logger
}

func NewCategoryService(categoryRepository CategoryRepository, logger config.Logger) CategoryService {
	return &categoryService{
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (s *categoryService) Create(ctx context.Context, req CategoryRequest) error {
	category := Category{
		Name: req.Name,
	}

	err := s.categoryRepository.CheckExists(ctx, req.Name)
	if err != nil {
		s.logger.Error("Failed to check if category exists", "error", err)
		return err
	}

	if req.Icon != "" {
		category.Icon = req.Icon
	}

	category.ID = common.GenerateUUID()

	err = s.categoryRepository.Create(ctx, category)
	if err != nil {
		s.logger.Error("Failed to create category", "error", err)
		return err
	}

	return nil
}

func (s *categoryService) Get(ctx context.Context, id string) (*CategoryDTO, error) {
	category, err := s.categoryRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get category", "error", err)
		return nil, err
	}
	categoryDTO := category.ToDTO()
	return &categoryDTO, nil
}

func (s *categoryService) Update(ctx context.Context, id string, req CategoryRequest) error {
	existingCategory, err := s.categoryRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get category", "error", err)
		return err
	}

	if req.Name != "" {
		existingCategory.Name = req.Name
		err = s.categoryRepository.CheckExists(ctx, req.Name)
		if err != nil {
			s.logger.Error("Failed to check if category exists", "error", err)
			return err
		}
	}
	if req.Icon != "" {
		existingCategory.Icon = req.Icon
	}

	err = s.categoryRepository.Update(ctx, existingCategory)
	if err != nil {
		s.logger.Error("Failed to update category", "error", err)
		return err
	}

	return nil
}

func (s *categoryService) Delete(ctx context.Context, id string) error {
	if err := s.categoryRepository.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete category", "error", err)
		return err
	}
	return nil
}

func (s *categoryService) List(ctx context.Context, filter common.Filter) ([]*CategoryDTO, error) {
	categories, err := s.categoryRepository.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list categories", "error", err)
		return nil, err
	}
	categoryDTOs := make([]*CategoryDTO, len(categories))
	for i, category := range categories {
		categoryDTO := category.ToDTO()
		categoryDTOs[i] = &categoryDTO
	}
	return categoryDTOs, nil
}

func (s *categoryService) UnDelete(ctx context.Context, id string) error {
	err := s.categoryRepository.UnDelete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to undelete category", "error", err)
		return err
	}
	return nil
}

func (s *categoryService) CheckExists(ctx context.Context, name string) error {
	err := s.categoryRepository.CheckExists(ctx, name)
	if err != nil {
		s.logger.Error("Failed to check if category exists", "error", err)
		return err
	}
	return nil
}
