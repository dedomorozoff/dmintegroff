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

// TestRealWorldScenarios тестирует реальные сценарии использования AI
func TestRealWorldScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		description string
		sampleData  string
		context     string
		validation  func(t *testing.T, result *ai.CreatedIntegration)
	}{
		{
			name:        "E-commerce: Уведомления о заказах в Slack",
			description: "Отправлять уведомления в Slack канал #orders когда приходит новый заказ с информацией о клиенте и сумме",
			sampleData: `{
				"order": {
					"id": "ORD-2024-12345",
					"number": "12345",
					"status": "pending",
					"created_at": "2024-12-16T10:30:00Z",
					"total": {
						"amount": 2500.00,
						"currency": "RUB"
					},
					"customer": {
						"id": "CUST-001",
						"first_name": "Иван",
						"last_name": "Петров",
						"email": "ivan.petrov@example.com",
						"phone": "+7 999 123-45-67"
					},
					"items": [
						{
							"id": "ITEM-001",
							"name": "Смартфон Samsung Galaxy",
							"quantity": 1,
							"price": 2500.00,
							"sku": "SAMSUNG-GALAXY-001"
						}
					],
					"shipping": {
						"method": "courier",
						"address": "Москва, ул. Тестовая, д. 1"
					}
				},
				"event": "order.created",
				"timestamp": "2024-12-16T10:30:00Z"
			}`,
			context: "Интернет-магазин электроники",
			validation: func(t *testing.T, result *ai.CreatedIntegration) {
				assert.Contains(t, strings.ToLower(result.TargetURL), "slack")
				assert.Contains(t, strings.ToLower(result.Name), "slack")
				
				// Проверяем шаблон
				var template map[string]interface{}
				err := json.Unmarshal([]byte(result.Template), &template)
				require.NoError(t, err)
				
				text, exists := template["text"]
				assert.True(t, exists, "Должно быть поле text")
				assert.Contains(t, strings.ToLower(text.(string)), "заказ", "Текст должен упоминать заказ")
			},
		},
		{
			name:        "CRM: Создание лидов в Salesforce",
			description: "Автоматически создавать лиды в Salesforce CRM когда пользователь заполняет форму обратной связи на сайте",
			sampleData: `{
				"form_submission": {
					"form_id": "contact_form_v2",
					"submitted_at": "2024-12-16T10:30:00Z",
					"page_url": "https://example.com/contact",
					"utm_source": "google_ads",
					"utm_campaign": "lead_generation_q4",
					"utm_medium": "cpc"
				},
				"contact": {
					"first_name": "Анна",
					"last_name": "Сидорова",
					"email": "anna.sidorova@company.ru",
					"phone": "+7 999 987-65-43",
					"company": "ООО Инновационные Технологии",
					"position": "Директор по развитию",
					"website": "https://innovation-tech.ru"
				},
				"inquiry": {
					"subject": "Интеграция с нашей CRM системой",
					"message": "Здравствуйте! Нас интересует возможность интеграции вашего решения с нашей CRM системой. У нас около 500 сотрудников и мы обрабатываем до 1000 лидов в месяц.",
					"budget_range": "500000-1000000",
					"timeline": "Q1 2025",
					"priority": "high"
				}
			}`,
			context: "B2B SaaS компания",
			validation: func(t *testing.T, result *ai.CreatedIntegration) {
				assert.Contains(t, strings.ToLower(result.TargetURL), "salesforce")
				assert.Equal(t, "POST", result.Method)
				
				// Проверяем шаблон
				var template map[string]interface{}
				err := json.Unmarshal([]byte(result.Template), &template)
				require.NoError(t, err)
				
				// Должны быть поля для Salesforce Lead
				records, exists := template["records"]
				assert.True(t, exists, "Должно быть поле records")
				
				recordsArray := records.([]interface{})
				assert.Greater(t, len(recordsArray), 0, "Должна быть хотя бы одна запись")
			},
		},
		{
			name:        "Support: Уведомления в Telegram о критических ошибках",
			description: "Отправлять срочные уведомления в Telegram чат техподдержки при возникновении критических ошибок в системе",
			sampleData: `{
				"alert": {
					"id": "ALERT-2024-001",
					"type": "system_error",
					"severity": "critical",
					"status": "firing",
					"created_at": "2024-12-16T10:30:00Z"
				},
				"service": {
					"name": "payment_processor",
					"version": "v2.1.3",
					"environment": "production",
					"region": "eu-west-1"
				},
				"error": {
					"code": "PAY_GATEWAY_TIMEOUT",
					"message": "Timeout connecting to payment gateway",
					"stack_trace": "PaymentProcessor.processPayment:142\nGatewayClient.connect:89",
					"first_occurred": "2024-12-16T10:28:15Z",
					"occurrence_count": 15
				},
				"impact": {
					"affected_users": 150,
					"failed_transactions": 23,
					"estimated_revenue_loss": 45000.00,
					"estimated_downtime": "5-10 minutes"
				},
				"metrics": {
					"cpu_usage": 85.5,
					"memory_usage": 92.1,
					"response_time": 5500,
					"error_rate": 15.8
				}
			}`,
			context: "Финтех платформа",
			validation: func(t *testing.T, result *ai.CreatedIntegration) {
				assert.Contains(t, strings.ToLower(result.TargetURL), "telegram")
				
				// Проверяем шаблон
				var template map[string]interface{}
				err := json.Unmarshal([]byte(result.Template), &template)
				require.NoError(t, err)
				
				text, exists := template["text"]
				assert.True(t, exists, "Должно быть поле text")
				textStr := text.(string)
				assert.Contains(t, strings.ToLower(textStr), "критич", "Должно упоминать критичность")
				
				chatID, exists := template["chat_id"]
				assert.True(t, exists, "Должно быть поле chat_id")
				assert.NotEmpty(t, chatID, "chat_id не должен быть пустым")
			},
		},
		{
			name:        "Marketing: Синхронизация лидов с HubSpot",
			description: "Синхронизировать новые лиды из лендинга с HubSpot CRM для дальнейшей обработки маркетинговой командой",
			sampleData: `{
				"lead": {
					"id": "LEAD-2024-5678",
					"source": "landing_page",
					"campaign": "black_friday_2024",
					"created_at": "2024-12-16T10:30:00Z",
					"score": 85,
					"status": "new"
				},
				"contact": {
					"email": "maria.kozlova@startup.com",
					"first_name": "Мария",
					"last_name": "Козлова",
					"phone": "+7 999 555-44-33",
					"company": "Стартап Решения",
					"job_title": "Маркетинг директор",
					"linkedin": "https://linkedin.com/in/maria-kozlova"
				},
				"engagement": {
					"pages_visited": 5,
					"time_on_site": 420,
					"downloads": ["whitepaper_crm.pdf", "case_study_retail.pdf"],
					"form_fills": 2,
					"email_opens": 3,
					"email_clicks": 1
				},
				"demographics": {
					"country": "Russia",
					"city": "Saint Petersburg",
					"company_size": "50-200",
					"industry": "Technology",
					"annual_revenue": "10M-50M"
				},
				"utm": {
					"source": "google",
					"medium": "cpc",
					"campaign": "black_friday_crm",
					"term": "crm integration",
					"content": "ad_variant_a"
				}
			}`,
			context: "SaaS маркетинг",
			validation: func(t *testing.T, result *ai.CreatedIntegration) {
				// HubSpot или общий CRM API
				urlLower := strings.ToLower(result.TargetURL)
				assert.True(t, 
					strings.Contains(urlLower, "hubspot") || 
					strings.Contains(urlLower, "crm") ||
					strings.Contains(urlLower, "api"),
					"URL должен указывать на CRM систему")
				
				assert.Equal(t, "POST", result.Method)
				
				// Проверяем, что есть маппинг основных полей
				assert.Contains(t, result.Mapping, "email", "Должно быть маппинг email")
			},
		},
		{
			name:        "Analytics: Отправка событий в Google Analytics",
			description: "Отправлять пользовательские события в Google Analytics 4 для отслеживания конверсий и поведения пользователей",
			sampleData: `{
				"event": {
					"name": "purchase",
					"timestamp": "2024-12-16T10:30:00Z",
					"session_id": "sess_abc123def456",
					"event_id": "evt_789xyz012"
				},
				"user": {
					"id": "user_12345",
					"client_id": "GA1.1.123456789.1234567890",
					"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
					"ip_address": "192.168.1.100",
					"language": "ru-RU",
					"country": "RU"
				},
				"ecommerce": {
					"transaction_id": "TXN-2024-001",
					"value": 2500.00,
					"currency": "RUB",
					"tax": 250.00,
					"shipping": 300.00,
					"items": [
						{
							"item_id": "PROD-001",
							"item_name": "Смартфон",
							"category": "Electronics",
							"quantity": 1,
							"price": 2500.00,
							"brand": "Samsung"
						}
					]
				},
				"page": {
					"url": "https://shop.example.com/checkout/success",
					"title": "Заказ оформлен",
					"referrer": "https://shop.example.com/cart"
				}
			}`,
			context: "E-commerce аналитика",
			validation: func(t *testing.T, result *ai.CreatedIntegration) {
				urlLower := strings.ToLower(result.TargetURL)
				assert.True(t, 
					strings.Contains(urlLower, "google") || 
					strings.Contains(urlLower, "analytics") ||
					strings.Contains(urlLower, "measurement"),
					"URL должен указывать на Google Analytics")
				
				// Проверяем шаблон
				var template map[string]interface{}
				err := json.Unmarshal([]byte(result.Template), &template)
				require.NoError(t, err)
				
				// Должны быть поля для GA4
				assert.NotEmpty(t, template, "Шаблон не должен быть пустым")
			},
		},
	}

	generator := createTestGenerator()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			req := &ai.CreateIntegrationRequest{
				Description: scenario.description,
				SampleData:  scenario.sampleData,
				ProjectID:   1,
			}

			result, err := generator.CreateIntegration(ctx, req)
			require.NoError(t, err, "Создание интеграции должно быть успешным")
			require.NotNil(t, result, "Результат не должен быть nil")

			// Базовые проверки
			assert.NotEmpty(t, result.Name, "Название должно быть заполнено")
			assert.NotEmpty(t, result.TargetURL, "URL должен быть заполнен")
			assert.NotEmpty(t, result.Template, "Шаблон должен быть заполнен")
			assert.NotEmpty(t, result.Explanation, "Объяснение должно быть заполнено")
			assert.NotEmpty(t, result.NextSteps, "Следующие шаги должны быть заполнены")
			assert.GreaterOrEqual(t, len(result.NextSteps), 3, "Должно быть минимум 3 следующих шага")

			// Проверяем валидность JSON шаблона
			if result.TemplateType == "json" || result.TemplateType == "" {
				var templateData interface{}
				err := json.Unmarshal([]byte(result.Template), &templateData)
				assert.NoError(t, err, "Шаблон должен быть валидным JSON")
			}

			// Специфичные проверки для сценария
			if scenario.validation != nil {
				scenario.validation(t, result)
			}

			// Выводим подробную информацию
			t.Logf("\n=== Результат для сценария: %s ===", scenario.name)
			t.Logf("Контекст: %s", scenario.context)
			t.Logf("Название: %s", result.Name)
			t.Logf("URL: %s", result.TargetURL)
			t.Logf("Метод: %s", result.Method)
			t.Logf("Тип авторизации: %s", result.AuthType)
			t.Logf("Объяснение: %s", result.Explanation)
			
			// Выводим шаблон в читаемом виде
			if result.TemplateType == "json" || result.TemplateType == "" {
				var prettyJSON interface{}
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
			t.Logf("=== Конец результата ===\n")
		})
	}
}

// TestComplexDataStructures тестирует обработку сложных структур данных
func TestComplexDataStructures(t *testing.T) {
	complexScenarios := []struct {
		name        string
		description string
		sampleData  string
	}{
		{
			name:        "Многоуровневая вложенность",
			description: "Обработка данных с глубокой вложенностью",
			sampleData: `{
				"level1": {
					"level2": {
						"level3": {
							"level4": {
								"data": "deep_value",
								"array": [
									{
										"nested_object": {
											"id": 1,
											"properties": {
												"name": "test",
												"values": [1, 2, 3]
											}
										}
									}
								]
							}
						}
					}
				}
			}`,
		},
		{
			name:        "Большие массивы данных",
			description: "Обработка больших массивов с различными типами данных",
			sampleData: generateComplexArrayJSON(),
		},
		{
			name:        "Смешанные типы данных",
			description: "Обработка данных со всеми типами JSON",
			sampleData: `{
				"string_field": "текстовое значение",
				"number_field": 123.456,
				"integer_field": 789,
				"boolean_true": true,
				"boolean_false": false,
				"null_field": null,
				"array_strings": ["один", "два", "три"],
				"array_numbers": [1, 2, 3, 4.5],
				"array_mixed": ["text", 123, true, null, {"nested": "object"}],
				"object_field": {
					"nested_string": "вложенная строка",
					"nested_number": 456.789,
					"nested_array": [{"id": 1}, {"id": 2}]
				},
				"empty_object": {},
				"empty_array": []
			}`,
		},
	}

	generator := createTestGenerator()

	for _, scenario := range complexScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			req := &ai.CreateIntegrationRequest{
				Description: scenario.description,
				SampleData:  scenario.sampleData,
				ProjectID:   1,
			}

			result, err := generator.CreateIntegration(ctx, req)
			require.NoError(t, err, "Должно успешно обработать сложные данные")
			require.NotNil(t, result, "Результат не должен быть nil")

			// Проверяем, что сложные данные не сломали генерацию
			assert.NotEmpty(t, result.Name, "Название должно быть сгенерировано")
			assert.NotEmpty(t, result.Template, "Шаблон должен быть сгенерирован")
			
			// Проверяем валидность JSON шаблона
			if result.TemplateType == "json" || result.TemplateType == "" {
				var templateData interface{}
				err := json.Unmarshal([]byte(result.Template), &templateData)
				assert.NoError(t, err, "Шаблон должен быть валидным JSON даже для сложных данных")
			}

			t.Logf("Успешно обработаны сложные данные для: %s", scenario.name)
			t.Logf("Размер исходных данных: %d символов", len(scenario.sampleData))
			t.Logf("Размер сгенерированного шаблона: %d символов", len(result.Template))
		})
	}
}

// Вспомогательные функции

func generateComplexArrayJSON() string {
	data := map[string]interface{}{
		"users": make([]interface{}, 0),
		"orders": make([]interface{}, 0),
		"products": make([]interface{}, 0),
	}

	// Генерируем массив пользователей
	for i := 0; i < 10; i++ {
		user := map[string]interface{}{
			"id":    fmt.Sprintf("user_%d", i),
			"name":  fmt.Sprintf("Пользователь %d", i),
			"email": fmt.Sprintf("user%d@example.com", i),
			"profile": map[string]interface{}{
				"age":     20 + i,
				"city":    []string{"Москва", "СПб", "Казань"}[i%3],
				"premium": i%2 == 0,
			},
		}
		data["users"] = append(data["users"].([]interface{}), user)
	}

	// Генерируем массив заказов
	for i := 0; i < 5; i++ {
		order := map[string]interface{}{
			"id":     fmt.Sprintf("order_%d", i),
			"amount": float64(1000 + i*500),
			"items": []interface{}{
				map[string]interface{}{
					"name":     fmt.Sprintf("Товар %d", i),
					"quantity": i + 1,
				},
			},
		}
		data["orders"] = append(data["orders"].([]interface{}), order)
	}

	bytes, _ := json.Marshal(data)
	return string(bytes)
}