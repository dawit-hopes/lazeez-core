package app

import (
	"database/sql"
	"fmt"

	"lazeez-core/config"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// initDB initializes the database connection
func initDB(logger config.Logger) (*sql.DB, error) {
	dsn := getDatabaseDSN()
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")
	return db, nil
}
