package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"dmintegroff/internal/ai"
	"dmintegroff/internal/config"
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAIIntegrationPrompts тестирует различные промпты для создания интеграций
func TestAIIntegrationPrompts(t *testing.T) {
	// Инициализируем тестовую базу данных
	setupTestDB(t)
	defer teardownTestDB(t)

	// Загружаем AI конфигурацию
	aiConfig := config.LoadAIConfig(database.DB)
	
	// Создаем AI клиент и генератор
	client := ai.NewClient(aiConfig)
	generator := ai.NewGenerator(client)

	// Тестовые сценарии
	testCases := []struct {
		name         string
		description  string
		sampleData   string
		expectedType string
		shouldWork   bool
	}{
		{
			name:        "Slack уведомления",
			description: "Отправлять уведомления в Slack при новых заказах",
			sampleData: `{
				"order_id": "12345",
				"customer_name": "Иван Иванов",
				"amount": 1500,
				"email": "ivan@example.com",
				"status": "new"
			}`,
			expectedType: "slack",
			shouldWork:   true,
		},
		{
			name:        "Telegram бот",
			description: "Отправлять сообщения в Telegram бот при получении заявок",
			sampleData: `{
				"lead_id": "67890",
				"name": "Петр Петров",
				"phone": "+7 999 123-45-67",
				"message": "Интересует ваш продукт",
				"source": "website"
			}`,
			expectedType: "telegram",
			shouldWork:   true,
		},
		{
			name:        "Discord webhook",
			description: "Уведомления в Discord о новых пользователях",
			sampleData: `{
				"user_id": "user_123",
				"username": "newuser",
				"email": "user@example.com",
				"registration_date": "2024-12-16T10:30:00Z"
			}`,
			expectedType: "discord",
			shouldWork:   true,
		},
		{
			name:        "CRM интеграция",
			description: "Создавать лиды в Salesforce CRM",
			sampleData: `{
				"first_name": "Анна",
				"last_name": "Сидорова",
				"email": "anna@example.com",
				"phone": "+7 999 987-65-43",
				"company": "ООО Тест",
				"source": "web_form"
			}`,
			expectedType: "salesforce",
			shouldWork:   true,
		},
		{
			name:        "Email уведомления",
			description: "Отправлять email уведомления о статусе заказа",
			sampleData: `{
				"order_id": "ORD-001",
				"customer_email": "customer@example.com",
				"status": "shipped",
				"tracking_number": "TRK123456789"
			}`,
			expectedType: "email",
			shouldWork:   true,
		},
		{
			name:        "Пользовательский webhook",
			description: "Отправлять данные на внешний API для аналитики",
			sampleData: `{
				"event": "page_view",
				"user_id": "user_456",
				"page": "/products",
				"timestamp": "2024-12-16T10:30:00Z",
				"ip": "192.168.1.1"
			}`,
			expectedType: "webhook",
			shouldWork:   true,
		},
		{
			name:        "Некорректные данные",
			description: "Тест с некорректным JSON",
			sampleData:  `{invalid json}`,
			shouldWork:  false,
		},
		{
			name:        "Пустое описание",
			description: "",
			sampleData: `{"test": "data"}`,
			shouldWork:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Создаем запрос на создание интеграции
			req := &ai.CreateIntegrationRequest{
				Description: tc.description,
				SampleData:  tc.sampleData,
				ProjectID:   1, // Тестовый проект
			}

			// Выполняем создание интеграции
			result, err := generator.CreateIntegration(ctx, req)

			if !tc.shouldWork {
				// Ожидаем ошибку для некорректных данных
				assert.Error(t, err, "Должна быть ошибка для некорректных данных")
				return
			}

			// Проверяем успешное выполнение
			require.NoError(t, err, "Не должно быть ошибки при создании интеграции")
			require.NotNil(t, result, "Результат не должен быть nil")

			// Проверяем основные поля результата
			assert.NotEmpty(t, result.Name, "Название интеграции не должно быть пустым")
			assert.NotEmpty(t, result.TargetURL, "URL не должен быть пустым")
			assert.NotEmpty(t, result.Method, "HTTP метод не должен быть пустым")
			assert.NotEmpty(t, result.Template, "Шаблон не должен быть пустым")
			assert.NotEmpty(t, result.Explanation, "Объяснение не должно быть пустым")
			assert.NotEmpty(t, result.NextSteps, "Следующие шаги не должны быть пустыми")

			// Проверяем корректность JSON шаблона
			if result.TemplateType == "json" || result.TemplateType == "" {
				var templateData interface{}
				err := json.Unmarshal([]byte(result.Template), &templateData)
				assert.NoError(t, err, "Шаблон должен быть валидным JSON")
			}

			// Проверяем соответствие ожидаемому типу
			if tc.expectedType != "" {
				targetURL := strings.ToLower(result.TargetURL)
				switch tc.expectedType {
				case "slack":
					assert.Contains(t, targetURL, "slack", "URL должен содержать 'slack'")
				case "telegram":
					assert.Contains(t, targetURL, "telegram", "URL должен содержать 'telegram'")
				case "discord":
					assert.Contains(t, targetURL, "discord", "URL должен содержать 'discord'")
				case "salesforce":
					assert.Contains(t, targetURL, "salesforce", "URL должен содержать 'salesforce'")
				case "email":
					assert.True(t, 
						strings.Contains(targetURL, "sendgrid") || 
						strings.Contains(targetURL, "mailgun") ||
						strings.Contains(targetURL, "mail"),
						"URL должен содержать email сервис")
				}
			}

			// Выводим результат для анализа
			t.Logf("Результат для сценария '%s':", tc.name)
			t.Logf("  Название: %s", result.Name)
			t.Logf("  URL: %s", result.TargetURL)
			t.Logf("  Метод: %s", result.Method)
			t.Logf("  Тип шаблона: %s", result.TemplateType)
			t.Logf("  Тип авторизации: %s", result.AuthType)
			t.Logf("  Объяснение: %s", result.Explanation)
			t.Logf("  Шаблон (первые 200 символов): %s", truncateString(result.Template, 200))
		})
	}
}

// TestAIPromptGeneration тестирует генерацию промптов
func TestAIPromptGeneration(t *testing.T) {
	// Тестируем различные типы промптов
	testCases := []struct {
		name        string
		promptType  string
		data        map[string]interface{}
		description string
	}{
		{
			name:       "Анализ данных заказа",
			promptType: "data_analysis",
			data: map[string]interface{}{
				"order_id":      "12345",
				"customer_name": "Иван Иванов",
				"amount":        1500.50,
				"items": []interface{}{
					map[string]interface{}{
						"name":     "Товар 1",
						"quantity": 2,
						"price":    750.25,
					},
				},
			},
		},
		{
			name:        "Генерация маппинга для Slack",
			promptType:  "mapping_generation",
			description: "Отправлять уведомления в Slack",
			data: map[string]interface{}{
				"event":   "new_order",
				"message": "Новый заказ получен",
				"amount":  1000,
			},
		},
		{
			name:        "Чат с AI",
			promptType:  "chat",
			description: "Помоги настроить интеграцию с Telegram",
			data: map[string]interface{}{
				"user_message": "Как настроить бота?",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := ai.NewChatPromptBuilder()

			var prompt string
			switch tc.promptType {
			case "data_analysis":
				prompt = builder.BuildDataAnalysisPrompt(tc.data, "json")
			case "mapping_generation":
				prompt = builder.BuildMappingPrompt(tc.data, "slack", tc.description, tc.description)
			case "chat":
				prompt = builder.BuildChatPrompt(tc.description, tc.data)
			}

			// Проверяем, что промпт не пустой
			assert.NotEmpty(t, prompt, "Промпт не должен быть пустым")
			
			// Проверяем, что промпт содержит ключевые элементы
			assert.Contains(t, prompt, tc.description, "Промпт должен содержать описание")
			
			// Выводим промпт для анализа
			t.Logf("Сгенерированный промпт для '%s':", tc.name)
			t.Logf("%s", truncateString(prompt, 500))
		})
	}
}

// TestPopularAPIs тестирует получение популярных API
func TestPopularAPIs(t *testing.T) {
	apis := ai.GetPopularAPIs()
	
	// Проверяем, что список не пустой
	assert.NotEmpty(t, apis, "Список популярных API не должен быть пустым")
	
	// Проверяем наличие основных API
	apiNames := make(map[string]bool)
	for _, api := range apis {
		apiNames[strings.ToLower(api.Name)] = true
		
		// Проверяем обязательные поля
		assert.NotEmpty(t, api.Name, "Название API не должно быть пустым")
		assert.NotEmpty(t, api.Description, "Описание API не должно быть пустым")
		assert.NotEmpty(t, api.AuthType, "Тип авторизации не должен быть пустым")
		
		// Проверяем шаблоны
		assert.NotEmpty(t, api.Templates, "У API должны быть шаблоны")
		
		for _, template := range api.Templates {
			assert.NotEmpty(t, template.Name, "Название шаблона не должно быть пустым")
			assert.NotEmpty(t, template.Method, "HTTP метод не должен быть пустым")
			assert.NotEmpty(t, template.BodyTemplate, "Шаблон тела не должен быть пустым")
		}
	}
	
	// Проверяем наличие ключевых API
	expectedAPIs := []string{"slack", "telegram", "discord"}
	for _, expected := range expectedAPIs {
		assert.True(t, apiNames[expected], fmt.Sprintf("Должен быть API для %s", expected))
	}
	
	t.Logf("Найдено %d популярных API", len(apis))
	for _, api := range apis {
		t.Logf("  - %s (%s, %d шаблонов)", api.Name, api.AuthType, len(api.Templates))
	}
}

// TestQuickSuggestions тестирует быстрые предложения
func TestQuickSuggestions(t *testing.T) {
	suggestions := ai.GetQuickSuggestions()
	
	// Проверяем, что список не пустой
	assert.NotEmpty(t, suggestions, "Список предложений не должен быть пустым")
	
	for _, suggestion := range suggestions {
		// Проверяем обязательные поля
		assert.NotEmpty(t, suggestion.Type, "Тип предложения не должен быть пустым")
		assert.NotEmpty(t, suggestion.Title, "Заголовок не должен быть пустым")
		assert.NotEmpty(t, suggestion.Description, "Описание не должно быть пустым")
		assert.True(t, suggestion.Confidence > 0, "Уверенность должна быть больше 0")
		assert.True(t, suggestion.Confidence <= 1, "Уверенность должна быть не больше 1")
	}
	
	t.Logf("Найдено %d быстрых предложений", len(suggestions))
	for _, suggestion := range suggestions {
		t.Logf("  - %s (%.1f%%)", suggestion.Title, suggestion.Confidence*100)
	}
}

// Вспомогательные функции

func setupTestDB(t *testing.T) {
	// Здесь должна быть инициализация тестовой БД
	// Для простоты используем существующую конфигурацию
	if database.DB == nil {
		t.Skip("База данных не инициализирована")
	}
	
	// Создаем тестовый проект
	testProject := models.Project{
		Name:        "Тестовый проект",
		Description: "Проект для тестирования AI",
		CreatedByID: 1,
	}
	database.DB.Create(&testProject)
}

func teardownTestDB(t *testing.T) {
	// Очищаем тестовые данные
	if database.DB != nil {
		database.DB.Where("name = ?", "Тестовый проект").Delete(&models.Project{})
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}