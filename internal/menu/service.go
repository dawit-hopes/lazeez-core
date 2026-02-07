package menu

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MenuService interface {
	Create(ctx context.Context, req MenuRequest) (Menu, error)
	Get(ctx context.Context, id string) (Menu, error)
	Update(ctx context.Context, id string, req MenuRequest) (Menu, error)
	Delete(ctx context.Context, id string) error
}

type menuService struct {
	menuRepository MenuRepository
	logger         config.Logger
}

func NewMenuService(menuRepository MenuRepository, logger config.Logger) MenuService {
	return &menuService{
		menuRepository: menuRepository,
		logger:         logger,
	}
}

func (s *menuService) Create(ctx context.Context, req MenuRequest) (Menu, error) {
	menu := Menu{
		Name: req.Name,
	}

	menu.ID = common.GenerateUUID()

	s.logger.Info("Creating menu", "menu", menu)

	createdMenu, err := s.menuRepository.Create(ctx, menu)
	if err != nil {
		s.logger.Error("Failed to create menu", "error", err)
		return menu, err
	}

	return createdMenu, nil
}

func (s *menuService) Get(ctx context.Context, id string) (Menu, error) {
	s.logger.Info("Getting menu by ID", "id", id)
	menu, err := s.menuRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return Menu{}, err
	}
	return menu, nil
}

func (s *menuService) Update(ctx context.Context, id string, req MenuRequest) (Menu, error) {
	s.logger.Info("Updating menu", "id", id)

	existingMenu, err := s.menuRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return Menu{}, err
	}

	if req.Name != "" {
		existingMenu.Name = req.Name
	}
	if req.Image != nil {
		// we will upload the image to the cloud storage and update the image url in the database
	}

	updatedMenu, err := s.menuRepository.Update(ctx, existingMenu)
	if err != nil {
		s.logger.Error("Failed to update menu", "error", err)
		return Menu{}, err
	}

	return updatedMenu, nil
}

func (s *menuService) Delete(ctx context.Context, id string) error {
	s.logger.Info("Deleting menu", "id", id)

	if err := s.menuRepository.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete menu", "error", err)
		return err
	}

	return nil
}
