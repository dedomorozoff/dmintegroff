package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/services"
	"dmintegroff/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// getAppURL возвращает базовый URL приложения из переменной окружения или из запроса
func getAppURL(c *gin.Context) string {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		// Fallback: используем текущий origin из запроса
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		appURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}
	return appURL
}

// getAppPath возвращает базовый путь приложения из переменной окружения
func getAppPath() string {
	appPath := os.Getenv("APP_PATH")
	if appPath == "" {
		return ""
	}
	return appPath
}

func IntegrationList(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	var integrations []models.Integration
	query := database.DB.Preload("Project").Preload("Project.CreatedBy")
	
	// Specialist видит только свои интеграции, admin видит все
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	
	query.Find(&integrations)
	
	// Get all projects for import modal
	var projects []models.Project
	projectQuery := database.DB.Order("name ASC")
	if role != "admin" {
		projectQuery = projectQuery.Where("created_by_id = ?", userID)
	}
	projectQuery.Find(&projects)

	c.HTML(http.StatusOK, "pages/integrations.html", gin.H{
		"title":        "Интеграции",
		"CurrentPage":  "integrations",
		"integrations": integrations,
		"projects":     projects,
		"username":     session.Get("username"),
		"role":         role,
		"appURL":       getAppURL(c),
		"appPath":      getAppPath(),
	})
}

// IntegrationsListAPI - API endpoint для получения списка интеграций в JSON
func IntegrationsListAPI(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	var integrations []models.Integration
	query := database.DB.Preload("Project")
	
	// Specialist видит только свои интеграции
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	
	query.Find(&integrations)

	// Возвращаем только необходимые данные для обновления статуса
	type IntegrationStatus struct {
		ID            uint   `json:"id"`
		Mode          string `json:"mode"`
		HasPayload    bool   `json:"has_payload"`
		SamplePayload string `json:"sample_payload"`
	}

	statuses := make([]IntegrationStatus, len(integrations))
	for i, integration := range integrations {
		statuses[i] = IntegrationStatus{
			ID:            integration.ID,
			Mode:          integration.Mode,
			HasPayload:    integration.SamplePayload != "",
			SamplePayload: integration.SamplePayload,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"integrations": statuses,
	})
}

func IntegrationCreate(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	var projects []models.Project
	query := database.DB
	
	// Specialist видит только свои проекты
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	
	query.Find(&projects)

	// Получаем информацию о пользователе для проверки демо-режима
	userIDInterface := session.Get("user_id")
	isDemo := false
	if userIDInterface != nil {
		if userID, ok := userIDInterface.(uint); ok {
			var user models.User
			if err := database.DB.First(&user, userID).Error; err == nil {
				isDemo = user.IsDemo
			}
		}
	}

	// Проверяем, есть ли уже интеграции у пользователя
	var integrationCount int64
	countQuery := database.DB.Model(&models.Integration{})
	if role != "admin" {
		countQuery = countQuery.Where("created_by_id = ?", userID)
	}
	countQuery.Count(&integrationCount)
	isFirstIntegration := integrationCount == 0

	c.HTML(http.StatusOK, "pages/integration_create.html", gin.H{
		"title":              "Создание интеграции",
		"CurrentPage":        "integration_create",
		"projects":           projects,
		"username":           session.Get("username"),
		"role":               session.Get("role"),
		"isDemo":             isDemo,
		"appURL":             getAppURL(c),
		"appPath":            getAppPath(),
		"isFirstIntegration": isFirstIntegration,
	})
}

func IntegrationStore(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	if userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user session"})
		return
	}

	// Generate unique webhook token
	token, err := utils.GenerateToken(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Проверяем, что проект выбран (обязательное поле)
	projectIDStr := c.PostForm("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Проект обязателен для создания интеграции"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID проекта"})
		return
	}

	// Проверяем, что проект существует и пользователь имеет к нему доступ
	role := session.Get("role")
	var project models.Project
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&project, uint(projectID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Проект не найден"})
		return
	}

	// Проверяем демо-режим и валидируем Target API URL
	var user models.User
	targetAPI := c.PostForm("target_api")
	if err := database.DB.First(&user, userID).Error; err == nil && user.IsDemo {
		// В демо-режиме разрешен только тестовый URL
		testURL := "/webhook/test"
		if targetAPI != testURL && !strings.HasSuffix(targetAPI, testURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "В демо-режиме разрешен только тестовый URL"})
			return
		}
	}

	httpMethod := c.PostForm("http_method")
	if httpMethod == "" {
		httpMethod = "POST" // Default to POST
	}

	// Webhook signature fields
	webhookSignatureEnabled := c.PostForm("webhook_signature_enabled") == "true"
	webhookSignatureHeader := c.PostForm("webhook_signature_header")
	if webhookSignatureHeader == "" {
		webhookSignatureHeader = "X-Webhook-Signature"
	}
	webhookSignatureAlgorithm := c.PostForm("webhook_signature_algorithm")
	if webhookSignatureAlgorithm == "" {
		webhookSignatureAlgorithm = "sha256"
	}

	integration := models.Integration{
		Name:         c.PostForm("name"),
		WebhookToken: token,
		SourceAPI:    c.PostForm("source_api"), // Optional
		TargetAPI:    targetAPI,
		HTTPMethod:   httpMethod,
		Mode:         "listening", // Start in listening mode
		CreatedByID:  userID,
		ProjectID:    uint(projectID),
		
		// OAuth and authentication fields
		AuthType:           c.PostForm("auth_type"),
		OAuth2TokenURL:     c.PostForm("oauth2_token_url"),
		OAuth2ClientID:     c.PostForm("oauth2_client_id"),
		OAuth2ClientSecret: c.PostForm("oauth2_client_secret"),
		OAuth2Scope:        c.PostForm("oauth2_scope"),
		OAuth2GrantType:    c.PostForm("oauth2_grant_type"),
		BearerToken:        c.PostForm("bearer_token"),
		BasicAuthUser:      c.PostForm("basic_auth_user"),
		BasicAuthPass:      c.PostForm("basic_auth_pass"),
		
		// Webhook signature fields
		WebhookSignatureEnabled:   webhookSignatureEnabled,
		WebhookSignatureSecret:    c.PostForm("webhook_signature_secret"),
		WebhookSignatureHeader:    webhookSignatureHeader,
		WebhookSignatureAlgorithm: webhookSignatureAlgorithm,
		
		// Custom headers
		CustomHeaders: c.PostForm("custom_headers"),
	}
	
	// Log custom headers for debugging
	logger.Log.WithFields(map[string]interface{}{
		"custom_headers": integration.CustomHeaders,
		"name":           integration.Name,
	}).Info("Creating integration with custom headers")
	
	// Validate signature config if enabled
	if err := services.ValidateSignatureConfig(&integration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook signature configuration error: " + err.Error()})
		return
	}

	if err := database.DB.Create(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/integrations")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func WebhookHandler(c *gin.Context) {
	token := c.Param("token")

	var integration models.Integration
	if err := database.DB.Where("webhook_token = ?", token).First(&integration).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Читаем тело запроса как байты
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Пытаемся распарсить как JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		// Логируем ошибку парсинга
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
			"error":          err.Error(),
			"body_preview":   string(bodyBytes[:min(len(bodyBytes), 100)]),
		}).Error("Failed to parse webhook body as JSON")
		
		// Если не JSON, возвращаем ошибку с подробностями
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	payloadJSON := bodyBytes // Используем оригинальные байты
	headersJSON, _ := json.Marshal(c.Request.Header)

	// Логируем входящий webhook
	incomingLog := models.RequestLog{
		IntegrationID:  integration.ID,
		Method:         c.Request.Method,
		URL:            c.Request.URL.Path,
		RequestBody:    string(payloadJSON),
		RequestHeaders: string(headersJSON),
		StatusCode:     200,
		LogType:        "webhook",
	}
	CreateLogWithLimit(&incomingLog)

	// If in listening mode, save sample payload
	if integration.Mode == "listening" {
		integration.SamplePayload = string(payloadJSON)
		database.DB.Save(&integration)

		c.JSON(http.StatusOK, gin.H{
			"status":  "captured",
			"message": "Sample data captured. Configure field mapping to activate integration.",
		})
		return
	}

	// If active, process the webhook
	if integration.Mode == "active" {
		if err := services.ProcessWebhook(integration.ID, payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
		return
	}

	// If inactive
	c.JSON(http.StatusOK, gin.H{"status": "inactive", "message": "Integration is inactive"})
}

func IntegrationConfigure(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Если нет SamplePayload, но есть MappingConfig, пытаемся использовать его для отображения
	// Это позволяет редактировать маппинг даже если нет свежих данных
	var fields []utils.FieldInfo
	var sampleData map[string]interface{}

	if integration.SamplePayload != "" {
		// Parse sample payload to show fields (recursive)
		var err error
		fields, err = utils.ParseJSONString(integration.SamplePayload)
		if err != nil {
			fields = []utils.FieldInfo{}
		}

		// Также сохраняем оригинальные данные для отображения
		json.Unmarshal([]byte(integration.SamplePayload), &sampleData)
	} else if integration.MappingConfig != "" {
		// Если нет SamplePayload, но есть MappingConfig, пытаемся восстановить структуру из маппинга
		// Это позволяет редактировать маппинг активной интеграции
		var mapping map[string]string
		if err := json.Unmarshal([]byte(integration.MappingConfig), &mapping); err == nil {
			// Создаем поля из маппинга (обратный маппинг)
			for _, sourceField := range mapping {
				fields = append(fields, utils.FieldInfo{
					Path:     sourceField,
					Value:    nil,
					Type:     "unknown",
					FullPath: sourceField,
				})
			}
		}
	}

	// Подготавливаем payload для JavaScript (экранируем JSON)
	var payloadJSON string
	if integration.SamplePayload != "" {
		payloadBytes, _ := json.Marshal(integration.SamplePayload)
		payloadJSON = string(payloadBytes)
	} else {
		payloadJSON = "null"
	}

	// Загружаем текущий маппинг для отображения в форме
	var currentMapping map[string]string
	if integration.MappingConfig != "" {
		json.Unmarshal([]byte(integration.MappingConfig), &currentMapping)
	}

	// Сортируем поля по алфавиту
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})

	c.HTML(http.StatusOK, "pages/integration_configure.html", gin.H{
		"title":          "Настройка маппинга",
		"CurrentPage":    "integrations",
		"integration":    integration,
		"sampleData":     sampleData,
		"fields":         fields,
		"payloadJSON":    payloadJSON,
		"currentMapping": currentMapping,
		"username":       session.Get("username"),
		"role":           role,
		"appURL":         getAppURL(c),
		"appPath":        getAppPath(),
	})
}

// IntegrationCheckUpdate - API endpoint для проверки обновлений интеграции
func IntegrationCheckUpdate(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Возвращаем информацию о том, есть ли SamplePayload
	hasPayload := integration.SamplePayload != ""

	c.JSON(http.StatusOK, gin.H{
		"has_payload":    hasPayload,
		"mode":           integration.Mode,
		"sample_payload": integration.SamplePayload,
	})
}

func IntegrationSaveMapping(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Проверяем, это обновление только payload или полное сохранение маппинга
	if c.PostForm("mapping_config") == "" && c.PostForm("output_template") == "" && c.PostForm("sample_payload") != "" {
		// Обновляем только SamplePayload
		newPayload := c.PostForm("sample_payload")
		// Проверяем валидность JSON
		var testData interface{}
		if err := json.Unmarshal([]byte(newPayload), &testData); err == nil {
			integration.SamplePayload = newPayload
			database.DB.Save(&integration)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("/integrations/%d/configure", id))
		return
	}

	// Обновляем SamplePayload, если он был изменен
	if newPayload := c.PostForm("sample_payload"); newPayload != "" {
		// Проверяем валидность JSON
		var testData interface{}
		if err := json.Unmarshal([]byte(newPayload), &testData); err == nil {
			integration.SamplePayload = newPayload
		}
	}

	// Получаем output_template или mapping_config
	outputTemplate := c.PostForm("output_template")
	mappingConfig := c.PostForm("mapping_config")
	templateType := c.PostForm("template_type")

	// Валидируем output_template, если он задан
	if outputTemplate != "" {
		// Для JSON шаблонов валидируем структуру
		if templateType == "json" || templateType == "" {
			processor := utils.NewTemplateProcessor()
			if err := processor.ValidateTemplate(outputTemplate); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid output template: " + err.Error()})
				return
			}
		}
		// Для других типов (xml, text, custom) просто сохраняем как есть
		integration.OutputTemplate = outputTemplate
		integration.TemplateType = templateType
		if integration.TemplateType == "" {
			integration.TemplateType = "json" // По умолчанию JSON
		}
		// Очищаем старый mapping_config, если используется шаблон
		integration.MappingConfig = ""
	} else if mappingConfig != "" {
		// Используем старый способ с mapping_config
		integration.MappingConfig = mappingConfig
		// Очищаем output_template
		integration.OutputTemplate = ""
		integration.TemplateType = "json" // Маппинг всегда генерирует JSON
	}

	// Активируем интеграцию только если она была в режиме listening или inactive
	// Если уже active, оставляем active
	if integration.Mode == "listening" || integration.Mode == "inactive" {
		integration.Mode = "active"
	}

	database.DB.Save(&integration)

	c.Redirect(http.StatusFound, "/integrations")
}

func IntegrationEdit(c *gin.Context) {
	session := sessions.Default(c)
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Получаем информацию о пользователе для проверки демо-режима
	userIDInterface := session.Get("user_id")
	isDemo := false
	if userIDInterface != nil {
		if userID, ok := userIDInterface.(uint); ok {
			var user models.User
			if err := database.DB.First(&user, userID).Error; err == nil {
				isDemo = user.IsDemo
			}
		}
	}

	// Check if enrichment feature is enabled
	enrichmentEnabled := os.Getenv("ENABLE_GRAPHQL_ENRICHMENT") == "true"

	c.HTML(http.StatusOK, "pages/integration_edit.html", gin.H{
		"title":             "Редактирование интеграции",
		"CurrentPage":       "integrations",
		"integration":       integration,
		"username":          session.Get("username"),
		"role":              session.Get("role"),
		"isDemo":            isDemo,
		"appURL":            getAppURL(c),
		"appPath":           getAppPath(),
		"enrichmentEnabled": enrichmentEnabled,
	})
}

func IntegrationUpdate(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Проверяем демо-режим и валидируем Target API URL
	targetAPI := c.PostForm("target_api")
	if userID != nil {
		if uid, ok := userID.(uint); ok {
			var user models.User
			if err := database.DB.First(&user, uid).Error; err == nil && user.IsDemo {
				// В демо-режиме разрешен только тестовый URL
				testURL := "/webhook/test"
				if targetAPI != testURL && !strings.HasSuffix(targetAPI, testURL) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "В демо-режиме разрешен только тестовый URL"})
					return
				}
			}
		}
	}

	httpMethod := c.PostForm("http_method")
	if httpMethod == "" {
		httpMethod = "POST" // Default to POST
	}

	integration.Name = c.PostForm("name")
	integration.SourceAPI = c.PostForm("source_api")
	integration.TargetAPI = targetAPI
	integration.HTTPMethod = httpMethod
	
	// Update OAuth and authentication fields
	integration.AuthType = c.PostForm("auth_type")
	integration.OAuth2TokenURL = c.PostForm("oauth2_token_url")
	integration.OAuth2ClientID = c.PostForm("oauth2_client_id")
	
	// Only update secret if provided (don't overwrite with empty)
	if newSecret := c.PostForm("oauth2_client_secret"); newSecret != "" {
		integration.OAuth2ClientSecret = newSecret
	}
	
	integration.OAuth2Scope = c.PostForm("oauth2_scope")
	integration.OAuth2GrantType = c.PostForm("oauth2_grant_type")
	
	// Update bearer token if provided
	if newToken := c.PostForm("bearer_token"); newToken != "" {
		integration.BearerToken = newToken
	}
	
	integration.BasicAuthUser = c.PostForm("basic_auth_user")
	
	// Only update password if provided
	if newPass := c.PostForm("basic_auth_pass"); newPass != "" {
		integration.BasicAuthPass = newPass
	}
	
	// Update webhook signature fields
	integration.WebhookSignatureEnabled = c.PostForm("webhook_signature_enabled") == "true"
	
	// Only update secret if provided (don't overwrite with empty)
	if newSecret := c.PostForm("webhook_signature_secret"); newSecret != "" {
		integration.WebhookSignatureSecret = newSecret
	}
	
	webhookSignatureHeader := c.PostForm("webhook_signature_header")
	if webhookSignatureHeader != "" {
		integration.WebhookSignatureHeader = webhookSignatureHeader
	} else {
		integration.WebhookSignatureHeader = "X-Webhook-Signature"
	}
	
	webhookSignatureAlgorithm := c.PostForm("webhook_signature_algorithm")
	if webhookSignatureAlgorithm != "" {
		integration.WebhookSignatureAlgorithm = webhookSignatureAlgorithm
	} else {
		integration.WebhookSignatureAlgorithm = "sha256"
	}
	
	// Update custom headers
	integration.CustomHeaders = c.PostForm("custom_headers")
	
	// Log custom headers for debugging
	logger.Log.WithFields(map[string]interface{}{
		"integration_id":  integration.ID,
		"custom_headers":  integration.CustomHeaders,
	}).Info("Updating integration with custom headers")
	
	// Validate signature config if enabled
	if err := services.ValidateSignatureConfig(&integration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook signature configuration error: " + err.Error()})
		return
	}

	database.DB.Save(&integration)

	c.Redirect(http.StatusFound, "/integrations")
}

// IntegrationRegenerateToken - генерация нового webhook токена
func IntegrationRegenerateToken(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Генерируем новый токен
	newToken, err := utils.GenerateToken(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Обновляем токен
	integration.WebhookToken = newToken
	if err := database.DB.Save(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token"})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/integrations/%d/edit", id))
}

func IntegrationDelete(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// Проверяем доступ перед удалением
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	database.DB.Delete(&integration)

	c.Redirect(http.StatusFound, "/integrations")
}

// IntegrationToggle - активация/деактивация интеграции
func IntegrationToggle(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Переключаем режим
	if integration.Mode == "active" {
		integration.Mode = "inactive"
	} else if integration.Mode == "inactive" || integration.Mode == "listening" {
		// Проверяем, что есть маппинг перед активацией
		if integration.MappingConfig == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot activate: mapping not configured"})
			return
		}
		integration.Mode = "active"
	}

	database.DB.Save(&integration)
	c.Redirect(http.StatusFound, "/integrations")
}

// IntegrationReconfigure - переход в режим переопределения маппинга
func IntegrationReconfigure(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Переводим в режим прослушивания для получения новых данных
	integration.Mode = "listening"
	integration.SamplePayload = "" // Очищаем старые данные

	database.DB.Save(&integration)
	c.Redirect(http.StatusFound, "/integrations")
}

// IntegrationCancelListening - отмена режима прослушивания
func IntegrationCancelListening(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Переводим в режим неактивна
	integration.Mode = "inactive"

	database.DB.Save(&integration)
	c.Redirect(http.StatusFound, "/integrations")
}

// IntegrationTestOAuth - тестирование OAuth конфигурации
func IntegrationTestOAuth(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	if integration.AuthType != "oauth2" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Integration is not configured for OAuth2"})
		return
	}

	// Test OAuth connection
	if err := services.TestOAuth2Connection(&integration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "OAuth2 connection successful",
	})
}

// IntegrationGraphQLConfigure shows GraphQL configuration page
func IntegrationGraphQLConfigure(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	integrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "pages/error.html", gin.H{
			"title":   "Ошибка",
			"message": "Неверный ID интеграции",
		})
		return
	}

	var integration models.Integration
	query := database.DB.Preload("Project").Preload("CreatedBy")

	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}

	if err := query.First(&integration, integrationID).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/error.html", gin.H{
			"title":   "Не найдено",
			"message": "Интеграция не найдена",
		})
		return
	}

	// Check if it's a GraphQL integration
	if integration.APIType != "graphql" {
		c.HTML(http.StatusBadRequest, "pages/error.html", gin.H{
			"title":   "Ошибка",
			"message": "Это не GraphQL интеграция",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/integration_graphql_configure.html", gin.H{
		"title":       "Настройка GraphQL",
		"integration": integration,
		"username":    session.Get("username"),
		"role":        role,
	})
}

// IntegrationGraphQLConfigureSave saves GraphQL configuration
func IntegrationGraphQLConfigureSave(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	integrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID интеграции"})
		return
	}

	var integration models.Integration
	query := database.DB

	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}

	if err := query.First(&integration, integrationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Интеграция не найдена"})
		return
	}

	// Check if it's a GraphQL integration
	if integration.APIType != "graphql" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Это не GraphQL интеграция"})
		return
	}

	var req struct {
		GraphQLQuery         string `json:"graphql_query"`
		GraphQLVariables     string `json:"graphql_variables"`
		GraphQLOperationName string `json:"graphql_operation_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate JSON if provided
	if req.GraphQLVariables != "" {
		var test map[string]interface{}
		if err := json.Unmarshal([]byte(req.GraphQLVariables), &test); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON в маппинге переменных"})
			return
		}
	}

	// Update integration
	integration.GraphQLQuery = req.GraphQLQuery
	integration.GraphQLVariables = req.GraphQLVariables
	integration.GraphQLOperationName = req.GraphQLOperationName

	if err := database.DB.Save(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// IntegrationEnrichmentConfigure shows enrichment configuration page
func IntegrationEnrichmentConfigure(c *gin.Context) {
	// Check if enrichment feature is enabled
	if os.Getenv("ENABLE_GRAPHQL_ENRICHMENT") != "true" {
		c.HTML(http.StatusForbidden, "pages/error.html", gin.H{
			"title":   "Функция недоступна",
			"message": "Функция обогащения данных отключена. Включите ENABLE_GRAPHQL_ENRICHMENT в .env",
		})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	integrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "pages/error.html", gin.H{
			"title":   "Ошибка",
			"message": "Неверный ID интеграции",
		})
		return
	}

	var integration models.Integration
	query := database.DB.Preload("Project").Preload("CreatedBy")

	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}

	if err := query.First(&integration, integrationID).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/error.html", gin.H{
			"title":   "Не найдено",
			"message": "Интеграция не найдена",
		})
		return
	}

	// Check if it's a REST integration
	if integration.APIType != "rest" && integration.APIType != "" {
		c.HTML(http.StatusBadRequest, "pages/error.html", gin.H{
			"title":   "Ошибка",
			"message": "Обогащение доступно только для REST интеграций",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/integration_enrichment_configure.html", gin.H{
		"title":       "Обогащение данных",
		"integration": integration,
		"username":    session.Get("username"),
		"role":        role,
	})
}

// IntegrationEnrichmentConfigureSave saves enrichment configuration
func IntegrationEnrichmentConfigureSave(c *gin.Context) {
	// Check if enrichment feature is enabled
	if os.Getenv("ENABLE_GRAPHQL_ENRICHMENT") != "true" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Функция обогащения данных отключена"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	integrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID интеграции"})
		return
	}

	var integration models.Integration
	query := database.DB

	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}

	if err := query.First(&integration, integrationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Интеграция не найдена"})
		return
	}

	var req struct {
		EnrichmentEnabled   bool   `json:"enrichment_enabled"`
		EnrichmentEndpoint  string `json:"enrichment_endpoint"`
		EnrichmentQuery     string `json:"enrichment_query"`
		EnrichmentVariables string `json:"enrichment_variables"`
		EnrichmentMergeMode string `json:"enrichment_merge_mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate JSON if provided
	if req.EnrichmentVariables != "" {
		var test map[string]interface{}
		if err := json.Unmarshal([]byte(req.EnrichmentVariables), &test); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON в маппинге переменных"})
			return
		}
	}

	// Update integration
	integration.EnrichmentEnabled = req.EnrichmentEnabled
	integration.EnrichmentEndpoint = req.EnrichmentEndpoint
	integration.EnrichmentQuery = req.EnrichmentQuery
	integration.EnrichmentVariables = req.EnrichmentVariables
	integration.EnrichmentMergeMode = req.EnrichmentMergeMode

	if err := database.DB.Save(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
