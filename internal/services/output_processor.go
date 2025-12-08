package services

import (
	"bytes"
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ProcessWebhookWithOutputs обрабатывает webhook с множественными выходами
// Всегда выполняет основной маппинг интеграции + все дополнительные выходы
func ProcessWebhookWithOutputs(integrationID uint, payload map[string]interface{}) error {
	var integration models.Integration
	if err := database.DB.First(&integration, integrationID).Error; err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	// Получаем все активные выходы для этой интеграции
	var outputs []models.IntegrationOutput
	database.DB.Where("integration_id = ? AND enabled = ?", integrationID, true).
		Order("priority ASC, id ASC").
		Find(&outputs)

	var wg sync.WaitGroup
	errors := make(chan error, len(outputs)+1) // +1 для основного маппинга

	// Сначала выполняем основной маппинг интеграции (если он настроен)
	if integration.MappingConfig != "" || integration.OutputTemplate != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ProcessWebhook(integrationID, payload); err != nil {
				logger.Log.WithFields(map[string]interface{}{
					"integration_id": integrationID,
					"error":          err.Error(),
				}).Error("Failed to process main integration mapping")
				errors <- fmt.Errorf("main mapping: %w", err)
			} else {
				logger.Log.WithFields(map[string]interface{}{
					"integration_id": integrationID,
				}).Info("Main integration mapping processed successfully")
			}
		}()
	}

	// Затем выполняем дополнительные выходы
	if len(outputs) > 0 {
		for _, output := range outputs {
			wg.Add(1)
			go func(out models.IntegrationOutput) {
				defer wg.Done()

				// Проверяем условие выполнения, если оно задано
				if out.Condition != "" {
					shouldExecute, err := evaluateCondition(out.Condition, payload)
					if err != nil {
						logger.Log.WithFields(map[string]interface{}{
							"output_id": out.ID,
							"error":     err.Error(),
						}).Error("Failed to evaluate condition")
						errors <- fmt.Errorf("output %s: condition evaluation failed: %w", out.Name, err)
						return
					}
					if !shouldExecute {
						logger.Log.WithFields(map[string]interface{}{
							"output_id":   out.ID,
							"output_name": out.Name,
						}).Info("Output skipped due to condition")
						return
					}
				}

				// Обрабатываем выход
				if err := processOutput(integration, out, payload); err != nil {
					logger.Log.WithFields(map[string]interface{}{
						"output_id":   out.ID,
						"output_name": out.Name,
						"error":       err.Error(),
					}).Error("Failed to process output")
					errors <- fmt.Errorf("output %s: %w", out.Name, err)
				} else {
					logger.Log.WithFields(map[string]interface{}{
						"output_id":   out.ID,
						"output_name": out.Name,
					}).Info("Output processed successfully")
				}
			}(output)
		}
	}

	// Ждем завершения всех горутин
	wg.Wait()
	close(errors)

	// Собираем ошибки
	var errorMessages []string
	for err := range errors {
		errorMessages = append(errorMessages, err.Error())
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("some outputs failed: %v", errorMessages)
	}

	return nil
}

// processMultipleOutputs обрабатывает несколько выходов параллельно
func processMultipleOutputs(integration models.Integration, outputs []models.IntegrationOutput, payload map[string]interface{}) error {
	var wg sync.WaitGroup
	errors := make(chan error, len(outputs))
	
	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integration.ID,
		"outputs_count":  len(outputs),
	}).Info("Processing webhook with multiple outputs")

	for _, output := range outputs {
		wg.Add(1)
		go func(out models.IntegrationOutput) {
			defer wg.Done()
			
			// Проверяем условие выполнения, если оно задано
			if out.Condition != "" {
				shouldExecute, err := evaluateCondition(out.Condition, payload)
				if err != nil {
					logger.Log.WithFields(map[string]interface{}{
						"output_id": out.ID,
						"error":     err.Error(),
					}).Error("Failed to evaluate condition")
					errors <- fmt.Errorf("output %s: condition evaluation failed: %w", out.Name, err)
					return
				}
				if !shouldExecute {
					logger.Log.WithFields(map[string]interface{}{
						"output_id": out.ID,
						"output_name": out.Name,
					}).Info("Output skipped due to condition")
					return
				}
			}

			// Обрабатываем выход
			if err := processOutput(integration, out, payload); err != nil {
				logger.Log.WithFields(map[string]interface{}{
					"output_id":   out.ID,
					"output_name": out.Name,
					"error":       err.Error(),
				}).Error("Failed to process output")
				errors <- fmt.Errorf("output %s: %w", out.Name, err)
			} else {
				logger.Log.WithFields(map[string]interface{}{
					"output_id":   out.ID,
					"output_name": out.Name,
				}).Info("Output processed successfully")
			}
		}(output)
	}

	// Ждем завершения всех горутин
	wg.Wait()
	close(errors)

	// Собираем ошибки
	var errorMessages []string
	for err := range errors {
		errorMessages = append(errorMessages, err.Error())
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("some outputs failed: %v", errorMessages)
	}

	return nil
}

// processOutput обрабатывает один выход
func processOutput(integration models.Integration, output models.IntegrationOutput, payload map[string]interface{}) error {
	// Трансформируем данные
	var transformed interface{}
	var err error

	if output.OutputTemplate != "" {
		processor := utils.NewTemplateProcessor()
		templateType := output.TemplateType
		if templateType == "" {
			templateType = "json"
		}

		if templateType == "json" {
			transformed, err = processor.ProcessTemplate(output.OutputTemplate, payload)
		} else {
			transformed, err = processor.ProcessTemplateString(output.OutputTemplate, payload)
		}
		
		if err != nil {
			return fmt.Errorf("template processing failed: %w", err)
		}
	} else if output.MappingConfig != "" {
		var mapping map[string]string
		if err := json.Unmarshal([]byte(output.MappingConfig), &mapping); err != nil {
			return fmt.Errorf("invalid mapping config: %w", err)
		}

		transformedMap := make(map[string]interface{})
		for targetField, sourceField := range mapping {
			if v, err := utils.GetValueByPath(payload, sourceField); err == nil {
				transformedMap[targetField] = v
			}
		}
		transformed = transformedMap
	} else {
		transformed = payload
	}

	// Отправляем на Target API
	return sendToTargetAPI(output, transformed, integration.ID)
}

// sendToTargetAPI отправляет данные на Target API выхода
func sendToTargetAPI(output models.IntegrationOutput, data interface{}, integrationID uint) error {
	// Подготавливаем данные для отправки
	var requestBody []byte
	var contentType string
	
	// Проверяем тип данных
	if strData, ok := data.(string); ok {
		// Строковые данные (XML, Text, Custom)
		requestBody = []byte(strData)
		
		templateType := output.TemplateType
		switch templateType {
		case "xml":
			contentType = "application/xml"
		case "text":
			contentType = "text/plain"
		case "custom":
			contentType = "text/plain"
		default:
			contentType = "application/json"
		}
	} else {
		// JSON данные
		var err error
		requestBody, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		contentType = "application/json"
	}
	
	// Используем HTTP метод из настроек выхода
	httpMethod := output.HTTPMethod
	if httpMethod == "" {
		httpMethod = "POST"
	}
	
	// Создаем временную интеграцию для использования существующих функций аутентификации
	tempIntegration := models.Integration{
		TargetAPI:          output.TargetAPI,
		HTTPMethod:         httpMethod,
		AuthType:           output.AuthType,
		OAuth2TokenURL:     output.OAuth2TokenURL,
		OAuth2ClientID:     output.OAuth2ClientID,
		OAuth2ClientSecret: output.OAuth2ClientSecret,
		OAuth2Scope:        output.OAuth2Scope,
		OAuth2GrantType:    output.OAuth2GrantType,
		BearerToken:        output.BearerToken,
		BasicAuthUser:      output.BasicAuthUser,
		BasicAuthPass:      output.BasicAuthPass,
		OAuth2AccessToken:  output.OAuth2AccessToken,
		OAuth2RefreshToken: output.OAuth2RefreshToken,
		OAuth2ExpiresAt:    output.OAuth2ExpiresAt,
		CustomHeaders:      output.CustomHeaders,
	}
	tempIntegration.ID = integrationID
	
	// Создаем HTTP запрос
	req, err := http.NewRequest(httpMethod, output.TargetAPI, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", contentType)
	
	// Добавляем аутентификацию
	if err := AddAuthHeaders(req, &tempIntegration); err != nil {
		return fmt.Errorf("failed to add auth headers: %w", err)
	}
	
	// Добавляем кастомные заголовки
	if err := AddCustomHeaders(req, &tempIntegration); err != nil {
		return fmt.Errorf("failed to add custom headers: %w", err)
	}
	
	// Логируем исходящий запрос
	headersJSON, _ := json.Marshal(req.Header)
	requestLog := models.RequestLog{
		IntegrationID:  integrationID,
		Method:         httpMethod,
		URL:            output.TargetAPI,
		RequestBody:    string(requestBody),
		RequestHeaders: string(headersJSON),
		LogType:        "webhook",
		OutputName:     output.Name,
	}
	
	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		requestLog.ErrorMessage = err.Error()
		requestLog.StatusCode = 0
		database.DB.Create(&requestLog)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Читаем тело ответа
	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)
	
	requestLog.ResponseBody = responseBody.String()
	requestLog.StatusCode = resp.StatusCode
	
	// Проверяем статус ответа
	if resp.StatusCode >= 400 {
		requestLog.ErrorMessage = fmt.Sprintf("Target API returned error status: %d", resp.StatusCode)
	}
	
	// Сохраняем лог
	database.DB.Create(&requestLog)
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("target API returned error status: %d", resp.StatusCode)
	}
	
	return nil
}

// evaluateCondition оценивает условие выполнения
// Простая реализация: поддержка базовых операторов ==, !=, >, <, >=, <=
func evaluateCondition(condition string, payload map[string]interface{}) (bool, error) {
	// TODO: Реализовать полноценный парсер условий
	// Пока возвращаем true (выполнять всегда)
	// В будущем можно использовать библиотеку для выражений типа govaluate или expr
	
	if condition == "" {
		return true, nil
	}

	// Простая реализация: если условие не пустое, считаем что нужно выполнять
	// В будущих версиях здесь будет полноценный парсер
	logger.Log.WithFields(map[string]interface{}{
		"condition": condition,
	}).Warn("Condition evaluation not fully implemented yet, executing output")
	
	return true, nil
}
