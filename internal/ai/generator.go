package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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