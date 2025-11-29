package controllers

import (
	"bufio"
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"encoding/json"
	"net/http"
	"os"
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

	// Сохраняем лог с заголовками
	payloadJSON, _ := json.Marshal(payload)
	headersJSON, _ := json.Marshal(c.Request.Header)

	log := models.RequestLog{
		Method:         c.Request.Method,
		URL:            c.Request.URL.Path,
		RequestBody:    string(payloadJSON),
		RequestHeaders: string(headersJSON),
		StatusCode:     200,
		LogType:        "request",
	}
	CreateLogWithLimit(&log)

	// Обновляем все интеграции в режиме прослушивания, которые еще не имеют данных
	var integrations []models.Integration
	database.DB.Where("mode = ? AND (sample_payload = '' OR sample_payload IS NULL)", "listening").Find(&integrations)

	for _, integration := range integrations {
		integration.SamplePayload = string(payloadJSON)
		database.DB.Save(&integration)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":               "success",
		"message":              "Test request received",
		"data":                 payload,
		"updated_integrations": len(integrations),
	})
}

// LogsPage - страница с логами
func LogsPage(c *gin.Context) {
	var logs []models.RequestLog
	database.DB.Order("created_at desc").Limit(100).Find(&logs)

	c.HTML(http.StatusOK, "pages/logs.html", gin.H{
		"title":       "Логи запросов",
		"CurrentPage": "logs",
		"logs":        logs,
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

// LogError - логирование ошибок сервера
func LogError(method, url, errorMsg string, statusCode int) {
	log := models.RequestLog{
		Method:       method,
		URL:          url,
		ErrorMessage: errorMsg,
		StatusCode:   statusCode,
		LogType:      "error",
	}
	CreateLogWithLimit(&log)
}

// CreateLogWithLimit - создает лог и удаляет старые, если их больше 50
func CreateLogWithLimit(log *models.RequestLog) {
	database.DB.Create(log)

	// Подсчитываем количество логов
	var count int64
	database.DB.Model(&models.RequestLog{}).Count(&count)

	// Если больше 50, удаляем самые старые
	if count > 50 {
		database.DB.Exec("DELETE FROM request_logs WHERE id IN (SELECT id FROM request_logs ORDER BY created_at ASC LIMIT ?)", count-50)
	}
}

// ServerLogsAPI - API endpoint для получения логов из файла dmIntegroff.log
func ServerLogsAPI(c *gin.Context) {
	// Читаем последние 100 строк из лог-файла
	logFile := "dmIntegroff.log"

	file, err := os.Open(logFile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"logs": []string{},
		})
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Берем последние 100 строк
	start := 0
	if len(lines) > 100 {
		start = len(lines) - 100
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": lines[start:],
	})
}
