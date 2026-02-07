package main

import (
	"log"

	"github.com/joho/godotenv"
	"lazeez-core/internal/app"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize application
	application, err := app.NewApp()
	if err != nil {
		panic(err)
	}
	defer application.Close()

	// Start server
	if err := application.Run(); err != nil {
		panic(err)
	}
}
