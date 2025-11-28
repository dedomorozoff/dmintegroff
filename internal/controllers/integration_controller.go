package controllers

import (
	"encoding/json"
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"dmintegroff/internal/services"
	"dmintegroff/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IntegrationList(c *gin.Context) {
	var integrations []models.Integration
	database.DB.Preload("Project").Find(&integrations)

	c.HTML(http.StatusOK, "integrations.html", gin.H{
		"title":        "Интеграции",
		"integrations": integrations,
	})
}

func IntegrationCreate(c *gin.Context) {
	var projects []models.Project
	database.DB.Find(&projects)

	c.HTML(http.StatusOK, "integration_create.html", gin.H{
		"title":    "Создание интеграции",
		"projects": projects,
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

	// If in listening mode, save sample payload
	if integration.Mode == "listening" {
		payloadJSON, _ := json.Marshal(payload)
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

	// Parse sample payload to show fields
	var sampleData map[string]interface{}
	if integration.SamplePayload != "" {
		json.Unmarshal([]byte(integration.SamplePayload), &sampleData)
	}

	c.HTML(http.StatusOK, "integration_configure.html", gin.H{
		"title":       "Настройка маппинга",
		"integration": integration,
		"sampleData":  sampleData,
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

	// Get mapping config from form
	mappingConfig := c.PostForm("mapping_config")
	integration.MappingConfig = mappingConfig
	integration.Mode = "active" // Activate integration

	database.DB.Save(&integration)

	c.Redirect(http.StatusFound, "/integrations")
}

func IntegrationEdit(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	c.HTML(http.StatusOK, "integration_edit.html", gin.H{
		"title":       "Редактирование интеграции",
		"integration": integration,
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
