package controllers

import (
	"gintegra/internal/database"
	"gintegra/internal/models"
	"gintegra/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IntegrationList(c *gin.Context) {
	var integrations []models.Integration
	database.DB.Find(&integrations)

	c.HTML(http.StatusOK, "integrations.html", gin.H{
		"title":        "Интеграции",
		"integrations": integrations,
	})
}

func IntegrationCreate(c *gin.Context) {
	c.HTML(http.StatusOK, "integration_create.html", gin.H{
		"title": "Создание интеграции",
	})
}

func IntegrationStore(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(uint)

	integration := models.Integration{
		Name:          c.PostForm("name"),
		SourceAPI:     c.PostForm("source_api"),
		TargetAPI:     c.PostForm("target_api"),
		MappingConfig: c.PostForm("mapping_config"),
		CreatedByID:   userID,
	}

	if err := database.DB.Create(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/integrations")
}

func WebhookHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := services.ProcessWebhook(uint(id), payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
