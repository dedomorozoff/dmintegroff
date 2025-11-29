package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ActivityItem представляет элемент активности для отображения
type ActivityItem struct {
	Status         string    `json:"status"`          // "success", "error", "warning"
	IntegrationName string   `json:"integration_name"`
	Message        string    `json:"message"`
	RequestURL     string    `json:"request_url"`
	Time           time.Time `json:"time"`
	TimeAgo        string    `json:"time_ago"`
}

// DashboardPage - главная страница с панелью управления
func DashboardPage(c *gin.Context) {
	session := sessions.Default(c)
	role := session.Get("role")

	// Получаем статистику интеграций
	var totalIntegrations int64
	var activeIntegrations int64
	database.DB.Model(&models.Integration{}).Count(&totalIntegrations)
	database.DB.Model(&models.Integration{}).Where("status = ?", "active").Count(&activeIntegrations)

	// Получаем последние логи для активности
	var logs []models.RequestLog
	database.DB.Order("created_at desc").Limit(10).Find(&logs)

	// Преобразуем логи в элементы активности
	activities := make([]ActivityItem, 0)
	for _, log := range logs {
		activity := ActivityItem{
			Time:    log.CreatedAt,
			TimeAgo: formatTimeAgo(log.CreatedAt),
			RequestURL: log.URL,
		}

		// Получаем имя интеграции, если есть IntegrationID
		if log.IntegrationID > 0 {
			var integration models.Integration
			if err := database.DB.First(&integration, log.IntegrationID).Error; err == nil {
				activity.IntegrationName = integration.Name
			}
		}

		// Определяем статус и сообщение на основе типа лога и кода ответа
		switch log.LogType {
		case "error":
			activity.Status = "error"
			if log.ErrorMessage != "" {
				activity.Message = log.ErrorMessage
			} else {
				activity.Message = "Произошла ошибка при обработке запроса"
			}
		case "webhook":
			if log.StatusCode >= 200 && log.StatusCode < 300 {
				activity.Status = "success"
				activity.Message = "Webhook успешно обработан"
			} else if log.StatusCode >= 400 && log.StatusCode < 500 {
				activity.Status = "warning"
				activity.Message = "Ошибка клиента при обработке webhook"
			} else if log.StatusCode >= 500 {
				activity.Status = "error"
				activity.Message = "Ошибка сервера при обработке webhook"
			} else {
				activity.Status = "success"
				activity.Message = "Webhook обработан"
			}
		case "test":
			if log.StatusCode >= 200 && log.StatusCode < 300 {
				activity.Status = "success"
				activity.Message = "Тестовый запрос успешно обработан"
			} else if log.StatusCode >= 400 && log.StatusCode < 500 {
				activity.Status = "warning"
				activity.Message = "Ошибка в тестовом запросе"
			} else if log.StatusCode >= 500 {
				activity.Status = "error"
				activity.Message = "Ошибка сервера при тестовом запросе"
			} else {
				activity.Status = "success"
				activity.Message = "Тестовый запрос обработан"
			}
		default:
			activity.Status = "success"
			activity.Message = "Операция выполнена"
		}

		// Если имя интеграции не найдено, пытаемся извлечь из URL или тела запроса
		if activity.IntegrationName == "" {
			if log.URL != "" {
				activity.IntegrationName = "Webhook"
			} else {
				activity.IntegrationName = "Система"
			}
		}

		activities = append(activities, activity)
	}

	// Подсчитываем общее количество обработанных запросов
	var totalRequests int64
	database.DB.Model(&models.RequestLog{}).Where("log_type = ?", "webhook").Count(&totalRequests)

	// Определяем прогресс пользователя
	progress := calculateProgress()
	
	// Подсчитываем количество завершенных шагов
	completedSteps := 0
	for _, step := range progress {
		if step.Completed {
			completedSteps++
		}
	}
	progressPercentage := (completedSteps * 100) / 5

	c.HTML(http.StatusOK, "pages/dashboard.html", gin.H{
		"title":               "Главная",
		"role":                role,
		"CurrentPage":         "dashboard",
		"total_integrations":  totalIntegrations,
		"active_integrations": activeIntegrations,
		"total_requests":      totalRequests,
		"activities":          activities,
		"progress":            progress,
		"progress_completed":  completedSteps,
		"progress_percentage": progressPercentage,
	})
}

// ProgressStep представляет шаг в прогрессе пользователя
type ProgressStep struct {
	Completed bool   `json:"completed"`
	Text      string `json:"text"`
}

// calculateProgress определяет прогресс пользователя
func calculateProgress() map[string]ProgressStep {
	progress := make(map[string]ProgressStep)

	// Шаг 1: Создан ли хотя бы один проект
	var projectCount int64
	database.DB.Model(&models.Project{}).Count(&projectCount)
	progress["step1"] = ProgressStep{
		Completed: projectCount > 0,
		Text:      "Создайте проект для группировки интеграций",
	}

	// Шаг 2: Создана ли хотя бы одна интеграция
	var integrationCount int64
	database.DB.Model(&models.Integration{}).Count(&integrationCount)
	progress["step2"] = ProgressStep{
		Completed: integrationCount > 0,
		Text:      "Создайте интеграцию в проекте",
	}

	// Шаг 3: Получен ли хотя бы один webhook запрос
	var webhookCount int64
	database.DB.Model(&models.RequestLog{}).Where("log_type = ?", "webhook").Count(&webhookCount)
	progress["step3"] = ProgressStep{
		Completed: webhookCount > 0,
		Text:      "Отправьте тестовый запрос на webhook URL",
	}

	// Шаг 4: Настроен ли маппинг хотя бы в одной интеграции
	var mappedIntegrationCount int64
	database.DB.Model(&models.Integration{}).Where("mapping_config != '' AND mapping_config IS NOT NULL").Count(&mappedIntegrationCount)
	progress["step4"] = ProgressStep{
		Completed: mappedIntegrationCount > 0,
		Text:      "Настройте маппинг полей",
	}

	// Шаг 5: Активирована ли хотя бы одна интеграция
	var activeCount int64
	database.DB.Model(&models.Integration{}).Where("mode = ?", "active").Count(&activeCount)
	progress["step5"] = ProgressStep{
		Completed: activeCount > 0,
		Text:      "Активируйте интеграцию",
	}

	return progress
}

// formatTimeAgo форматирует время в человекочитаемый формат "X минут назад"
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration.Seconds() < 60 {
		seconds := int(duration.Seconds())
		if seconds <= 1 {
			return "только что"
		}
		return formatPlural(seconds, "секунду", "секунды", "секунд") + " назад"
	}

	if duration.Minutes() < 60 {
		minutes := int(duration.Minutes())
		return formatPlural(minutes, "минуту", "минуты", "минут") + " назад"
	}

	if duration.Hours() < 24 {
		hours := int(duration.Hours())
		return formatPlural(hours, "час", "часа", "часов") + " назад"
	}

	days := int(duration.Hours() / 24)
	if days < 30 {
		return formatPlural(days, "день", "дня", "дней") + " назад"
	}

	months := days / 30
	if months < 12 {
		return formatPlural(months, "месяц", "месяца", "месяцев") + " назад"
	}

	years := months / 12
	return formatPlural(years, "год", "года", "лет") + " назад"
}

// formatPlural форматирует число с правильным окончанием для русского языка
func formatPlural(n int, one, few, many string) string {
	nStr := formatInt(n)
	if n%10 == 1 && n%100 != 11 {
		return nStr + " " + one
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return nStr + " " + few
	}
	return nStr + " " + many
}

// formatInt преобразует int в string
func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}

// GetRequestStats - API endpoint для получения статистики запросов по дням
func GetRequestStats(c *gin.Context) {
	// Получаем статистику за последние 30 дней
	type DayStats struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	var stats []DayStats
	
	// SQL запрос для группировки по дням
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as count
		FROM request_logs
		WHERE log_type = 'webhook'
		AND created_at >= datetime('now', '-30 days')
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`
	
	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stat DayStats
		if err := rows.Scan(&stat.Date, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	// Если нет данных, возвращаем пустой массив
	if stats == nil {
		stats = []DayStats{}
	}

	// Логируем для отладки
	fmt.Printf("Stats API: found %d records\n", len(stats))

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetRecentActivity - API endpoint для получения последней активности
func GetRecentActivity(c *gin.Context) {
	var logs []models.RequestLog
	database.DB.Order("created_at desc").Limit(10).Find(&logs)

	activities := make([]ActivityItem, 0)
	for _, log := range logs {
		activity := ActivityItem{
			Time:    log.CreatedAt,
			TimeAgo: formatTimeAgo(log.CreatedAt),
			RequestURL: log.URL,
		}

		if log.IntegrationID > 0 {
			var integration models.Integration
			if err := database.DB.First(&integration, log.IntegrationID).Error; err == nil {
				activity.IntegrationName = integration.Name
			}
		}

		switch log.LogType {
		case "error":
			activity.Status = "error"
			if log.ErrorMessage != "" {
				activity.Message = log.ErrorMessage
			} else {
				activity.Message = "Произошла ошибка при обработке запроса"
			}
		case "webhook":
			if log.StatusCode >= 200 && log.StatusCode < 300 {
				activity.Status = "success"
				activity.Message = "Webhook успешно обработан"
			} else if log.StatusCode >= 400 && log.StatusCode < 500 {
				activity.Status = "warning"
				activity.Message = "Ошибка клиента при обработке webhook"
			} else if log.StatusCode >= 500 {
				activity.Status = "error"
				activity.Message = "Ошибка сервера при обработке webhook"
			}
		case "test":
			if log.StatusCode >= 200 && log.StatusCode < 300 {
				activity.Status = "success"
				activity.Message = "Тестовый запрос успешно обработан"
			} else {
				activity.Status = "error"
				activity.Message = "Ошибка при обработке тестового запроса"
			}
		}

		if activity.IntegrationName == "" {
			activity.IntegrationName = "Система"
		}

		activities = append(activities, activity)
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
	})
}
