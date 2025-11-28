package main

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/routes"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default/env vars")
	}

	logger.Init()
	logger.Log.Info("Starting dmIntegroff server...")

	database.Connect()
	database.Migrate(&models.User{}, &models.Project{}, &models.Integration{}, &models.RequestLog{})
	database.SeedAdmin()

	r := routes.SetupRouter()

	r.Run(":8080")
}
