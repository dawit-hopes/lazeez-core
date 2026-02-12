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
	"lazeez-core/internal/session"
	"lazeez-core/internal/users"

	"github.com/go-chi/chi/v5"
)

// Dependencies holds all initialized dependencies
type Dependencies struct {
	KeyService key.KeyService

	// Repositories
	UserRepo     users.UserRepository
	BranchRepo   branch.BranchRepository
	MerchantRepo merchant.MerchantRepository
	MenuRepo     menu.MenuRepository

	// Services
	AuthService     auth.AuthService
	UserService     users.UserService
	BranchService   branch.BranchService
	MerchantService merchant.MerchantService
	MenuService     menu.MenuService

	// Handlers
	AuthHandler     auth.AuthHandler
	UserHandler     users.UserHandler
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
	userDAL := common.NewDAL[*users.User](db, func() *users.User { return &users.User{} })
	branchDAL := common.NewDAL[*branch.Branch](db, func() *branch.Branch { return &branch.Branch{} })
	merchantDAL := common.NewDAL[*merchant.Merchant](db, func() *merchant.Merchant { return &merchant.Merchant{} })
	menuDAL := common.NewDAL[*menu.Menu](db, func() *menu.Menu { return &menu.Menu{} })
	sessionDAL := common.NewDAL[*session.Session](db, func() *session.Session { return &session.Session{} })

	// Initialize repositories
	userRepo := users.NewUserRepository(userDAL, logger)
	branchRepo := branch.NewBranchRepository(branchDAL, logger)
	merchantRepo := merchant.NewMerchantRepository(merchantDAL, logger)
	menuRepo := menu.NewMenuRepository(menuDAL, logger)
	sessionRepo := session.NewSessionRepository(sessionDAL, logger)

	// Initialize services
	branchService := branch.NewBranchService(branchRepo, logger)
	sessionService := session.NewSessionService(sessionRepo, logger)
	userService := users.NewUserService(userRepo, branchService, logger)
	authService := auth.NewAuthService(userService, sessionService, keyService, logger, secretKey)
	merchantService := merchant.NewMerchantService(merchantRepo, logger)
	menuService := menu.NewMenuService(menuRepo, logger)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService, logger)
	userHandler := users.NewUserHandler(userService, logger)
	branchHandler := branch.NewBranchHandler(branchService, logger)
	merchantHandler := merchant.NewMerchantHandler(merchantService, logger)
	menuHandler := menu.NewMenuHandler(menuService, logger)

	middleware := middleware.NewMiddleware(keyService, sessionService, logger)

	return &Dependencies{
		KeyService: keyService,

		UserRepo:     userRepo,
		BranchRepo:   branchRepo,
		MerchantRepo: merchantRepo,
		MenuRepo:     menuRepo,

		AuthService:     authService,
		UserService:     userService,
		BranchService:   branchService,
		MerchantService: merchantService,
		MenuService:     menuService,

		AuthHandler:     authHandler,
		UserHandler:     userHandler,
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
	users.NewUserRoutes(router, deps.UserHandler, deps.Middleware)
	branch.NewBranchRoutes(router, deps.BranchHandler, deps.Middleware)
	merchant.NewMerchantRoutes(router, deps.MerchantHandler, deps.Middleware)
	menu.NewMenuRoutes(router, deps.MenuHandler, deps.Middleware)
}
