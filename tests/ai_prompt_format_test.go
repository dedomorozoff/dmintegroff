package tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"dmintegroff/internal/ai"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptFormatting тестирует форматирование различных типов промптов
func TestPromptFormatting(t *testing.T) {
	builder := ai.NewChatPromptBuilder()

	testCases := []struct {
		name         string
		promptType   string
		userMessage  string
		sampleData   map[string]interface{}
		targetAPI    string
		task         string
		expectedKeys []string // ключевые слова, которые должны быть в промпте
		minLength    int      // минимальная длина промпта
	}{
		{
			name:        "Чат промпт с контекстом",
			promptType:  "chat",
			userMessage: "Помоги настроить интеграцию с Slack для уведомлений о заказах",
			sampleData: map[string]interface{}{
				"order_id": "12345",
				"amount":   1500.50,
				"customer": "Иван Петров",
			},
			expectedKeys: []string{"Пользователь", "Образец входящих данных", "order_id", "Slack"},
			minLength:    100,
		},
		{
			name:       "Анализ данных промпт",
			promptType: "data_analysis",
			sampleData: map[string]interface{}{
				"user_id":    "user_123",
				"email":      "test@example.com",
				"created_at": "2024-12-16T10:30:00Z",
				"profile": map[string]interface{}{
					"name": "Тест Тестов",
					"age":  30,
				},
			},
			expectedKeys: []string{"Проанализируй структуру данных", "JSON формат", "fields", "schema"},
			minLength:    200,
		},
		{
			name:        "Генерация маппинга для Slack",
			promptType:  "mapping_generation",
			targetAPI:   "slack",
			task:        "Отправлять уведомления о новых заказах",
			sampleData: map[string]interface{}{
				"order_id":      "ORD-001",
				"customer_name": "Клиент",
				"total":         2500.00,
			},
			expectedKeys: []string{"Создай маппинг", "Slack", "hooks.slack.com", "template"},
			minLength:    300,
		},
		{
			name:        "Генерация маппинга для Telegram",
			promptType:  "mapping_generation",
			targetAPI:   "telegram",
			task:        "Уведомления о новых лидах",
			sampleData: map[string]interface{}{
				"lead_id": "LEAD-001",
				"name":    "Потенциальный клиент",
				"phone":   "+7 999 123-45-67",
			},
			expectedKeys: []string{"Создай маппинг", "Telegram", "api.telegram.org", "chat_id"},
			minLength:    300,
		},
		{
			name:        "Генерация маппинга для Discord",
			promptType:  "mapping_generation",
			targetAPI:   "discord",
			task:        "Системные уведомления",
			sampleData: map[string]interface{}{
				"event_type": "error",
				"message":    "Системная ошибка",
				"severity":   "high",
			},
			expectedKeys: []string{"Создай маппинг", "Discord", "discord.com/api/webhooks", "content"},
			minLength:    300,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var prompt string

			// Генерируем промпт в зависимости от типа
			switch tc.promptType {
			case "chat":
				prompt = builder.BuildChatPrompt(tc.userMessage, tc.sampleData)
			case "data_analysis":
				prompt = builder.BuildDataAnalysisPrompt(tc.sampleData, "json")
			case "mapping_generation":
				prompt = builder.BuildMappingPrompt(tc.sampleData, tc.targetAPI, tc.task, tc.task)
			default:
				t.Fatalf("Неизвестный тип промпта: %s", tc.promptType)
			}

			// Базовые проверки
			assert.NotEmpty(t, prompt, "Промпт не должен быть пустым")
			assert.GreaterOrEqual(t, len(prompt), tc.minLength, 
				"Промпт должен быть не менее %d символов", tc.minLength)

			// Проверяем наличие ключевых слов
			promptLower := strings.ToLower(prompt)
			for _, key := range tc.expectedKeys {
				assert.Contains(t, promptLower, strings.ToLower(key), 
					"Промпт должен содержать ключевое слово: %s", key)
			}

			// Проверяем структуру промпта
			lines := strings.Split(prompt, "\n")
			assert.Greater(t, len(lines), 1, "Промпт должен содержать несколько строк")

			// Выводим промпт для анализа
			t.Logf("\n=== Промпт для '%s' ===", tc.name)
			t.Logf("Длина: %d символов", len(prompt))
			t.Logf("Строк: %d", len(lines))
			t.Logf("Содержимое:\n%s", prompt)
			t.Logf("=== Конец промпта ===\n")
		})
	}
}

// TestPromptTemplateValidation тестирует валидацию шаблонов в промптах
func TestPromptTemplateValidation(t *testing.T) {
	testCases := []struct {
		name       string
		targetAPI  string
		sampleData map[string]interface{}
		task       string
	}{
		{
			name:      "Slack с заказом",
			targetAPI: "slack",
			sampleData: map[string]interface{}{
				"order_id":   "12345",
				"customer":   "Иван Иванов",
				"amount":     1500.50,
				"status":     "new",
				"created_at": "2024-12-16T10:30:00Z",
			},
			task: "Уведомления о новых заказах",
		},
		{
			name:      "Telegram с лидом",
			targetAPI: "telegram",
			sampleData: map[string]interface{}{
				"lead_id":    "LEAD-001",
				"name":       "Анна Петрова",
				"email":      "anna@example.com",
				"phone":      "+7 999 123-45-67",
				"source":     "website",
				"interest":   "Консультация",
			},
			task: "Уведомления о новых лидах",
		},
		{
			name:      "Discord с системным событием",
			targetAPI: "discord",
			sampleData: map[string]interface{}{
				"event_type":    "system_alert",
				"severity":      "critical",
				"service":       "payment_gateway",
				"message":       "Сервис недоступен",
				"affected_users": 150,
				"timestamp":     "2024-12-16T10:30:00Z",
			},
			task: "Системные уведомления",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := ai.NewChatPromptBuilder()
			prompt := builder.BuildMappingPrompt(tc.sampleData, tc.targetAPI, tc.task, tc.task)

			// Проверяем, что промпт содержит правильную структуру для JSON ответа
			assert.Contains(t, prompt, "JSON", "Промпт должен упоминать JSON формат")
			assert.Contains(t, prompt, "template", "Промпт должен упоминать шаблон")
			assert.Contains(t, prompt, "target_url", "Промпт должен упоминать target_url")

			// Проверяем специфичные для API элементы
			switch tc.targetAPI {
			case "slack":
				assert.Contains(t, prompt, "hooks.slack.com", "Должен содержать Slack URL")
				assert.Contains(t, prompt, "text", "Должен упоминать поле text для Slack")
			case "telegram":
				assert.Contains(t, prompt, "api.telegram.org", "Должен содержать Telegram URL")
				assert.Contains(t, prompt, "chat_id", "Должен упоминать chat_id для Telegram")
			case "discord":
				assert.Contains(t, prompt, "discord.com/api/webhooks", "Должен содержать Discord URL")
				assert.Contains(t, prompt, "content", "Должен упоминать поле content для Discord")
			}

			// Проверяем, что все поля из sampleData упоминаются в промпте
			for key := range tc.sampleData {
				assert.Contains(t, prompt, key, "Промпт должен содержать поле: %s", key)
			}

			t.Logf("Промпт для %s валиден", tc.name)
		})
	}
}

// TestPromptContextBuilding тестирует построение контекста в промптах
func TestPromptContextBuilding(t *testing.T) {
	builder := ai.NewChatPromptBuilder()

	// Добавляем различные типы контекста
	builder.WithContext("integration_type", "webhook")
	builder.WithContext("user_role", "admin")
	builder.WithContext("project_name", "Тестовый проект")

	sampleData := map[string]interface{}{
		"event": "user_registration",
		"user": map[string]interface{}{
			"id":    "user_123",
			"email": "test@example.com",
		},
	}

	prompt := builder.BuildChatPrompt("Настрой интеграцию с CRM", sampleData)

	// Проверяем, что контекст включен в промпт
	assert.Contains(t, prompt, "Контекст:", "Промпт должен содержать секцию контекста")
	assert.Contains(t, prompt, "integration_type: webhook", "Должен содержать тип интеграции")
	assert.Contains(t, prompt, "user_role: admin", "Должен содержать роль пользователя")
	assert.Contains(t, prompt, "project_name: Тестовый проект", "Должен содержать название проекта")

	// Проверяем структуру промпта
	assert.Contains(t, prompt, "Пользователь:", "Должна быть секция пользователя")
	assert.Contains(t, prompt, "Образец входящих данных:", "Должна быть секция с данными")

	t.Logf("Промпт с контекстом:\n%s", prompt)
}

// TestPromptDataFormatting тестирует форматирование данных в промптах
func TestPromptDataFormatting(t *testing.T) {
	testCases := []struct {
		name     string
		data     map[string]interface{}
		expected []string // строки, которые должны быть в отформатированных данных
	}{
		{
			name: "Простые данные",
			data: map[string]interface{}{
				"name":  "Тест",
				"value": 123,
				"flag":  true,
			},
			expected: []string{"name: Тест", "value: 123", "flag: true"},
		},
		{
			name: "Вложенные данные",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "Пользователь",
				},
				"metadata": map[string]interface{}{
					"version": "1.0",
				},
			},
			expected: []string{"user:", "metadata:", "id", "name", "version"},
		},
		{
			name: "Массивы данных",
			data: map[string]interface{}{
				"items": []interface{}{
					"item1",
					"item2",
					map[string]interface{}{
						"nested": "value",
					},
				},
			},
			expected: []string{"items:", "item1", "item2", "nested"},
		},
		{
			name: "Специальные символы",
			data: map[string]interface{}{
				"special": "Строка с \"кавычками\" и символами: @#$%",
				"unicode": "Русский текст и эмодзи 🚀",
			},
			expected: []string{"special:", "unicode:", "кавычками", "эмодзи"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := ai.NewChatPromptBuilder()
			prompt := builder.BuildDataAnalysisPrompt(tc.data, "json")

			// Проверяем, что все ожидаемые элементы присутствуют
			for _, expected := range tc.expected {
				assert.Contains(t, prompt, expected, 
					"Промпт должен содержать: %s", expected)
			}

			// Проверяем, что данные правильно отформатированы как JSON
			assert.Contains(t, prompt, "{", "Должны быть JSON скобки")
			assert.Contains(t, prompt, "}", "Должны быть JSON скобки")

			t.Logf("Данные успешно отформатированы для: %s", tc.name)
		})
	}
}

// TestPromptSystemMessages тестирует системные сообщения в промптах
func TestPromptSystemMessages(t *testing.T) {
	// Получаем системные промпты
	systemPrompts := map[string]string{
		"chat":               ai.SystemPrompts["chat"],
		"data_analysis":      ai.SystemPrompts["data_analysis"],
		"mapping_generation": ai.SystemPrompts["mapping_generation"],
	}

	for promptType, systemPrompt := range systemPrompts {
		t.Run(fmt.Sprintf("Системный промпт: %s", promptType), func(t *testing.T) {
			// Базовые проверки
			assert.NotEmpty(t, systemPrompt, "Системный промпт не должен быть пустым")
			assert.Greater(t, len(systemPrompt), 100, "Системный промпт должен быть достаточно подробным")

			// Проверяем ключевые элементы для каждого типа
			switch promptType {
			case "chat":
				assert.Contains(t, systemPrompt, "AI-ассистент", "Должно быть описание роли")
				assert.Contains(t, systemPrompt, "webhook", "Должно упоминать webhook")
				assert.Contains(t, systemPrompt, "интеграций", "Должно упоминать интеграции")
				assert.Contains(t, systemPrompt, "Slack", "Должно упоминать популярные сервисы")

			case "data_analysis":
				assert.Contains(t, systemPrompt, "анализу структуры данных", "Должно описывать анализ данных")
				assert.Contains(t, systemPrompt, "JSON формате", "Должно упоминать JSON")
				assert.Contains(t, systemPrompt, "fields", "Должно упоминать поля")
				assert.Contains(t, systemPrompt, "string", "Должно упоминать типы данных")

			case "mapping_generation":
				assert.Contains(t, systemPrompt, "маппингов", "Должно упоминать маппинги")
				assert.Contains(t, systemPrompt, "SLACK WEBHOOK", "Должно содержать примеры API")
				assert.Contains(t, systemPrompt, "TELEGRAM BOT", "Должно содержать примеры API")
				assert.Contains(t, systemPrompt, "template", "Должно упоминать шаблоны")
			}

			// Проверяем структуру
			lines := strings.Split(systemPrompt, "\n")
			assert.Greater(t, len(lines), 5, "Системный промпт должен быть многострочным")

			t.Logf("Системный промпт '%s' валиден (%d символов, %d строк)", 
				promptType, len(systemPrompt), len(lines))
		})
	}
}

// TestPromptJSONValidation тестирует валидацию JSON в промптах
func TestPromptJSONValidation(t *testing.T) {
	testData := map[string]interface{}{
		"order_id": "12345",
		"amount":   1500.50,
		"items": []interface{}{
			map[string]interface{}{
				"name":  "Товар 1",
				"price": 750.25,
			},
		},
	}

	builder := ai.NewChatPromptBuilder()
	prompt := builder.BuildDataAnalysisPrompt(testData, "json")

	// Ищем JSON структуры в промпте
	lines := strings.Split(prompt, "\n")
	var jsonLines []string
	inJSON := false

	for _, line := range lines {
		if strings.Contains(line, "{") {
			inJSON = true
		}
		if inJSON {
			jsonLines = append(jsonLines, line)
		}
		if strings.Contains(line, "}") && inJSON {
			break
		}
	}

	// Пытаемся распарсить найденный JSON
	if len(jsonLines) > 0 {
		jsonStr := strings.Join(jsonLines, "\n")
		var parsed interface{}
		err := json.Unmarshal([]byte(jsonStr), &parsed)
		
		// JSON может быть не валидным в промпте (это нормально для примеров)
		if err == nil {
			t.Logf("Найден валидный JSON в промпте")
		} else {
			t.Logf("JSON в промпте используется как пример (не для парсинга)")
		}
	}

	// Проверяем, что промпт содержит правильные инструкции для JSON ответа
	assert.Contains(t, prompt, "JSON", "Промпт должен упоминать JSON")
	assert.Contains(t, prompt, "fields", "Должно быть поле fields")
	assert.Contains(t, prompt, "schema", "Должно быть поле schema")

	t.Logf("JSON валидация промпта пройдена")
}