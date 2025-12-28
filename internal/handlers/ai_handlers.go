package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AIGenerateMappingRequest представляет запрос на генерацию маппинга
type AIGenerateMappingRequest struct {
	SamplePayload   string `json:"sample_payload" binding:"required"`
	TargetSystem    string `json:"target_system"`
	Description     string `json:"description"`
	ExistingMapping string `json:"existing_mapping,omitempty"`
}

// AIGenerateMappingResponse представляет ответ с сгенерированным маппингом
type AIGenerateMappingResponse struct {
	Status      string                 `json:"status"`
	Mapping     map[string]string      `json:"mapping"`
	Template    string                 `json:"template,omitempty"`
	TemplateType string                `json:"template_type,omitempty"`
	Explanation string                 `json:"explanation"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// AICreateIntegrationRequest представляет запрос на создание интеграции
type AICreateIntegrationRequest struct {
	Description   string `json:"description" binding:"required"`
	SourceSystem  string `json:"source_system,omitempty"`
	TargetSystem  string `json:"target_system,omitempty"`
	SampleData    string `json:"sample_data,omitempty"`
	ProjectID     int    `json:"project_id,omitempty"`
}

// AICreateIntegrationResponse представляет ответ с созданной интеграцией
type AICreateIntegrationResponse struct {
	Status        string                 `json:"status"`
	Integration   map[string]interface{} `json:"integration"`
	Mapping       map[string]string      `json:"mapping,omitempty"`
	Template      string                 `json:"template,omitempty"`
	TemplateType  string                 `json:"template_type,omitempty"`
	Explanation   string                 `json:"explanation"`
	NextSteps     []string               `json:"next_steps"`
}

// AIGenerateMapping генерирует маппинг полей с помощью AI
func (h *Handler) AIGenerateMapping(c *gin.Context) {
	var req AIGenerateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Неверный формат запроса: " + err.Error(),
		})
		return
	}

	// Получаем ID интеграции из URL
	integrationIDStr := c.Param("id")
	integrationID, err := strconv.Atoi(integrationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Неверный ID интеграции",
		})
		return
	}

	// Получаем интеграцию из базы данных
	integration, err := h.db.GetIntegration(integrationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Интеграция не найдена",
		})
		return
	}

	// Генерируем маппинг с помощью AI
	mapping, template, templateType, explanation, suggestions, err := h.generateMappingWithAI(req, integration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Ошибка генерации маппинга: " + err.Error(),
		})
		return
	}

	response := AIGenerateMappingResponse{
		Status:       "success",
		Mapping:      mapping,
		Template:     template,
		TemplateType: templateType,
		Explanation:  explanation,
		Suggestions:  suggestions,
	}

	c.JSON(http.StatusOK, response)
}

// AICreateIntegration создает интеграцию с помощью AI
func (h *Handler) AICreateIntegration(c *gin.Context) {
	var req AICreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Неверный формат запроса: " + err.Error(),
		})
		return
	}

	// Создаем интеграцию с помощью AI
	integration, mapping, template, templateType, explanation, nextSteps, err := h.createIntegrationWithAI(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Ошибка создания интеграции: " + err.Error(),
		})
		return
	}

	response := AICreateIntegrationResponse{
		Status:       "success",
		Integration:  integration,
		Mapping:      mapping,
		Template:     template,
		TemplateType: templateType,
		Explanation:  explanation,
		NextSteps:    nextSteps,
	}

	c.JSON(http.StatusOK, response)
}

// generateMappingWithAI генерирует маппинг с помощью AI
func (h *Handler) generateMappingWithAI(req AIGenerateMappingRequest, integration interface{}) (map[string]string, string, string, string, []string, error) {
	// Парсим sample payload для извлечения полей
	var sampleData interface{}
	if err := json.Unmarshal([]byte(req.SamplePayload), &sampleData); err != nil {
		return nil, "", "", "", nil, fmt.Errorf("неверный формат JSON: %v", err)
	}

	// Извлекаем поля из sample data
	fields := extractFieldsFromData(sampleData, "")
	
	// Определяем целевую систему и создаем маппинг
	targetSystem := strings.ToLower(req.TargetSystem)
	mapping := make(map[string]string)
	var template string
	var templateType string
	var explanation string
	var suggestions []string

	// Генерируем маппинг на основе целевой системы
	switch {
	case strings.Contains(targetSystem, "salesforce"):
		mapping, template, templateType, explanation, suggestions = h.generateSalesforceMapping(fields, req.Description)
	case strings.Contains(targetSystem, "hubspot"):
		mapping, template, templateType, explanation, suggestions = h.generateHubSpotMapping(fields, req.Description)
	case strings.Contains(targetSystem, "slack"):
		mapping, template, templateType, explanation, suggestions = h.generateSlackMapping(fields, req.Description)
	case strings.Contains(targetSystem, "telegram"):
		mapping, template, templateType, explanation, suggestions = h.generateTelegramMapping(fields, req.Description)
	case strings.Contains(targetSystem, "webhook"):
		mapping, template, templateType, explanation, suggestions = h.generateWebhookMapping(fields, req.Description)
	default:
		mapping, template, templateType, explanation, suggestions = h.generateGenericMapping(fields, req.Description, targetSystem)
	}

	return mapping, template, templateType, explanation, suggestions, nil
}

// createIntegrationWithAI создает интеграцию с помощью AI
func (h *Handler) createIntegrationWithAI(req AICreateIntegrationRequest) (map[string]interface{}, map[string]string, string, string, string, []string, error) {
	// Анализируем описание для определения типа интеграции
	description := strings.ToLower(req.Description)
	
	var integrationName string
	var targetAPI string
	var httpMethod string
	var mapping map[string]string
	var template string
	var templateType string
	var explanation string
	var nextSteps []string

	// Определяем тип интеграции на основе описания
	switch {
	case strings.Contains(description, "slack"):
		integrationName = "Slack уведомления"
		targetAPI = "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
		httpMethod = "POST"
		mapping, template, templateType, explanation, nextSteps = h.generateSlackIntegration(req)
		
	case strings.Contains(description, "telegram"):
		integrationName = "Telegram бот"
		targetAPI = "https://api.telegram.org/botYOUR_BOT_TOKEN/sendMessage"
		httpMethod = "POST"
		mapping, template, templateType, explanation, nextSteps = h.generateTelegramIntegration(req)
		
	case strings.Contains(description, "salesforce") || strings.Contains(description, "crm"):
		integrationName = "Salesforce CRM"
		targetAPI = "https://your-instance.salesforce.com/services/data/v52.0/sobjects/Lead"
		httpMethod = "POST"
		mapping, template, templateType, explanation, nextSteps = h.generateSalesforceIntegration(req)
		
	case strings.Contains(description, "email") || strings.Contains(description, "mail"):
		integrationName = "Email уведомления"
		targetAPI = "https://api.sendgrid.com/v3/mail/send"
		httpMethod = "POST"
		mapping, template, templateType, explanation, nextSteps = h.generateEmailIntegration(req)
		
	default:
		integrationName = "Пользовательская интеграция"
		targetAPI = "https://api.example.com/webhook"
		httpMethod = "POST"
		mapping, template, templateType, explanation, nextSteps = h.generateGenericIntegration(req)
	}

	// Создаем объект интеграции
	integration := map[string]interface{}{
		"name":            integrationName,
		"target_api":      targetAPI,
		"http_method":     httpMethod,
		"project_id":      req.ProjectID,
		"mode":           "inactive",
		"sample_payload":  req.SampleData,
		"output_template": template,
		"template_type":   templateType,
		"description":     req.Description,
	}

	return integration, mapping, template, templateType, explanation, nextSteps, nil
}

// Вспомогательные функции для генерации маппингов различных систем

func (h *Handler) generateSalesforceMapping(fields []string, description string) (map[string]string, string, string, string, []string) {
	mapping := make(map[string]string)
	
	// Стандартные маппинги для Salesforce
	for _, field := range fields {
		fieldLower := strings.ToLower(field)
		switch {
		case strings.Contains(fieldLower, "email"):
			mapping["Email"] = field
		case strings.Contains(fieldLower, "first") && strings.Contains(fieldLower, "name"):
			mapping["FirstName"] = field
		case strings.Contains(fieldLower, "last") && strings.Contains(fieldLower, "name"):
			mapping["LastName"] = field
		case strings.Contains(fieldLower, "name") && !strings.Contains(fieldLower, "first") && !strings.Contains(fieldLower, "last"):
			mapping["LastName"] = field
		case strings.Contains(fieldLower, "phone"):
			mapping["Phone"] = field
		case strings.Contains(fieldLower, "company"):
			mapping["Company"] = field
		}
	}

	template := `{
  "records": [{
    "attributes": {
      "type": "Lead"
    },
    "FirstName": "{{FirstName}}",
    "LastName": "{{LastName}}",
    "Email": "{{Email}}",
    "Phone": "{{Phone}}",
    "Company": "{{Company}}",
    "Status": "Open - Not Contacted",
    "LeadSource": "Web"
  }]
}`

	explanation := "Создан маппинг для Salesforce Lead API. Поля автоматически сопоставлены с стандартными полями Salesforce."
	suggestions := []string{
		"Добавьте авторизацию OAuth 2.0 для Salesforce",
		"Настройте обработку ошибок для дублирующихся записей",
		"Рассмотрите добавление кастомных полей",
	}

	return mapping, template, "json", explanation, suggestions
}

func (h *Handler) generateSlackMapping(fields []string, description string) (map[string]string, string, string, string, []string) {
	mapping := make(map[string]string)
	
	// Стандартные маппинги для Slack
	for _, field := range fields {
		fieldLower := strings.ToLower(field)
		switch {
		case strings.Contains(fieldLower, "message") || strings.Contains(fieldLower, "text"):
			mapping["text"] = field
		case strings.Contains(fieldLower, "channel"):
			mapping["channel"] = field
		case strings.Contains(fieldLower, "user") || strings.Contains(fieldLower, "name"):
			mapping["username"] = field
		}
	}

	template := `{
  "text": "Новое уведомление: {{text}}",
  "channel": "#general",
  "username": "dmIntegroff Bot",
  "icon_emoji": ":robot_face:"
}`

	explanation := "Создан маппинг для Slack Webhook API. Сообщения будут отправляться в указанный канал."
	suggestions := []string{
		"Настройте канал по умолчанию",
		"Добавьте форматирование сообщений",
		"Рассмотрите использование Slack App вместо webhook",
	}

	return mapping, template, "json", explanation, suggestions
}

func (h *Handler) generateTelegramMapping(fields []string, description string) (map[string]string, string, string, string, []string) {
	mapping := make(map[string]string)
	
	// Стандартные маппинги для Telegram
	for _, field := range fields {
		fieldLower := strings.ToLower(field)
		switch {
		case strings.Contains(fieldLower, "message") || strings.Contains(fieldLower, "text"):
			mapping["text"] = field
		case strings.Contains(fieldLower, "chat") || strings.Contains(fieldLower, "id"):
			mapping["chat_id"] = field
		}
	}

	template := `{
  "chat_id": "YOUR_CHAT_ID",
  "text": "📢 Новое уведомление:\n\n{{text}}",
  "parse_mode": "HTML"
}`

	explanation := "Создан маппинг для Telegram Bot API. Сообщения будут отправляться в указанный чат."
	suggestions := []string{
		"Замените YOUR_CHAT_ID на реальный ID чата",
		"Настройте токен бота в URL",
		"Добавьте обработку ошибок для заблокированных ботов",
	}

	return mapping, template, "json", explanation, suggestions
}

func (h *Handler) generateWebhookMapping(fields []string, description string) (map[string]string, string, string, string, []string) {
	mapping := make(map[string]string)
	
	// Простой маппинг 1:1 для webhook
	for _, field := range fields {
		mapping[field] = field
	}

	template := `{
  "timestamp": "{{timestamp}}",
  "source": "dmIntegroff",
  "data": {
    {{range $key, $value := .}}
    "{{$key}}": "{{$value}}"{{if not (last $key)}}{{end}}
    {{end}}
  }
}`

	explanation := "Создан простой маппинг для webhook. Все поля передаются как есть с добавлением метаданных."
	suggestions := []string{
		"Настройте фильтрацию полей при необходимости",
		"Добавьте валидацию данных",
		"Рассмотрите добавление подписи для безопасности",
	}

	return mapping, template, "json", explanation, suggestions
}

func (h *Handler) generateGenericMapping(fields []string, description string, targetSystem string) (map[string]string, string, string, string, []string) {
	mapping := make(map[string]string)
	
	// Простой маппинг 1:1
	for _, field := range fields {
		mapping[field] = field
	}

	template := `{
  {{range $i, $field := .Fields}}
  "{{$field}}": "{{index $.Values $field}}"{{if not (last $i)}}{{end}}
  {{end}}
}`

	explanation := fmt.Sprintf("Создан базовый маппинг для системы '%s'. Все поля сопоставлены 1:1.", targetSystem)
	suggestions := []string{
		"Настройте маппинг в соответствии с API целевой системы",
		"Добавьте валидацию и трансформацию данных",
		"Проверьте документацию API для правильных имен полей",
	}

	return mapping, template, "json", explanation, suggestions
}

// Функции для создания интеграций

func (h *Handler) generateSlackIntegration(req AICreateIntegrationRequest) (map[string]string, string, string, string, []string) {
	mapping := map[string]string{
		"text":     "message",
		"channel":  "#general",
		"username": "dmIntegroff Bot",
	}

	template := `{
  "text": "{{message}}",
  "channel": "#general",
  "username": "dmIntegroff Bot",
  "icon_emoji": ":robot_face:"
}`

	explanation := "Создана интеграция для отправки уведомлений в Slack через Incoming Webhook."
	nextSteps := []string{
		"1. Создайте Incoming Webhook в настройках Slack",
		"2. Замените URL на реальный webhook URL",
		"3. Настройте канал по умолчанию",
		"4. Протестируйте отправку сообщения",
		"5. Активируйте интеграцию",
	}

	return mapping, template, "json", explanation, nextSteps
}

func (h *Handler) generateTelegramIntegration(req AICreateIntegrationRequest) (map[string]string, string, string, string, []string) {
	mapping := map[string]string{
		"text":    "message",
		"chat_id": "YOUR_CHAT_ID",
	}

	template := `{
  "chat_id": "YOUR_CHAT_ID",
  "text": "📢 {{message}}",
  "parse_mode": "HTML"
}`

	explanation := "Создана интеграция для отправки сообщений через Telegram Bot API."
	nextSteps := []string{
		"1. Создайте бота через @BotFather в Telegram",
		"2. Получите токен бота и замените в URL",
		"3. Получите chat_id и замените в шаблоне",
		"4. Протестируйте отправку сообщения",
		"5. Активируйте интеграцию",
	}

	return mapping, template, "json", explanation, nextSteps
}

func (h *Handler) generateSalesforceIntegration(req AICreateIntegrationRequest) (map[string]string, string, string, string, []string) {
	mapping := map[string]string{
		"FirstName": "first_name",
		"LastName":  "last_name",
		"Email":     "email",
		"Phone":     "phone",
		"Company":   "company",
	}

	template := `{
  "records": [{
    "attributes": {
      "type": "Lead"
    },
    "FirstName": "{{first_name}}",
    "LastName": "{{last_name}}",
    "Email": "{{email}}",
    "Phone": "{{phone}}",
    "Company": "{{company}}",
    "Status": "Open - Not Contacted",
    "LeadSource": "Web"
  }]
}`

	explanation := "Создана интеграция для создания лидов в Salesforce CRM."
	nextSteps := []string{
		"1. Настройте OAuth 2.0 авторизацию в Salesforce",
		"2. Замените URL на ваш Salesforce instance",
		"3. Добавьте необходимые заголовки авторизации",
		"4. Протестируйте создание лида",
		"5. Активируйте интеграцию",
	}

	return mapping, template, "json", explanation, nextSteps
}

func (h *Handler) generateEmailIntegration(req AICreateIntegrationRequest) (map[string]string, string, string, string, []string) {
	mapping := map[string]string{
		"to":      "email",
		"subject": "subject",
		"content": "message",
	}

	template := `{
  "personalizations": [{
    "to": [{"email": "{{email}}"}],
    "subject": "{{subject}}"
  }],
  "from": {"email": "noreply@yourdomain.com"},
  "content": [{
    "type": "text/plain",
    "value": "{{message}}"
  }]
}`

	explanation := "Создана интеграция для отправки email через SendGrid API."
	nextSteps := []string{
		"1. Получите API ключ SendGrid",
		"2. Добавьте заголовок Authorization с API ключом",
		"3. Настройте адрес отправителя",
		"4. Протестируйте отправку email",
		"5. Активируйте интеграцию",
	}

	return mapping, template, "json", explanation, nextSteps
}

func (h *Handler) generateGenericIntegration(req AICreateIntegrationRequest) (map[string]string, string, string, string, []string) {
	mapping := map[string]string{
		"data": "payload",
	}

	template := `{
  "timestamp": "{{timestamp}}",
  "source": "dmIntegroff",
  "data": "{{payload}}"
}`

	explanation := "Создана базовая интеграция для отправки данных на внешний API."
	nextSteps := []string{
		"1. Замените URL на реальный API endpoint",
		"2. Настройте необходимые заголовки авторизации",
		"3. Адаптируйте шаблон под формат API",
		"4. Протестируйте отправку данных",
		"5. Активируйте интеграцию",
	}

	return mapping, template, "json", explanation, nextSteps
}

// extractFieldsFromData извлекает все поля из JSON данных
func extractFieldsFromData(data interface{}, prefix string) []string {
	var fields []string
	
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fieldPath := key
			if prefix != "" {
				fieldPath = prefix + "." + key
			}
			
			// Добавляем текущее поле
			fields = append(fields, fieldPath)
			
			// Рекурсивно обрабатываем вложенные объекты
			if nestedFields := extractFieldsFromData(value, fieldPath); len(nestedFields) > 0 {
				fields = append(fields, nestedFields...)
			}
		}
	case []interface{}:
		if len(v) > 0 {
			// Обрабатываем первый элемент массива
			arrayFields := extractFieldsFromData(v[0], prefix+"[0]")
			fields = append(fields, arrayFields...)
		}
	}
	
	return fields
}