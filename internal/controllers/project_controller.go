package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ProjectList(c *gin.Context) {
	var projects []models.Project
	database.DB.Preload("Integrations").Preload("CreatedBy").Find(&projects)

	c.HTML(http.StatusOK, "projects.html", gin.H{
		"title":    "Проекты",
		"projects": projects,
	})
}

func ProjectCreate(c *gin.Context) {
	c.HTML(http.StatusOK, "project_create.html", gin.H{
		"title": "Создание проекта",
	})
}

func ProjectStore(c *gin.Context) {
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

	project := models.Project{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		CreatedByID: userID,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/projects")
}

func ProjectView(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var project models.Project
	if err := database.DB.Preload("Integrations").Preload("CreatedBy").First(&project, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.HTML(http.StatusOK, "project_view.html", gin.H{
		"title":   project.Name,
		"project": project,
	})
}

// ProjectCreateIntegration - создание интеграции в контексте проекта
func ProjectCreateIntegration(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 32)

	var project models.Project
	if err := database.DB.First(&project, uint(projectID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.HTML(http.StatusOK, "project_integration_create.html", gin.H{
		"title":   "Создание интеграции",
		"project": project,
	})
}

// ProjectStoreIntegration - сохранение интеграции в контексте проекта
func ProjectStoreIntegration(c *gin.Context) {
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

	projectIDStr := c.Param("id")
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 32)

	// Проверяем, что проект существует
	var project models.Project
	if err := database.DB.First(&project, uint(projectID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Generate unique webhook token
	token, err := utils.GenerateToken(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	integration := models.Integration{
		Name:         c.PostForm("name"),
		WebhookToken: token,
		SourceAPI:    c.PostForm("source_api"),
		TargetAPI:    c.PostForm("target_api"),
		Mode:         "listening",
		CreatedByID:  userID,
		ProjectID:    uint(projectID),
	}

	if err := database.DB.Create(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectIDStr)
}

// ProjectDeleteIntegration - удаление интеграции из проекта
func ProjectDeleteIntegration(c *gin.Context) {
	projectIDStr := c.Param("id")
	integrationIDStr := c.Param("integration_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)

	// Удаляем интеграцию
	database.DB.Delete(&models.Integration{}, uint(integrationID))

	c.Redirect(http.StatusFound, "/projects/"+projectIDStr)
}

func ProjectEdit(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var project models.Project
	if err := database.DB.First(&project, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.HTML(http.StatusOK, "project_edit.html", gin.H{
		"title":   "Редактирование проекта",
		"project": project,
	})
}

func ProjectUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var project models.Project
	if err := database.DB.First(&project, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	project.Name = c.PostForm("name")
	project.Description = c.PostForm("description")

	database.DB.Save(&project)

	c.Redirect(http.StatusFound, "/projects/"+idStr)
}

func ProjectDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// Проверяем количество интеграций в проекте
	var integrationCount int64
	database.DB.Model(&models.Integration{}).Where("project_id = ?", uint(id)).Count(&integrationCount)

	// Удаляем проект (интеграции удалятся автоматически благодаря ON DELETE CASCADE)
	database.DB.Delete(&models.Project{}, uint(id))

	c.Redirect(http.StatusFound, "/projects")
}


