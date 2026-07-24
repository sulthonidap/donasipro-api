package main

import (
	"clean-api/database"
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

	// Setup Router
	r := routes.SetupRoutes()

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server running on port " + port)
	r.Run(":" + port)
}
