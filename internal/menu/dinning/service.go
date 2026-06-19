package menu

import (
	"context"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/ingredient"
	modgroup "lazeez-core/internal/modifiers/group"
	modoption "lazeez-core/internal/modifiers/option"
	"lazeez-core/internal/promotion"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/room"
	"lazeez-core/internal/users"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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

	// public services
	ListMenus(ctx context.Context, filter common.Filter, reference, referenceType string) (*PublicMenuCatalogResponse, error)

	GetBranchMenuSnapshots(ctx context.Context, branchID string, menuIDs []string) (map[string]BranchMenuSnapshot, error)
}

type menuService struct {
	menuRepository        MenuRepository
	categoryService       category.CategoryService
	branchService         branch.BranchService
	ingredientService     ingredient.IngredientService
	modifierGroupService  modgroup.ModifierGroupService
	modifierOptionService modoption.ModifierOptionService
	roomService           room.RoomService
	bookingService        booking.BookingService
	promotionService      promotion.PromotionService
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
	roomService room.RoomService,
	bookingService booking.BookingService,
	promotionService promotion.PromotionService,
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
		roomService:           roomService,
		bookingService:        bookingService,
		promotionService:      promotionService,
	}
}

func isMasterMenuCreate(role string, branchID string) bool {
	return branchID == "" && (users.IsSuperBranchAdminRoleString(role) || users.IsSuperAdminRoleString(role))
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
	if isMaster && req.MerchantID == "" && !users.IsSuperAdminRoleString(role) {
		return nil, common.ErrBranchAdminMissingMerchant
	}
	if users.IsBranchStaffRoleString(role) && req.BranchID == "" {
		return nil, common.ErrUnAuthorized
	}
	if users.IsSuperBranchAdminRoleString(role) && req.BranchID == "" {
		// Restaurant owner creates master menu for their merchant.
	} else if !users.IsSuperAdminRoleString(role) && !users.IsBranchStaffRoleString(role) {
		return nil, common.ErrUnAuthorized
	}

	if err := s.validateMenu(ctx, req, isMaster); err != nil {
		s.logger.Error("Failed to validate menu", "error", err)
		return nil, err
	}

	modifierGroupIDs, modifierGroupDTOs, err := s.createModifierGroupsFromRequest(ctx, req.Modifiers)
	if err != nil {
		return nil, err
	}

	menu := req.ToModel(isMaster)
	req.applyDiscountTo(&menu)
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
		if users.IsBranchStaffRoleString(role) {
			return s.updateMasterFromBranch(ctx, id, branchID, req)
		}
		if !users.CanManageMerchantMaster(role, req.MerchantID, existingMenu.MerchantIDString()) {
			return common.ErrUnAuthorized
		}
	} else if users.IsBranchStaffRoleString(role) {
		if existingMenu.BranchIDString() != branchID {
			return common.ErrUnAuthorized
		}
	} else if !users.CanManageMerchantMaster(role, req.MerchantID, existingMenu.MerchantIDString()) && !users.IsSuperAdminRoleString(role) {
		return common.ErrUnAuthorized
	}

	if req.Name != "" {
		newName := common.FormatText(req.Name)
		if newName != existingMenu.Name {
			if existingMenu.IsMaster() {
				if err := s.menuRepository.CheckMasterExists(ctx, newName, existingMenu.MerchantIDString()); err != nil {
					return err
				}
			} else if err := s.menuRepository.CheckExists(ctx, newName, existingMenu.BranchIDString()); err != nil {
				return err
			}
		}
		existingMenu.Name = newName
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

	if req.IsChefsChoice != nil {
		existingMenu.IsChefsChoice = *req.IsChefsChoice
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

	if req.DiscountSet {
		if req.Discount != nil && !req.Discount.IsValid(existingMenu.Price) {
			if req.Discount.Type == DiscountTypePercentage {
				return validation.NewError("validation", "percentage discount must be greater than 0 and at most 100")
			}
			return validation.NewError("validation", "fixed discount must be greater than 0 and less than price")
		}
		applyDiscountToModel(&existingMenu, req.Discount)
	}

	if req.ModifiersSet {
		modifierGroupIDs, _, err := s.createModifierGroupsFromRequest(ctx, req.Modifiers)
		if err != nil {
			return err
		}
		existingMenu.Modifiers = modifierGroupIDs
	}

	return s.menuRepository.Update(ctx, existingMenu)
}

func (s *menuService) createModifierGroupsFromRequest(ctx context.Context, modifiers []modgroup.ModifierGroupRequest) (pq.StringArray, []modgroup.ModifierGroupDTO, error) {
	modifierGroupIDs := make(pq.StringArray, 0, len(modifiers))
	modifierGroupDTOs := make([]modgroup.ModifierGroupDTO, 0, len(modifiers))

	for _, mgReq := range modifiers {
		if err := mgReq.Validate(); err != nil {
			s.logger.Error("Failed to validate modifier group", "error", err)
			return nil, nil, err
		}

		optionIDs := make(pq.StringArray, 0, len(mgReq.Options))
		optionDTOs := make([]modoption.ModifierOptionDTO, 0, len(mgReq.Options))
		for _, optReq := range mgReq.Options {
			if err := optReq.Validate(); err != nil {
				s.logger.Error("Failed to validate modifier option", "error", err)
				return nil, nil, err
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
				return nil, nil, err
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
			MinSelections: mgReq.MinSelections.Int(),
			MaxSelections: mgReq.MaxSelections.Int(),
			Options:       optionIDs,
		}

		if err := s.modifierGroupService.Create(ctx, groupModel); err != nil {
			s.logger.Error("Failed to create modifier group", "error", err)
			return nil, nil, err
		}

		modifierGroupIDs = append(modifierGroupIDs, groupModel.ID)
		modifierGroupDTOs = append(modifierGroupDTOs, groupModel.ToDTO(optionDTOs))
	}

	return modifierGroupIDs, modifierGroupDTOs, nil
}

func (s *menuService) updateMasterFromBranch(ctx context.Context, id, branchID string, req MenuRequest) error {
	if branchID == "" {
		return common.ErrUnAuthorized
	}
	if req.IsAvailable == nil {
		return common.ErrNoDataToUpdate
	}
	// Branch users may only toggle availability on inherited master items.
	if len(req.Ingredients) > 0 || req.ModifiersSet || req.Name != "" || req.CategoryID != "" ||
		req.Image != nil || req.IsFasting != nil || req.IsChefsChoice != nil || req.Description != "" || req.Price != 0 ||
		req.PreparationTime != 0 {
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
		if users.IsBranchStaffRoleString(role) {
			if branchID == "" {
				return common.ErrUnAuthorized
			}
			return s.menuRepository.SetBranchExcluded(ctx, branchID, id, true)
		}
		if !users.CanManageMerchantMaster(role, merchantID, existing.MerchantIDString()) {
			return common.ErrUnAuthorized
		}
		return s.menuRepository.Delete(ctx, id, "")
	}

	if users.IsBranchStaffRoleString(role) && existing.BranchIDString() != branchID {
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
	if users.IsBranchStaffRoleString(role) {
		if branchID == "" {
			return common.ErrUnAuthorized
		}
		br, err := s.branchService.Get(ctx, branchID)
		if err != nil {
			return err
		}
		// Restore excluded master item — loadMenuForAction skips excluded rows, so use GetMaster.
		master, err := s.menuRepository.GetMaster(ctx, id, br.MerchantID)
		if err == nil && master.IsMaster() {
			return s.menuRepository.SetBranchExcluded(ctx, branchID, id, false)
		}
		if err != nil && !errors.Is(err, common.ErrMenuNotFound) {
			return err
		}
		return s.menuRepository.UnDelete(ctx, id, branchID)
	}

	existing, err := s.loadMenuForAction(ctx, id, branchID, merchantID)
	if err != nil {
		return err
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
		case users.IsSuperBranchAdminRoleString(role) && branchID == "":
			scope = ScopeMaster
		case users.IsSuperAdminRoleString(role) && branchID == "":
			scope = ScopeAllBranches
		case users.IsSuperBranchAdminRoleString(role):
			scope = ScopeAllBranches
		case users.IsBranchStaffRoleString(role):
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

	if scope == ScopeBranchManage && branchID == "" {
		return nil, common.ErrUnAuthorized
	}

	if scope == ScopeAllBranches && users.IsSuperBranchAdminRoleString(role) && merchantID == "" {
		return nil, common.ErrBranchAdminMissingMerchant
	}

	if branchID != "" && (scope == ScopeBranchManage || scope == ScopeBranchEffective) {
		if br, err := s.branchService.Get(ctx, branchID); err == nil && br.MerchantID != "" {
			if err := s.menuRepository.AssignOrphanMasterMenus(ctx, br.MerchantID); err != nil {
				s.logger.Error("Failed to assign orphan master menus", "error", err)
			}
		}
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

func (s *menuService) ListMenus(ctx context.Context, filter common.Filter, reference, referenceType string) (*PublicMenuCatalogResponse, error) {
	var catalog *PublicMenuCatalogResponse
	var err error

	switch referenceType {
	case "table":
		catalog, err = s.menuRepository.ListMenusForTables(ctx, filter, reference)
	case "room":
		rm, roomErr := s.roomService.GetRoomByReference(ctx, reference)
		if roomErr != nil {
			return nil, roomErr
		}

		activeBooking, bookingErr := s.bookingService.GetBookingByRoom(ctx, rm.ID)
		if bookingErr != nil {
			s.logger.Error("no active booking for room", "room_id", rm.ID, "status", rm.Status, "error", bookingErr)
			return nil, bookingErr
		}

		catalog, err = s.menuRepository.ListMenusForRooms(ctx, filter, reference)
		if err != nil {
			return nil, err
		}
		if catalog.Guest == nil && activeBooking.GuestName != "" {
			catalog.Guest = &booking.GuestResponseSimplified{GuestName: activeBooking.GuestName}
		}
	default:
		s.logger.Error("Invalid reference type", "reference_type", referenceType)
		return nil, common.ErrInvalidRequest
	}
	if err != nil {
		return nil, err
	}

	promotions, err := s.promotionService.ListActivePublicByReference(ctx, reference, referenceType)
	if err != nil {
		s.logger.Error("failed to list public promotions", "reference_type", referenceType, "error", err)
		return nil, err
	}
	if promotions == nil {
		promotions = []*promotion.PromotionPublicDTO{}
	}
	catalog.Promotions = promotions
	return catalog, nil
}
