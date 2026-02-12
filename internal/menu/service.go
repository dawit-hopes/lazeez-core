package menu

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
)

type MenuService interface {
	Create(ctx context.Context, req MenuRequest) (Menu, error)
	Get(ctx context.Context, id string) (Menu, error)
	Update(ctx context.Context, id string, req MenuRequest) (Menu, error)
	Delete(ctx context.Context, id string) error
}

type menuService struct {
	menuRepository MenuRepository
	fileService    files.FileService
	logger         config.Logger
}

func NewMenuService(menuRepository MenuRepository, fileService files.FileService, logger config.Logger) MenuService {
	return &menuService{
		menuRepository: menuRepository,
		logger:         logger,
		fileService:    fileService,
	}
}

func (s *menuService) Create(ctx context.Context, req MenuRequest) (Menu, error) {
	menu := Menu{
		Name: req.Name,
	}

	if req.Image != nil {
		imageURL, err := s.fileService.UploadFile(ctx, &req.ImageHeader)
		if err != nil {
			s.logger.Error("Failed to upload image", "error", err)
			return menu, err
		}
		menu.Image = imageURL
	}
	menu.ID = common.GenerateUUID()

	createdMenu, err := s.menuRepository.Create(ctx, menu)
	if err != nil {
		s.logger.Error("Failed to create menu", "error", err)
		return menu, err
	}

	return createdMenu, nil
}

func (s *menuService) Get(ctx context.Context, id string) (Menu, error) {
	menu, err := s.menuRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return Menu{}, err
	}
	return menu, nil
}

func (s *menuService) Update(ctx context.Context, id string, req MenuRequest) (Menu, error) {
	existingMenu, err := s.menuRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return Menu{}, err
	}

	if req.Name != "" {
		existingMenu.Name = req.Name
	}
	if req.Image != nil {
		imageURL, err := s.fileService.UploadFile(ctx, &req.ImageHeader)
		if err != nil {
			s.logger.Error("Failed to upload image", "error", err)
			return Menu{}, err
		}
		existingMenu.Image = imageURL
	}

	updatedMenu, err := s.menuRepository.Update(ctx, existingMenu)
	if err != nil {
		s.logger.Error("Failed to update menu", "error", err)
		return Menu{}, err
	}

	return updatedMenu, nil
}

func (s *menuService) Delete(ctx context.Context, id string) error {
	if err := s.menuRepository.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete menu", "error", err)
		return err
	}

	return nil
}
