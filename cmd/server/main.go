package main

import (
	"dmintegroff/internal/cache"
	"dmintegroff/internal/config"
	"dmintegroff/internal/controllers"
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/middleware"
	"dmintegroff/internal/models"
	"dmintegroff/internal/routes"
	"dmintegroff/internal/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default/env vars")
	}

	logger.Init()
	logger.Log.Info("Starting dmIntegroff server...")

	// Database connection first
	database.Connect()
	
	// Initialize AI configuration (after database connection)
	aiConfig := config.LoadAIConfig(database.DB)
	if aiConfig.IsConfigured() {
		logger.Log.Info("AI configured with provider: " + aiConfig.GetCurrentProvider())
	} else {
		logger.Log.Warn("AI not configured - set OPENROUTER_API_KEY or OPENAI_API_KEY in settings")
	}

	// Initialize Redis (optional)
	if err := cache.InitRedis(); err != nil {
		logger.Log.Warn("Redis initialization failed: " + err.Error())
	}

	database.Migrate(&models.User{}, &models.Project{}, &models.Integration{}, &models.IntegrationOutput{}, &models.RequestLog{}, &models.WebhookTest{}, &models.WebhookTestRequest{}, &models.AISettings{})
	database.SeedAdmin()

	// Initialize rate limiter
	rateLimitRate := 60  // requests per minute
	rateLimitBurst := 10 // burst capacity
	
	if rateStr := os.Getenv("RATE_LIMIT_RATE"); rateStr != "" {
		if rate, err := strconv.Atoi(rateStr); err == nil {
			rateLimitRate = rate
		}
	}
	
	if burstStr := os.Getenv("RATE_LIMIT_BURST"); burstStr != "" {
		if burst, err := strconv.Atoi(burstStr); err == nil {
			rateLimitBurst = burst
		}
	}
	
	middleware.InitRateLimiter(rateLimitRate, rateLimitBurst)

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
	
	// Use http.Server to ensure HOST binding is respected
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("Server failed to start: " + err.Error())
	}
}
