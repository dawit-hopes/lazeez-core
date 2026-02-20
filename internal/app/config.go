package app

import (
	"fmt"
	"os"

	"lazeez-core/internal/common"
)

// getDatabaseDSN returns the database connection string from environment.
// Prefers individual DB_* vars over DATABASE_URL because the key=value format
// correctly handles special characters in passwords (=, +, ?, @, etc.).
func getDatabaseDSN() string {
	host := getEnv("DB_HOST", defaultDBHost())
	// Use individual vars when we have a host - avoids URL parsing issues with
	// special characters in passwords (e.g. =, +, ? break postgres:// URL format)
	if host != "" {
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", getEnv("POSTGRES_USER", "postgres"))
		password := getEnv("DB_PASSWORD", os.Getenv("POSTGRES_PASSWORD"))
		dbname := getEnv("DB_NAME", getEnv("POSTGRES_DB", "lazeez"))
		sslmode := getEnv("DB_SSLMODE", "disable")
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	// Fallback to DATABASE_URL (password must be URL-encoded if it contains =, +, ?, @)
	return os.Getenv("DATABASE_URL")
}

// defaultDBHost returns "db" when running in Docker (for docker-compose), "localhost" otherwise.
func defaultDBHost() string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "db"
	}
	return "localhost"
}

// getSecretKey returns the JWT secret key from environment
func getSecretKey() (string, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		fmt.Println("JWT_SECRET_KEY is not set")
		return "", common.ErrSecretKeyNotProvided
	}
	return secretKey, nil
}

// getPort returns the server port from environment
func getPort() string {
	return getEnv("PORT", "8080")
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
