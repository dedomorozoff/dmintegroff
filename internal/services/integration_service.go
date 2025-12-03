package services

import (
	"bytes"
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

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

// ValidateDemoMode - проверяет ограничения демо-режима
func ValidateDemoMode(integration *models.Integration) error {
	demoMode := os.Getenv("DEMO_MODE") == "true"
	if !demoMode {
		return nil // Демо-режим выключен, ограничений нет
	}

	// Загружаем пользователя, создавшего интеграцию
	var user models.User
	if err := database.DB.First(&user, integration.CreatedByID).Error; err != nil {
		return err
	}

	// Если пользователь не демо, ограничений нет
	if !user.IsDemo {
		return nil
	}

	// Для демо-пользователей проверяем target URL
	demoTargetURL := os.Getenv("DEMO_TARGET_URL")
	if demoTargetURL == "" {
		return errors.New("demo mode enabled but DEMO_TARGET_URL not configured")
	}

	// Проверяем, что target API начинается с разрешенного URL
	if !strings.HasPrefix(integration.TargetAPI, demoTargetURL) {
		return errors.New("demo users can only send webhooks to demo target URL")
	}

	return nil
}

func ProcessWebhook(integrationID uint, payload map[string]interface{}) error {
	var integration models.Integration
	if err := database.DB.Preload("CreatedBy").First(&integration, integrationID).Error; err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"error":          err.Error(),
		}).Error("Failed to find integration")
		return err
	}

	// Проверяем демо-режим
	if err := ValidateDemoMode(&integration); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"error":          err.Error(),
		}).Warn("Demo mode validation failed")
		return err
	}

	var transformed map[string]interface{}

	// Приоритет 1: Используем OutputTemplate, если он задан
	if integration.OutputTemplate != "" {
		processor := utils.NewTemplateProcessor()
		var err error
		transformed, err = processor.ProcessTemplate(integration.OutputTemplate, payload)
		if err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"integration_id": integrationID,
				"error":          err.Error(),
			}).Error("Failed to process output template")
			return err
		}
	} else if integration.MappingConfig != "" {
		// Приоритет 2: Используем MappingConfig (старый способ)
		var mapping map[string]string
		if err := json.Unmarshal([]byte(integration.MappingConfig), &mapping); err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"integration_id": integrationID,
				"error":          err.Error(),
			}).Warn("Invalid mapping config, using passthrough")
			transformed = payload
		} else {
			// Transform data using mapping
			transformed = make(map[string]interface{})
			for targetField, sourceField := range mapping {
				// Поддерживаем как простые поля, так и вложенные пути
				var val interface{}
				var found bool

				// Сначала пробуем как простое поле (для обратной совместимости)
				if v, ok := payload[sourceField]; ok {
					val = v
					found = true
				} else {
					// Пробуем извлечь по пути (для вложенных полей)
					if v, err := utils.GetValueByPath(payload, sourceField); err == nil {
						val = v
						found = true
					}
				}

				if found {
					transformed[targetField] = val
				}
			}
		}
	} else {
		// Приоритет 3: Если ничего не задано, передаем данные как есть
		transformed = payload
	}

	// Send to target
	jsonData, _ := json.Marshal(transformed)
	
	// Используем HTTP метод из настроек интеграции
	httpMethod := integration.HTTPMethod
	if httpMethod == "" {
		httpMethod = "POST" // Default to POST
	}
	
	req, err := http.NewRequest(httpMethod, integration.TargetAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"target_api":     integration.TargetAPI,
			"error":          err.Error(),
		}).Error("Failed to create request to target API")
		
		log := models.RequestLog{
			IntegrationID: integrationID,
			Method:        httpMethod,
			URL:           integration.TargetAPI,
			RequestBody:   string(jsonData),
			StatusCode:    500,
			LogType:       "webhook",
			ErrorMessage:  err.Error(),
		}
		CreateLogWithLimit(&log)
		
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Add authentication headers
	if err := AddAuthHeaders(req, &integration); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"error":          err.Error(),
		}).Error("Failed to add authentication headers")
		
		log := models.RequestLog{
			IntegrationID: integrationID,
			Method:        httpMethod,
			URL:           integration.TargetAPI,
			RequestBody:   string(jsonData),
			StatusCode:    500,
			LogType:       "webhook",
			ErrorMessage:  "Authentication failed: " + err.Error(),
		}
		CreateLogWithLimit(&log)
		
		return err
	}
	
	// Execute request with retry logic
	client := &http.Client{}
	retryConfig := DefaultRetryConfig()
	resp, err := retryWithBackoff(req, client, retryConfig)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"target_api":     integration.TargetAPI,
			"error":          err.Error(),
		}).Error("Failed to send webhook to target API after retries")
		
		// Логируем ошибку отправки
		log := models.RequestLog{
			IntegrationID: integrationID,
			Method:        httpMethod,
			URL:           integration.TargetAPI,
			RequestBody:   string(jsonData),
			StatusCode:    500,
			LogType:       "webhook",
			ErrorMessage:  err.Error(),
		}
		CreateLogWithLimit(&log)
		
		return err
	}
	defer resp.Body.Close()

	// Читаем ответ от целевого API
	var responseBody []byte
	if resp.Body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		responseBody = buf.Bytes()
	}

	// Логируем успешную отправку к target API
	log := models.RequestLog{
		IntegrationID: integrationID,
		Method:        httpMethod,
		URL:           integration.TargetAPI,
		RequestBody:   string(jsonData),
		ResponseBody:  string(responseBody),
		StatusCode:    resp.StatusCode,
		LogType:       "webhook",
	}
	CreateLogWithLimit(&log)

	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integrationID,
		"target_api":     integration.TargetAPI,
		"status_code":    resp.StatusCode,
	}).Info("Webhook processed successfully")

	return nil
}
