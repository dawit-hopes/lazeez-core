package menu

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/ingredient"
	modgroup "lazeez-core/internal/modifiers/group"
	modoption "lazeez-core/internal/modifiers/option"

	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"
)

type MenuService interface {
	Create(ctx context.Context, req MenuRequest) (*MenuDTO, error)
	Get(ctx context.Context, id string, branchID string) (*MenuDTO, error)
	List(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*MenuDTO], error)
	Update(ctx context.Context, id string, req MenuRequest) error
	Delete(ctx context.Context, id string, branchID string) error
	UnDelete(ctx context.Context, id string, branchID string) error
}

type menuService struct {
	menuRepository        MenuRepository
	categoryService       category.CategoryService
	branchService         branch.BranchService
	ingredientService     ingredient.IngredientService
	modifierGroupService  modgroup.ModifierGroupService
	modifierOptionService modoption.ModifierOptionService
	fileService           files.FileService
	logger                config.Logger
}

func NewMenuService(menuRepository MenuRepository,
	fileService files.FileService,
	categoryService category.CategoryService,
	branchService branch.BranchService,
	ingredientService ingredient.IngredientService,
	modifierGroupService modgroup.ModifierGroupService,
	modifierOptionService modoption.ModifierOptionService,
	logger config.Logger) MenuService {
	return &menuService{
		menuRepository:        menuRepository,
		logger:                logger,
		fileService:           fileService,
		categoryService:       categoryService,
		branchService:         branchService,
		ingredientService:     ingredientService,
		modifierGroupService:  modifierGroupService,
		modifierOptionService: modifierOptionService,
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

func (s *menuService) Create(ctx context.Context, req MenuRequest) (*MenuDTO, error) {
	if err := s.validateMenu(ctx, req); err != nil {
		s.logger.Error("Failed to validate menu", "error", err)
		return nil, err
	}

	// Create modifier groups and options (if any) and collect their IDs.
	modifierGroupIDs := make(pq.StringArray, 0, len(req.Modifiers))
	modifierGroupDTOs := make([]modgroup.ModifierGroupDTO, 0, len(req.Modifiers))

	for _, mgReq := range req.Modifiers {
		// Validate group and options using their own validators
		if err := mgReq.Validate(); err != nil {
			s.logger.Error("Failed to validate modifier group", "error", err)
			return nil, err
		}

		optionIDs := make(pq.StringArray, 0, len(mgReq.Options))
		optionDTOs := make([]modoption.ModifierOptionDTO, 0, len(mgReq.Options))
		for _, optReq := range mgReq.Options {
			if err := optReq.Validate(); err != nil {
				s.logger.Error("Failed to validate modifier option", "error", err)
				return nil, err
			}

			optionModel := modoption.ModifierOption{
				Base: common.Base{
					ID:        common.GenerateUUID(),
					IsDeleted: false,
				},
				Name:            common.FormatText(optReq.Name),
				PriceAdjustment: optReq.PriceAdjustment,
				IsDefault:       optReq.IsDefault,
				IsAvailable:     optReq.IsAvailable,
			}

			if err := s.modifierOptionService.Create(ctx, optionModel); err != nil {
				s.logger.Error("Failed to create modifier option", "error", err)
				return nil, err
			}

			optionIDs = append(optionIDs, optionModel.ID)
			optionDTOs = append(optionDTOs, optionModel.ToDTO())
		}

		groupModel := modgroup.ModifierGroup{
			Base: common.Base{
				ID:        common.GenerateUUID(),
				IsDeleted: false,
			},
			Name:          common.FormatText(mgReq.Name),
			SelectionType: string(mgReq.SelectionType),
			IsRequired:    mgReq.IsRequired,
			MinSelections: mgReq.MinSelections,
			MaxSelections: mgReq.MaxSelections,
			Options:       optionIDs,
		}

		if err := s.modifierGroupService.Create(ctx, groupModel); err != nil {
			s.logger.Error("Failed to create modifier group", "error", err)
			return nil, err
		}

		modifierGroupIDs = append(modifierGroupIDs, groupModel.ID)
		modifierGroupDTOs = append(modifierGroupDTOs, groupModel.ToDTO(optionDTOs))
	}

	menu := req.ToModel()
	if req.Image != nil {
		imageURL, err := s.fileService.UploadFile(ctx, &req.ImageHeader)
		if err != nil {
			s.logger.Error("Failed to upload image", "error", err)
			return nil, err
		}
		menu.Image = imageURL
	}
	menu.ID = common.GenerateUUID()
	menu.Modifiers = modifierGroupIDs

	if err := s.menuRepository.Create(ctx, menu); err != nil {
		s.logger.Error("Failed to create menu", "error", err)
		return nil, err
	}

	// Build response DTO, including category, ingredients, and modifier groups/options.
	menuDTO := menu.ToDTO()

	// Enrich category and ingredients as in List/Get
	s.enrichMenuDTO(ctx, &menuDTO, menuDTO.CategoryID, menu.Ingredients)
	menuDTO.Modifiers = modifierGroupDTOs

	return &menuDTO, nil
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

		existingMenu.CategoryID = common.ParseStringToUUID(req.CategoryID)
	}

	if req.BranchID != "" {
		_, err := s.branchService.Get(ctx, req.BranchID)
		if err != nil {
			s.logger.Error("Failed to get branch", "error", err)
			return err
		}
		existingMenu.BranchID = common.ParseStringToUUID(req.BranchID)
	}

	if len(req.Ingredients) > 0 {
		for _, ingredient := range req.Ingredients {
			_, err := s.ingredientService.Get(ctx, ingredient)
			if err != nil {
				s.logger.Error("Failed to get ingredient", "error", err)
				return err
			}

		}

		ingredients := make(pq.StringArray, len(req.Ingredients))
		copy(ingredients, req.Ingredients)
		existingMenu.Ingredients = ingredients
	}

	if req.Image != nil {
		imageURL, err := s.fileService.UploadFile(ctx, &req.ImageHeader)
		if err != nil {
			s.logger.Error("Failed to upload image", "error", err)
			return err
		}
		existingMenu.Image = imageURL
	}

	if req.IsFasting != nil && *req.IsFasting != existingMenu.IsFasting {
		existingMenu.IsFasting = *req.IsFasting
	}

	if req.IsAvailable != nil && *req.IsAvailable != existingMenu.IsAvailable {
		existingMenu.IsAvailable = *req.IsAvailable
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

func (s *menuService) enrichMenuDTO(ctx context.Context, dto *MenuDTO, categoryID string, ingredientIDs []string) {
	if categoryID != "" {
		if cat, err := s.categoryService.Get(ctx, categoryID); err == nil {
			dto.Category = cat
		}
	}
	ingredients := make([]*ingredient.IngredientDTO, 0, len(ingredientIDs))
	for _, id := range ingredientIDs {
		if id == "" {
			continue
		}
		if ing, err := s.ingredientService.Get(ctx, id); err == nil {
			ingredients = append(ingredients, ing)
		}
	}
	dto.Ingredients = ingredients
}

func (s *menuService) buildModifierGroups(ctx context.Context, modifierGroupIDs []string) ([]modgroup.ModifierGroupDTO, error) {
	groups := make([]modgroup.ModifierGroupDTO, 0, len(modifierGroupIDs))
	for _, groupID := range modifierGroupIDs {
		if groupID == "" {
			continue
		}

		groupModel, err := s.modifierGroupService.Get(ctx, groupID)
		if err != nil {
			s.logger.Error("Failed to get modifier group", "id", groupID, "error", err)
			return nil, err
		}

		options := make([]modoption.ModifierOptionDTO, 0, len(groupModel.Options))
		for _, optionID := range groupModel.Options {
			if optionID == "" {
				continue
			}
			optionModel, err := s.modifierOptionService.Get(ctx, optionID)
			if err != nil {
				s.logger.Error("Failed to get modifier option", "id", optionID, "error", err)
				return nil, err
			}
			options = append(options, optionModel.ToDTO())
		}

		groups = append(groups, groupModel.ToDTO(options))
	}

	return groups, nil
}

func (s *menuService) Get(ctx context.Context, id string, branchID string) (*MenuDTO, error) {
	menu, err := s.menuRepository.Get(ctx, id, branchID)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return nil, err
	}
	menuDTO := menu.ToDTO()
	s.enrichMenuDTO(ctx, &menuDTO, menuDTO.CategoryID, menu.Ingredients)

	if len(menu.Modifiers) > 0 {
		modifiers, err := s.buildModifierGroups(ctx, menu.Modifiers)
		if err != nil {
			return nil, err
		}
		menuDTO.Modifiers = modifiers
	}

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

func (s *menuService) List(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*MenuDTO], error) {
	result, err := s.menuRepository.List(ctx, filter, branchID)
	if err != nil {
		s.logger.Error("Failed to list menus", "error", err)
		return nil, err
	}
	menus := result.Data
	// Collect unique category and ingredient IDs for batch lookup
	uniqueCategoryIDs := make(map[string]struct{})
	uniqueIngredientIDs := make(map[string]struct{})
	for _, menu := range menus {
		uniqueCategoryIDs[common.ParseUUIDToString(menu.CategoryID)] = struct{}{}
		for _, id := range menu.Ingredients {
			if id != "" {
				uniqueIngredientIDs[id] = struct{}{}
			}
		}
	}
	categoryMap := make(map[string]*category.CategoryDTO)
	for id := range uniqueCategoryIDs {
		if cat, err := s.categoryService.Get(ctx, id); err == nil {
			categoryMap[id] = cat
		}
	}
	ingredientMap := make(map[string]*ingredient.IngredientDTO)
	for id := range uniqueIngredientIDs {
		if ing, err := s.ingredientService.Get(ctx, id); err == nil {
			ingredientMap[id] = ing
		}
	}
	menuDTOs := make([]*MenuDTO, len(menus))
	for i, menu := range menus {
		dto := menu.ToDTO()
		dto.Category = categoryMap[dto.CategoryID]
		ingredients := make([]*ingredient.IngredientDTO, 0, len(menu.Ingredients))
		for _, id := range menu.Ingredients {
			if ing := ingredientMap[id]; ing != nil {
				ingredients = append(ingredients, ing)
			}
		}
		dto.Ingredients = ingredients

		if len(menu.Modifiers) > 0 {
			if modifiers, err := s.buildModifierGroups(ctx, menu.Modifiers); err == nil {
				dto.Modifiers = modifiers
			} else {
				// If we fail to build modifiers for a specific menu, log and continue
				s.logger.Error("Failed to build modifiers for menu", "menu_id", menu.ID, "error", err)
			}
		}

		menuDTOs[i] = &dto
	}
	return &common.PaginatedResponse[[]*MenuDTO]{
		Data: menuDTOs,
		Meta: result.Meta,
	}, nil
}
