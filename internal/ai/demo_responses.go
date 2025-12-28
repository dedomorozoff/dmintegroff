package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DemoResponses содержит предопределенные ответы для демо режима
type DemoResponses struct{}

// NewDemoResponses создает новый экземпляр демо-ответов
func NewDemoResponses() *DemoResponses {
	return &DemoResponses{}
}

// GetDemoChatResponse возвращает демо-ответ для чата
func (d *DemoResponses) GetDemoChatResponse(message string, targetAPI string, sampleData map[string]interface{}) *ChatResponse {
	// Анализируем сообщение для выбора подходящего ответа
	messageLower := strings.ToLower(message)
	
	var response string
	var suggestions []AISuggestion
	var nextSteps []string
	
	// Определяем тип запроса и формируем соответствующий ответ
	if strings.Contains(messageLower, "slack") || targetAPI == "slack" {
		response = d.getSlackDemoResponse()
		suggestions = d.getSlackDemoSuggestions()
		nextSteps = d.getSlackDemoNextSteps()
	} else if strings.Contains(messageLower, "telegram") || targetAPI == "telegram" {
		response = d.getTelegramDemoResponse()
		suggestions = d.getTelegramDemoSuggestions()
		nextSteps = d.getTelegramDemoNextSteps()
	} else if strings.Contains(messageLower, "discord") || targetAPI == "discord" {
		response = d.getDiscordDemoResponse()
		suggestions = d.getDiscordDemoSuggestions()
		nextSteps = d.getDiscordDemoNextSteps()
	} else if strings.Contains(messageLower, "amocrm") || strings.Contains(messageLower, "амо") {
		response = d.getAmoCRMDemoResponse()
		suggestions = d.getAmoCRMDemoSuggestions()
		nextSteps = d.getAmoCRMDemoNextSteps()
	} else if strings.Contains(messageLower, "email") || strings.Contains(messageLower, "почт") {
		response = d.getEmailDemoResponse()
		suggestions = d.getEmailDemoSuggestions()
		nextSteps = d.getEmailDemoNextSteps()
	} else {
		response = d.getGenericDemoResponse(message)
		suggestions = d.getGenericDemoSuggestions()
		nextSteps = d.getGenericDemoNextSteps()
	}
	
	return &ChatResponse{
		Response:    response,
		Suggestions: suggestions,
		NextSteps:   nextSteps,
		Confidence:  0.9, // Высокая уверенность для демо
	}
}

// GetDemoDataAnalysis возвращает демо-анализ данных
func (d *DemoResponses) GetDemoDataAnalysis(data map[string]interface{}, format string) *DataAnalysisResponse {
	// Создаем базовый анализ на основе входных данных
	fields := make([]FieldInfo, 0)
	suggestions := make([]FieldMapping, 0)
	
	// Анализируем поля данных
	for key, value := range data {
		fieldType := d.detectFieldType(key, value)
		description := d.generateFieldDescription(key, fieldType)
		
		field := FieldInfo{
			Name:        key,
			Type:        fieldType,
			Required:    d.isRequiredField(key),
			Examples:    d.generateExamples(value),
			Description: description,
			Pattern:     d.generatePattern(fieldType),
		}
		fields = append(fields, field)
		
		// Создаем предложения маппинга
		if targetField := d.suggestTargetField(key, fieldType); targetField != "" {
			suggestion := FieldMapping{
				SourceField: key,
				TargetField: targetField,
				Transform:   d.suggestTransform(fieldType),
				Confidence:  0.8,
				Reasoning:   fmt.Sprintf("Поле '%s' подходит для маппинга в '%s'", key, targetField),
			}
			suggestions = append(suggestions, suggestion)
		}
	}
	
	// Определяем тип данных
	dataType := d.detectDataType(data)
	
	return &DataAnalysisResponse{
		Fields:      fields,
		Schema:      d.generateSchema(fields, dataType),
		Suggestions: suggestions,
		DataType:    dataType,
		Confidence:  0.85,
	}
}

// GetDemoGeneratedMapping возвращает демо-маппинг
func (d *DemoResponses) GetDemoGeneratedMapping(req *MappingGenerationRequest) *GeneratedMapping {
	targetAPI := strings.ToLower(req.TargetAPI)
	
	var mapping *GeneratedMapping
	
	switch {
	case strings.Contains(targetAPI, "slack"):
		mapping = d.generateSlackMapping(req.SourceData)
	case strings.Contains(targetAPI, "telegram"):
		mapping = d.generateTelegramMapping(req.SourceData)
	case strings.Contains(targetAPI, "discord"):
		mapping = d.generateDiscordMapping(req.SourceData)
	case strings.Contains(targetAPI, "amocrm") || strings.Contains(targetAPI, "амо"):
		mapping = d.generateAmoCRMMapping(req.SourceData)
	case strings.Contains(targetAPI, "email"):
		mapping = d.generateEmailMapping(req.SourceData)
	default:
		mapping = d.generateGenericMapping(req.SourceData, req.TargetAPI)
	}
	
	// Добавляем информацию о демо режиме
	mapping.Description += " (Демо режим - настройте реальные параметры перед использованием)"
	mapping.Reasoning = "Это демонстрационный маппинг. В реальном режиме AI анализирует ваши данные и создает оптимальную конфигурацию."
	
	return mapping
}

// GetDemoCreatedIntegration возвращает демо-интеграцию
func (d *DemoResponses) GetDemoCreatedIntegration(req *CreateIntegrationRequest) *CreatedIntegration {
	description := strings.ToLower(req.Description)
	
	var integration *CreatedIntegration
	
	// Определяем тип интеграции по описанию
	if strings.Contains(description, "slack") {
		integration = d.createSlackDemoIntegration()
	} else if strings.Contains(description, "telegram") {
		integration = d.createTelegramDemoIntegration()
	} else if strings.Contains(description, "discord") {
		integration = d.createDiscordDemoIntegration()
	} else if strings.Contains(description, "amocrm") || strings.Contains(description, "амо") {
		integration = d.createAmoCRMDemoIntegration()
	} else if strings.Contains(description, "email") {
		integration = d.createEmailDemoIntegration()
	} else {
		integration = d.createGenericDemoIntegration(req.Description)
	}
	
	// Добавляем информацию о демо режиме
	integration.Explanation += "\n\n⚠️ ДЕМО РЕЖИМ: Это демонстрационная интеграция. Для реальной работы необходимо настроить API ключи и параметры."
	integration.NextSteps = append([]string{"🔧 Выйти из демо режима и настроить реальные API ключи"}, integration.NextSteps...)
	
	return integration
}

// Вспомогательные методы для генерации демо-ответов

func (d *DemoResponses) getSlackDemoResponse() string {
	return `🚀 Отлично! Я помогу вам настроить интеграцию со Slack.

Для отправки уведомлений в Slack вам понадобится:
1. Создать Incoming Webhook в настройках Slack
2. Настроить формат сообщений
3. Выбрать канал для уведомлений

Я могу создать готовый шаблон с красивым форматированием, эмодзи и структурированной подачей информации.

💡 В демо режиме я покажу пример конфигурации, но для реальной работы потребуется настроить webhook URL.`
}

func (d *DemoResponses) getSlackDemoSuggestions() []AISuggestion {
	return []AISuggestion{
		{
			Type:        "mapping_generation",
			Title:       "Создать Slack маппинг",
			Description: "Сгенерировать шаблон для отправки сообщений в Slack",
			Data: map[string]interface{}{
				"action":     "generate_mapping",
				"target_api": "slack",
			},
			Confidence: 0.9,
		},
		{
			Type:        "template",
			Title:       "Настроить формат сообщения",
			Description: "Выбрать стиль и структуру уведомлений",
			Data: map[string]interface{}{
				"action": "configure_template",
				"type":   "slack_message",
			},
			Confidence: 0.8,
		},
	}
}

func (d *DemoResponses) getSlackDemoNextSteps() []string {
	return []string{
		"Создать Incoming Webhook в Slack",
		"Выбрать канал для уведомлений",
		"Настроить формат сообщений",
		"Протестировать отправку",
		"Активировать интеграцию",
	}
}

func (d *DemoResponses) getTelegramDemoResponse() string {
	return `🤖 Создаем интеграцию с Telegram!

Telegram Bot API позволяет отправлять уведомления прямо в чат или канал. Поддерживается:
- HTML и Markdown форматирование
- Кнопки и inline клавиатуры  
- Отправка файлов и изображений
- Групповые чаты и каналы

Для настройки потребуется:
1. Создать бота через @BotFather
2. Получить токен бота
3. Определить chat_id получателя

💡 В демо режиме показываю примеры, реальная настройка требует токена бота.`
}

func (d *DemoResponses) getTelegramDemoSuggestions() []AISuggestion {
	return []AISuggestion{
		{
			Type:        "mapping_generation",
			Title:       "Создать Telegram маппинг",
			Description: "Настроить отправку сообщений через Telegram Bot API",
			Data: map[string]interface{}{
				"action":     "generate_mapping",
				"target_api": "telegram",
			},
			Confidence: 0.9,
		},
	}
}

func (d *DemoResponses) getTelegramDemoNextSteps() []string {
	return []string{
		"Создать бота через @BotFather",
		"Получить токен бота",
		"Определить chat_id",
		"Настроить формат сообщений",
		"Протестировать отправку",
	}
}

func (d *DemoResponses) getDiscordDemoResponse() string {
	return `🎮 Настраиваем Discord интеграцию!

Discord Webhooks позволяют отправлять красивые сообщения с:
- Rich Embeds с цветами и полями
- Аватары и имена ботов
- Упоминания пользователей
- Файлы и изображения

Преимущества Discord:
- Простая настройка webhook
- Богатое форматирование
- Поддержка Markdown
- Интеграция с игровыми сообществами

💡 Демо режим: показываю структуру, для работы нужен реальный webhook URL.`
}

func (d *DemoResponses) getDiscordDemoSuggestions() []AISuggestion {
	return []AISuggestion{
		{
			Type:        "mapping_generation",
			Title:       "Создать Discord маппинг",
			Description: "Настроить webhook для Discord с rich embeds",
			Data: map[string]interface{}{
				"action":     "generate_mapping",
				"target_api": "discord",
			},
			Confidence: 0.9,
		},
	}
}

func (d *DemoResponses) getDiscordDemoNextSteps() []string {
	return []string{
		"Создать webhook в настройках канала",
		"Скопировать webhook URL",
		"Настроить embed сообщения",
		"Выбрать цвета и стиль",
		"Протестировать отправку",
	}
}

func (d *DemoResponses) getAmoCRMDemoResponse() string {
	return `💼 Интеграция с AmoCRM!

AmoCRM API позволяет автоматически:
- Создавать лиды и контакты
- Обновлять сделки
- Добавлять примечания
- Устанавливать теги и статусы

Поддерживаемые операции:
- Создание лидов из форм
- Обновление контактной информации
- Привязка к воронкам продаж
- Автоматизация процессов

💡 Демо: показываю структуру API, для работы нужен Access Token AmoCRM.`
}

func (d *DemoResponses) getAmoCRMDemoSuggestions() []AISuggestion {
	return []AISuggestion{
		{
			Type:        "mapping_generation",
			Title:       "Создать AmoCRM маппинг",
			Description: "Настроить создание лидов и контактов в AmoCRM",
			Data: map[string]interface{}{
				"action":     "generate_mapping",
				"target_api": "amocrm",
			},
			Confidence: 0.9,
		},
	}
}

func (d *DemoResponses) getAmoCRMDemoNextSteps() []string {
	return []string{
		"Получить Access Token в AmoCRM",
		"Определить ID полей и воронок",
		"Настроить маппинг данных",
		"Протестировать создание лида",
		"Настроить автоматизацию",
	}
}

func (d *DemoResponses) getEmailDemoResponse() string {
	return `📧 Email интеграция готова!

Настройка email уведомлений через SendGrid API:
- HTML и текстовые сообщения
- Персонализация контента
- Отслеживание доставки
- Шаблоны писем

Возможности:
- Автоматические уведомления
- Транзакционные письма
- Маркетинговые рассылки
- Аналитика открытий

💡 Демо режим: показываю формат, для отправки нужен API ключ SendGrid.`
}

func (d *DemoResponses) getEmailDemoSuggestions() []AISuggestion {
	return []AISuggestion{
		{
			Type:        "mapping_generation",
			Title:       "Создать Email маппинг",
			Description: "Настроить отправку email через SendGrid API",
			Data: map[string]interface{}{
				"action":     "generate_mapping",
				"target_api": "email",
			},
			Confidence: 0.9,
		},
	}
}

func (d *DemoResponses) getEmailDemoNextSteps() []string {
	return []string{
		"Получить API ключ SendGrid",
		"Настроить адрес отправителя",
		"Создать шаблон письма",
		"Протестировать отправку",
		"Настроить аналитику",
	}
}

func (d *DemoResponses) getGenericDemoResponse(message string) string {
	return fmt.Sprintf(`🤖 Демо режим AI ассистента активен!

Ваш запрос: "%s"

В демо режиме я показываю примеры того, как AI может помочь с интеграциями:
- Анализ структуры данных
- Генерация маппингов
- Создание шаблонов
- Настройка популярных API

Поддерживаемые интеграции:
• Slack - уведомления в каналы
• Telegram - сообщения через бота  
• Discord - webhook сообщения
• AmoCRM - создание лидов
• Email - отправка писем
• И многие другие...

💡 Для полной функциональности настройте AI API ключи в разделе Настройки.`, message)
}

func (d *DemoResponses) getGenericDemoSuggestions() []AISuggestion {
	return []AISuggestion{
		{
			Type:        "data_analysis",
			Title:       "Анализ данных (демо)",
			Description: "Показать пример анализа структуры данных",
			Data: map[string]interface{}{
				"action": "demo_analyze",
			},
			Confidence: 0.8,
		},
		{
			Type:        "mapping_generation",
			Title:       "Создать маппинг (демо)",
			Description: "Показать пример генерации маппинга",
			Data: map[string]interface{}{
				"action": "demo_mapping",
			},
			Confidence: 0.8,
		},
	}
}

func (d *DemoResponses) getGenericDemoNextSteps() []string {
	return []string{
		"Выберите целевой API для интеграции",
		"Предоставьте образец данных",
		"Настройте AI API ключи для полной функциональности",
		"Протестируйте созданную интеграцию",
	}
}

// Методы для генерации демо-маппингов

func (d *DemoResponses) generateSlackMapping(sourceData map[string]interface{}) *GeneratedMapping {
	template := map[string]interface{}{
		"text": "🔔 Новое уведомление\n{{message}}\n\nДетали:\n{{#each data}}• {{@key}}: {{this}}\n{{/each}}",
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	
	return &GeneratedMapping{
		Type:        "json_template",
		Template:    string(templateJSON),
		TargetURL:   "https://hooks.slack.com/services/DEMO/WEBHOOK/URL",
		Method:      "POST",
		Headers:     map[string]string{"Content-Type": "application/json"},
		AuthType:    "none",
		AuthConfig:  map[string]interface{}{},
		Description: "Slack уведомления (демо)",
	}
}

func (d *DemoResponses) generateTelegramMapping(sourceData map[string]interface{}) *GeneratedMapping {
	template := map[string]interface{}{
		"chat_id":    "DEMO_CHAT_ID",
		"text":       "🤖 <b>Уведомление</b>\n\n{{message}}\n\n<i>Отправлено из dmIntegroff</i>",
		"parse_mode": "HTML",
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	
	return &GeneratedMapping{
		Type:        "json_template",
		Template:    string(templateJSON),
		TargetURL:   "https://api.telegram.org/botDEMO_TOKEN/sendMessage",
		Method:      "POST",
		Headers:     map[string]string{"Content-Type": "application/json"},
		AuthType:    "none",
		AuthConfig:  map[string]interface{}{},
		Description: "Telegram уведомления (демо)",
	}
}

func (d *DemoResponses) generateDiscordMapping(sourceData map[string]interface{}) *GeneratedMapping {
	template := map[string]interface{}{
		"content": "Новое уведомление",
		"embeds": []map[string]interface{}{
			{
				"title":       "{{title}}",
				"description": "{{message}}",
				"color":       5814783,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
		},
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	
	return &GeneratedMapping{
		Type:        "json_template",
		Template:    string(templateJSON),
		TargetURL:   "https://discord.com/api/webhooks/DEMO/WEBHOOK",
		Method:      "POST",
		Headers:     map[string]string{"Content-Type": "application/json"},
		AuthType:    "none",
		AuthConfig:  map[string]interface{}{},
		Description: "Discord webhook (демо)",
	}
}

func (d *DemoResponses) generateAmoCRMMapping(sourceData map[string]interface{}) *GeneratedMapping {
	template := []map[string]interface{}{
		{
			"name":  "Лид из dmIntegroff - {{name}}",
			"price": 0,
			"_embedded": map[string]interface{}{
				"contacts": []map[string]interface{}{
					{
						"name": "{{name}}",
						"custom_fields_values": []map[string]interface{}{
							{
								"field_id": 123456,
								"values": []map[string]interface{}{
									{"value": "{{phone}}"},
								},
							},
						},
					},
				},
			},
		},
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	
	return &GeneratedMapping{
		Type:       "json_template",
		Template:   string(templateJSON),
		TargetURL:  "https://DEMO.amocrm.ru/api/v4/leads",
		Method:     "POST",
		Headers:    map[string]string{"Content-Type": "application/json"},
		AuthType:   "bearer",
		AuthConfig: map[string]interface{}{"token": "DEMO_ACCESS_TOKEN"},
		Description: "AmoCRM лиды (демо)",
	}
}

func (d *DemoResponses) generateEmailMapping(sourceData map[string]interface{}) *GeneratedMapping {
	template := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]interface{}{
					{"email": "{{email}}"},
				},
			},
		},
		"from": map[string]interface{}{
			"email": "noreply@demo.com",
		},
		"subject": "Уведомление: {{subject}}",
		"content": []map[string]interface{}{
			{
				"type":  "text/html",
				"value": "<h1>Уведомление</h1><p>{{message}}</p>",
			},
		},
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	
	return &GeneratedMapping{
		Type:       "json_template",
		Template:   string(templateJSON),
		TargetURL:  "https://api.sendgrid.com/v3/mail/send",
		Method:     "POST",
		Headers:    map[string]string{"Content-Type": "application/json"},
		AuthType:   "bearer",
		AuthConfig: map[string]interface{}{"token": "DEMO_SENDGRID_KEY"},
		Description: "Email уведомления (демо)",
	}
}

func (d *DemoResponses) generateGenericMapping(sourceData map[string]interface{}, targetAPI string) *GeneratedMapping {
	template := make(map[string]interface{})
	for key := range sourceData {
		template[key] = fmt.Sprintf("{{%s}}", key)
	}
	
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	
	return &GeneratedMapping{
		Type:        "json_template",
		Template:    string(templateJSON),
		TargetURL:   fmt.Sprintf("https://api.%s.com/webhook", strings.ToLower(targetAPI)),
		Method:      "POST",
		Headers:     map[string]string{"Content-Type": "application/json"},
		AuthType:    "bearer",
		AuthConfig:  map[string]interface{}{"token": "DEMO_API_TOKEN"},
		Description: fmt.Sprintf("%s интеграция (демо)", targetAPI),
	}
}

// Методы для создания демо-интеграций

func (d *DemoResponses) createSlackDemoIntegration() *CreatedIntegration {
	return &CreatedIntegration{
		Name:         "Slack уведомления (демо)",
		TargetURL:    "https://hooks.slack.com/services/DEMO/WEBHOOK/URL",
		Method:       "POST",
		Template:     `{"text": "🔔 {{message}}"}`,
		TemplateType: "json_template",
		Mapping:      map[string]string{"message": "message"},
		AuthType:     "none",
		AuthConfig:   map[string]interface{}{},
		Headers:      map[string]string{"Content-Type": "application/json"},
		Explanation:  "Демонстрационная интеграция для отправки уведомлений в Slack",
		NextSteps: []string{
			"Создать Incoming Webhook в Slack",
			"Заменить DEMO URL на реальный",
			"Настроить канал по умолчанию",
			"Протестировать отправку",
		},
	}
}

func (d *DemoResponses) createTelegramDemoIntegration() *CreatedIntegration {
	return &CreatedIntegration{
		Name:         "Telegram бот (демо)",
		TargetURL:    "https://api.telegram.org/botDEMO_TOKEN/sendMessage",
		Method:       "POST",
		Template:     `{"chat_id": "DEMO_CHAT_ID", "text": "🤖 {{message}}", "parse_mode": "HTML"}`,
		TemplateType: "json_template",
		Mapping:      map[string]string{"message": "message"},
		AuthType:     "none",
		AuthConfig:   map[string]interface{}{},
		Headers:      map[string]string{"Content-Type": "application/json"},
		Explanation:  "Демонстрационная интеграция для отправки сообщений через Telegram Bot API",
		NextSteps: []string{
			"Создать бота через @BotFather",
			"Получить токен бота",
			"Определить chat_id получателя",
			"Заменить DEMO значения на реальные",
		},
	}
}

func (d *DemoResponses) createDiscordDemoIntegration() *CreatedIntegration {
	return &CreatedIntegration{
		Name:         "Discord webhook (демо)",
		TargetURL:    "https://discord.com/api/webhooks/DEMO/WEBHOOK",
		Method:       "POST",
		Template:     `{"content": "🎮 {{message}}"}`,
		TemplateType: "json_template",
		Mapping:      map[string]string{"message": "message"},
		AuthType:     "none",
		AuthConfig:   map[string]interface{}{},
		Headers:      map[string]string{"Content-Type": "application/json"},
		Explanation:  "Демонстрационная интеграция для отправки сообщений в Discord",
		NextSteps: []string{
			"Создать webhook в настройках канала Discord",
			"Скопировать webhook URL",
			"Заменить DEMO URL на реальный",
			"Настроить rich embeds при необходимости",
		},
	}
}

func (d *DemoResponses) createAmoCRMDemoIntegration() *CreatedIntegration {
	return &CreatedIntegration{
		Name:         "AmoCRM интеграция (демо)",
		TargetURL:    "https://DEMO.amocrm.ru/api/v4/leads",
		Method:       "POST",
		Template:     `[{"name": "Лид - {{name}}", "price": 0}]`,
		TemplateType: "json_template",
		Mapping:      map[string]string{"name": "name", "phone": "phone", "email": "email"},
		AuthType:     "bearer",
		AuthConfig:   map[string]interface{}{"token": "DEMO_ACCESS_TOKEN"},
		Headers:      map[string]string{"Content-Type": "application/json"},
		Explanation:  "Демонстрационная интеграция для создания лидов в AmoCRM",
		NextSteps: []string{
			"Получить Access Token в AmoCRM",
			"Заменить DEMO на ваш поддомен",
			"Настроить ID полей",
			"Протестировать создание лида",
		},
	}
}

func (d *DemoResponses) createEmailDemoIntegration() *CreatedIntegration {
	return &CreatedIntegration{
		Name:         "Email уведомления (демо)",
		TargetURL:    "https://api.sendgrid.com/v3/mail/send",
		Method:       "POST",
		Template:     `{"personalizations":[{"to":[{"email":"{{email}}"}]}],"from":{"email":"demo@example.com"},"subject":"{{subject}}","content":[{"type":"text/html","value":"<p>{{message}}</p>"}]}`,
		TemplateType: "json_template",
		Mapping:      map[string]string{"email": "email", "subject": "subject", "message": "message"},
		AuthType:     "bearer",
		AuthConfig:   map[string]interface{}{"token": "DEMO_SENDGRID_KEY"},
		Headers:      map[string]string{"Content-Type": "application/json"},
		Explanation:  "Демонстрационная интеграция для отправки email через SendGrid",
		NextSteps: []string{
			"Получить API ключ SendGrid",
			"Настроить адрес отправителя",
			"Заменить DEMO ключ на реальный",
			"Протестировать отправку email",
		},
	}
}

func (d *DemoResponses) createGenericDemoIntegration(description string) *CreatedIntegration {
	return &CreatedIntegration{
		Name:         "Пользовательская интеграция (демо)",
		TargetURL:    "https://api.example.com/webhook",
		Method:       "POST",
		Template:     `{"message": "{{message}}", "data": "{{data}}"}`,
		TemplateType: "json_template",
		Mapping:      map[string]string{"message": "message", "data": "data"},
		AuthType:     "bearer",
		AuthConfig:   map[string]interface{}{"token": "DEMO_API_TOKEN"},
		Headers:      map[string]string{"Content-Type": "application/json"},
		Explanation:  fmt.Sprintf("Демонстрационная интеграция на основе описания: %s", description),
		NextSteps: []string{
			"Определить реальный API endpoint",
			"Настроить аутентификацию",
			"Адаптировать шаблон под API",
			"Протестировать интеграцию",
		},
	}
}

// Вспомогательные методы для анализа данных

func (d *DemoResponses) detectFieldType(key string, value interface{}) string {
	keyLower := strings.ToLower(key)
	
	// Проверяем по имени поля
	if strings.Contains(keyLower, "email") || strings.Contains(keyLower, "mail") {
		return "email"
	}
	if strings.Contains(keyLower, "phone") || strings.Contains(keyLower, "tel") {
		return "phone"
	}
	if strings.Contains(keyLower, "date") || strings.Contains(keyLower, "time") {
		return "datetime"
	}
	if strings.Contains(keyLower, "url") || strings.Contains(keyLower, "link") {
		return "url"
	}
	if strings.Contains(keyLower, "id") {
		return "number"
	}
	
	// Проверяем по типу значения
	switch value.(type) {
	case bool:
		return "boolean"
	case int, int32, int64, float32, float64:
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "string"
	}
}

func (d *DemoResponses) generateFieldDescription(key, fieldType string) string {
	descriptions := map[string]string{
		"email":    "Адрес электронной почты",
		"phone":    "Номер телефона",
		"datetime": "Дата и время",
		"url":      "URL адрес",
		"number":   "Числовое значение",
		"boolean":  "Логическое значение",
		"array":    "Массив данных",
		"object":   "Объект с вложенными полями",
		"string":   "Текстовая строка",
	}
	
	if desc, exists := descriptions[fieldType]; exists {
		return desc
	}
	return "Поле данных"
}

func (d *DemoResponses) isRequiredField(key string) bool {
	requiredFields := []string{"id", "name", "email", "message", "title"}
	keyLower := strings.ToLower(key)
	
	for _, required := range requiredFields {
		if strings.Contains(keyLower, required) {
			return true
		}
	}
	return false
}

func (d *DemoResponses) generateExamples(value interface{}) []string {
	switch v := value.(type) {
	case string:
		return []string{v}
	case int, int32, int64:
		return []string{fmt.Sprintf("%v", v)}
	case float32, float64:
		return []string{fmt.Sprintf("%.2f", v)}
	case bool:
		return []string{fmt.Sprintf("%v", v)}
	default:
		return []string{"пример значения"}
	}
}

func (d *DemoResponses) generatePattern(fieldType string) string {
	patterns := map[string]string{
		"email":    `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		"phone":    `^\+?[1-9]\d{1,14}$`,
		"url":      `^https?://[^\s/$.?#].[^\s]*$`,
		"number":   `^\d+(\.\d+)?$`,
		"datetime": `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`,
	}
	
	if pattern, exists := patterns[fieldType]; exists {
		return pattern
	}
	return ""
}

func (d *DemoResponses) suggestTargetField(sourceField, fieldType string) string {
	suggestions := map[string]string{
		"email":   "recipient_email",
		"phone":   "contact_phone",
		"name":    "contact_name",
		"message": "notification_text",
		"title":   "subject",
		"id":      "external_id",
	}
	
	sourceLower := strings.ToLower(sourceField)
	for key, target := range suggestions {
		if strings.Contains(sourceLower, key) {
			return target
		}
	}
	
	return ""
}

func (d *DemoResponses) suggestTransform(fieldType string) string {
	transforms := map[string]string{
		"email":    "lowercase",
		"phone":    "normalize_phone",
		"datetime": "format_datetime",
		"string":   "trim",
		"number":   "to_number",
	}
	
	if transform, exists := transforms[fieldType]; exists {
		return transform
	}
	return "none"
}

func (d *DemoResponses) detectDataType(data map[string]interface{}) string {
	// Простая эвристика для определения типа данных
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, strings.ToLower(key))
	}
	
	keyString := strings.Join(keys, " ")
	
	if strings.Contains(keyString, "order") || strings.Contains(keyString, "purchase") {
		return "order"
	}
	if strings.Contains(keyString, "user") || strings.Contains(keyString, "customer") {
		return "user"
	}
	if strings.Contains(keyString, "lead") || strings.Contains(keyString, "contact") {
		return "lead"
	}
	if strings.Contains(keyString, "payment") || strings.Contains(keyString, "transaction") {
		return "payment"
	}
	if strings.Contains(keyString, "error") || strings.Contains(keyString, "exception") {
		return "error"
	}
	
	return "notification"
}

func (d *DemoResponses) generateSchema(fields []FieldInfo, dataType string) string {
	return fmt.Sprintf("Схема данных типа '%s' с %d полями. Демо режим: показывает примерную структуру.", dataType, len(fields))
}