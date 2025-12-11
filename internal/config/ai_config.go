package config

import (
	"os"
	"strconv"

	"dmintegroff/internal/ai"
)

// LoadAIConfig загружает конфигурацию AI из переменных окружения
func LoadAIConfig() *ai.AIConfig {
	config := &ai.AIConfig{
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterModel:  getEnvOrDefault("OPENROUTER_MODEL", "anthropic/claude-3.5-sonnet"),
		OpenRouterURL:    getEnvOrDefault("OPENROUTER_URL", "https://openrouter.ai/api/v1"),
		
		MaxTokens:       getEnvIntOrDefault("AI_MAX_TOKENS", 4000),
		Temperature:     getEnvFloatOrDefault("AI_TEMPERATURE", 0.3),
		RequestTimeout:  getEnvIntOrDefault("AI_REQUEST_TIMEOUT", 30),
		
		LocalLLMEnabled: getEnvBoolOrDefault("LOCAL_LLM_ENABLED", false),
		LocalLLMURL:     getEnvOrDefault("LOCAL_LLM_URL", "http://localhost:11434"),
		
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:     getEnvOrDefault("OPENAI_MODEL", "gpt-4-1106-preview"),
	}
	
	return config
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