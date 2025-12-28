package tests

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"dmintegroff/internal/ai"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptValidation тестирует валидацию и качество промптов
func TestPromptValidation(t *testing.T) {
	testCases := []struct {
		name           string
		description    string
		sampleData     string
		expectedFields []string // поля, которые должны быть в результате
		expectedURL    string   // часть URL, которая должна присутствовать
		minSteps       int      // минимальное количество следующих шагов
	}{
		{
			name:        "Slack с полными данными заказа",
			description: "Отправлять уведомления в Slack канал #orders при получении новых заказов с деталями клиента",
			sampleData: `{
				"order_id": "ORD-2024-001",
				"customer": {
					"name": "Иван Петров",
					"email": "ivan@example.com",
					"phone": "+7 999 123-45-67"
				},
				"items": [
					{
						"name": "Товар 1",
						"quantity": 2,
						"price": 1500.00
					}
				],
				"total_amount": 3000.00,
				"status": "pending",
				"created_at": "2024-12-16T10:30:00Z"
			}`,
			expectedFields: []string{"text"},
			expectedURL:    "slack",
			minSteps:       4,
		},
		{
			name:        "Telegram с данными лида",
			description: "Уведомлять в Telegram о новых лидах с контактной информацией",
			sampleData: `{
				"lead_id": "LEAD-001",
				"source": "landing_page",
				"contact": {
					"first_name": "Анна",
					"last_name": "Сидорова",
					"email": "anna@company.com",
					"phone": "+7 999 987-65-43",
					"company": "ООО Инновации"
				},
				"interest": "Консультация по продукту",
				"budget": "100000-500000",
				"timeline": "1-3 месяца"
			}`,
			expectedFields: []string{"chat_id", "text"},
			expectedURL:    "telegram",
			minSteps:       4,
		},
		{
			name:        "Discord с событиями системы",
			description: "Отправлять в Discord уведомления о критических событиях системы",
			sampleData: `{
				"event_type": "system_error",
				"severity": "critical",
				"service": "payment_processor",
				"message": "Ошибка подключения к платежному шлюзу",
				"error_code": "PAY_001",
				"timestamp": "2024-12-16T10:30:00Z",
				"affected_users": 150,
				"estimated_downtime": "5-10 минут"
			}`,
			expectedFields: []string{"content"},
			expectedURL:    "discord",
			minSteps:       4,
		},
		{
			name:        "Salesforce CRM с лидом",
			description: "Создавать лиды в Salesforce при заполнении формы на сайте",
			sampleData: `{
				"form_id": "contact_form_v2",
				"submitted_at": "2024-12-16T10:30:00Z",
				"contact_info": {
					"first_name": "Михаил",
					"last_name": "Козлов",
					"email": "mikhail@startup.ru",
					"phone": "+7 999 555-44-33",
					"company": "Стартап Технологии",
					"position": "CTO"
				},
				"inquiry": {
					"subject": "Интеграция с нашей системой",
					"message": "Интересует возможность интеграции вашего решения с нашей CRM",
					"budget": "от 500000 руб",
					"timeline": "Q1 2025"
				},
				"utm_source": "google_ads",
				"utm_campaign": "crm_integration"
			}`,
			expectedFields: []string{"FirstName", "LastName", "Email"},
			expectedURL:    "salesforce",
			minSteps:       4,
		},
		{
			name:        "Email уведомления",
			description: "Отправлять email уведомления клиентам об изменении статуса заказа",
			sampleData: `{
				"order_id": "ORD-2024-002",
				"status_change": {
					"from": "processing",
					"to": "shipped",
					"timestamp": "2024-12-16T10:30:00Z"
				},
				"customer": {
					"email": "customer@example.com",
					"name": "Елена Васильева"
				},
				"shipping": {
					"tracking_number": "TRK123456789",
					"carrier": "СДЭК",
					"estimated_delivery": "2024-12-18"
				},
				"order_details": {
					"total": 2500.00,
					"items_count": 3
				}
			}`,
			expectedFields: []string{"to", "subject"},
			expectedURL:    "sendgrid",
			minSteps:       4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем генератор (без реального AI клиента для тестирования)
			generator := createTestGenerator()
			
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Создаем запрос
			req := &ai.CreateIntegrationRequest{
				Description: tc.description,
				SampleData:  tc.sampleData,
				ProjectID:   1,
			}

			// Выполняем создание интеграции
			result, err := generator.CreateIntegration(ctx, req)
			require.NoError(t, err, "Создание интеграции должно быть успешным")
			require.NotNil(t, result, "Результат не должен быть nil")

			// Проверяем базовые поля
			assert.NotEmpty(t, result.Name, "Название должно быть заполнено")
			assert.NotEmpty(t, result.TargetURL, "URL должен быть заполнен")
			assert.NotEmpty(t, result.Template, "Шаблон должен быть заполнен")
			assert.NotEmpty(t, result.Explanation, "Объяснение должно быть заполнено")

			// Проверяем URL
			if tc.expectedURL != "" {
				assert.Contains(t, strings.ToLower(result.TargetURL), tc.expectedURL,
					"URL должен содержать ожидаемый сервис")
			}

			// Проверяем количество следующих шагов
			assert.GreaterOrEqual(t, len(result.NextSteps), tc.minSteps,
				"Должно быть минимум %d следующих шагов", tc.minSteps)

			// Проверяем валидность JSON шаблона
			if result.TemplateType == "json" || result.TemplateType == "" {
				var templateData map[string]interface{}
				err := json.Unmarshal([]byte(result.Template), &templateData)
				require.NoError(t, err, "Шаблон должен быть валидным JSON")

				// Проверяем наличие ожидаемых полей в шаблоне
				for _, expectedField := range tc.expectedFields {
					_, exists := templateData[expectedField]
					assert.True(t, exists, "Шаблон должен содержать поле '%s'", expectedField)
				}
			}

			// Проверяем маппинг полей
			assert.NotEmpty(t, result.Mapping, "Маппинг полей не должен быть пустым")

			// Выводим детальную информацию для анализа
			t.Logf("\n=== Результат для '%s' ===", tc.name)
			t.Logf("Название: %s", result.Name)
			t.Logf("URL: %s", result.TargetURL)
			t.Logf("Метод: %s", result.Method)
			t.Logf("Тип авторизации: %s", result.AuthType)
			t.Logf("Объяснение: %s", result.Explanation)
			t.Logf("Количество полей в маппинге: %d", len(result.Mapping))
			t.Logf("Количество следующих шагов: %d", len(result.NextSteps))
			
			// Выводим шаблон в читаемом виде
			if result.TemplateType == "json" || result.TemplateType == "" {
				var prettyJSON map[string]interface{}
				if json.Unmarshal([]byte(result.Template), &prettyJSON) == nil {
					prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")
					t.Logf("Шаблон:\n%s", string(prettyBytes))
				}
			} else {
				t.Logf("Шаблон (%s):\n%s", result.TemplateType, result.Template)
			}
			
			t.Logf("Следующие шаги:")
			for i, step := range result.NextSteps {
				t.Logf("  %d. %s", i+1, step)
			}
		})
	}
}

// TestPromptEdgeCases тестирует граничные случаи и обработку ошибок
func TestPromptEdgeCases(t *testing.T) {
	generator := createTestGenerator()
	ctx := context.Background()

	testCases := []struct {
		name        string
		description string
		sampleData  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Очень длинное описание",
			description: strings.Repeat("Очень длинное описание интеграции ", 100),
			sampleData:  `{"test": "data"}`,
			expectError: false,
		},
		{
			name:        "Пустой JSON",
			description: "Тест с пустым JSON",
			sampleData:  `{}`,
			expectError: false,
		},
		{
			name:        "Сложная вложенная структура",
			description: "Интеграция со сложными данными",
			sampleData: `{
				"level1": {
					"level2": {
						"level3": {
							"data": "deep_value",
							"array": [1, 2, 3, {"nested": true}]
						}
					}
				},
				"metadata": {
					"version": "1.0",
					"tags": ["tag1", "tag2"]
				}
			}`,
			expectError: false,
		},
		{
			name:        "Специальные символы в данных",
			description: "Тест с специальными символами: @#$%^&*()",
			sampleData: `{
				"special_chars": "!@#$%^&*()",
				"unicode": "Тест с русскими символами и эмодзи 🚀",
				"quotes": "Строка с \"кавычками\" и 'апострофами'"
			}`,
			expectError: false,
		},
		{
			name:        "Большой объем данных",
			description: "Интеграция с большим объемом данных",
			sampleData:  generateLargeJSON(1000), // 1000 полей
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ai.CreateIntegrationRequest{
				Description: tc.description,
				SampleData:  tc.sampleData,
				ProjectID:   1,
			}

			result, err := generator.CreateIntegration(ctx, req)

			if tc.expectError {
				assert.Error(t, err, "Ожидалась ошибка")
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Сообщение об ошибке должно содержать ожидаемый текст")
				}
			} else {
				assert.NoError(t, err, "Не должно быть ошибки")
				assert.NotNil(t, result, "Результат не должен быть nil")
				
				// Базовые проверки для успешных случаев
				assert.NotEmpty(t, result.Name, "Название должно быть заполнено")
				assert.NotEmpty(t, result.TargetURL, "URL должен быть заполнен")
				
				t.Logf("Успешно обработан случай: %s", tc.name)
				t.Logf("  Название: %s", result.Name)
				t.Logf("  Размер шаблона: %d символов", len(result.Template))
			}
		})
	}
}

// TestPromptConsistency тестирует консистентность результатов
func TestPromptConsistency(t *testing.T) {
	generator := createTestGenerator()
	ctx := context.Background()

	// Один и тот же запрос должен давать похожие результаты
	req := &ai.CreateIntegrationRequest{
		Description: "Отправлять уведомления в Slack о новых заказах",
		SampleData: `{
			"order_id": "12345",
			"customer_name": "Тест",
			"amount": 1000
		}`,
		ProjectID: 1,
	}

	var results []*ai.CreatedIntegration
	
	// Выполняем несколько раз
	for i := 0; i < 3; i++ {
		result, err := generator.CreateIntegration(ctx, req)
		require.NoError(t, err, "Итерация %d должна быть успешной", i+1)
		results = append(results, result)
	}

	// Проверяем консистентность
	firstResult := results[0]
	for i, result := range results[1:] {
		assert.Equal(t, firstResult.Method, result.Method, 
			"HTTP метод должен быть одинаковым (итерация %d)", i+2)
		
		assert.Contains(t, strings.ToLower(result.TargetURL), "slack",
			"URL должен содержать 'slack' (итерация %d)", i+2)
		
		// Шаблоны должны быть валидными JSON
		var template1, template2 map[string]interface{}
		err1 := json.Unmarshal([]byte(firstResult.Template), &template1)
		err2 := json.Unmarshal([]byte(result.Template), &template2)
		
		assert.NoError(t, err1, "Первый шаблон должен быть валидным JSON")
		assert.NoError(t, err2, "Шаблон итерации %d должен быть валидным JSON", i+2)
		
		// Должны содержать поле 'text' для Slack
		_, hasText1 := template1["text"]
		_, hasText2 := template2["text"]
		assert.True(t, hasText1, "Первый шаблон должен содержать поле 'text'")
		assert.True(t, hasText2, "Шаблон итерации %d должен содержать поле 'text'", i+2)
	}

	t.Logf("Проверена консистентность %d результатов", len(results))
}

// Вспомогательные функции

func createTestGenerator() *ai.Generator {
	// Создаем тестовый AI клиент без реального подключения
	config := &ai.AIConfig{
		Enabled: false, // Отключаем реальный AI для тестов
	}
	client := ai.NewClient(config)
	return ai.NewGenerator(client)
}

func generateLargeJSON(fieldCount int) string {
	data := make(map[string]interface{})
	
	for i := 0; i < fieldCount; i++ {
		data[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	
	bytes, _ := json.Marshal(data)
	return string(bytes)
}