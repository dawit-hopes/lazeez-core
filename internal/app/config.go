package app

import (
	"fmt"
	"lazeez-core/internal/common"
	"os"
)

// getDatabaseDSN returns the database connection string from environment
func getDatabaseDSN() string {
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		return dsn
	}

	// Fallback to individual environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "lazeez")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
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
