package main

import (
	"dmintegroff/internal/cache"
	"dmintegroff/internal/controllers"
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/routes"
	"dmintegroff/internal/services"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default/env vars")
	}

	logger.Init()
	logger.Log.Info("Starting dmIntegroff server...")

	// Initialize Redis (optional)
	if err := cache.InitRedis(); err != nil {
		logger.Log.Warn("Redis initialization failed: " + err.Error())
	}

	database.Connect()
	database.Migrate(&models.User{}, &models.Project{}, &models.Integration{}, &models.IntegrationOutput{}, &models.RequestLog{}, &models.WebhookTest{}, &models.WebhookTestRequest{})
	database.SeedAdmin()

	// Cleanup expired test webhooks on startup
	controllers.CleanupExpiredWebhooks()

	// Schedule cleanup every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			controllers.CleanupExpiredWebhooks()
		}
	}()

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
