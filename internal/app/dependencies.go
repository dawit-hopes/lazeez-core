package app

import (
	"database/sql"

	"lazeez-core/config"
	"lazeez-core/internal/auth"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
	"lazeez-core/internal/menu"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// Dependencies holds all initialized dependencies
type Dependencies struct {
	KeyService key.KeyService

	// Repositories
	AuthRepo     auth.AuthRepository
	BranchRepo   branch.BranchRepository
	MerchantRepo merchant.MerchantRepository
	MenuRepo     menu.MenuRepository

	// Services
	AuthService     auth.AuthService
	BranchService   branch.BranchService
	MerchantService merchant.MerchantService
	MenuService     menu.MenuService

	// Handlers
	AuthHandler     auth.AuthHandler
	BranchHandler   branch.BranchHandler
	MerchantHandler merchant.MerchantHandler
	MenuHandler     menu.MenuHandler

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
	authDAL := common.NewDAL[*auth.User](db, func() *auth.User { return &auth.User{} })
	branchDAL := common.NewDAL[*branch.Branch](db, func() *branch.Branch { return &branch.Branch{} })
	merchantDAL := common.NewDAL[*merchant.Merchant](db, func() *merchant.Merchant { return &merchant.Merchant{} })
	menuDAL := common.NewDAL[*menu.Menu](db, func() *menu.Menu { return &menu.Menu{} })

	// Initialize repositories
	authRepo := auth.NewAuthRepository(authDAL, logger)
	branchRepo := branch.NewBranchRepository(branchDAL, logger)
	merchantRepo := merchant.NewMerchantRepository(merchantDAL, logger)
	menuRepo := menu.NewMenuRepository(menuDAL, logger)

	// Initialize services
	authService := auth.NewAuthService(authRepo, branchRepo, keyService, logger)
	branchService := branch.NewBranchService(branchRepo, logger)
	merchantService := merchant.NewMerchantService(merchantRepo, logger)
	menuService := menu.NewMenuService(menuRepo, logger)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService, logger)
	branchHandler := branch.NewBranchHandler(branchService, logger)
	merchantHandler := merchant.NewMerchantHandler(merchantService, logger)
	menuHandler := menu.NewMenuHandler(menuService, logger)

	middleware := middleware.NewMiddleware(keyService, logger)

	return &Dependencies{
		KeyService: keyService,

		AuthRepo:     authRepo,
		BranchRepo:   branchRepo,
		MerchantRepo: merchantRepo,
		MenuRepo:     menuRepo,

		AuthService:     authService,
		BranchService:   branchService,
		MerchantService: merchantService,
		MenuService:     menuService,

		AuthHandler:     authHandler,
		BranchHandler:   branchHandler,
		MerchantHandler: merchantHandler,
		MenuHandler:     menuHandler,

		Middleware: middleware,
	}, nil
}

// registerRoutes registers all routes with the router
func registerRoutes(router chi.Router, deps *Dependencies) {
	router.Use(deps.Middleware.CORSHandler)
	router.MethodNotAllowed(deps.Middleware.MethodNotAllowedHandler)
	router.NotFound(deps.Middleware.NotFoundHandler)

	auth.NewAuthRoutes(router, deps.AuthHandler, deps.Middleware)
	branch.NewBranchRoutes(router, deps.BranchHandler, deps.Middleware)
	merchant.NewMerchantRoutes(router, deps.MerchantHandler, deps.Middleware)
	menu.NewMenuRoutes(router, deps.MenuHandler, deps.Middleware)
}
