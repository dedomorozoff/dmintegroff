package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequestHistoryData представляет данные для истории запросов
type RequestHistoryData struct {
	Requests    []RequestHistoryItem `json:"requests"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PerPage     int                  `json:"per_page"`
	TotalPages  int                  `json:"total_pages"`
	HasNext     bool                 `json:"has_next"`
	HasPrev     bool                 `json:"has_prev"`
	Filters     RequestFilters       `json:"filters"`
}

// RequestHistoryItem представляет элемент истории запросов
type RequestHistoryItem struct {
	ID               uint      `json:"id"`
	IntegrationID    uint      `json:"integration_id"`
	IntegrationName  string    `json:"integration_name"`
	Method           string    `json:"method"`
	URL              string    `json:"url"`
	StatusCode       int       `json:"status_code"`
	LogType          string    `json:"log_type"`
	ErrorMessage     string    `json:"error_message"`
	OutputName       string    `json:"output_name"`
	CreatedAt        time.Time `json:"created_at"`
	ResponseTime     int64     `json:"response_time,omitempty"` // в миллисекундах
	RequestSize      int       `json:"request_size"`
	ResponseSize     int       `json:"response_size"`
	HasRequestBody   bool      `json:"has_request_body"`
	HasResponseBody  bool      `json:"has_response_body"`
	HasRequestHeaders bool     `json:"has_request_headers"`
}

// RequestFilters представляет фильтры для поиска
type RequestFilters struct {
	IntegrationID string `json:"integration_id"`
	Status        string `json:"status"`        // success, error, all
	Method        string `json:"method"`        // GET, POST, PUT, DELETE, all
	LogType       string `json:"log_type"`      // webhook, request, error, all
	DateFrom      string `json:"date_from"`     // YYYY-MM-DD
	DateTo        string `json:"date_to"`       // YYYY-MM-DD
	Search        string `json:"search"`        // поиск по URL, error_message
	OutputName    string `json:"output_name"`   // фильтр по названию выхода
}

// RequestHistoryPage отображает страницу истории запросов
func RequestHistoryPage(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	userID := session.Get("user_id")

	// Получаем список интеграций для фильтра
	var integrations []models.Integration
	query := database.DB.Model(&models.Integration{})
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	query.Find(&integrations)

	// Генерируем breadcrumbs
	breadcrumbs := utils.NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddActiveWithIcon("История запросов", "history").
		Build()

	c.HTML(http.StatusOK, "pages/request_history.html", gin.H{
		"title":        "История запросов",
		"username":     username,
		"role":         role,
		"CurrentPage":  "request_history",
		"integrations": integrations,
		"breadcrumbs":  breadcrumbs,
	})
}

// GetRequestHistory возвращает историю запросов с фильтрацией и пагинацией
func GetRequestHistory(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 10 || perPage > 1000 {
		perPage = 50
	}

	// Фильтры
	filters := RequestFilters{
		IntegrationID: c.Query("integration_id"),
		Status:        c.DefaultQuery("status", "all"),
		Method:        c.DefaultQuery("method", "all"),
		LogType:       c.DefaultQuery("log_type", "all"),
		DateFrom:      c.Query("date_from"),
		DateTo:        c.Query("date_to"),
		Search:        c.Query("search"),
		OutputName:    c.Query("output_name"),
	}

	// Базовый запрос
	query := database.DB.Model(&models.RequestLog{}).
		Preload("Integration").
		Order("created_at DESC")

	// Права доступа
	if role != "admin" {
		query = query.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.created_by_id = ?", userID)
	}

	// Применяем фильтры
	query = applyRequestFilters(query, filters)

	// Подсчет общего количества
	var total int64
	query.Count(&total)

	// Пагинация
	offset := (page - 1) * perPage
	var requestLogs []models.RequestLog
	query.Offset(offset).Limit(perPage).Find(&requestLogs)

	// Преобразуем в формат для фронтенда
	var items []RequestHistoryItem
	for _, log := range requestLogs {
		item := RequestHistoryItem{
			ID:               log.ID,
			IntegrationID:    log.IntegrationID,
			Method:           log.Method,
			URL:              log.URL,
			StatusCode:       log.StatusCode,
			LogType:          log.LogType,
			ErrorMessage:     log.ErrorMessage,
			OutputName:       log.OutputName,
			CreatedAt:        log.CreatedAt,
			RequestSize:      len(log.RequestBody),
			ResponseSize:     len(log.ResponseBody),
			HasRequestBody:   log.RequestBody != "",
			HasResponseBody:  log.ResponseBody != "",
			HasRequestHeaders: log.RequestHeaders != "",
		}

		// Добавляем название интеграции
		if log.Integration != nil {
			item.IntegrationName = log.Integration.Name
		}

		items = append(items, item)
	}

	// Расчет пагинации
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	hasNext := page < totalPages
	hasPrev := page > 1

	response := RequestHistoryData{
		Requests:   items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Filters:    filters,
	}

	c.JSON(http.StatusOK, response)
}

// GetRequestDetails возвращает детальную информацию о запросе
func GetRequestDetails(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID запроса не указан"})
		return
	}

	var requestLog models.RequestLog
	query := database.DB.Preload("Integration")

	// Права доступа
	if role != "admin" {
		query = query.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.created_by_id = ?", userID)
	}

	if err := query.First(&requestLog, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Запрос не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных"})
		}
		return
	}

	// Парсим заголовки из JSON
	var requestHeaders map[string][]string
	var responseHeaders map[string][]string

	if requestLog.RequestHeaders != "" {
		json.Unmarshal([]byte(requestLog.RequestHeaders), &requestHeaders)
	}

	// Формируем детальный ответ
	response := gin.H{
		"id":               requestLog.ID,
		"integration_id":   requestLog.IntegrationID,
		"integration_name": "",
		"method":           requestLog.Method,
		"url":              requestLog.URL,
		"status_code":      requestLog.StatusCode,
		"log_type":         requestLog.LogType,
		"error_message":    requestLog.ErrorMessage,
		"output_name":      requestLog.OutputName,
		"created_at":       requestLog.CreatedAt,
		"request_body":     requestLog.RequestBody,
		"response_body":    requestLog.ResponseBody,
		"request_headers":  requestHeaders,
		"response_headers": responseHeaders,
	}

	if requestLog.Integration != nil {
		response["integration_name"] = requestLog.Integration.Name
	}

	c.JSON(http.StatusOK, response)
}

// ExportRequestHistory экспортирует историю запросов в CSV
func ExportRequestHistory(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")

	// Получаем те же фильтры, что и для обычного запроса
	filters := RequestFilters{
		IntegrationID: c.Query("integration_id"),
		Status:        c.DefaultQuery("status", "all"),
		Method:        c.DefaultQuery("method", "all"),
		LogType:       c.DefaultQuery("log_type", "all"),
		DateFrom:      c.Query("date_from"),
		DateTo:        c.Query("date_to"),
		Search:        c.Query("search"),
		OutputName:    c.Query("output_name"),
	}

	// Базовый запрос
	query := database.DB.Model(&models.RequestLog{}).
		Preload("Integration").
		Order("created_at DESC")

	// Права доступа
	if role != "admin" {
		query = query.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.created_by_id = ?", userID)
	}

	// Применяем фильтры
	query = applyRequestFilters(query, filters)

	// Ограничиваем экспорт (максимум 10000 записей)
	var requestLogs []models.RequestLog
	query.Limit(10000).Find(&requestLogs)

	// Устанавливаем заголовки для скачивания файла
	filename := fmt.Sprintf("request_history_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Создаем CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Заголовки CSV
	headers := []string{
		"ID", "Дата/Время", "Интеграция", "Метод", "URL", "Статус", 
		"Тип лога", "Название выхода", "Ошибка", "Размер запроса", "Размер ответа",
	}
	writer.Write(headers)

	// Записываем данные
	for _, log := range requestLogs {
		integrationName := ""
		if log.Integration != nil {
			integrationName = log.Integration.Name
		}

		record := []string{
			strconv.FormatUint(uint64(log.ID), 10),
			log.CreatedAt.Format("2006-01-02 15:04:05"),
			integrationName,
			log.Method,
			log.URL,
			strconv.Itoa(log.StatusCode),
			log.LogType,
			log.OutputName,
			log.ErrorMessage,
			strconv.Itoa(len(log.RequestBody)),
			strconv.Itoa(len(log.ResponseBody)),
		}
		writer.Write(record)
	}
}

// ClearRequestHistory очищает историю запросов (только для админов)
func ClearRequestHistory(c *gin.Context) {
	session := sessions.Default(c)
	role := session.Get("role")

	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
		return
	}

	// Получаем параметры очистки
	olderThan := c.DefaultQuery("older_than", "30") // дни
	days, err := strconv.Atoi(olderThan)
	if err != nil || days < 1 {
		days = 30
	}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Удаляем старые записи
	result := database.DB.Where("created_at < ?", cutoffDate).Delete(&models.RequestLog{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка очистки истории"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Удалено %d записей старше %d дней", result.RowsAffected, days),
		"deleted": result.RowsAffected,
	})
}

// applyRequestFilters применяет фильтры к запросу
func applyRequestFilters(query *gorm.DB, filters RequestFilters) *gorm.DB {
	// Фильтр по интеграции
	if filters.IntegrationID != "" && filters.IntegrationID != "all" {
		if id, err := strconv.ParseUint(filters.IntegrationID, 10, 32); err == nil {
			query = query.Where("integration_id = ?", id)
		}
	}

	// Фильтр по статусу
	switch filters.Status {
	case "success":
		query = query.Where("status_code >= 200 AND status_code < 400")
	case "error":
		query = query.Where("status_code >= 400")
	}

	// Фильтр по методу
	if filters.Method != "" && filters.Method != "all" {
		query = query.Where("method = ?", strings.ToUpper(filters.Method))
	}

	// Фильтр по типу лога
	if filters.LogType != "" && filters.LogType != "all" {
		query = query.Where("log_type = ?", filters.LogType)
	}

	// Фильтр по дате от
	if filters.DateFrom != "" {
		if dateFrom, err := time.Parse("2006-01-02", filters.DateFrom); err == nil {
			query = query.Where("created_at >= ?", dateFrom)
		}
	}

	// Фильтр по дате до
	if filters.DateTo != "" {
		if dateTo, err := time.Parse("2006-01-02", filters.DateTo); err == nil {
			// Добавляем 23:59:59 к дате окончания
			dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("created_at <= ?", dateTo)
		}
	}

	// Поиск по URL и сообщению об ошибке
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("url LIKE ? OR error_message LIKE ?", searchTerm, searchTerm)
	}

	// Фильтр по названию выхода
	if filters.OutputName != "" && filters.OutputName != "all" {
		query = query.Where("output_name = ?", filters.OutputName)
	}

	return query
}