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
	"lazeez-core/internal/feedback"
	"lazeez-core/internal/files"
	"lazeez-core/internal/ingredient"
	"lazeez-core/internal/key"
	menu "lazeez-core/internal/menu/dinning"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/middleware"
	modgroup "lazeez-core/internal/modifiers/group"
	modoption "lazeez-core/internal/modifiers/option"
	item "lazeez-core/internal/order/Item"
	"lazeez-core/internal/order/check"
	roomorder "lazeez-core/internal/order/room"
	order "lazeez-core/internal/order/table"
	"lazeez-core/internal/payment"
	"lazeez-core/internal/promotion"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/folio"
	"lazeez-core/internal/rooms/room"
	rooms "lazeez-core/internal/rooms/roomtype"
	"lazeez-core/internal/roomsession"
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
	RoomRepo           rooms.RoomRepository
	SingleRoomRepo     room.RoomRepository
	BookingRepo        booking.BookingRepository
	ClientSessionRepo  clientsession.ClientSessionRepository
	FolioRepo          folio.FolioRepository
	RoomSessionRepo    roomsession.RoomSessionRepository
	RoomOrderRepo      roomorder.RoomOrderRepository
	PromotionRepo      promotion.PromotionRepository
	FeedbackRepo       feedback.FeedbackRepository

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
	RoomTypeService       rooms.RoomTypesService
	SingleRoomService     room.RoomService
	BookingService        booking.BookingService
	ClientSessionService  clientsession.ClientSessionService
	FolioService          folio.FolioService
	RoomSessionService    roomsession.RoomSessionService
	RoomOrderService      roomorder.RoomOrderService
	CheckService          check.CheckService
	PaymentService        payment.PaymentService
	PromotionService      promotion.PromotionService
	FeedbackService       feedback.FeedbackService

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
	RoomHandler          rooms.RoomHandler
	SingleRoomHandler    room.RoomHandler
	BookingHandler       booking.BookingHandler
	ClientSessionHandler clientsession.ClientSessionHandler
	FolioHandler         folio.FolioHandler
	RoomSessionHandler   roomsession.RoomSessionHandler
	RoomOrderHandler     roomorder.RoomOrderHandler
	CheckHandler         check.CheckHandler
	PromotionHandler     promotion.PromotionHandler
	FeedbackHandler      feedback.FeedbackHandler

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
	tableCheckDAL := common.NewDAL(db, func() *check.TableCheck { return &check.TableCheck{} })
	tableDAL := common.NewDAL(db, func() *table.Table { return &table.Table{} })
	roomDAL := common.NewDAL(db, func() *rooms.Room { return &rooms.Room{} })
	singleRoomDAL := common.NewDAL(db, func() *room.Room { return &room.Room{} })
	bookingDAL := common.NewDAL(db, func() *booking.Booking { return &booking.Booking{} })
	clientSessionDAL := common.NewDAL(db, func() *clientsession.ClientSession { return &clientsession.ClientSession{} })
	roomBillDAL := common.NewDAL(db, func() *folio.Bill { return &folio.Bill{} })
	roomSessionDAL := common.NewDAL(db, func() *roomsession.RoomSession { return &roomsession.RoomSession{} })
	roomOrderDAL := common.NewDAL(db, func() *roomorder.RoomOrder { return &roomorder.RoomOrder{} })
	roomOrderItemDAL := common.NewDAL(db, func() *roomorder.RoomOrderItem { return &roomorder.RoomOrderItem{} })
	promotionDAL := common.NewDAL(db, func() *promotion.Promotion { return &promotion.Promotion{} })
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
	checkRepo := check.NewCheckRepository(tableCheckDAL, joinDAL, logger)
	tableRepo := table.NewTableRepository(tableDAL, logger)
	roomRepo := rooms.NewRoomRepository(roomDAL, joinDAL, logger)
	singleRoomRepo := room.NewRoomRepository(singleRoomDAL, joinDAL, logger)
	bookingRepo := booking.NewBookingRepository(bookingDAL, joinDAL, logger)
	clientSessionRepo := clientsession.NewClientSessionRepository(clientSessionDAL, joinDAL, logger)
	folioRepo := folio.NewFolioRepository(roomBillDAL, logger)
	roomSessionRepo := roomsession.NewRoomSessionRepository(roomSessionDAL, joinDAL, logger)
	roomOrderRepo := roomorder.NewRoomOrderRepository(roomOrderDAL, roomOrderItemDAL, joinDAL, logger)
	promotionRepo := promotion.NewPromotionRepository(promotionDAL, joinDAL, logger)
	feedbackRepo := feedback.NewFeedbackRepository(joinDAL, logger)

	// Initialize services
	paymentService := payment.NewPaymentService(chapaSecretKey, chapaInitialURL, verifyURL, webhookSecret, logger)
	branchService := branch.NewBranchService(branchRepo, logger)
	sessionService := session.NewSessionService(sessionRepo, logger)
	userService := users.NewUserService(userRepo, branchService, keyService, logger)
	authService := auth.NewAuthService(userService, sessionService, keyService, logger, secretKey)
	merchantService := merchant.NewMerchantService(merchantRepo, fileService, logger)
	roomTypeService := rooms.NewRoomTypeService(roomRepo, branchService, merchantService, fileService, logger)
	categoryService := category.NewCategoryService(categoryRepo, logger)
	ingredientService := ingredient.NewIngredientService(ingredientRepo, logger)
	modifierGroupService := modgroup.NewModifierGroupService(modifierGroupRepo, logger)
	modifierOptionService := modoption.NewModifierOptionService(modifierOptionRepo, logger)
	singleRoomService := room.NewRoomService(singleRoomRepo, roomRepo, branchService, merchantService, fileService, logger)
	folioService := folio.NewFolioService(folioRepo, logger)
	bookingService := booking.NewBookingService(bookingRepo, singleRoomRepo, branchService, merchantService, keyService, folioService, logger)
	promotionService := promotion.NewPromotionService(promotionRepo, fileService, logger)
	menuService := menu.NewMenuService(menuRepo, fileService, categoryService, branchService, ingredientService, modifierGroupService, modifierOptionService, singleRoomService, bookingService, promotionService, logger)
	orderItemService := item.NewOrderItemService(orderItemRepo, logger)
	clientSessionService := clientsession.NewClientSessionService(clientSessionRepo, logger)
	checkService := check.NewCheckService(checkRepo, logger)
	orderService := order.NewOrderService(orderRepo, orderItemService, menuService, modifierOptionService, clientSessionService, paymentService, userService, checkService, logger, callbackURL, menuBaseURL)
	feedbackService := feedback.NewFeedbackService(feedbackRepo, orderRepo, clientSessionService, logger)
	tableService := table.NewTableService(tableRepo, fileService, branchService, logger)
	roomSessionService := roomsession.NewRoomSessionService(roomSessionRepo, singleRoomRepo, bookingService, logger)
	roomOrderService := roomorder.NewRoomOrderService(roomOrderRepo, menuService, modifierOptionService, folioService, singleRoomService, bookingService, logger)

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
	roomHandler := rooms.NewRoomHandler(roomTypeService, logger)
	singleRoomHandler := room.NewRoomHandler(singleRoomService, logger)
	bookingHandler := booking.NewBookingHandler(bookingService, logger)
	clientSessionHandler := clientsession.NewClientSessionHandler(clientSessionService, logger)
	folioHandler := folio.NewFolioHandler(folioService, logger)
	roomSessionHandler := roomsession.NewRoomSessionHandler(roomSessionService, logger)
	roomOrderHandler := roomorder.NewRoomOrderHandler(roomOrderService, logger)
	checkHandler := check.NewCheckHandler(checkService, logger)
	promotionHandler := promotion.NewPromotionHandler(promotionService, userService, logger)
	feedbackHandler := feedback.NewFeedbackHandler(feedbackService, logger)

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
		RoomRepo:              roomRepo,
		SingleRoomRepo:        singleRoomRepo,
		BookingRepo:           bookingRepo,
		ClientSessionRepo:     clientSessionRepo,
		FolioRepo:             folioRepo,
		RoomSessionRepo:       roomSessionRepo,
		RoomOrderRepo:         roomOrderRepo,
		PromotionRepo:         promotionRepo,
		FeedbackRepo:          feedbackRepo,
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
		RoomTypeService:       roomTypeService,
		SingleRoomService:     singleRoomService,
		BookingService:        bookingService,
		ClientSessionService:  clientSessionService,
		FolioService:          folioService,
		RoomSessionService:    roomSessionService,
		RoomOrderService:      roomOrderService,
		CheckService:          checkService,
		PaymentService:        paymentService,
		PromotionService:      promotionService,
		FeedbackService:       feedbackService,
		AuthHandler:           authHandler,
		UserHandler:           userHandler,
		BranchHandler:         branchHandler,
		MerchantHandler:       merchantHandler,
		MenuHandler:           menuHandler,
		CategoryHandler:       categoryHandler,
		IngredientHandler:     ingredientHandler,
		OrderHandler:          orderHandler,
		TableHandler:          tableHandler,
		RoomHandler:           roomHandler,
		SingleRoomHandler:     singleRoomHandler,
		BookingHandler:        bookingHandler,
		ClientSessionHandler:  clientSessionHandler,
		FolioHandler:          folioHandler,
		RoomSessionHandler:    roomSessionHandler,
		RoomOrderHandler:      roomOrderHandler,
		CheckHandler:          checkHandler,
		PromotionHandler:      promotionHandler,
		FeedbackHandler:       feedbackHandler,
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
	rooms.NewRoomRoutes(router, deps.RoomHandler, deps.Middleware)
	room.NewRoomRoutes(router, deps.SingleRoomHandler, deps.Middleware)
	booking.NewBookingRoutes(router, deps.BookingHandler, deps.Middleware)
	folio.NewFolioRoutes(router, deps.FolioHandler, deps.Middleware)
	roomsession.NewRoomSessionRoutes(router, deps.RoomSessionHandler, deps.Middleware)
	roomorder.NewRoomOrderRoutes(router, deps.RoomOrderHandler, deps.Middleware)
	check.NewCheckRoutes(router, deps.CheckHandler, deps.Middleware)
	promotion.NewPromotionRoutes(router, deps.PromotionHandler, deps.Middleware)
	feedback.NewFeedbackRoutes(router, deps.FeedbackHandler, deps.Middleware)
}
