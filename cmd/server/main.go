package main

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/routes"
	"dmintegroff/internal/services"
	"fmt"
	"log"
	"os"

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

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	logger.Log.Info("Server starting on " + addr)
	r.Run(addr)
}
