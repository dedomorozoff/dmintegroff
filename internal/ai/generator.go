package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"dmintegroff/internal/logger"
)

// Generator генератор маппингов для интеграций
type Generator struct {
	client   *Client
	analyzer *Analyzer
}

// NewGenerator создает новый генератор
func NewGenerator(client *Client) *Generator {
	return &Generator{
		client:   client,
		analyzer: NewAnalyzer(client),
	}
}

// GenerateMapping генерирует маппинг для интеграции
func (g *Generator) GenerateMapping(ctx context.Context, req *MappingGenerationRequest) (*GeneratedMapping, error) {
	// Сначала анализируем исходные данные
	analysis, err := g.analyzer.AnalyzeDataStructure(ctx, req.SourceData, "json")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze source data: %w", err)
	}
	
	// Если AI настроен, используем его для генерации
	if g.client.IsConfigured() {
		mapping, err := g.client.GenerateMapping(ctx, req)
		if err == nil {
			// Дополняем маппинг данными анализа
			return g.enhanceMapping(mapping, analysis), nil
		}
		// Если AI недоступен, используем локальную генерацию
		fmt.Printf("AI mapping generation failed, using local generation: %v\n", err)
	}
	
	// Локальная генерация маппинга
	return g.generateMappingLocally(req, analysis)
}

// generateMappingLocally генерирует маппинг локально без AI
func (g *Generator) generateMappingLocally(req *MappingGenerationRequest, analysis *DataAnalysisResponse) (*GeneratedMapping, error) {
	// Определяем целевой API
	apiInfo := g.detectTargetAPI(req.TargetAPI, req.Task)
	
	// Создаем базовый маппинг
	mapping := &GeneratedMapping{
		Type:        "json_template",
		Method:      "POST",
		Headers:     make(map[string]string),
		AuthType:    apiInfo.AuthType,
		AuthConfig:  make(map[string]interface{}),
		Description: fmt.Sprintf("Интеграция с %s", apiInfo.Name),
		Reasoning:   "Автоматически сгенерированный маппинг на основе анализа данных",
	}
	
	// Настраиваем заголовки
	mapping.Headers["Content-Type"] = "application/json"
	
	// Генерируем шаблон на основе типа API
	template, targetURL := g.generateTemplateForAPI(apiInfo, req.SourceData, analysis)
	mapping.Template = template
	mapping.TargetURL = targetURL
	
	// Настраиваем аутентификацию
	g.configureAuth(mapping, apiInfo)
	
	return mapping, nil
}

// detectTargetAPI определяет целевой API
func (g *Generator) detectTargetAPI(targetAPI, task string) PopularAPI {
	targetAPI = strings.ToLower(targetAPI)
	task = strings.ToLower(task)
	
	// Проверяем популярные API
	popularAPIs := GetPopularAPIs()
	
	for _, api := range popularAPIs {
		apiName := strings.ToLower(api.Name)
		if strings.Contains(targetAPI, apiName) || strings.Contains(task, apiName) {
			return api
		}
	}
	
	// Если не найден, создаем общий API
	return PopularAPI{
		Name:        "Custom API",
		BaseURL:     "",
		AuthType:    "bearer",
		Description: "Пользовательский API",
		CommonFields: make(map[string]string),
	}
}

// generateTemplateForAPI генерирует шаблон для конкретного API
func (g *Generator) generateTemplateForAPI(api PopularAPI, sourceData map[string]interface{}, analysis *DataAnalysisResponse) (string, string) {
	switch strings.ToLower(api.Name) {
	case "slack":
		return g.generateSlackTemplate(sourceData, analysis)
	case "telegram":
		return g.generateTelegramTemplate(sourceData, analysis)
	case "discord":
		return g.generateDiscordTemplate(sourceData, analysis)
	case "amocrm":
		return g.generateAmoCRMTemplate(sourceData, analysis)
	default:
		return g.generateGenericTemplate(sourceData, analysis)
	}
}

// generateSlackTemplate генерирует шаблон для Slack
func (g *Generator) generateSlackTemplate(sourceData map[string]interface{}, analysis *DataAnalysisResponse) (string, string) {
	// Находим поле для текста сообщения
	textField := g.findBestField(analysis.Fields, []string{"message", "text", "content", "description"})
	if textField == "" {
		textField = g.getFirstStringField(analysis.Fields)
	}
	
	// Создаем шаблон сообщения
	var messageParts []string
	
	// Добавляем эмодзи в зависимости от типа данных
	emoji := g.getEmojiForDataType(analysis.DataType)
	if emoji != "" {
		messageParts = append(messageParts, emoji)
	}
	
	// Добавляем заголовок
	title := g.generateTitle(analysis.DataType)
	messageParts = append(messageParts, title)
	
	// Добавляем основной текст
	if textField != "" {
		messageParts = append(messageParts, fmt.Sprintf("{{%s}}", textField))
	}
	
	// Добавляем дополнительные поля
	additionalFields := g.getImportantFields(analysis.Fields, []string{textField})
	for _, field := range additionalFields {
		messageParts = append(messageParts, fmt.Sprintf("%s: {{%s}}", field.Description, field.Name))
	}
	
	template := map[string]interface{}{
		"text": strings.Join(messageParts, "\n"),
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	return string(templateJSON), "https://hooks.slack.com/services/YOUR_WEBHOOK_PATH"
}

// generateTelegramTemplate генерирует шаблон для Telegram
func (g *Generator) generateTelegramTemplate(sourceData map[string]interface{}, analysis *DataAnalysisResponse) (string, string) {
	// Находим поле для текста сообщения
	textField := g.findBestField(analysis.Fields, []string{"message", "text", "content", "description"})
	if textField == "" {
		textField = g.getFirstStringField(analysis.Fields)
	}
	
	// Создаем шаблон сообщения
	var messageParts []string
	
	// Добавляем заголовок с эмодзи
	emoji := g.getEmojiForDataType(analysis.DataType)
	title := g.generateTitle(analysis.DataType)
	messageParts = append(messageParts, fmt.Sprintf("%s <b>%s</b>", emoji, title))
	
	// Добавляем основной текст
	if textField != "" {
		messageParts = append(messageParts, fmt.Sprintf("{{%s}}", textField))
	}
	
	// Добавляем дополнительные поля
	additionalFields := g.getImportantFields(analysis.Fields, []string{textField})
	for _, field := range additionalFields {
		messageParts = append(messageParts, fmt.Sprintf("• <i>%s:</i> {{%s}}", field.Description, field.Name))
	}
	
	template := map[string]interface{}{
		"chat_id":    "YOUR_CHAT_ID",
		"text":       strings.Join(messageParts, "\n\n"),
		"parse_mode": "HTML",
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	return string(templateJSON), "https://api.telegram.org/botYOUR_BOT_TOKEN/sendMessage"
}

// generateDiscordTemplate генерирует шаблон для Discord
func (g *Generator) generateDiscordTemplate(sourceData map[string]interface{}, analysis *DataAnalysisResponse) (string, string) {
	// Находим поле для текста сообщения
	textField := g.findBestField(analysis.Fields, []string{"message", "text", "content", "description"})
	if textField == "" {
		textField = g.getFirstStringField(analysis.Fields)
	}
	
	// Создаем шаблон сообщения
	var messageParts []string
	
	// Добавляем заголовок с эмодзи
	emoji := g.getEmojiForDataType(analysis.DataType)
	title := g.generateTitle(analysis.DataType)
	messageParts = append(messageParts, fmt.Sprintf("%s **%s**", emoji, title))
	
	// Добавляем основной текст
	if textField != "" {
		messageParts = append(messageParts, fmt.Sprintf("{{%s}}", textField))
	}
	
	// Добавляем дополнительные поля как embed
	additionalFields := g.getImportantFields(analysis.Fields, []string{textField})
	if len(additionalFields) > 0 {
		embed := map[string]interface{}{
			"title": title,
			"color": g.getColorForDataType(analysis.DataType),
			"fields": []map[string]interface{}{},
		}
		
		for _, field := range additionalFields {
			embedField := map[string]interface{}{
				"name":   field.Description,
				"value":  fmt.Sprintf("{{%s}}", field.Name),
				"inline": true,
			}
			embed["fields"] = append(embed["fields"].([]map[string]interface{}), embedField)
		}
		
		template := map[string]interface{}{
			"content": strings.Join(messageParts, "\n"),
			"embeds":  []map[string]interface{}{embed},
		}
		
		templateJSON, _ := json.MarshalIndent(template, "", "  ")
		return string(templateJSON), "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
	}
	
	template := map[string]interface{}{
		"content": strings.Join(messageParts, "\n"),
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	return string(templateJSON), "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
}

// generateAmoCRMTemplate генерирует шаблон для AmoCRM
func (g *Generator) generateAmoCRMTemplate(sourceData map[string]interface{}, analysis *DataAnalysisResponse) (string, string) {
	// Определяем тип операции - создание контакта или лида
	isLead := false
	for _, field := range analysis.Fields {
		name := strings.ToLower(field.Name)
		if strings.Contains(name, "amount") || strings.Contains(name, "price") || strings.Contains(name, "sum") {
			isLead = true
			break
		}
	}
	
	// Находим основные поля
	nameField := g.findBestField(analysis.Fields, []string{"name", "client", "customer", "first_name", "full_name"})
	phoneField := g.findBestField(analysis.Fields, []string{"phone", "tel", "telephone", "mobile"})
	emailField := g.findBestField(analysis.Fields, []string{"email", "mail", "e_mail"})
	
	if nameField == "" {
		nameField = g.getFirstStringField(analysis.Fields)
	}
	
	if isLead {
		// Генерируем шаблон для создания лида
		amountField := g.findBestField(analysis.Fields, []string{"amount", "price", "sum", "total", "cost"})
		sourceField := g.findBestField(analysis.Fields, []string{"source", "utm_source", "channel", "origin"})
		
		template := []map[string]interface{}{
			{
				"name": fmt.Sprintf("Лид из dmIntegroff - {{%s}}", nameField),
				"price": func() interface{} {
					if amountField != "" {
						return fmt.Sprintf("{{%s}}", amountField)
					}
					return 0
				}(),
				"custom_fields_values": []map[string]interface{}{
					{
						"field_id": 123458, // ID поля "Описание" (нужно заменить на реальный)
						"values": []map[string]interface{}{
							{
								"value": func() string {
									if sourceField != "" {
										return fmt.Sprintf("Источник: {{%s}}", sourceField)
									}
									return "Лид создан через dmIntegroff"
								}(),
							},
						},
					},
				},
				"_embedded": map[string]interface{}{
					"contacts": []map[string]interface{}{
						{
							"name": fmt.Sprintf("{{%s}}", nameField),
							"custom_fields_values": func() []map[string]interface{} {
								var fields []map[string]interface{}
								
								if phoneField != "" {
									fields = append(fields, map[string]interface{}{
										"field_id": 123456, // ID поля "Телефон" (нужно заменить на реальный)
										"values": []map[string]interface{}{
											{
												"value":     fmt.Sprintf("{{%s}}", phoneField),
												"enum_code": "WORK",
											},
										},
									})
								}
								
								if emailField != "" {
									fields = append(fields, map[string]interface{}{
										"field_id": 123457, // ID поля "Email" (нужно заменить на реальный)
										"values": []map[string]interface{}{
											{
												"value":     fmt.Sprintf("{{%s}}", emailField),
												"enum_code": "WORK",
											},
										},
									})
								}
								
								return fields
							}(),
						},
					},
				},
			},
		}
		
		templateJSON, _ := json.MarshalIndent(template, "", "  ")
		return string(templateJSON), "https://SUBDOMAIN.amocrm.ru/api/v4/leads"
	} else {
		// Генерируем шаблон для создания контакта
		template := []map[string]interface{}{
			{
				"name": fmt.Sprintf("{{%s}}", nameField),
				"custom_fields_values": func() []map[string]interface{} {
					var fields []map[string]interface{}
					
					if phoneField != "" {
						fields = append(fields, map[string]interface{}{
							"field_id": 123456, // ID поля "Телефон" (нужно заменить на реальный)
							"values": []map[string]interface{}{
								{
									"value":     fmt.Sprintf("{{%s}}", phoneField),
									"enum_code": "WORK",
								},
							},
						})
					}
					
					if emailField != "" {
						fields = append(fields, map[string]interface{}{
							"field_id": 123457, // ID поля "Email" (нужно заменить на реальный)
							"values": []map[string]interface{}{
								{
									"value":     fmt.Sprintf("{{%s}}", emailField),
									"enum_code": "WORK",
								},
							},
						})
					}
					
					return fields
				}(),
			},
		}
		
		templateJSON, _ := json.MarshalIndent(template, "", "  ")
		return string(templateJSON), "https://SUBDOMAIN.amocrm.ru/api/v4/contacts"
	}
}

// generateGenericTemplate генерирует общий шаблон
func (g *Generator) generateGenericTemplate(sourceData map[string]interface{}, analysis *DataAnalysisResponse) (string, string) {
	template := make(map[string]interface{})
	
	// Добавляем все поля как есть
	for _, field := range analysis.Fields {
		template[field.Name] = fmt.Sprintf("{{%s}}", field.Name)
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	return string(templateJSON), "https://api.example.com/webhook"
}

// findBestField находит лучшее поле из списка кандидатов
func (g *Generator) findBestField(fields []FieldInfo, candidates []string) string {
	for _, candidate := range candidates {
		for _, field := range fields {
			if strings.ToLower(field.Name) == candidate {
				return field.Name
			}
		}
	}
	
	// Поиск по частичному совпадению
	for _, candidate := range candidates {
		for _, field := range fields {
			if strings.Contains(strings.ToLower(field.Name), candidate) {
				return field.Name
			}
		}
	}
	
	return ""
}

// getFirstStringField возвращает первое строковое поле
func (g *Generator) getFirstStringField(fields []FieldInfo) string {
	for _, field := range fields {
		if field.Type == "string" {
			return field.Name
		}
	}
	return ""
}

// getImportantFields возвращает важные поля (исключая уже использованные)
func (g *Generator) getImportantFields(fields []FieldInfo, exclude []string) []FieldInfo {
	important := make([]FieldInfo, 0)
	
	excludeMap := make(map[string]bool)
	for _, ex := range exclude {
		excludeMap[ex] = true
	}
	
	for _, field := range fields {
		if excludeMap[field.Name] {
			continue
		}
		
		// Важные поля
		name := strings.ToLower(field.Name)
		if strings.Contains(name, "id") || 
		   strings.Contains(name, "name") ||
		   strings.Contains(name, "email") ||
		   strings.Contains(name, "amount") ||
		   strings.Contains(name, "status") ||
		   strings.Contains(name, "type") ||
		   field.Type == "number" ||
		   field.Type == "email" {
			important = append(important, field)
		}
		
		// Ограничиваем количество полей
		if len(important) >= 5 {
			break
		}
	}
	
	return important
}

// getEmojiForDataType возвращает эмодзи для типа данных
func (g *Generator) getEmojiForDataType(dataType string) string {
	emojis := map[string]string{
		"order":        "🛒",
		"user":         "👤",
		"event":        "📅",
		"notification": "🔔",
		"payment":      "💳",
		"lead":         "🎯",
		"error":        "❌",
		"success":      "✅",
		"warning":      "⚠️",
		"info":         "ℹ️",
	}
	
	if emoji, exists := emojis[dataType]; exists {
		return emoji
	}
	return "📢"
}

// generateTitle генерирует заголовок для типа данных
func (g *Generator) generateTitle(dataType string) string {
	titles := map[string]string{
		"order":        "Новый заказ",
		"user":         "Новый пользователь",
		"event":        "Событие",
		"notification": "Уведомление",
		"payment":      "Платеж",
		"lead":         "Новый лид",
		"error":        "Ошибка",
		"success":      "Успех",
		"warning":      "Предупреждение",
		"info":         "Информация",
	}
	
	if title, exists := titles[dataType]; exists {
		return title
	}
	return "Уведомление"
}

// getColorForDataType возвращает цвет для типа данных (для Discord embeds)
func (g *Generator) getColorForDataType(dataType string) int {
	colors := map[string]int{
		"order":        0x00ff00, // зеленый
		"user":         0x0099ff, // синий
		"event":        0xffaa00, // оранжевый
		"notification": 0x9900ff, // фиолетовый
		"payment":      0x00ff99, // мятный
		"lead":         0xff6600, // оранжево-красный
		"error":        0xff0000, // красный
		"success":      0x00ff00, // зеленый
		"warning":      0xffff00, // желтый
		"info":         0x0099ff, // синий
	}
	
	if color, exists := colors[dataType]; exists {
		return color
	}
	return 0x999999 // серый
}

// configureAuth настраивает аутентификацию для маппинга
func (g *Generator) configureAuth(mapping *GeneratedMapping, api PopularAPI) {
	switch api.AuthType {
	case "bearer":
		mapping.AuthConfig["token_field"] = "Authorization"
		mapping.AuthConfig["token_prefix"] = "Bearer "
	case "basic":
		mapping.AuthConfig["username_field"] = "username"
		mapping.AuthConfig["password_field"] = "password"
	case "oauth":
		mapping.AuthConfig["token_field"] = "Authorization"
		mapping.AuthConfig["token_prefix"] = "Bearer "
		mapping.AuthConfig["token_type"] = "oauth"
	case "webhook":
		// Для webhook аутентификация не нужна
		mapping.AuthType = "none"
	default:
		mapping.AuthType = "none"
	}
}

// CreateIntegration создает интеграцию с помощью AI
func (g *Generator) CreateIntegration(ctx context.Context, req *CreateIntegrationRequest) (*CreatedIntegration, error) {
	logger.Log.WithFields(map[string]interface{}{
		"action":      "ai_generator_create_integration_start",
		"description": req.Description,
		"project_id":  req.ProjectID,
		"has_sample":  req.SampleData != "",
		"sample_size": len(req.SampleData),
	}).Info("AI Generator: Starting integration creation")

	// Парсим образец данных если есть
	var sampleData map[string]interface{}
	if req.SampleData != "" {
		if err := json.Unmarshal([]byte(req.SampleData), &sampleData); err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"action": "ai_generator_parse_sample",
				"error":  "json_parse_failed",
				"sample": func() string {
					if len(req.SampleData) > 200 {
						return req.SampleData[:200] + "..."
					}
					return req.SampleData
				}(),
			}).Warn("AI Generator: Failed to parse sample as JSON, using as plain text")
			
			// Если не JSON, создаем простую структуру
			sampleData = map[string]interface{}{
				"data": req.SampleData,
			}
		} else {
			logger.Log.WithFields(map[string]interface{}{
				"action":      "ai_generator_parse_sample",
				"field_count": len(sampleData),
			}).Info("AI Generator: Successfully parsed sample data as JSON")
		}
	} else {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_generator_parse_sample",
		}).Info("AI Generator: No sample data provided, using default structure")
		
		// Создаем пустую структуру для анализа
		sampleData = map[string]interface{}{
			"message": "sample message",
		}
	}

	// Анализируем описание для определения типа интеграции
	description := strings.ToLower(req.Description)
	
	var integrationName string
	var targetSystem string
	
	// Определяем тип интеграции на основе описания
	if strings.Contains(description, "slack") {
		integrationName = "Slack уведомления"
		targetSystem = "slack"
	} else if strings.Contains(description, "telegram") {
		integrationName = "Telegram бот"
		targetSystem = "telegram"
	} else if strings.Contains(description, "discord") {
		integrationName = "Discord webhook"
		targetSystem = "discord"
	} else if strings.Contains(description, "amocrm") || strings.Contains(description, "амо") {
		integrationName = "AmoCRM интеграция"
		targetSystem = "amocrm"
	} else if strings.Contains(description, "salesforce") {
		integrationName = "Salesforce CRM"
		targetSystem = "salesforce"
	} else if strings.Contains(description, "email") || strings.Contains(description, "mail") {
		integrationName = "Email уведомления"
		targetSystem = "email"
	} else {
		integrationName = "Пользовательская интеграция"
		targetSystem = "webhook"
	}

	logger.Log.WithFields(map[string]interface{}{
		"action":           "ai_generator_detect_system",
		"detected_system":  targetSystem,
		"integration_name": integrationName,
		"description":      req.Description,
	}).Info("AI Generator: Detected target system from description")

	// Анализируем данные
	analysis, err := g.analyzer.AnalyzeDataStructure(ctx, sampleData, "json")
	if err != nil {
		// Создаем базовый анализ если AI недоступен
		analysis = &DataAnalysisResponse{
			Fields: []FieldInfo{
				{Name: "message", Type: "string", Description: "Сообщение"},
			},
			DataType:   "notification",
			Confidence: 0.5,
		}
	}

	// Генерируем маппинг
	mappingReq := &MappingGenerationRequest{
		SourceData: sampleData,
		TargetAPI:  targetSystem,
		Task:       req.Description,
		UserPrompt: req.Description,
	}

	mapping, err := g.GenerateMapping(ctx, mappingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mapping: %w", err)
	}

	// Создаем маппинг полей
	fieldMapping := make(map[string]string)
	for _, field := range analysis.Fields {
		fieldMapping[field.Name] = field.Name
	}

	// Генерируем следующие шаги
	nextSteps := g.generateNextStepsForIntegration(targetSystem, mapping.AuthType)

	// Создаем результат
	result := &CreatedIntegration{
		Name:         integrationName,
		TargetURL:    mapping.TargetURL,
		Method:       mapping.Method,
		Template:     mapping.GetTemplateString(),
		TemplateType: mapping.Type,
		Mapping:      fieldMapping,
		AuthType:     mapping.AuthType,
		AuthConfig:   mapping.AuthConfig,
		Headers:      mapping.Headers,
		Explanation:  g.generateExplanation(targetSystem, req.Description),
		NextSteps:    nextSteps,
	}

	logger.Log.WithFields(map[string]interface{}{
		"action":        "ai_generator_create_integration_success",
		"target_system": targetSystem,
		"name":          result.Name,
		"target_url":    result.TargetURL,
		"method":        result.Method,
		"template_type": result.TemplateType,
		"auth_type":     result.AuthType,
		"field_count":   len(result.Mapping),
		"template_size": len(result.Template),
		"template_preview": func() string {
			if len(result.Template) > 150 {
				return result.Template[:150] + "..."
			}
			return result.Template
		}(),
	}).Info("AI Generator: Integration creation completed successfully")

	return result, nil
}

// generateNextStepsForIntegration генерирует следующие шаги для интеграции
func (g *Generator) generateNextStepsForIntegration(targetSystem, authType string) []string {
	steps := make([]string, 0)

	switch targetSystem {
	case "slack":
		steps = append(steps, "1. Создайте Incoming Webhook в настройках Slack")
		steps = append(steps, "2. Замените URL на реальный webhook URL")
		steps = append(steps, "3. Настройте канал по умолчанию")
		steps = append(steps, "4. Протестируйте отправку сообщения")
		steps = append(steps, "5. Активируйте интеграцию")
	case "telegram":
		steps = append(steps, "1. Создайте бота через @BotFather в Telegram")
		steps = append(steps, "2. Получите токен бота и замените в URL")
		steps = append(steps, "3. Получите chat_id и замените в шаблоне")
		steps = append(steps, "4. Протестируйте отправку сообщения")
		steps = append(steps, "5. Активируйте интеграцию")
	case "discord":
		steps = append(steps, "1. Создайте webhook в настройках канала Discord")
		steps = append(steps, "2. Замените URL на реальный webhook URL")
		steps = append(steps, "3. Настройте формат сообщений")
		steps = append(steps, "4. Протестируйте отправку сообщения")
		steps = append(steps, "5. Активируйте интеграцию")
	case "salesforce":
		steps = append(steps, "1. Настройте OAuth 2.0 авторизацию в Salesforce")
		steps = append(steps, "2. Замените URL на ваш Salesforce instance")
		steps = append(steps, "3. Добавьте необходимые заголовки авторизации")
		steps = append(steps, "4. Протестируйте создание лида")
		steps = append(steps, "5. Активируйте интеграцию")
	case "email":
		steps = append(steps, "1. Получите API ключ SendGrid")
		steps = append(steps, "2. Добавьте заголовок Authorization с API ключом")
		steps = append(steps, "3. Настройте адрес отправителя")
		steps = append(steps, "4. Протестируйте отправку email")
		steps = append(steps, "5. Активируйте интеграцию")
	case "amocrm":
		steps = append(steps, "1. Получите Access Token в настройках AmoCRM")
		steps = append(steps, "2. Замените SUBDOMAIN на ваш поддомен AmoCRM")
		steps = append(steps, "3. Замените field_id на реальные ID полей из вашей AmoCRM")
		steps = append(steps, "4. Протестируйте создание контакта/лида")
		steps = append(steps, "5. Активируйте интеграцию")
	default:
		steps = append(steps, "1. Замените URL на реальный API endpoint")
		steps = append(steps, "2. Настройте необходимые заголовки авторизации")
		steps = append(steps, "3. Адаптируйте шаблон под формат API")
		steps = append(steps, "4. Протестируйте отправку данных")
		steps = append(steps, "5. Активируйте интеграцию")
	}

	return steps
}

// generateExplanation генерирует объяснение для интеграции
func (g *Generator) generateExplanation(targetSystem, description string) string {
	switch targetSystem {
	case "slack":
		return "Создана интеграция для отправки уведомлений в Slack через Incoming Webhook. " +
			"Сообщения будут отправляться в указанный канал с автоматическим форматированием."
	case "telegram":
		return "Создана интеграция для отправки сообщений через Telegram Bot API. " +
			"Сообщения будут отправляться в указанный чат с поддержкой HTML форматирования."
	case "discord":
		return "Создана интеграция для отправки сообщений в Discord через webhook. " +
			"Поддерживается отправка обычных сообщений и rich embeds."
	case "salesforce":
		return "Создана интеграция для создания лидов в Salesforce CRM. " +
			"Данные будут автоматически преобразованы в формат Salesforce Lead API."
	case "email":
		return "Создана интеграция для отправки email уведомлений через SendGrid API. " +
			"Поддерживается отправка текстовых и HTML сообщений."
	case "amocrm":
		return "Создана интеграция для работы с AmoCRM API. " +
			"Данные будут автоматически преобразованы в формат AmoCRM для создания контактов или лидов."
	default:
		return fmt.Sprintf("Создана пользовательская интеграция на основе описания: %s. " +
			"Данные будут отправляться в JSON формате на указанный endpoint.", description)
	}
}

// enhanceMapping дополняет AI-сгенерированный маппинг данными анализа
func (g *Generator) enhanceMapping(mapping *GeneratedMapping, analysis *DataAnalysisResponse) *GeneratedMapping {
	// Добавляем информацию о типе данных в описание
	if analysis.DataType != "unknown" {
		mapping.Description += fmt.Sprintf(" (тип данных: %s)", analysis.DataType)
	}
	
	// Добавляем информацию об уверенности
	if analysis.Confidence > 0 {
		mapping.Description += fmt.Sprintf(" [уверенность: %.0f%%]", analysis.Confidence*100)
	}
	
	return mapping
}