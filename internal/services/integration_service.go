package services

import (
	"bytes"
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"encoding/json"
	"net/http"
)

func ProcessWebhook(integrationID uint, payload map[string]interface{}) error {
	var integration models.Integration
	if err := database.DB.First(&integration, integrationID).Error; err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"error":          err.Error(),
		}).Error("Failed to find integration")
		return err
	}

	// Parse mapping config
	// Expected format: {"target_field": "source_field"}
	var mapping map[string]string
	if integration.MappingConfig != "" {
		if err := json.Unmarshal([]byte(integration.MappingConfig), &mapping); err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"integration_id": integrationID,
				"error":          err.Error(),
			}).Warn("Invalid mapping config, using passthrough")
		}
	}

	// Transform data
	transformed := make(map[string]interface{})
	if len(mapping) > 0 {
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
	} else {
		// If no mapping, pass through all data
		transformed = payload
	}

	// Send to target
	jsonData, _ := json.Marshal(transformed)
	resp, err := http.Post(integration.TargetAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"target_api":     integration.TargetAPI,
			"error":          err.Error(),
		}).Error("Failed to send webhook to target API")
		
		// Логируем ошибку отправки
		log := models.RequestLog{
			IntegrationID: integrationID,
			Method:        "POST",
			URL:           integration.TargetAPI,
			RequestBody:   string(jsonData),
			StatusCode:    500,
			LogType:       "outgoing",
			ErrorMessage:  err.Error(),
		}
		database.DB.Create(&log)
		
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
		Method:        "POST",
		URL:           integration.TargetAPI,
		RequestBody:   string(jsonData),
		ResponseBody:  string(responseBody),
		StatusCode:    resp.StatusCode,
		LogType:       "outgoing",
	}
	database.DB.Create(&log)

	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integrationID,
		"target_api":     integration.TargetAPI,
		"status_code":    resp.StatusCode,
	}).Info("Webhook processed successfully")

	return nil
}
