package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"lazeez-core/config"

	"github.com/go-chi/chi/v5"
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

	// Initialize database connection
	db, err := initDB(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize router
	router := chi.NewRouter()

	// Initialize all dependencies
	deps, err := initializeDependencies(db, logger)
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

// Run starts the HTTP server
func (a *App) Run() error {
	port := getPort()
	a.Logger.Info("Starting server", "port", port)
	return http.ListenAndServe(":"+port, a.Router)
}

// Close gracefully closes all connections
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
