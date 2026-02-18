package app

import (
	"database/sql"

	"lazeez-core/config"
	"lazeez-core/internal/auth"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/ingredient"
	"lazeez-core/internal/key"
	"lazeez-core/internal/menu"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/middleware"
	modgroup "lazeez-core/internal/modifiers/group"
	modoption "lazeez-core/internal/modifiers/option"
	"lazeez-core/internal/session"
	"lazeez-core/internal/users"

	"github.com/go-chi/chi/v5"
)

// Dependencies holds all initialized dependencies
type Dependencies struct {
	KeyService key.KeyService

	// Repositories
	UserRepo           users.UserRepository
	BranchRepo         branch.BranchRepository
	MerchantRepo       merchant.MerchantRepository
	MenuRepo           menu.MenuRepository
	CategoryRepo       category.CategoryRepository
	IngredientRepo     ingredient.IngredientRepository
	ModifierGroupRepo  modgroup.ModifierGroupRepository
	ModifierOptionRepo modoption.ModifierOptionRepository

	// Services
	AuthService           auth.AuthService
	UserService           users.UserService
	BranchService         branch.BranchService
	MerchantService       merchant.MerchantService
	MenuService           menu.MenuService
	CategoryService       category.CategoryService
	IngredientService     ingredient.IngredientService
	ModifierGroupService  modgroup.ModifierGroupService
	ModifierOptionService modoption.ModifierOptionService

	// Handlers
	AuthHandler       auth.AuthHandler
	UserHandler       users.UserHandler
	BranchHandler     branch.BranchHandler
	MerchantHandler   merchant.MerchantHandler
	MenuHandler       menu.MenuHandler
	CategoryHandler   category.CategoryHandler
	IngredientHandler ingredient.IngredientHandler

	Middleware middleware.Middleware
}

// initializeDependencies initializes all dependencies in the correct order
func initializeDependencies(db *sql.DB, logger config.Logger) (*Dependencies, error) {
	// Initialize shared services
	secretKey, err := getSecretKey()
	if err != nil {
		return nil, err
	}
	keyService := key.NewKeyService(logger, secretKey)

	// Initialize DAL instances
	userDAL := common.NewDAL(db, func() *users.User { return &users.User{} })
	branchDAL := common.NewDAL(db, func() *branch.Branch { return &branch.Branch{} })
	merchantDAL := common.NewDAL(db, func() *merchant.Merchant { return &merchant.Merchant{} })
	menuDAL := common.NewDAL(db, func() *menu.Menu { return &menu.Menu{} })
	categoryDAL := common.NewDAL(db, func() *category.Category { return &category.Category{} })
	ingredientDAL := common.NewDAL(db, func() *ingredient.Ingredient { return &ingredient.Ingredient{} })
	sessionDAL := common.NewDAL(db, func() *session.Session { return &session.Session{} })
	modifierGroupDAL := common.NewDAL(db, func() *modgroup.ModifierGroup { return &modgroup.ModifierGroup{} })
	modifierOptionDAL := common.NewDAL(db, func() *modoption.ModifierOption { return &modoption.ModifierOption{} })

	joinDAL := common.NewJoinDAL(db)
	cld, err := initCloudinary(logger)
	if err != nil {
		return nil, err
	}
	fileService := files.NewFileService(logger, cld)

	// Initialize repositories
	userRepo := users.NewUserRepository(userDAL, logger)
	branchRepo := branch.NewBranchRepository(branchDAL, joinDAL, logger)
	merchantRepo := merchant.NewMerchantRepository(merchantDAL, joinDAL, logger)
	menuRepo := menu.NewMenuRepository(menuDAL, logger)
	categoryRepo := category.NewCategoryRepository(categoryDAL, logger)
	ingredientRepo := ingredient.NewIngredientRepository(ingredientDAL, logger)
	sessionRepo := session.NewSessionRepository(sessionDAL, logger)
	modifierGroupRepo := modgroup.NewModifierGroupRepository(modifierGroupDAL, logger)
	modifierOptionRepo := modoption.NewModifierOptionRepository(modifierOptionDAL, logger)

	// Initialize services
	branchService := branch.NewBranchService(branchRepo, logger)
	sessionService := session.NewSessionService(sessionRepo, logger)
	userService := users.NewUserService(userRepo, branchService, logger)
	authService := auth.NewAuthService(userService, sessionService, keyService, logger, secretKey)
	merchantService := merchant.NewMerchantService(merchantRepo, fileService, logger)
	categoryService := category.NewCategoryService(categoryRepo, logger)
	ingredientService := ingredient.NewIngredientService(ingredientRepo, logger)
	modifierGroupService := modgroup.NewModifierGroupService(modifierGroupRepo, logger)
	modifierOptionService := modoption.NewModifierOptionService(modifierOptionRepo, logger)
	menuService := menu.NewMenuService(menuRepo, fileService, categoryService, branchService, ingredientService, modifierGroupService, modifierOptionService, logger)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService, logger)
	userHandler := users.NewUserHandler(userService, logger)
	branchHandler := branch.NewBranchHandler(branchService, logger)
	merchantHandler := merchant.NewMerchantHandler(merchantService, logger)
	menuHandler := menu.NewMenuHandler(menuService, logger)
	categoryHandler := category.NewCategoryHandler(categoryService, logger)
	ingredientHandler := ingredient.NewIngredientHandler(ingredientService, logger)

	middleware := middleware.NewMiddleware(keyService, sessionService, logger)

	return &Dependencies{
		KeyService: keyService,

		UserRepo:           userRepo,
		BranchRepo:         branchRepo,
		MerchantRepo:       merchantRepo,
		MenuRepo:           menuRepo,
		CategoryRepo:       categoryRepo,
		IngredientRepo:     ingredientRepo,
		ModifierGroupRepo:  modifierGroupRepo,
		ModifierOptionRepo: modifierOptionRepo,

		AuthService:           authService,
		UserService:           userService,
		BranchService:         branchService,
		MerchantService:       merchantService,
		MenuService:           menuService,
		CategoryService:       categoryService,
		IngredientService:     ingredientService,
		ModifierGroupService:  modifierGroupService,
		ModifierOptionService: modifierOptionService,

		AuthHandler:       authHandler,
		UserHandler:       userHandler,
		BranchHandler:     branchHandler,
		MerchantHandler:   merchantHandler,
		MenuHandler:       menuHandler,
		CategoryHandler:   categoryHandler,
		IngredientHandler: ingredientHandler,

		Middleware: middleware,
	}, nil
}

// registerRoutes registers all routes with the router
func registerRoutes(router chi.Router, deps *Dependencies) {
	router.Use(deps.Middleware.CORSHandler)
	router.MethodNotAllowed(deps.Middleware.MethodNotAllowedHandler)
	router.NotFound(deps.Middleware.NotFoundHandler)

	auth.NewAuthRoutes(router, deps.AuthHandler, deps.Middleware)
	users.NewUserRoutes(router, deps.UserHandler, deps.Middleware)
	branch.NewBranchRoutes(router, deps.BranchHandler, deps.Middleware)
	merchant.NewMerchantRoutes(router, deps.MerchantHandler, deps.Middleware)
	menu.NewMenuRoutes(router, deps.MenuHandler, deps.Middleware)
	category.NewCategoryRoutes(router, deps.CategoryHandler, deps.Middleware)
	ingredient.NewIngredientRoutes(router, deps.IngredientHandler, deps.Middleware)
}

