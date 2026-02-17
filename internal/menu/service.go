package menu

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/ingredient"

	"golang.org/x/sync/errgroup"
)

type MenuService interface {
	Create(ctx context.Context, req MenuRequest) error
	Get(ctx context.Context, id string, branchID string) (*MenuDTO, error)
	List(ctx context.Context, filter common.Filter, branchID string) ([]*MenuDTO, error)
	Update(ctx context.Context, id string, req MenuRequest) error
	Delete(ctx context.Context, id string, branchID string) error
	UnDelete(ctx context.Context, id string, branchID string) error
}

type menuService struct {
	menuRepository    MenuRepository
	categoryService   category.CategoryService
	branchService     branch.BranchService
	ingredientService ingredient.IngredientService
	fileService       files.FileService
	logger            config.Logger
}

func NewMenuService(menuRepository MenuRepository,
	fileService files.FileService,
	categoryService category.CategoryService,
	branchService branch.BranchService,
	ingredientService ingredient.IngredientService,
	logger config.Logger) MenuService {
	return &menuService{
		menuRepository:    menuRepository,
		logger:            logger,
		fileService:       fileService,
		categoryService:   categoryService,
		branchService:     branchService,
		ingredientService: ingredientService,
	}
}

func (s *menuService) validateMenu(ctx context.Context, req MenuRequest) error {
	g, ctx := errgroup.WithContext(ctx)

	// category
	g.Go(func() error {
		_, err := s.categoryService.Get(ctx, req.CategoryID)
		if err != nil {
			s.logger.Error("Failed to get category", "error", err)
			return err
		}
		return nil
	})

	// branch
	g.Go(func() error {
		_, err := s.branchService.Get(ctx, req.BranchID)
		if err != nil {
			s.logger.Error("Failed to get branch", "error", err)
			return err
		}
		return nil
	})

	// ingredients
	g.Go(func() error {
		ingGroup, ingCtx := errgroup.WithContext(ctx)
		for _, id := range req.Ingredients {
			ingredientID := id
			ingGroup.Go(func() error {
				_, err := s.ingredientService.Get(ingCtx, ingredientID)
				if err != nil {
					s.logger.Error("Failed to get ingredient", "error", err)
					return err
				}
				return nil
			})
		}
		return ingGroup.Wait()
	})

	// name
	g.Go(func() error {
		if err := s.menuRepository.CheckExists(ctx, req.Name, req.BranchID); err != nil {
			s.logger.Error("Failed to check if menu exists", "error", err)
			return err
		}
		return nil
	})

	return g.Wait()
}

func (s *menuService) Create(ctx context.Context, req MenuRequest) error {
	if err := s.validateMenu(ctx, req); err != nil {
		s.logger.Error("Failed to validate menu", "error", err)
		return err
	}

	menu := req.ToModel()
	if req.Image != nil {
		imageURL, err := s.fileService.UploadFile(ctx, &req.ImageHeader)
		if err != nil {
			s.logger.Error("Failed to upload image", "error", err)
			return err
		}
		menu.Image = imageURL
	}
	menu.ID = common.GenerateUUID()

	err := s.menuRepository.Create(ctx, menu)
	if err != nil {
		s.logger.Error("Failed to create menu", "error", err)
		return err
	}

	return nil
}

func (s *menuService) Update(ctx context.Context, id string, req MenuRequest) error {
	existingMenu, err := s.menuRepository.Get(ctx, id, req.BranchID)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return err
	}

	if req.Name != "" {
		if err := s.menuRepository.CheckExists(ctx, req.Name, req.BranchID); err != nil {
			s.logger.Error("Failed to check if menu exists", "error", err)
			return err
		}
		existingMenu.Name = req.Name
	}

	if req.CategoryID != "" {
		_, err := s.categoryService.Get(ctx, req.CategoryID)
		if err != nil {
			s.logger.Error("Failed to get category", "error", err)
			return err
		}

		existingMenu.CategoryID = req.CategoryID
	}

	if req.BranchID != "" {
		_, err := s.branchService.Get(ctx, req.BranchID)
		if err != nil {
			s.logger.Error("Failed to get branch", "error", err)
			return err
		}
		existingMenu.BranchID = req.BranchID
	}

	if len(req.Ingredients) > 0 {
		for _, ingredient := range req.Ingredients {
			_, err := s.ingredientService.Get(ctx, ingredient)
			if err != nil {
				s.logger.Error("Failed to get ingredient", "error", err)
				return err
			}

		}

		existingMenu.Ingredients = req.Ingredients
	}

	if req.Image != nil {
		imageURL, err := s.fileService.UploadFile(ctx, &req.ImageHeader)
		if err != nil {
			s.logger.Error("Failed to upload image", "error", err)
			return err
		}
		existingMenu.Image = imageURL
	}

	if req.IsFasting != existingMenu.IsFasting {
		existingMenu.IsFasting = req.IsFasting
	}

	if req.IsAvailable != existingMenu.IsAvailable {
		existingMenu.IsAvailable = req.IsAvailable
	}

	if req.Description != existingMenu.Description {
		existingMenu.Description = req.Description
	}

	if req.Price != existingMenu.Price {
		existingMenu.Price = req.Price
	}

	err = s.menuRepository.Update(ctx, existingMenu)
	if err != nil {
		s.logger.Error("Failed to update menu", "error", err)
		return err
	}

	return nil
}

func (s *menuService) Get(ctx context.Context, id string, branchID string) (*MenuDTO, error) {
	menu, err := s.menuRepository.Get(ctx, id, branchID)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return nil, err
	}
	menuDTO := menu.ToDTO()
	return &menuDTO, nil
}

func (s *menuService) Delete(ctx context.Context, id string, branchID string) error {
	if err := s.menuRepository.Delete(ctx, id, branchID); err != nil {
		s.logger.Error("Failed to delete menu", "error", err)
		return err
	}

	return nil
}

func (s *menuService) UnDelete(ctx context.Context, id string, branchID string) error {
	s.logger.Info("Undeleting menu", "id", id)
	err := s.menuRepository.UnDelete(ctx, id, branchID)
	if err != nil {
		s.logger.Error("Failed to undelete menu", "error", err)
		return err
	}
	return nil
}

func (s *menuService) List(ctx context.Context, filter common.Filter, branchID string) ([]*MenuDTO, error) {
	menus, err := s.menuRepository.List(ctx, filter, branchID)
	if err != nil {
		s.logger.Error("Failed to list menus", "error", err)
		return nil, err
	}
	menuDTOs := make([]*MenuDTO, len(menus))
	for i, menu := range menus {
		menuDTO := menu.ToDTO()
		menuDTOs[i] = &menuDTO
	}
	return menuDTOs, nil
}
