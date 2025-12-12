package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"github.com/joho/godotenv"
)

// Утилита для миграции AI настроек из .env в базу данных
func main() {
	fmt.Println("🔄 Миграция AI настроек из .env в базу данных...")
	
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env файл не найден, используем переменные окружения")
	}

	// Подключаемся к базе данных
	database.Connect()
	
	// Проверяем, есть ли уже настройки в БД
	var existingSettings models.AISettings
	if err := database.DB.Where("is_active = ?", true).First(&existingSettings).Error; err == nil {
		fmt.Println("✅ AI настройки уже существуют в базе данных:")
		fmt.Printf("   Провайдер: %s\n", existingSettings.GetCurrentProvider())
		fmt.Printf("   Модель: %s\n", existingSettings.OpenRouterModel)
		fmt.Printf("   Настроен: %v\n", existingSettings.IsConfigured())
		return
	}

	// Читаем настройки из переменных окружения
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	openRouterModel := getEnvOrDefault("OPENROUTER_MODEL", "anthropic/claude-3.5-sonnet")
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	openAIModel := getEnvOrDefault("OPENAI_MODEL", "gpt-4-1106-preview")
	
	maxTokens := getEnvIntOrDefault("AI_MAX_TOKENS", 4000)
	temperature := getEnvFloatOrDefault("AI_TEMPERATURE", 0.3)
	requestTimeout := getEnvIntOrDefault("AI_REQUEST_TIMEOUT", 30)

	// Создаем настройки для сохранения в БД
	aiSettings := &models.AISettings{
		OpenRouterAPIKey: openRouterAPIKey,
		OpenRouterModel:  openRouterModel,
		OpenRouterURL:    "https://openrouter.ai/api/v1",
		OpenAIAPIKey:     openAIAPIKey,
		OpenAIModel:      openAIModel,
		MaxTokens:        maxTokens,
		Temperature:      temperature,
		RequestTimeout:   requestTimeout,
		LocalLLMEnabled:  getEnvBoolOrDefault("LOCAL_LLM_ENABLED", false),
		LocalLLMURL:      getEnvOrDefault("LOCAL_LLM_URL", "http://localhost:11434"),
	}

	// Проверяем, есть ли что мигрировать
	if !aiSettings.IsConfigured() {
		fmt.Println("⚠️  AI настройки не найдены в переменных окружения")
		fmt.Println("   Создаем настройки по умолчанию...")
	} else {
		fmt.Println("📦 Найдены AI настройки в переменных окружения:")
		if openRouterAPIKey != "" {
			fmt.Printf("   OpenRouter API Key: %s...%s\n", openRouterAPIKey[:8], openRouterAPIKey[len(openRouterAPIKey)-4:])
			fmt.Printf("   OpenRouter Model: %s\n", openRouterModel)
		}
		if openAIAPIKey != "" {
			fmt.Printf("   OpenAI API Key: %s...%s\n", openAIAPIKey[:8], openAIAPIKey[len(openAIAPIKey)-4:])
			fmt.Printf("   OpenAI Model: %s\n", openAIModel)
		}
		fmt.Printf("   Max Tokens: %d\n", maxTokens)
		fmt.Printf("   Temperature: %.1f\n", temperature)
		fmt.Printf("   Request Timeout: %d сек\n", requestTimeout)
	}

	// Сохраняем в базу данных
	if err := models.SaveAISettings(database.DB, aiSettings); err != nil {
		log.Fatalf("❌ Ошибка сохранения настроек в БД: %v", err)
	}

	fmt.Println("✅ AI настройки успешно мигрированы в базу данных!")
	fmt.Println("   Теперь вы можете управлять ими через веб-интерфейс в разделе 'Настройки'")
	fmt.Println("   Можно удалить AI настройки из .env файла")
}

// Вспомогательные функции
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloatOrDefault(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}