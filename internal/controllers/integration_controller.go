package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"dmintegroff/internal/services"
	"dmintegroff/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IntegrationList(c *gin.Context) {
	session := sessions.Default(c)
	var integrations []models.Integration
	database.DB.Preload("Project").Find(&integrations)

	c.HTML(http.StatusOK, "pages/integrations.html", gin.H{
		"title":        "Интеграции",
		"CurrentPage":  "integrations",
		"integrations": integrations,
		"username":     session.Get("username"),
		"role":         session.Get("role"),
	})
}

// IntegrationsListAPI - API endpoint для получения списка интеграций в JSON
func IntegrationsListAPI(c *gin.Context) {
	var integrations []models.Integration
	database.DB.Preload("Project").Find(&integrations)

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
	var projects []models.Project
	database.DB.Find(&projects)

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

	c.HTML(http.StatusOK, "pages/integration_create.html", gin.H{
		"title":       "Создание интеграции",
		"CurrentPage": "integration_create",
		"projects":    projects,
		"username":    session.Get("username"),
		"role":        session.Get("role"),
		"isDemo":      isDemo,
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

	// Проверяем, что проект существует
	var project models.Project
	if err := database.DB.First(&project, uint(projectID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Проект не найден"})
		return
	}

	integration := models.Integration{
		Name:         c.PostForm("name"),
		WebhookToken: token,
		SourceAPI:    c.PostForm("source_api"), // Optional
		TargetAPI:    c.PostForm("target_api"),
		Mode:         "listening", // Start in listening mode
		CreatedByID:  userID,
		ProjectID:    uint(projectID),
	}

	if err := database.DB.Create(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/integrations")
}

func WebhookHandler(c *gin.Context) {
	token := c.Param("token")

	var integration models.Integration
	if err := database.DB.Where("webhook_token = ?", token).First(&integration).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	payloadJSON, _ := json.Marshal(payload)
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
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
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

	session := sessions.Default(c)
	c.HTML(http.StatusOK, "pages/integration_configure.html", gin.H{
		"title":          "Настройка маппинга",
		"CurrentPage":    "integrations",
		"integration":    integration,
		"sampleData":     sampleData,
		"fields":         fields,
		"payloadJSON":    payloadJSON,
		"currentMapping": currentMapping,
		"username":       session.Get("username"),
		"role":           session.Get("role"),
	})
}

// IntegrationCheckUpdate - API endpoint для проверки обновлений интеграции
func IntegrationCheckUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
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
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
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

	// Валидируем output_template, если он задан
	if outputTemplate != "" {
		processor := utils.NewTemplateProcessor()
		if err := processor.ValidateTemplate(outputTemplate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid output template: " + err.Error()})
			return
		}
		integration.OutputTemplate = outputTemplate
		// Очищаем старый mapping_config, если используется шаблон
		integration.MappingConfig = ""
	} else if mappingConfig != "" {
		// Используем старый способ с mapping_config
		integration.MappingConfig = mappingConfig
		// Очищаем output_template
		integration.OutputTemplate = ""
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

	c.HTML(http.StatusOK, "pages/integration_edit.html", gin.H{
		"title":       "Редактирование интеграции",
		"CurrentPage": "integrations",
		"integration": integration,
		"username":    session.Get("username"),
		"role":        session.Get("role"),
		"isDemo":      isDemo,
	})
}

func IntegrationUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	integration.Name = c.PostForm("name")
	integration.SourceAPI = c.PostForm("source_api")
	integration.TargetAPI = c.PostForm("target_api")

	database.DB.Save(&integration)

	c.Redirect(http.StatusFound, "/integrations")
}

// IntegrationRegenerateToken - генерация нового webhook токена
func IntegrationRegenerateToken(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
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
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	database.DB.Delete(&models.Integration{}, uint(id))

	c.Redirect(http.StatusFound, "/integrations")
}

// IntegrationToggle - активация/деактивация интеграции
func IntegrationToggle(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
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
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
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
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Переводим в режим неактивна
	integration.Mode = "inactive"

	database.DB.Save(&integration)
	c.Redirect(http.StatusFound, "/integrations")
}
