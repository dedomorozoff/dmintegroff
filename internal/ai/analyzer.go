package ai

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"dmintegroff/internal/logger"
)

// Analyzer анализатор данных для AI
type Analyzer struct {
	client *Client
}

// NewAnalyzer создает новый анализатор
func NewAnalyzer(client *Client) *Analyzer {
	return &Analyzer{client: client}
}

// AnalyzeDataStructure анализирует структуру данных
func (a *Analyzer) AnalyzeDataStructure(ctx context.Context, data map[string]interface{}, format string) (*DataAnalysisResponse, error) {
	startTime := time.Now()
	
	// Логируем начало анализа
	logger.Log.WithFields(map[string]interface{}{
		"action": "analyzer_start",
		"format": format,
		"data_fields": len(data),
		"ai_configured": a.client.IsConfigured(),
	}).Info("AI Analyzer: Starting data structure analysis")
	
	// Сначала делаем локальный анализ
	localAnalysis := a.analyzeLocally(data)
	
	logger.Log.WithFields(map[string]interface{}{
		"action": "analyzer_local_complete",
		"detected_type": localAnalysis.DataType,
		"fields_count": len(localAnalysis.Fields),
		"confidence": localAnalysis.Confidence,
		"suggestions_count": len(localAnalysis.Suggestions),
	}).Info("AI Analyzer: Local analysis completed")
	
	// Если AI настроен, дополняем анализом от AI
	if a.client.IsConfigured() {
		logger.Log.WithFields(map[string]interface{}{
			"action": "analyzer_ai_start",
		}).Info("AI Analyzer: Starting AI analysis")
		
		aiAnalysis, err := a.client.AnalyzeData(ctx, data, format)
		if err == nil {
			// Объединяем результаты
			mergedAnalysis := a.mergeAnalysis(localAnalysis, aiAnalysis)
			
			logger.Log.WithFields(map[string]interface{}{
				"action": "analyzer_success",
				"analysis_type": "ai_enhanced",
				"final_type": mergedAnalysis.DataType,
				"final_confidence": mergedAnalysis.Confidence,
				"fields_count": len(mergedAnalysis.Fields),
				"suggestions_count": len(mergedAnalysis.Suggestions),
				"duration": time.Since(startTime).String(),
			}).Info("AI Analyzer: Analysis completed with AI enhancement")
			
			return mergedAnalysis, nil
		}
		
		// Если AI недоступен, используем только локальный анализ
		logger.Log.WithFields(map[string]interface{}{
			"action": "analyzer_ai_fallback",
			"error": err.Error(),
			"fallback_to": "local_only",
		}).Warn("AI Analyzer: AI analysis failed, using local analysis only")
	} else {
		logger.Log.WithFields(map[string]interface{}{
			"action": "analyzer_local_only",
			"reason": "ai_not_configured",
		}).Info("AI Analyzer: Using local analysis only (AI not configured)")
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"action": "analyzer_success",
		"analysis_type": "local_only",
		"final_type": localAnalysis.DataType,
		"final_confidence": localAnalysis.Confidence,
		"fields_count": len(localAnalysis.Fields),
		"suggestions_count": len(localAnalysis.Suggestions),
		"duration": time.Since(startTime).String(),
	}).Info("AI Analyzer: Analysis completed with local analysis only")
	
	return localAnalysis, nil
}

// analyzeLocally выполняет локальный анализ данных без AI
func (a *Analyzer) analyzeLocally(data map[string]interface{}) *DataAnalysisResponse {
	fields := make([]FieldInfo, 0)
	
	// Анализируем каждое поле
	for key, value := range data {
		field := a.analyzeField(key, value)
		fields = append(fields, field)
	}
	
	// Определяем тип данных на основе полей
	dataType := a.detectDataType(fields)
	
	// Создаем базовые предложения маппинга
	suggestions := a.createBasicSuggestions(fields, dataType)
	
	return &DataAnalysisResponse{
		Fields:      fields,
		Schema:      a.generateSchema(fields),
		Suggestions: suggestions,
		DataType:    dataType,
		Confidence:  0.7, // Локальный анализ менее точен чем AI
	}
}

// analyzeField анализирует отдельное поле
func (a *Analyzer) analyzeField(name string, value interface{}) FieldInfo {
	field := FieldInfo{
		Name:     name,
		Required: true, // По умолчанию считаем обязательным
		Examples: []string{fmt.Sprintf("%v", value)},
	}
	
	// Определяем тип поля
	switch v := value.(type) {
	case string:
		field.Type = a.detectStringType(v)
		field.Description = a.generateFieldDescription(name, field.Type)
	case float64, int, int64:
		field.Type = "number"
		field.Description = "Числовое значение"
	case bool:
		field.Type = "boolean"
		field.Description = "Логическое значение (true/false)"
	case []interface{}:
		field.Type = "array"
		field.Description = "Массив значений"
	case map[string]interface{}:
		field.Type = "object"
		field.Description = "Вложенный объект"
	case nil:
		field.Type = "string"
		field.Required = false
		field.Description = "Необязательное поле"
	default:
		field.Type = "string"
		field.Description = "Строковое значение"
	}
	
	return field
}

// detectStringType определяет тип строкового поля
func (a *Analyzer) detectStringType(value string) string {
	// Email
	if a.isEmail(value) {
		return "email"
	}
	
	// Phone
	if a.isPhone(value) {
		return "phone"
	}
	
	// URL
	if a.isURL(value) {
		return "url"
	}
	
	// Date/DateTime
	if a.isDateTime(value) {
		return "datetime"
	}
	
	if a.isDate(value) {
		return "date"
	}
	
	// Number as string
	if a.isNumericString(value) {
		return "number"
	}
	
	return "string"
}

// isEmail проверяет, является ли строка email
func (a *Analyzer) isEmail(value string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(value)
}

// isPhone проверяет, является ли строка телефоном
func (a *Analyzer) isPhone(value string) bool {
	phoneRegex := regexp.MustCompile(`^[\+]?[1-9][\d]{0,15}$`)
	cleanValue := regexp.MustCompile(`[^\d+]`).ReplaceAllString(value, "")
	return len(cleanValue) >= 7 && phoneRegex.MatchString(cleanValue)
}

// isURL проверяет, является ли строка URL
func (a *Analyzer) isURL(value string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(value)
}

// isDateTime проверяет, является ли строка датой и временем
func (a *Analyzer) isDateTime(value string) bool {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"02.01.2006 15:04:05",
		"02/01/2006 15:04:05",
	}
	
	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}
	return false
}

// isDate проверяет, является ли строка датой
func (a *Analyzer) isDate(value string) bool {
	formats := []string{
		"2006-01-02",
		"02.01.2006",
		"02/01/2006",
		"01/02/2006",
	}
	
	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}
	return false
}

// isNumericString проверяет, является ли строка числом
func (a *Analyzer) isNumericString(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

// generateFieldDescription генерирует описание поля
func (a *Analyzer) generateFieldDescription(name, fieldType string) string {
	name = strings.ToLower(name)
	
	// Специальные поля
	descriptions := map[string]string{
		"id":         "Уникальный идентификатор",
		"email":      "Email адрес",
		"phone":      "Номер телефона",
		"name":       "Имя",
		"title":      "Заголовок",
		"message":    "Сообщение",
		"text":       "Текстовое содержимое",
		"amount":     "Сумма",
		"price":      "Цена",
		"total":      "Общая сумма",
		"count":      "Количество",
		"status":     "Статус",
		"type":       "Тип",
		"category":   "Категория",
		"created_at": "Дата создания",
		"updated_at": "Дата обновления",
		"timestamp":  "Временная метка",
		"url":        "URL адрес",
		"link":       "Ссылка",
	}
	
	if desc, exists := descriptions[name]; exists {
		return desc
	}
	
	// Описание по типу
	typeDescriptions := map[string]string{
		"email":    "Email адрес",
		"phone":    "Номер телефона",
		"url":      "URL адрес",
		"date":     "Дата",
		"datetime": "Дата и время",
		"number":   "Числовое значение",
		"boolean":  "Логическое значение",
		"array":    "Массив значений",
		"object":   "Вложенный объект",
	}
	
	if desc, exists := typeDescriptions[fieldType]; exists {
		return desc
	}
	
	return "Строковое значение"
}

// detectDataType определяет тип данных на основе полей
func (a *Analyzer) detectDataType(fields []FieldInfo) string {
	fieldNames := make([]string, len(fields))
	for i, field := range fields {
		fieldNames[i] = strings.ToLower(field.Name)
	}
	
	// Паттерны для определения типа данных
	patterns := map[string][]string{
		"order": {"order_id", "amount", "total", "customer", "product"},
		"user":  {"user_id", "email", "name", "phone", "username"},
		"event": {"event", "type", "timestamp", "data", "source"},
		"notification": {"message", "title", "text", "alert"},
		"payment": {"amount", "currency", "payment_id", "transaction"},
		"lead": {"name", "email", "phone", "company", "source"},
	}
	
	maxScore := 0
	detectedType := "unknown"
	
	for dataType, keywords := range patterns {
		score := 0
		for _, keyword := range keywords {
			for _, fieldName := range fieldNames {
				if strings.Contains(fieldName, keyword) {
					score++
				}
			}
		}
		
		if score > maxScore {
			maxScore = score
			detectedType = dataType
		}
	}
	
	return detectedType
}

// createBasicSuggestions создает базовые предложения маппинга
func (a *Analyzer) createBasicSuggestions(fields []FieldInfo, dataType string) []FieldMapping {
	suggestions := make([]FieldMapping, 0)
	
	// Предложения на основе типа данных
	switch dataType {
	case "order":
		suggestions = append(suggestions, a.createOrderSuggestions(fields)...)
	case "user":
		suggestions = append(suggestions, a.createUserSuggestions(fields)...)
	case "event":
		suggestions = append(suggestions, a.createEventSuggestions(fields)...)
	case "notification":
		suggestions = append(suggestions, a.createNotificationSuggestions(fields)...)
	}
	
	return suggestions
}

// createOrderSuggestions создает предложения для заказов
func (a *Analyzer) createOrderSuggestions(fields []FieldInfo) []FieldMapping {
	suggestions := make([]FieldMapping, 0)
	
	for _, field := range fields {
		name := strings.ToLower(field.Name)
		
		switch {
		case strings.Contains(name, "order_id") || strings.Contains(name, "id"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "order_number",
				Transform:   "none",
				Confidence:  0.9,
				Reasoning:   "Идентификатор заказа",
			})
		case strings.Contains(name, "amount") || strings.Contains(name, "total"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "amount",
				Transform:   "format_currency",
				Confidence:  0.9,
				Reasoning:   "Сумма заказа",
			})
		case strings.Contains(name, "customer") || strings.Contains(name, "name"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "customer_name",
				Transform:   "none",
				Confidence:  0.8,
				Reasoning:   "Имя клиента",
			})
		}
	}
	
	return suggestions
}

// createUserSuggestions создает предложения для пользователей
func (a *Analyzer) createUserSuggestions(fields []FieldInfo) []FieldMapping {
	suggestions := make([]FieldMapping, 0)
	
	for _, field := range fields {
		name := strings.ToLower(field.Name)
		
		switch {
		case field.Type == "email":
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "email",
				Transform:   "lowercase",
				Confidence:  0.95,
				Reasoning:   "Email адрес пользователя",
			})
		case strings.Contains(name, "name"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "full_name",
				Transform:   "none",
				Confidence:  0.9,
				Reasoning:   "Полное имя пользователя",
			})
		case field.Type == "phone":
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "phone",
				Transform:   "format_phone",
				Confidence:  0.9,
				Reasoning:   "Номер телефона",
			})
		}
	}
	
	return suggestions
}

// createEventSuggestions создает предложения для событий
func (a *Analyzer) createEventSuggestions(fields []FieldInfo) []FieldMapping {
	suggestions := make([]FieldMapping, 0)
	
	for _, field := range fields {
		name := strings.ToLower(field.Name)
		
		switch {
		case strings.Contains(name, "event") || strings.Contains(name, "type"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "event_type",
				Transform:   "none",
				Confidence:  0.9,
				Reasoning:   "Тип события",
			})
		case field.Type == "datetime" || strings.Contains(name, "timestamp"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "timestamp",
				Transform:   "format_datetime",
				Confidence:  0.9,
				Reasoning:   "Время события",
			})
		}
	}
	
	return suggestions
}

// createNotificationSuggestions создает предложения для уведомлений
func (a *Analyzer) createNotificationSuggestions(fields []FieldInfo) []FieldMapping {
	suggestions := make([]FieldMapping, 0)
	
	for _, field := range fields {
		name := strings.ToLower(field.Name)
		
		switch {
		case strings.Contains(name, "message") || strings.Contains(name, "text"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "text",
				Transform:   "none",
				Confidence:  0.95,
				Reasoning:   "Текст уведомления",
			})
		case strings.Contains(name, "title"):
			suggestions = append(suggestions, FieldMapping{
				SourceField: field.Name,
				TargetField: "title",
				Transform:   "none",
				Confidence:  0.9,
				Reasoning:   "Заголовок уведомления",
			})
		}
	}
	
	return suggestions
}

// generateSchema генерирует описание схемы данных
func (a *Analyzer) generateSchema(fields []FieldInfo) string {
	var parts []string
	
	for _, field := range fields {
		required := ""
		if field.Required {
			required = " (обязательное)"
		}
		parts = append(parts, fmt.Sprintf("- %s: %s%s", field.Name, field.Type, required))
	}
	
	return strings.Join(parts, "\n")
}

// mergeAnalysis объединяет локальный и AI анализ
func (a *Analyzer) mergeAnalysis(local, ai *DataAnalysisResponse) *DataAnalysisResponse {
	// Используем AI результат как основу, дополняя локальными данными
	result := *ai
	
	// Если AI не смог определить тип поля, используем локальный
	for i, aiField := range result.Fields {
		for _, localField := range local.Fields {
			if aiField.Name == localField.Name && aiField.Type == "string" && localField.Type != "string" {
				result.Fields[i].Type = localField.Type
			}
		}
	}
	
	// Повышаем уверенность если AI и локальный анализ согласны
	if result.DataType == local.DataType {
		result.Confidence = (result.Confidence + local.Confidence) / 2 + 0.1
		if result.Confidence > 1.0 {
			result.Confidence = 1.0
		}
	}
	
	return &result
}