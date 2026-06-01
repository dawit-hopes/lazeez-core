package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"lazeez-core/config"

	"log"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

// App holds all application dependencies
type App struct {
	DB     *sql.DB
	Logger config.Logger
	Router chi.Router
}

// NewApp initializes and returns a new App instance
func NewApp() (*App, error) {
	// Initialize logger
	logger := config.NewLogger()
	logger.Info("Initializing application...")

	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	// Initialize database connection
	db, err := initDB(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize router
	router := chi.NewRouter()

	// Initialize all dependencies
	callbackURL := getEnv("CHAPA_CALLBACK_URL", "")
	verifyURL := getEnv("CHAPA_VERIFY_URL", "https://api.chapa.co/v1/transaction/verify")
	chapaSecretKey := getEnv("CHAPA_SECRET_KEY", getEnv("CHAPA_SECRET", ""))
	chapaInitialURL := getEnv("CHAPA_INITIAL_URL", "https://api.chapa.co/v1/transaction/initialize")
	webhookSecret := getEnv("CHAPA_WEBHOOK_SECRET", chapaSecretKey)
	menuBaseURL := getEnv("LAZEEZ_MENU_BASE_URL", getEnv("LAZEEZ_TABLE_BASE_URL", "http://localhost:8082/"))
	logger.Info("Chapa payment return base URL", "menu_base_url", menuBaseURL)

	if callbackURL == "" || menuBaseURL == "" || verifyURL == "" || chapaSecretKey == "" || chapaInitialURL == "" || webhookSecret == "" {
		return nil, fmt.Errorf("missing required Chapa environment variables")
	}

	deps, err := initializeDependencies(db, logger, callbackURL, menuBaseURL, verifyURL, chapaSecretKey, chapaInitialURL, webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Register routes
	registerRoutes(router, deps)

	logger.Info("Application initialized successfully")

	return &App{
		DB:     db,
		Logger: logger,
		Router: router,
	}, nil
}

// Run starts the HTTP server with proper timeout configurations
func (a *App) Run() error {
	port := getPort()

	// Configure server with security timeouts
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,   // Time to read request headers
		ReadTimeout:       15 * time.Second,  // Time to read entire request body
		WriteTimeout:      15 * time.Second,  // Time to write response
		IdleTimeout:       120 * time.Second, // Time to keep idle connections open
		MaxHeaderBytes:    1 << 20,           // 1MB max header size
	}

	a.Logger.Info("Starting server", "port", port)
	return server.ListenAndServe()
}

// Close gracefully closes all connections
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
