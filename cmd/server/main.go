package main

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/routes"
	"dmintegroff/internal/services"
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

	// Получаем *sql.DB из GORM для health service
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatal("Failed to get database connection:", err)
	}

	// Создание сервисов метрик и здоровья
	healthService := services.NewHealthService(sqlDB, "1.0.0")

	r := routes.SetupRouter(healthService)

	logger.Log.Info("Server starting on :8080")
	r.Run(":8080")
}
