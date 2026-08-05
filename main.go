package main

import (
	"clean-api/database"
	"clean-api/internal/storage"
	"clean-api/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	// Connect to Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "root:root@tcp(localhost:3306)/clean_app_db?parseTime=true"
	}
	database.Connect(dbURL)

	// Migrate Models (using the new database migration function)
	database.Migrate()

	// Connect to MinIO (optional — donation photo upload stays disabled if unset)
	minioStorage, err := storage.NewMinioStorage()
	if err != nil {
		log.Println("MinIO tidak aktif:", err)
	}

	// Setup Router
	r := routes.SetupRoutes(minioStorage)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server running on port " + port)
	r.Run(":" + port)
}
