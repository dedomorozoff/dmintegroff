package controllers

import (
	"encoding/json"
	"gintegra/internal/database"
	"gintegra/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TestEndpoint - тестовый endpoint для приема запросов
func TestEndpoint(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Сохраняем лог
	payloadJSON, _ := json.Marshal(payload)
	log := models.RequestLog{
		Method:      c.Request.Method,
		URL:         c.Request.URL.Path,
		RequestBody: string(payloadJSON),
		StatusCode:  200,
	}
	database.DB.Create(&log)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Test request received",
		"data":    payload,
	})
}

// LogsPage - страница с логами
func LogsPage(c *gin.Context) {
	var logs []models.RequestLog
	database.DB.Order("created_at desc").Limit(100).Find(&logs)

	c.HTML(http.StatusOK, "logs.html", gin.H{
		"title": "Логи запросов",
		"logs":  logs,
	})
}

// ClearLogs - очистка всех логов
func ClearLogs(c *gin.Context) {
	database.DB.Exec("DELETE FROM request_logs")
	c.Redirect(http.StatusFound, "/logs")
}

// DeleteLog - удаление конкретного лога
func DeleteLog(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	
	database.DB.Delete(&models.RequestLog{}, uint(id))
	c.Redirect(http.StatusFound, "/logs")
}

// LogsAPI - API endpoint для получения логов в JSON
func LogsAPI(c *gin.Context) {
	var logs []models.RequestLog
	database.DB.Order("created_at desc").Limit(100).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}
