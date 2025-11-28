package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
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

	// Получаем интеграции без проекта для возможности добавления
	var availableIntegrations []models.Integration
	database.DB.Where("project_id IS NULL OR project_id = ?", project.ID).Find(&availableIntegrations)

	c.HTML(http.StatusOK, "project_view.html", gin.H{
		"title":                 project.Name,
		"project":               project,
		"availableIntegrations": availableIntegrations,
	})
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

	// Удаляем проект (интеграции останутся, но project_id станет NULL благодаря ON DELETE SET NULL)
	database.DB.Delete(&models.Project{}, uint(id))

	c.Redirect(http.StatusFound, "/projects")
}

func ProjectAddIntegration(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 32)

	integrationIDStr := c.PostForm("integration_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	projectIDUint := uint(projectID)
	integration.ProjectID = &projectIDUint
	database.DB.Save(&integration)

	c.Redirect(http.StatusFound, "/projects/"+projectIDStr)
}

func ProjectRemoveIntegration(c *gin.Context) {
	projectIDStr := c.Param("id")
	integrationIDStr := c.Param("integration_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)

	var integration models.Integration
	if err := database.DB.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	integration.ProjectID = nil
	database.DB.Save(&integration)

	c.Redirect(http.StatusFound, "/projects/"+projectIDStr)
}
