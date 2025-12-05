package controllers

import (
	"bufio"
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// TestEndpoint - тестовый endpoint для приема запросов
func TestEndpoint(c *gin.Context) {
	// Читаем тело запроса как байты (поддержка любых форматов)
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Пытаемся распарсить как JSON для обратной совместимости
	var payload map[string]interface{}
	var payloadJSON []byte
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		// Если не JSON, сохраняем как есть
		payloadJSON = bodyBytes
		payload = map[string]interface{}{
			"_raw_body": string(bodyBytes),
		}
	} else {
		payloadJSON = bodyBytes
	}

	// Сохраняем лог с заголовками
	headersJSON, _ := json.Marshal(c.Request.Header)

	log := models.RequestLog{
		Method:         c.Request.Method,
		URL:            c.Request.URL.Path,
		RequestBody:    string(bodyBytes), // Сохраняем оригинальное тело
		RequestHeaders: string(headersJSON),
		StatusCode:     200,
		LogType:        "test",
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
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	query := database.DB.Preload("Integration").Preload("Integration.Project").Order("request_logs.created_at desc")

	// Specialist видит только логи своих интеграций
	if role != "admin" {
		query = query.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.created_by_id = ?", userID)
	}

	// Фильтры
	if integrationID := c.Query("integration_id"); integrationID != "" {
		query = query.Where("request_logs.integration_id = ?", integrationID)
	}
	if projectID := c.Query("project_id"); projectID != "" {
		if role != "admin" {
			query = query.Joins("JOIN integrations i2 ON i2.id = request_logs.integration_id").
				Where("i2.project_id = ? AND i2.created_by_id = ?", projectID, userID)
		} else {
			query = query.Joins("JOIN integrations i2 ON i2.id = request_logs.integration_id").
				Where("i2.project_id = ?", projectID)
		}
	}
	if search := c.Query("search"); search != "" {
		if role != "admin" {
			query = query.Joins("JOIN integrations i3 ON i3.id = request_logs.integration_id").
				Where("(i3.name LIKE ? OR request_logs.id = ?) AND i3.created_by_id = ?", "%"+search+"%", search, userID)
		} else {
			query = query.Joins("JOIN integrations i3 ON i3.id = request_logs.integration_id").
				Where("i3.name LIKE ? OR request_logs.id = ?", "%"+search+"%", search)
		}
	}
	if status := c.Query("status"); status != "" {
		if status == "error" {
			query = query.Where("error_message != '' AND error_message IS NOT NULL")
		} else if status == "success" {
			query = query.Where("error_message = '' OR error_message IS NULL")
		}
	}

	var logs []models.RequestLog
	query.Limit(100).Find(&logs)

	// Получаем список проектов для фильтра (только свои для specialist)
	var projects []models.Project
	projectQuery := database.DB.Order("name")
	if role != "admin" {
		projectQuery = projectQuery.Where("created_by_id = ?", userID)
	}
	projectQuery.Find(&projects)

	c.HTML(http.StatusOK, "pages/logs.html", gin.H{
		"title":              "Тесты интеграций",
		"CurrentPage":        "logs",
		"logs":               logs,
		"projects":           projects,
		"filter_integration": c.Query("integration_id"),
		"filter_project":     c.Query("project_id"),
		"filter_search":      c.Query("search"),
		"filter_status":      c.Query("status"),
		"username":           session.Get("username"),
		"role":               role,
	})
}

// ClearLogs - очистка всех логов
func ClearLogs(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	// Admin может удалить все логи, specialist только свои
	if role == "admin" {
		database.DB.Exec("DELETE FROM request_logs")
	} else {
		database.DB.Exec("DELETE FROM request_logs WHERE integration_id IN (SELECT id FROM integrations WHERE created_by_id = ?)", userID)
	}
	c.Redirect(http.StatusFound, "/logs")
}

// DeleteLog - удаление конкретного лога
func DeleteLog(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// Проверяем доступ к логу
	var log models.RequestLog
	query := database.DB.Preload("Integration")
	if err := query.First(&log, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Лог не найден"})
		return
	}

	// Specialist может удалять только логи своих интеграций
	if role != "admin" && log.Integration.CreatedByID != userID.(uint) {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Лог не найден"})
		return
	}

	database.DB.Delete(&log)
	c.Redirect(http.StatusFound, "/logs")
}

// LogsAPI - API endpoint для получения логов в JSON
func LogsAPI(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	query := database.DB.Order("request_logs.created_at desc").Limit(100)
	
	// Specialist видит только логи своих интеграций
	if role != "admin" {
		query = query.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.created_by_id = ?", userID)
	}
	
	var logs []models.RequestLog
	query.Find(&logs)

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
