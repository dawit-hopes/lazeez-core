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
	Create(ctx context.Context, req MenuRequest, role string) (*MenuDTO, error)
	Get(ctx context.Context, id string, branchID, merchantID string) (*MenuDTO, error)
	List(ctx context.Context, filter common.Filter, branchID, merchantID string, role string, scope ListScope) (*common.PaginatedResponse[[]*MenuDTO], error)
	Update(ctx context.Context, id string, req MenuRequest, role string) error
	Delete(ctx context.Context, id string, branchID, merchantID string, role string) error
	UnDelete(ctx context.Context, id string, branchID, merchantID string, role string) error
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

func isSuperAdmin(role string) bool {
	return role == "super_admin"
}

func isBranchUser(role string) bool {
	return role == "branch_manager" || role == "super_branch_admin" || role == "branch_staff"
}

func isSuperBranchAdmin(role string) bool {
	return role == "super_branch_admin"
}

func isBranchManager(role string) bool {
	return role == "branch_manager" || role == "branch_staff"
}

func canManageMasterMenu(role, userMerchantID, menuMerchantID string) bool {
	if isSuperAdmin(role) {
		return true
	}
	if isSuperBranchAdmin(role) && userMerchantID != "" && userMerchantID == menuMerchantID {
		return true
	}
	return false
}

func isMasterMenuCreate(role string, branchID string) bool {
	return branchID == "" && (isSuperBranchAdmin(role) || isSuperAdmin(role))
}

func (s *menuService) validateMenu(ctx context.Context, req MenuRequest, isMaster bool) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := s.categoryService.Get(ctx, req.CategoryID)
		if err != nil {
			s.logger.Error("Failed to get category", "error", err)
			return err
		}
		return nil
	})

	if !isMaster {
		g.Go(func() error {
			_, err := s.branchService.Get(ctx, req.BranchID)
			if err != nil {
				s.logger.Error("Failed to get branch", "error", err)
				return err
			}
			return nil
		})
	}

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

	g.Go(func() error {
		if isMaster {
			if err := s.menuRepository.CheckMasterExists(ctx, req.Name, req.MerchantID); err != nil {
				s.logger.Error("Failed to check if master menu exists", "error", err)
				return err
			}
		} else if err := s.menuRepository.CheckExists(ctx, req.Name, req.BranchID); err != nil {
			s.logger.Error("Failed to check if menu exists", "error", err)
			return err
		}
		return nil
	})

	return g.Wait()
}

func (s *menuService) Create(ctx context.Context, req MenuRequest, role string) (*MenuDTO, error) {
	isMaster := isMasterMenuCreate(role, req.BranchID)
	if isMaster && req.MerchantID == "" && !isSuperAdmin(role) {
		return nil, common.ErrBranchAdminMissingMerchant
	}
	if isBranchManager(role) && req.BranchID == "" {
		return nil, common.ErrUnAuthorized
	}
	if isSuperBranchAdmin(role) && req.BranchID == "" {
		// Restaurant owner creates master menu for their merchant.
	} else if !isSuperAdmin(role) && !isBranchManager(role) {
		return nil, common.ErrUnAuthorized
	}

	if err := s.validateMenu(ctx, req, isMaster); err != nil {
		s.logger.Error("Failed to validate menu", "error", err)
		return nil, err
	}

	modifierGroupIDs := make(pq.StringArray, 0, len(req.Modifiers))
	modifierGroupDTOs := make([]modgroup.ModifierGroupDTO, 0, len(req.Modifiers))

	for _, mgReq := range req.Modifiers {
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

	menu := req.ToModel(isMaster)
	if !isMaster && req.BranchID != "" {
		br, err := s.branchService.Get(ctx, req.BranchID)
		if err != nil {
			return nil, err
		}
		menu.MerchantID = common.ToNUllString(br.MerchantID)
	}
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

	menuDTO := menu.ToDTO()
	s.enrichMenuDTO(ctx, &menuDTO, menuDTO.CategoryID, menu.Ingredients)
	menuDTO.Modifiers = modifierGroupDTOs

	return &menuDTO, nil
}

func (s *menuService) Update(ctx context.Context, id string, req MenuRequest, role string) error {
	branchID := req.BranchID
	existingMenu, err := s.loadMenuForAction(ctx, id, branchID, req.MerchantID)
	if err != nil {
		s.logger.Error("Failed to get menu", "error", err)
		return err
	}

	if existingMenu.IsMaster() {
		if isBranchManager(role) {
			return s.updateMasterFromBranch(ctx, id, branchID, req)
		}
		if !canManageMasterMenu(role, req.MerchantID, existingMenu.MerchantIDString()) {
			return common.ErrUnAuthorized
		}
	} else if isBranchManager(role) {
		if existingMenu.BranchIDString() != branchID {
			return common.ErrUnAuthorized
		}
	} else if !canManageMasterMenu(role, req.MerchantID, existingMenu.MerchantIDString()) && !isSuperAdmin(role) {
		return common.ErrUnAuthorized
	}

	if req.Name != "" {
		if existingMenu.IsMaster() {
			if err := s.menuRepository.CheckMasterExists(ctx, req.Name, existingMenu.MerchantIDString()); err != nil {
				return err
			}
		} else if err := s.menuRepository.CheckExists(ctx, req.Name, existingMenu.BranchIDString()); err != nil {
			return err
		}
		existingMenu.Name = req.Name
	}

	if req.CategoryID != "" {
		if _, err := s.categoryService.Get(ctx, req.CategoryID); err != nil {
			return err
		}
		existingMenu.CategoryID = common.ParseStringToUUID(req.CategoryID)
	}

	if len(req.Ingredients) > 0 {
		for _, ingredientID := range req.Ingredients {
			if _, err := s.ingredientService.Get(ctx, ingredientID); err != nil {
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
			return err
		}
		existingMenu.Image = imageURL
	}

	if req.IsFasting != nil {
		existingMenu.IsFasting = *req.IsFasting
	}

	if req.IsAvailable != nil {
		existingMenu.IsAvailable = *req.IsAvailable
	}

	if req.Description != "" {
		existingMenu.Description = req.Description
	}

	if req.Price != 0 {
		existingMenu.Price = req.Price
	}

	if req.PreparationTime != 0 {
		existingMenu.PreparationTime = req.PreparationTime
	}

	return s.menuRepository.Update(ctx, existingMenu)
}

func (s *menuService) updateMasterFromBranch(ctx context.Context, id, branchID string, req MenuRequest) error {
	if branchID == "" {
		return common.ErrUnAuthorized
	}
	if req.IsAvailable == nil {
		return common.ErrNoDataToUpdate
	}
	// Branch users may only toggle availability on inherited master items.
	if len(req.Ingredients) > 0 || req.Name != "" || req.CategoryID != "" || req.Image != nil ||
		req.IsFasting != nil || req.Description != "" || req.Price != 0 || req.PreparationTime != 0 {
		return common.ErrUnAuthorized
	}
	return s.menuRepository.SetBranchOverride(ctx, branchID, id, *req.IsAvailable)
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

func (s *menuService) Get(ctx context.Context, id string, branchID, merchantID string) (*MenuDTO, error) {
	var menu Menu
	var err error
	if branchID != "" {
		menu, err = s.menuRepository.Get(ctx, id, branchID)
	} else if merchantID != "" {
		menu, err = s.menuRepository.GetMaster(ctx, id, merchantID)
	} else {
		return nil, common.ErrUnAuthorized
	}
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

func (s *menuService) Delete(ctx context.Context, id string, branchID, merchantID string, role string) error {
	existing, err := s.loadMenuForAction(ctx, id, branchID, merchantID)
	if err != nil {
		return err
	}

	if existing.IsMaster() {
		if isBranchManager(role) {
			if branchID == "" {
				return common.ErrUnAuthorized
			}
			return s.menuRepository.SetBranchExcluded(ctx, branchID, id, true)
		}
		if !canManageMasterMenu(role, merchantID, existing.MerchantIDString()) {
			return common.ErrUnAuthorized
		}
		return s.menuRepository.Delete(ctx, id, "")
	}

	if isBranchManager(role) && existing.BranchIDString() != branchID {
		return common.ErrUnAuthorized
	}

	return s.menuRepository.Delete(ctx, id, existing.BranchIDString())
}

func (s *menuService) loadMenuForAction(ctx context.Context, id, branchID, merchantID string) (Menu, error) {
	if branchID != "" {
		return s.menuRepository.Get(ctx, id, branchID)
	}
	if merchantID != "" {
		return s.menuRepository.GetMaster(ctx, id, merchantID)
	}
	return Menu{}, common.ErrUnAuthorized
}

func (s *menuService) UnDelete(ctx context.Context, id string, branchID, merchantID string, role string) error {
	existing, err := s.loadMenuForAction(ctx, id, branchID, merchantID)
	if err != nil {
		return err
	}

	if existing.IsMaster() && isBranchManager(role) {
		if branchID == "" {
			return common.ErrUnAuthorized
		}
		return s.menuRepository.SetBranchExcluded(ctx, branchID, id, false)
	}

	deleteBranchID := existing.BranchIDString()
	if existing.IsMaster() {
		deleteBranchID = ""
	}
	return s.menuRepository.UnDelete(ctx, id, deleteBranchID)
}

func (s *menuService) List(ctx context.Context, filter common.Filter, branchID, merchantID string, role string, scope ListScope) (*common.PaginatedResponse[[]*MenuDTO], error) {
	if scope == "" {
		switch {
		case isSuperBranchAdmin(role) && branchID == "":
			scope = ScopeMaster
		case isSuperAdmin(role) && branchID == "":
			scope = ScopeAllBranches
		case isSuperBranchAdmin(role):
			scope = ScopeAllBranches
		case isBranchManager(role):
			scope = ScopeBranchManage
		default:
			scope = ScopeBranchEffective
		}
	}

	if scope == ScopeMaster && merchantID == "" {
		return nil, common.ErrBranchAdminMissingMerchant
	}

	if scope == ScopeBranchEffective && branchID == "" {
		return nil, common.ErrUnAuthorized
	}

	if scope == ScopeAllBranches && isSuperBranchAdmin(role) && merchantID == "" {
		return nil, common.ErrBranchAdminMissingMerchant
	}

	result, err := s.menuRepository.ListScoped(ctx, filter, scope, branchID, merchantID)
	if err != nil {
		s.logger.Error("Failed to list menus", "error", err)
		return nil, err
	}
	menus := result.Data

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
