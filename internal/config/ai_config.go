package config

import (
	"fmt"
	"os"
	"strconv"

	"dmintegroff/internal/ai"
	"dmintegroff/internal/models"
	"gorm.io/gorm"
)

// LoadAIConfig загружает конфигурацию AI из базы данных
func LoadAIConfig(db *gorm.DB) *ai.AIConfig {
	// Проверяем глобальный переключатель AI
	aiEnabled := getEnvBoolOrDefault("AI_ENABLED", true)
	
	// Если AI отключен глобально, возвращаем отключенную конфигурацию
	if !aiEnabled {
		return &ai.AIConfig{
			Enabled: false,
		}
	}
	
	// Загружаем из базы данных
	if db != nil {
		if dbSettings, err := models.GetActiveAISettings(db); err == nil {
			config := &ai.AIConfig{
				OpenRouterAPIKey: dbSettings.OpenRouterAPIKey,
				OpenRouterModel:  dbSettings.OpenRouterModel,
				OpenRouterURL:    dbSettings.OpenRouterURL,
				
				MaxTokens:       dbSettings.MaxTokens,
				Temperature:     dbSettings.Temperature,
				RequestTimeout:  dbSettings.RequestTimeout,
				
				LocalLLMEnabled: dbSettings.LocalLLMEnabled,
				LocalLLMURL:     dbSettings.LocalLLMURL,
				
				OpenAIAPIKey:    dbSettings.OpenAIAPIKey,
				OpenAIModel:     dbSettings.OpenAIModel,
				
				Enabled: aiEnabled,
			}
			
			fmt.Printf("AI Config loaded from DB: OpenRouter=%v, OpenAI=%v, Model=%s\n", 
				dbSettings.OpenRouterAPIKey != "", 
				dbSettings.OpenAIAPIKey != "", 
				dbSettings.OpenRouterModel)
			
			return config
		} else {
			fmt.Printf("Failed to load AI config from DB: %v\n", err)
		}
	}
	
	// Если не удалось загрузить из БД, возвращаем пустую конфигурацию
	fmt.Printf("AI Config: No settings found in DB, AI not configured\n")
	return &ai.AIConfig{
		Enabled: aiEnabled,
		OpenRouterModel: "anthropic/claude-3.5-sonnet",
		OpenRouterURL: "https://openrouter.ai/api/v1",
		OpenAIModel: "gpt-4-1106-preview",
		MaxTokens: 4000,
		Temperature: 0.3,
		RequestTimeout: 30,
		LocalLLMURL: "http://localhost:11434",
	}
}



// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault возвращает int значение переменной окружения или значение по умолчанию
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvFloatOrDefault возвращает float32 значение переменной окружения или значение по умолчанию
func getEnvFloatOrDefault(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault возвращает bool значение переменной окружения или значение по умолчанию
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}