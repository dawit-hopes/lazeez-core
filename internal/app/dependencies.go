package app

import (
	"database/sql"
	"net/http"

	"lazeez-core/config"
	"lazeez-core/internal/auth"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/category"
	"lazeez-core/internal/clientsession"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/ingredient"
	"lazeez-core/internal/key"
	"lazeez-core/internal/menu"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/middleware"
	modgroup "lazeez-core/internal/modifiers/group"
	modoption "lazeez-core/internal/modifiers/option"
	"lazeez-core/internal/order"
	item "lazeez-core/internal/order/Item"
	"lazeez-core/internal/payment"
	"lazeez-core/internal/session"
	"lazeez-core/internal/table"
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
	OrderRepo          order.OrderRepository
	TableRepo          table.TableRepository
	ClientSessionRepo  clientsession.ClientSessionRepository

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
	OrderService          order.OrderService
	TableService          table.TableService
	ClientSessionService  clientsession.ClientSessionService
	PaymentService        payment.PaymentService

	// Handlers
	AuthHandler          auth.AuthHandler
	UserHandler          users.UserHandler
	BranchHandler        branch.BranchHandler
	MerchantHandler      merchant.MerchantHandler
	MenuHandler          menu.MenuHandler
	CategoryHandler      category.CategoryHandler
	IngredientHandler    ingredient.IngredientHandler
	OrderHandler         order.OrderHandler
	TableHandler         table.TableHandler
	ClientSessionHandler clientsession.ClientSessionHandler

	Middleware middleware.Middleware
}

// initializeDependencies initializes all dependencies in the correct order
func initializeDependencies(db *sql.DB, logger config.Logger, callbackURL string, menuBaseURL string, verifyURL string, chapaSecretKey string, chapaInitialURL string, webhookSecret string) (*Dependencies, error) {
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
	orderDAL := common.NewDAL(db, func() *order.Order { return &order.Order{} })
	orderItemDAL := common.NewDAL(db, func() *item.OrderItem { return &item.OrderItem{} })
	tableDAL := common.NewDAL(db, func() *table.Table { return &table.Table{} })
	clientSessionDAL := common.NewDAL(db, func() *clientsession.ClientSession { return &clientsession.ClientSession{} })
	joinDAL := common.NewJoinDAL(db)
	cld, err := initCloudinary(logger)
	if err != nil {
		return nil, err
	}
	fileService := files.NewFileService(logger, cld)

	// Initialize repositories
	userRepo := users.NewUserRepository(userDAL, joinDAL, logger)
	branchRepo := branch.NewBranchRepository(branchDAL, joinDAL, logger)
	merchantRepo := merchant.NewMerchantRepository(merchantDAL, joinDAL, logger)
	menuRepo := menu.NewMenuRepository(menuDAL, joinDAL, logger)
	categoryRepo := category.NewCategoryRepository(categoryDAL, logger)
	ingredientRepo := ingredient.NewIngredientRepository(ingredientDAL, logger)
	sessionRepo := session.NewSessionRepository(sessionDAL, logger)
	modifierGroupRepo := modgroup.NewModifierGroupRepository(modifierGroupDAL, logger)
	modifierOptionRepo := modoption.NewModifierOptionRepository(modifierOptionDAL, joinDAL, logger)
	orderRepo := order.NewOrderRepository(orderDAL, joinDAL, logger)
	orderItemRepo := item.NewOrderItemRepository(orderItemDAL, logger)
	tableRepo := table.NewTableRepository(tableDAL, logger)
	clientSessionRepo := clientsession.NewClientSessionRepository(clientSessionDAL, joinDAL, logger)


	// Initialize services
	paymentService := payment.NewPaymentService(chapaSecretKey, chapaInitialURL, verifyURL, webhookSecret, logger)
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
	orderItemService := item.NewOrderItemService(orderItemRepo, logger)
	clientSessionService := clientsession.NewClientSessionService(clientSessionRepo, logger)
	orderService := order.NewOrderService(orderRepo, orderItemService, menuService, modifierOptionService, clientSessionService, paymentService, logger, callbackURL, menuBaseURL)
	tableService := table.NewTableService(tableRepo, fileService, branchService, logger)


	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService, logger)
	userHandler := users.NewUserHandler(userService, logger)
	branchHandler := branch.NewBranchHandler(branchService, logger)
	merchantHandler := merchant.NewMerchantHandler(merchantService, logger)
	menuHandler := menu.NewMenuHandler(menuService, userService, logger)
	categoryHandler := category.NewCategoryHandler(categoryService, logger)
	ingredientHandler := ingredient.NewIngredientHandler(ingredientService, logger)
	orderHandler := order.NewOrderHandler(orderService, logger)
	tableHandler := table.NewTableHandler(tableService, logger)
	clientSessionHandler := clientsession.NewClientSessionHandler(clientSessionService, logger)

	middleware := middleware.NewMiddleware(keyService, sessionService, logger)

	return &Dependencies{
		KeyService: keyService,

		UserRepo:              userRepo,
		BranchRepo:            branchRepo,
		MerchantRepo:          merchantRepo,
		MenuRepo:              menuRepo,
		CategoryRepo:          categoryRepo,
		IngredientRepo:        ingredientRepo,
		ModifierGroupRepo:     modifierGroupRepo,
		ModifierOptionRepo:    modifierOptionRepo,
		OrderRepo:             orderRepo,
		TableRepo:             tableRepo,
		ClientSessionRepo:     clientSessionRepo,
		AuthService:           authService,
		UserService:           userService,
		BranchService:         branchService,
		MerchantService:       merchantService,
		MenuService:           menuService,
		CategoryService:       categoryService,
		IngredientService:     ingredientService,
		ModifierGroupService:  modifierGroupService,
		ModifierOptionService: modifierOptionService,
		OrderService:          orderService,
		TableService:          tableService,
		ClientSessionService:  clientSessionService,
		PaymentService:        paymentService,
		AuthHandler:           authHandler,
		UserHandler:           userHandler,
		BranchHandler:         branchHandler,
		MerchantHandler:       merchantHandler,
		MenuHandler:           menuHandler,
		CategoryHandler:       categoryHandler,
		IngredientHandler:     ingredientHandler,
		OrderHandler:          orderHandler,
		TableHandler:          tableHandler,
		ClientSessionHandler:  clientSessionHandler,
		Middleware:            middleware,
	}, nil
}

// registerRoutes registers all routes with the router
func registerRoutes(router chi.Router, deps *Dependencies) {
	router.Use(deps.Middleware.CORSHandler)
	router.MethodNotAllowed(deps.Middleware.MethodNotAllowedHandler)
	router.NotFound(deps.Middleware.NotFoundHandler)

	// Health check for cloud/orchestrators (no auth)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	auth.NewAuthRoutes(router, deps.AuthHandler, deps.Middleware)
	users.NewUserRoutes(router, deps.UserHandler, deps.Middleware)
	branch.NewBranchRoutes(router, deps.BranchHandler, deps.Middleware)
	merchant.NewMerchantRoutes(router, deps.MerchantHandler, deps.Middleware)
	menu.NewMenuRoutes(router, deps.MenuHandler, deps.Middleware)
	category.NewCategoryRoutes(router, deps.CategoryHandler, deps.Middleware)
	ingredient.NewIngredientRoutes(router, deps.IngredientHandler, deps.Middleware)
	order.NewOrderRoutes(router, deps.OrderHandler, deps.Middleware)
	clientsession.NewClientSessionRoutes(router, deps.ClientSessionHandler, deps.Middleware)
	table.NewTableRoutes(router, deps.TableHandler, deps.Middleware)
}
