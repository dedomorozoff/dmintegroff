package controllers

import (
	"bufio"
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
			Where("integrations.created_by_id = ? AND integrations.hide_in_logs = ?", userID, false)
	} else {
		// Admin тоже не видит скрытые логи
		query = query.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.hide_in_logs = ?", false)
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
	
	// Получаем параметры
	limitStr := c.DefaultQuery("limit", "100")
	logType := c.Query("type")
	
	limit, _ := strconv.Atoi(limitStr)
	if limit > 100 {
		limit = 100
	}
	
	query := database.DB.Order("request_logs.created_at desc").Limit(limit)
	
	// Фильтр по типу лога
	if logType != "" {
		query = query.Where("log_type = ?", logType)
	}
	
	// Specialist видит только логи своих интеграций (кроме тестовых)
	if role != "admin" && logType != "test" {
		query = query.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.created_by_id = ?", userID)
	}
	
	var logs []models.RequestLog
	query.Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}

// WebhookTestSimplePage - страница для тестирования произвольных вебхуков
func WebhookTestSimplePage(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	// Получаем базовый URL
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	testURL := fmt.Sprintf("%s/webhook/test", baseURL)

	c.HTML(http.StatusOK, "pages/webhook_test_simple.html", gin.H{
		"title":        "Отправка запросов",
		"username":     username,
		"role":         role,
		"CurrentPage":  "webhook-test-simple",
		"test_url":     testURL,
		"current_time": time.Now().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// HTTPProxyRequest - прокси для HTTP запросов (обход CORS)
func HTTPProxyRequest(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var requestData struct {
		URL     string            `json:"url" binding:"required"`
		Method  string            `json:"method" binding:"required"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Создаем запрос
	var bodyReader io.Reader
	if requestData.Body != "" {
		bodyReader = strings.NewReader(requestData.Body)
	}

	req, err := http.NewRequest(requestData.Method, requestData.URL, bodyReader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL or method"})
		return
	}

	// Добавляем заголовки
	for key, value := range requestData.Headers {
		req.Header.Set(key, value)
	}

	// Добавляем User-Agent если не указан
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "dmIntegroff-Proxy/1.0")
	}

	// Выполняем запрос
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":    "Request failed",
			"details":  err.Error(),
			"duration": duration.Milliseconds(),
		})
		return
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":    "Failed to read response",
			"details":  err.Error(),
			"duration": duration.Milliseconds(),
		})
		return
	}

	// Собираем заголовки ответа
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, gin.H{
		"status":           resp.StatusCode,
		"statusText":       resp.Status,
		"headers":          responseHeaders,
		"body":             string(responseBody),
		"duration":         duration.Milliseconds(),
		"contentLength":    len(responseBody),
		"url":              requestData.URL,
		"method":           requestData.Method,
		"success":          resp.StatusCode >= 200 && resp.StatusCode < 400,
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
