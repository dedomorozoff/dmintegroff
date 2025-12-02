package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// getProjectAppURL возвращает базовый URL приложения из переменной окружения или из запроса
func getProjectAppURL(c *gin.Context) string {
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

// getProjectAppPath возвращает базовый путь приложения из переменной окружения
func getProjectAppPath() string {
	appPath := os.Getenv("APP_PATH")
	if appPath == "" {
		return ""
	}
	return appPath
}

func ProjectList(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	// Отладка
	fmt.Printf("DEBUG ProjectList: userID=%v, role=%v, roleType=%T\n", userID, role, role)
	
	var projects []models.Project
	query := database.DB.Preload("Integrations").Preload("CreatedBy")
	
	// Specialist видит только свои проекты, admin видит все
	if role != "admin" {
		fmt.Printf("DEBUG: Applying filter for non-admin user\n")
		query = query.Where("created_by_id = ?", userID)
	} else {
		fmt.Printf("DEBUG: Admin user, showing all projects\n")
	}
	
	query.Find(&projects)
	fmt.Printf("DEBUG: Found %d projects\n", len(projects))

	c.HTML(http.StatusOK, "pages/projects.html", gin.H{
		"title":       "Проекты",
		"CurrentPage": "projects",
		"projects":    projects,
		"username":    session.Get("username"),
		"role":        role,
	})
}

func ProjectCreate(c *gin.Context) {
	session := sessions.Default(c)
	c.HTML(http.StatusOK, "pages/project_create.html", gin.H{
		"title":       "Создание проекта",
		"CurrentPage": "projects",
		"username":    session.Get("username"),
		"role":        session.Get("role"),
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
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var project models.Project
	query := database.DB.Preload("Integrations").Preload("CreatedBy")
	
	// Specialist может видеть только свои проекты
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	
	if err := query.First(&project, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Проект не найден"})
		return
	}

	c.HTML(http.StatusOK, "pages/project_view.html", gin.H{
		"title":       project.Name,
		"CurrentPage": "projects",
		"project":     project,
		"username":    session.Get("username"),
		"role":        role,
		"appURL":      getProjectAppURL(c),
		"appPath":     getProjectAppPath(),
	})
}

// ProjectCreateIntegration - создание интеграции в контексте проекта
func ProjectCreateIntegration(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	projectIDStr := c.Param("id")
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 32)

	var project models.Project
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&project, uint(projectID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Проект не найден"})
		return
	}

	// Получаем информацию о пользователе для проверки демо-режима
	isDemo := false
	if userID != nil {
		if uid, ok := userID.(uint); ok {
			var user models.User
			if err := database.DB.First(&user, uid).Error; err == nil {
				isDemo = user.IsDemo
			}
		}
	}

	c.HTML(http.StatusOK, "pages/project_integration_create.html", gin.H{
		"title":       "Создание интеграции",
		"CurrentPage": "projects",
		"project":     project,
		"username":    session.Get("username"),
		"role":        role,
		"isDemo":      isDemo,
		"appURL":      getAppURL(c),
		"appPath":     getAppPath(),
	})
}

// ProjectStoreIntegration - сохранение интеграции в контексте проекта
func ProjectStoreIntegration(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	role := session.Get("role")
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

	// Проверяем, что проект существует и пользователь имеет к нему доступ
	var project models.Project
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&project, uint(projectID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Проект не найден"})
		return
	}

	// Generate unique webhook token
	token, err := utils.GenerateToken(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Проверяем демо-режим и валидируем Target API URL
	targetAPI := c.PostForm("target_api")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err == nil && user.IsDemo {
		// В демо-режиме разрешен только тестовый URL
		testURL := "/webhook/test"
		if targetAPI != testURL && !strings.HasSuffix(targetAPI, testURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "В демо-режиме разрешен только тестовый URL"})
			return
		}
	}

	integration := models.Integration{
		Name:         c.PostForm("name"),
		WebhookToken: token,
		SourceAPI:    c.PostForm("source_api"),
		TargetAPI:    targetAPI,
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
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	projectIDStr := c.Param("id")
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 32)
	integrationIDStr := c.Param("integration_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)

	// Проверяем доступ к проекту
	var project models.Project
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&project, uint(projectID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Проект не найден"})
		return
	}

	// Удаляем интеграцию только если она принадлежит этому проекту
	database.DB.Where("project_id = ?", projectID).Delete(&models.Integration{}, uint(integrationID))

	c.Redirect(http.StatusFound, "/projects/"+projectIDStr)
}

func ProjectEdit(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var project models.Project
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&project, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Проект не найден"})
		return
	}

	c.HTML(http.StatusOK, "pages/project_edit.html", gin.H{
		"title":       "Редактирование проекта",
		"CurrentPage": "projects",
		"project":     project,
		"username":    session.Get("username"),
		"role":        role,
	})
}

func ProjectUpdate(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var project models.Project
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&project, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Проект не найден"})
		return
	}

	project.Name = c.PostForm("name")
	project.Description = c.PostForm("description")

	database.DB.Save(&project)

	c.Redirect(http.StatusFound, "/projects/"+idStr)
}

func ProjectDelete(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// Проверяем доступ к проекту перед удалением
	var project models.Project
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&project, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Проект не найден"})
		return
	}

	// Удаляем проект (все связанные интеграции удалятся автоматически благодаря ON DELETE CASCADE)
	database.DB.Delete(&project)

	c.Redirect(http.StatusFound, "/projects")
}


