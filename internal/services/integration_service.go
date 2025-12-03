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

	// Check if this is a GraphQL integration
	if integration.APIType == "graphql" {
		return ProcessGraphQLWebhook(integrationID, &integration, payload)
	}

	// Check if enrichment is enabled for REST integration
	if integration.EnrichmentEnabled && integration.EnrichmentEndpoint != "" {
		enrichedPayload, err := EnrichPayloadWithGraphQL(&integration, payload)
		if err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"integration_id": integrationID,
				"error":          err.Error(),
			}).Error("Failed to enrich payload with GraphQL")
			// Continue with original payload if enrichment fails
		} else {
			payload = enrichedPayload
		}
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
	
	// Add webhook signature
	if err := AddSignatureToRequest(req, jsonData, &integration); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"error":          err.Error(),
		}).Error("Failed to add webhook signature")
		
		log := models.RequestLog{
			IntegrationID: integrationID,
			Method:        httpMethod,
			URL:           integration.TargetAPI,
			RequestBody:   string(jsonData),
			StatusCode:    500,
			LogType:       "webhook",
			ErrorMessage:  "Signature generation failed: " + err.Error(),
		}
		CreateLogWithLimit(&log)
		
		return err
	}
	
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

// ProcessGraphQLWebhook processes webhook for GraphQL integrations
func ProcessGraphQLWebhook(integrationID uint, integration *models.Integration, payload map[string]interface{}) error {
	graphqlService := NewGraphQLService()
	
	// Execute GraphQL query
	result, err := graphqlService.ExecuteQuery(integration, payload)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"error":          err.Error(),
		}).Error("Failed to execute GraphQL query")
		
		// Log error
		log := models.RequestLog{
			IntegrationID: integrationID,
			Method:        "POST",
			URL:           integration.GraphQLEndpoint,
			RequestBody:   integration.GraphQLQuery,
			StatusCode:    500,
			LogType:       "graphql",
			ErrorMessage:  err.Error(),
		}
		CreateLogWithLimit(&log)
		
		return err
	}
	
	// Log successful GraphQL request
	resultJSON, _ := json.Marshal(result)
	log := models.RequestLog{
		IntegrationID: integrationID,
		Method:        "POST",
		URL:           integration.GraphQLEndpoint,
		RequestBody:   integration.GraphQLQuery,
		ResponseBody:  string(resultJSON),
		StatusCode:    200,
		LogType:       "graphql",
	}
	CreateLogWithLimit(&log)
	
	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integrationID,
		"endpoint":       integration.GraphQLEndpoint,
	}).Info("GraphQL query executed successfully")
	
	return nil
}

// EnrichPayloadWithGraphQL enriches payload with data from GraphQL query
func EnrichPayloadWithGraphQL(integration *models.Integration, payload map[string]interface{}) (map[string]interface{}, error) {
	graphqlService := NewGraphQLService()
	
	// Create temporary integration for GraphQL query
	enrichmentIntegration := &models.Integration{
		GraphQLEndpoint:      integration.EnrichmentEndpoint,
		GraphQLQuery:         integration.EnrichmentQuery,
		GraphQLVariables:     integration.EnrichmentVariables,
		AuthType:             integration.AuthType,
		BearerToken:          integration.BearerToken,
		BasicAuthUser:        integration.BasicAuthUser,
		BasicAuthPass:        integration.BasicAuthPass,
		OAuth2AccessToken:    integration.OAuth2AccessToken,
		OAuth2ExpiresAt:      integration.OAuth2ExpiresAt,
	}
	
	// Execute GraphQL query
	result, err := graphqlService.ExecuteQuery(enrichmentIntegration, payload)
	if err != nil {
		return nil, err
	}
	
	// Merge results based on merge mode
	switch integration.EnrichmentMergeMode {
	case "replace":
		// Replace entire payload with GraphQL result
		return result, nil
		
	case "append":
		// Add GraphQL result as a new field
		enriched := make(map[string]interface{})
		for k, v := range payload {
			enriched[k] = v
		}
		enriched["enrichment"] = result
		return enriched, nil
		
	case "merge":
		fallthrough
	default:
		// Merge GraphQL result into payload (default)
		enriched := make(map[string]interface{})
		// First copy original payload
		for k, v := range payload {
			enriched[k] = v
		}
		// Then merge GraphQL result (overwrites existing keys)
		for k, v := range result {
			enriched[k] = v
		}
		return enriched, nil
	}
}
