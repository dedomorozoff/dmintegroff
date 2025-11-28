package main

import (
	"gintegra/internal/database"
	"gintegra/internal/logger"
	"gintegra/internal/models"
	"gintegra/internal/routes"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default/env vars")
	}

	logger.Init()
	logger.Log.Info("Starting GIntegra server...")

	database.Connect()
	database.Migrate(&models.User{}, &models.Integration{}, &models.RequestLog{})
	database.SeedAdmin()

	r := routes.SetupRouter()

	r.Run(":8080")
}
