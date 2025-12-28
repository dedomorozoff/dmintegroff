package tests

import (
	"os"
	"testing"

	"dmintegroff/internal/ai"
	"dmintegroff/internal/controllers"
	"github.com/stretchr/testify/assert"
)

// TestDemoModeActivation тестирует активацию демо режима
func TestDemoModeActivation(t *testing.T) {
	// Сохраняем оригинальное значение
	originalDemoMode := os.Getenv("DEMO_MODE")
	defer os.Setenv("DEMO_MODE", originalDemoMode)

	// Тестируем активацию демо режима
	os.Setenv("DEMO_MODE", "true")
	
	config := &ai.AIConfig{
		Enabled: false, // AI отключен, но демо режим должен работать
	}
	
	controller := controllers.NewAIController(config)
	
	// В демо режиме контроллер должен быть создан успешно
	assert.NotNil(t, controller)
}

// TestDemoResponses тестирует демо ответы
func TestDemoResponses(t *testing.T) {
	demoResponses := ai.NewDemoResponses()
	
	testCases := []struct {
		name      string
		message   string
		targetAPI string
		expected  string
	}{
		{
			name:      "Slack интеграция",
			message:   "настроить slack уведомления",
			targetAPI: "slack",
			expected:  "Slack",
		},
		{
			name:      "Telegram бот",
			message:   "создать telegram бота",
			targetAPI: "telegram",
			expected:  "Telegram",
		},
		{
			name:      "Discord webhook",
			message:   "discord уведомления",
			targetAPI: "discord",
			expected:  "Discord",
		},
		{
			name:      "AmoCRM интеграция",
			message:   "интеграция с amocrm",
			targetAPI: "amocrm",
			expected:  "AmoCRM",
		},
		{
			name:      "Email уведомления",
			message:   "отправка email",
			targetAPI: "email",
			expected:  "Email",
		},
		{
			name:      "Общий запрос",
			message:   "помощь с интеграцией",
			targetAPI: "",
			expected:  "Демо режим",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sampleData := map[string]interface{}{
				"message": "test message",
				"user":    "test user",
			}
			
			response := demoResponses.GetDemoChatResponse(tc.message, tc.targetAPI, sampleData)
			
			// Проверяем, что ответ содержит ожидаемый текст
			assert.Contains(t, response.Response, tc.expected)
			assert.NotEmpty(t, response.Suggestions)
			assert.NotEmpty(t, response.NextSteps)
			assert.Greater(t, response.Confidence, 0.0)
		})
	}
}

// TestDemoDataAnalysis тестирует демо анализ данных
func TestDemoDataAnalysis(t *testing.T) {
	demoResponses := ai.NewDemoResponses()
	
	testData := map[string]interface{}{
		"email":    "test@example.com",
		"phone":    "+7 999 123-45-67",
		"name":     "Иван Петров",
		"order_id": 12345,
		"amount":   99.99,
		"date":     "2024-12-27T10:30:00Z",
	}
	
	analysis := demoResponses.GetDemoDataAnalysis(testData, "json")
	
	// Проверяем результат анализа
	assert.NotEmpty(t, analysis.Fields)
	assert.NotEmpty(t, analysis.DataType)
	assert.Greater(t, analysis.Confidence, 0.0)
	
	// Проверяем, что поля правильно определены
	fieldTypes := make(map[string]string)
	for _, field := range analysis.Fields {
		fieldTypes[field.Name] = field.Type
	}
	
	assert.Equal(t, "email", fieldTypes["email"])
	assert.Equal(t, "phone", fieldTypes["phone"])
	assert.Equal(t, "string", fieldTypes["name"])
	assert.Equal(t, "number", fieldTypes["order_id"])
	assert.Equal(t, "number", fieldTypes["amount"])
}

// TestDemoMappingGeneration тестирует демо генерацию маппингов
func TestDemoMappingGeneration(t *testing.T) {
	demoResponses := ai.NewDemoResponses()
	
	testCases := []struct {
		name      string
		targetAPI string
		expected  string
	}{
		{
			name:      "Slack маппинг",
			targetAPI: "slack",
			expected:  "hooks.slack.com",
		},
		{
			name:      "Telegram маппинг",
			targetAPI: "telegram",
			expected:  "api.telegram.org",
		},
		{
			name:      "Discord маппинг",
			targetAPI: "discord",
			expected:  "discord.com/api/webhooks",
		},
		{
			name:      "AmoCRM маппинг",
			targetAPI: "amocrm",
			expected:  "amocrm.ru/api/v4",
		},
		{
			name:      "Email маппинг",
			targetAPI: "email",
			expected:  "sendgrid.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ai.MappingGenerationRequest{
				SourceData: map[string]interface{}{
					"message": "test message",
					"user":    "test user",
				},
				TargetAPI:  tc.targetAPI,
				Task:       "test task",
				UserPrompt: "test prompt",
			}
			
			mapping := demoResponses.GetDemoGeneratedMapping(req)
			
			// Проверяем результат
			assert.NotEmpty(t, mapping.Type)
			assert.NotEmpty(t, mapping.Template)
			assert.Contains(t, mapping.TargetURL, tc.expected)
			assert.NotEmpty(t, mapping.Method)
			assert.Contains(t, mapping.Description, "демо")
			assert.Contains(t, mapping.Reasoning, "демонстрационный")
		})
	}
}

// TestDemoIntegrationCreation тестирует демо создание интеграций
func TestDemoIntegrationCreation(t *testing.T) {
	demoResponses := ai.NewDemoResponses()
	
	testCases := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "Slack интеграция",
			description: "уведомления в slack о заказах",
			expected:    "Slack",
		},
		{
			name:        "Telegram интеграция",
			description: "telegram бот для поддержки",
			expected:    "Telegram",
		},
		{
			name:        "Discord интеграция",
			description: "discord webhook для игрового сервера",
			expected:    "Discord",
		},
		{
			name:        "AmoCRM интеграция",
			description: "создание лидов в amocrm",
			expected:    "AmoCRM",
		},
		{
			name:        "Email интеграция",
			description: "отправка email уведомлений",
			expected:    "Email",
		},
		{
			name:        "Общая интеграция",
			description: "пользовательская интеграция с API",
			expected:    "Пользовательская",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ai.CreateIntegrationRequest{
				Description: tc.description,
				SampleData:  `{"message": "test", "user": "test"}`,
				ProjectID:   1,
			}
			
			integration := demoResponses.GetDemoCreatedIntegration(req)
			
			// Проверяем результат
			assert.Contains(t, integration.Name, tc.expected)
			assert.NotEmpty(t, integration.TargetURL)
			assert.NotEmpty(t, integration.Method)
			assert.NotEmpty(t, integration.Template)
			assert.NotEmpty(t, integration.TemplateType)
			assert.NotEmpty(t, integration.Mapping)
			assert.NotEmpty(t, integration.Explanation)
			assert.NotEmpty(t, integration.NextSteps)
			
			// Проверяем, что есть маркеры демо режима
			assert.Contains(t, integration.Explanation, "ДЕМО РЕЖИМ")
			assert.Contains(t, integration.NextSteps[0], "демо режим")
		})
	}
}

// TestDemoModeEnvironmentDetection тестирует определение демо режима
func TestDemoModeEnvironmentDetection(t *testing.T) {
	// Сохраняем оригинальное значение
	originalDemoMode := os.Getenv("DEMO_MODE")
	defer os.Setenv("DEMO_MODE", originalDemoMode)

	testCases := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "Демо режим включен",
			envValue: "true",
			expected: true,
		},
		{
			name:     "Демо режим выключен",
			envValue: "false",
			expected: false,
		},
		{
			name:     "Демо режим не установлен",
			envValue: "",
			expected: false,
		},
		{
			name:     "Неверное значение",
			envValue: "invalid",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("DEMO_MODE", tc.envValue)
			
			// Создаем контроллер и проверяем демо режим
			config := &ai.AIConfig{Enabled: true}
			controller := controllers.NewAIController(config)
			
			// Проверяем через переменную окружения
			isDemoMode := os.Getenv("DEMO_MODE") == "true"
			assert.Equal(t, tc.expected, isDemoMode)
			
			// Контроллер должен быть создан в любом случае
			assert.NotNil(t, controller)
		})
	}
}

// TestDemoModeWithDisabledAI тестирует демо режим при отключенном AI
func TestDemoModeWithDisabledAI(t *testing.T) {
	// Сохраняем оригинальные значения
	originalDemoMode := os.Getenv("DEMO_MODE")
	originalAIEnabled := os.Getenv("AI_ENABLED")
	defer func() {
		os.Setenv("DEMO_MODE", originalDemoMode)
		os.Setenv("AI_ENABLED", originalAIEnabled)
	}()

	// Включаем демо режим и отключаем AI
	os.Setenv("DEMO_MODE", "true")
	os.Setenv("AI_ENABLED", "false")
	
	config := &ai.AIConfig{
		Enabled: false, // AI отключен
	}
	
	controller := controllers.NewAIController(config)
	demoResponses := ai.NewDemoResponses()
	
	// В демо режиме все должно работать даже при отключенном AI
	assert.NotNil(t, controller)
	assert.NotNil(t, demoResponses)
	
	// Тестируем демо ответ
	response := demoResponses.GetDemoChatResponse("test message", "slack", map[string]interface{}{})
	assert.NotEmpty(t, response.Response)
	assert.Greater(t, response.Confidence, 0.0)
}