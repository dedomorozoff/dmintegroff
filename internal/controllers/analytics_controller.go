package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AnalyticsData представляет данные для аналитики
type AnalyticsData struct {
	// Общая статистика
	TotalRequests     int64                    `json:"total_requests"`
	SuccessfulRequests int64                   `json:"successful_requests"`
	FailedRequests    int64                    `json:"failed_requests"`
	SuccessRate       float64                  `json:"success_rate"`
	
	// Статистика по времени
	RequestsByHour    []HourlyStats            `json:"requests_by_hour"`
	RequestsByDay     []DailyStats             `json:"requests_by_day"`
	RequestsByWeek    []WeeklyStats            `json:"requests_by_week"`
	
	// Статистика по интеграциям
	TopIntegrations   []IntegrationStats       `json:"top_integrations"`
	
	// Статистика по ошибкам
	ErrorsByType      []ErrorStats             `json:"errors_by_type"`
	ErrorsByIntegration []IntegrationErrorStats `json:"errors_by_integration"`
	
	// Производительность
	AverageResponseTime float64                `json:"average_response_time"`
	ResponseTimeByHour  []ResponseTimeStats    `json:"response_time_by_hour"`
	
	// Статистика по HTTP методам
	RequestsByMethod  []MethodStats            `json:"requests_by_method"`
	
	// Статистика по Content-Type
	RequestsByContentType []ContentTypeStats   `json:"requests_by_content_type"`
}

// HourlyStats представляет статистику по часам
type HourlyStats struct {
	Hour     int   `json:"hour"`
	Requests int64 `json:"requests"`
	Errors   int64 `json:"errors"`
}

// DailyStats представляет статистику по дням
type DailyStats struct {
	Date     string `json:"date"`
	Requests int64  `json:"requests"`
	Errors   int64  `json:"errors"`
	Success  int64  `json:"success"`
}

// WeeklyStats представляет статистику по неделям
type WeeklyStats struct {
	Week     string `json:"week"`
	Requests int64  `json:"requests"`
	Errors   int64  `json:"errors"`
}

// IntegrationStats представляет статистику по интеграциям
type IntegrationStats struct {
	IntegrationID   uint   `json:"integration_id"`
	IntegrationName string `json:"integration_name"`
	Requests        int64  `json:"requests"`
	Errors          int64  `json:"errors"`
	SuccessRate     float64 `json:"success_rate"`
	LastRequest     time.Time `json:"last_request"`
}

// ErrorStats представляет статистику по типам ошибок
type ErrorStats struct {
	StatusCode int    `json:"status_code"`
	Count      int64  `json:"count"`
	Percentage float64 `json:"percentage"`
}

// IntegrationErrorStats представляет ошибки по интеграциям
type IntegrationErrorStats struct {
	IntegrationID   uint   `json:"integration_id"`
	IntegrationName string `json:"integration_name"`
	ErrorCount      int64  `json:"error_count"`
	LastError       time.Time `json:"last_error"`
	CommonErrors    []string `json:"common_errors"`
}

// ResponseTimeStats представляет статистику времени ответа
type ResponseTimeStats struct {
	Hour            int     `json:"hour"`
	AverageTime     float64 `json:"average_time"`
	MinTime         float64 `json:"min_time"`
	MaxTime         float64 `json:"max_time"`
}

// MethodStats представляет статистику по HTTP методам
type MethodStats struct {
	Method   string `json:"method"`
	Count    int64  `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ContentTypeStats представляет статистику по Content-Type
type ContentTypeStats struct {
	ContentType string `json:"content_type"`
	Count       int64  `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// GetAnalytics возвращает аналитические данные
func GetAnalytics(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	
	// Получаем параметры запроса
	period := c.DefaultQuery("period", "7d") // 1d, 7d, 30d, 90d
	integrationID := c.Query("integration_id")
	
	// Определяем временной диапазон
	var startTime time.Time
	switch period {
	case "1d":
		startTime = time.Now().AddDate(0, 0, -1)
	case "7d":
		startTime = time.Now().AddDate(0, 0, -7)
	case "30d":
		startTime = time.Now().AddDate(0, 0, -30)
	case "90d":
		startTime = time.Now().AddDate(0, 0, -90)
	default:
		startTime = time.Now().AddDate(0, 0, -7)
	}
	
	analytics := &AnalyticsData{}
	
	// Базовый запрос с учетом прав доступа
	baseQuery := database.DB.Model(&models.RequestLog{}).Where("created_at >= ?", startTime)
	if role != "admin" {
		baseQuery = baseQuery.Joins("JOIN integrations ON integrations.id = request_logs.integration_id").
			Where("integrations.created_by_id = ?", userID)
	}
	
	// Фильтр по интеграции
	if integrationID != "" {
		if id, err := strconv.ParseUint(integrationID, 10, 32); err == nil {
			baseQuery = baseQuery.Where("integration_id = ?", id)
		}
	}
	
	// Общая статистика
	baseQuery.Where("log_type = ?", "webhook").Count(&analytics.TotalRequests)
	baseQuery.Where("log_type = ? AND status_code >= 200 AND status_code < 400", "webhook").Count(&analytics.SuccessfulRequests)
	baseQuery.Where("log_type = ? AND status_code >= 400", "webhook").Count(&analytics.FailedRequests)
	
	if analytics.TotalRequests > 0 {
		analytics.SuccessRate = float64(analytics.SuccessfulRequests) / float64(analytics.TotalRequests) * 100
	}
	
	// Статистика по дням
	analytics.RequestsByDay = getRequestsByDay(baseQuery, startTime, period)
	
	// Статистика по часам (только для периода 1d и 7d)
	if period == "1d" || period == "7d" {
		analytics.RequestsByHour = getRequestsByHour(baseQuery)
	}
	
	// Топ интеграций
	analytics.TopIntegrations = getTopIntegrations(baseQuery, role, userID)
	
	// Статистика по ошибкам
	analytics.ErrorsByType = getErrorsByType(baseQuery)
	analytics.ErrorsByIntegration = getErrorsByIntegration(baseQuery, role, userID)
	
	// Статистика по HTTP методам
	analytics.RequestsByMethod = getRequestsByMethod(baseQuery)
	
	// Статистика по Content-Type (из заголовков)
	analytics.RequestsByContentType = getRequestsByContentType(baseQuery)
	
	c.JSON(http.StatusOK, analytics)
}

// getRequestsByDay возвращает статистику запросов по дням
func getRequestsByDay(baseQuery *gorm.DB, startTime time.Time, period string) []DailyStats {
	var stats []DailyStats
	
	// Определяем формат даты в зависимости от периода
	// dateFormat := "2006-01-02"
	// if period == "90d" {
	// 	dateFormat = "2006-01" // По месяцам для 90 дней
	// }
	
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as requests,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as errors,
			SUM(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 ELSE 0 END) as success
		FROM request_logs 
		WHERE created_at >= ? AND log_type = 'webhook'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`
	
	rows, err := database.DB.Raw(query, startTime).Rows()
	if err != nil {
		return stats
	}
	defer rows.Close()
	
	for rows.Next() {
		var stat DailyStats
		if err := rows.Scan(&stat.Date, &stat.Requests, &stat.Errors, &stat.Success); err != nil {
			continue
		}
		stats = append(stats, stat)
	}
	
	return stats
}

// getRequestsByHour возвращает статистику запросов по часам
func getRequestsByHour(baseQuery *gorm.DB) []HourlyStats {
	var stats []HourlyStats
	
	query := `
		SELECT 
			strftime('%H', created_at) as hour,
			COUNT(*) as requests,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as errors
		FROM request_logs 
		WHERE log_type = 'webhook'
		GROUP BY strftime('%H', created_at)
		ORDER BY hour ASC
	`
	
	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		return stats
	}
	defer rows.Close()
	
	for rows.Next() {
		var stat HourlyStats
		var hourStr string
		if err := rows.Scan(&hourStr, &stat.Requests, &stat.Errors); err != nil {
			continue
		}
		if hour, err := strconv.Atoi(hourStr); err == nil {
			stat.Hour = hour
			stats = append(stats, stat)
		}
	}
	
	return stats
}

// getTopIntegrations возвращает топ интеграций по количеству запросов
func getTopIntegrations(baseQuery *gorm.DB, role interface{}, userID interface{}) []IntegrationStats {
	var stats []IntegrationStats
	
	query := `
		SELECT 
			i.id as integration_id,
			i.name as integration_name,
			COUNT(rl.id) as requests,
			SUM(CASE WHEN rl.status_code >= 400 THEN 1 ELSE 0 END) as errors,
			MAX(rl.created_at) as last_request
		FROM integrations i
		LEFT JOIN request_logs rl ON rl.integration_id = i.id AND rl.log_type = 'webhook'
	`
	
	if role != "admin" {
		query += " WHERE i.created_by_id = ?"
	}
	
	query += `
		GROUP BY i.id, i.name
		ORDER BY requests DESC
		LIMIT 10
	`
	
	var rows interface{ Close() error; Next() bool; Scan(...interface{}) error }
	var err error
	
	if role != "admin" {
		rows, err = database.DB.Raw(query, userID).Rows()
	} else {
		rows, err = database.DB.Raw(query).Rows()
	}
	
	if err != nil {
		return stats
	}
	defer rows.Close()
	
	for rows.Next() {
		var stat IntegrationStats
		if err := rows.Scan(&stat.IntegrationID, &stat.IntegrationName, &stat.Requests, &stat.Errors, &stat.LastRequest); err != nil {
			continue
		}
		
		if stat.Requests > 0 {
			stat.SuccessRate = float64(stat.Requests-stat.Errors) / float64(stat.Requests) * 100
		}
		
		stats = append(stats, stat)
	}
	
	return stats
}

// getErrorsByType возвращает статистику ошибок по типам (HTTP коды)
func getErrorsByType(baseQuery *gorm.DB) []ErrorStats {
	var stats []ErrorStats
	
	query := `
		SELECT 
			status_code,
			COUNT(*) as count
		FROM request_logs 
		WHERE log_type = 'webhook' AND status_code >= 400
		GROUP BY status_code
		ORDER BY count DESC
		LIMIT 10
	`
	
	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		return stats
	}
	defer rows.Close()
	
	var totalErrors int64
	database.DB.Model(&models.RequestLog{}).Where("log_type = ? AND status_code >= 400", "webhook").Count(&totalErrors)
	
	for rows.Next() {
		var stat ErrorStats
		if err := rows.Scan(&stat.StatusCode, &stat.Count); err != nil {
			continue
		}
		
		if totalErrors > 0 {
			stat.Percentage = float64(stat.Count) / float64(totalErrors) * 100
		}
		
		stats = append(stats, stat)
	}
	
	return stats
}

// getErrorsByIntegration возвращает статистику ошибок по интеграциям
func getErrorsByIntegration(baseQuery *gorm.DB, role interface{}, userID interface{}) []IntegrationErrorStats {
	var stats []IntegrationErrorStats
	
	query := `
		SELECT 
			i.id as integration_id,
			i.name as integration_name,
			COUNT(rl.id) as error_count,
			MAX(rl.created_at) as last_error
		FROM integrations i
		LEFT JOIN request_logs rl ON rl.integration_id = i.id AND rl.log_type = 'webhook' AND rl.status_code >= 400
	`
	
	if role != "admin" {
		query += " WHERE i.created_by_id = ?"
	}
	
	query += `
		GROUP BY i.id, i.name
		HAVING error_count > 0
		ORDER BY error_count DESC
		LIMIT 10
	`
	
	var rows interface{ Close() error; Next() bool; Scan(...interface{}) error }
	var err error
	
	if role != "admin" {
		rows, err = database.DB.Raw(query, userID).Rows()
	} else {
		rows, err = database.DB.Raw(query).Rows()
	}
	
	if err != nil {
		return stats
	}
	defer rows.Close()
	
	for rows.Next() {
		var stat IntegrationErrorStats
		if err := rows.Scan(&stat.IntegrationID, &stat.IntegrationName, &stat.ErrorCount, &stat.LastError); err != nil {
			continue
		}
		stats = append(stats, stat)
	}
	
	return stats
}

// getRequestsByMethod возвращает статистику по HTTP методам
func getRequestsByMethod(baseQuery *gorm.DB) []MethodStats {
	var stats []MethodStats
	
	query := `
		SELECT 
			method,
			COUNT(*) as count
		FROM request_logs 
		WHERE log_type = 'webhook'
		GROUP BY method
		ORDER BY count DESC
	`
	
	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		return stats
	}
	defer rows.Close()
	
	var totalRequests int64
	database.DB.Model(&models.RequestLog{}).Where("log_type = ?", "webhook").Count(&totalRequests)
	
	for rows.Next() {
		var stat MethodStats
		if err := rows.Scan(&stat.Method, &stat.Count); err != nil {
			continue
		}
		
		if totalRequests > 0 {
			stat.Percentage = float64(stat.Count) / float64(totalRequests) * 100
		}
		
		stats = append(stats, stat)
	}
	
	return stats
}

// getRequestsByContentType возвращает статистику по Content-Type
func getRequestsByContentType(baseQuery *gorm.DB) []ContentTypeStats {
	var stats []ContentTypeStats
	
	// Извлекаем Content-Type из JSON заголовков
	query := `
		SELECT 
			CASE 
				WHEN request_headers LIKE '%"Content-Type":["%application/json%"]%' THEN 'application/json'
				WHEN request_headers LIKE '%"Content-Type":["%application/x-www-form-urlencoded%"]%' THEN 'application/x-www-form-urlencoded'
				WHEN request_headers LIKE '%"Content-Type":["%multipart/form-data%"]%' THEN 'multipart/form-data'
				WHEN request_headers LIKE '%"Content-Type":["%application/xml%"]%' THEN 'application/xml'
				WHEN request_headers LIKE '%"Content-Type":["%text/xml%"]%' THEN 'text/xml'
				WHEN request_headers LIKE '%"Content-Type":["%text/plain%"]%' THEN 'text/plain'
				ELSE 'other'
			END as content_type,
			COUNT(*) as count
		FROM request_logs 
		WHERE log_type = 'webhook' AND request_headers IS NOT NULL
		GROUP BY content_type
		ORDER BY count DESC
	`
	
	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		return stats
	}
	defer rows.Close()
	
	var totalRequests int64
	database.DB.Model(&models.RequestLog{}).Where("log_type = ? AND request_headers IS NOT NULL", "webhook").Count(&totalRequests)
	
	for rows.Next() {
		var stat ContentTypeStats
		if err := rows.Scan(&stat.ContentType, &stat.Count); err != nil {
			continue
		}
		
		if totalRequests > 0 {
			stat.Percentage = float64(stat.Count) / float64(totalRequests) * 100
		}
		
		stats = append(stats, stat)
	}
	
	return stats
}

// AnalyticsPage отображает страницу аналитики
func AnalyticsPage(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	
	c.HTML(http.StatusOK, "pages/analytics.html", gin.H{
		"title":       "Аналитика",
		"username":    username,
		"role":        role,
		"CurrentPage": "analytics",
	})
}