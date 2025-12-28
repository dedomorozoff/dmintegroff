package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"dmintegroff/internal/ai"
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
)

// AIController контроллер для AI функций
type AIController struct {
	client       *ai.Client
	generator    *ai.Generator
	analyzer     *ai.Analyzer
	demoResponses *ai.DemoResponses
}

// NewAIController создает новый AI контроллер
func NewAIController(config *ai.AIConfig) *AIController {
	client := ai.NewClient(config)
	return &AIController{
		client:        client,
		generator:     ai.NewGenerator(client),
		analyzer:      ai.NewAnalyzer(client),
		demoResponses: ai.NewDemoResponses(),
	}
}

// isDemoMode проверяет, включен ли демо режим
func (c *AIController) isDemoMode() bool {
	return os.Getenv("DEMO_MODE") == "true"
}

// Chat обрабатывает запросы к AI чату
func (c *AIController) Chat(ctx *gin.Context) {
	startTime := time.Now()
	
	// Получаем информацию о пользователе из сессии
	session := sessions.Default(ctx)
	userID := session.Get("user_id")
	username := session.Get("username")
	
	var req ai.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_chat",
			"user_id": userID,
			"username": username,
			"error": "invalid_request_format",
			"details": err.Error(),
			"ip": ctx.ClientIP(),
		}).Error("AI Chat: Invalid request format")
		
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
			"details": err.Error(),
		})
		return
	}

	// Логируем запрос пользователя к AI чату
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_chat_request",
		"user_id": userID,
		"username": username,
		"message": req.Message,
		"target_api": req.TargetAPI,
		"has_sample_data": len(req.SampleData) > 0,
		"sample_data_size": len(req.SampleData),
		"has_context": len(req.Context) > 0,
		"context_keys": func() []string {
			keys := make([]string, 0, len(req.Context))
			for k := range req.Context {
				keys = append(keys, k)
			}
			return keys
		}(),
		"history_length": len(req.History),
		"demo_mode": c.isDemoMode(),
		"ip": ctx.ClientIP(),
	}).Info("AI Chat: User chat request received")

	// Проверяем демо режим
	if c.isDemoMode() {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_chat_demo",
			"user_id": userID,
			"username": username,
			"message": req.Message,
			"target_api": req.TargetAPI,
		}).Info("AI Chat: Demo mode - returning demo response")
		
		// Возвращаем демо ответ
		demoResponse := c.demoResponses.GetDemoChatResponse(req.Message, req.TargetAPI, req.SampleData)
		
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_chat_demo_success",
			"user_id": userID,
			"username": username,
			"suggestions_count": len(demoResponse.Suggestions),
			"duration": time.Since(startTime).String(),
		}).Info("AI Chat: Demo response sent successfully")
		
		ctx.JSON(http.StatusOK, demoResponse)
		return
	}

	// Проверяем, настроен ли AI (только если не демо режим)
	if !c.client.IsConfigured() {
		provider := c.client.GetCurrentProvider()
		
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_chat",
			"user_id": userID,
			"username": username,
			"error": "ai_not_configured",
			"provider": provider,
			"ip": ctx.ClientIP(),
		}).Error("AI Chat: AI not configured")
		
		if provider == "Disabled (AI_ENABLED=false)" {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "AI отключен",
				"details": "AI функциональность отключена в настройках сервера (AI_ENABLED=false)",
			})
		} else {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "AI не настроен",
				"details": "Необходимо настроить OPENROUTER_API_KEY или OPENAI_API_KEY в разделе Настройки",
			})
		}
		return
	}

	// Создаем контекст с таймаутом из конфигурации AI
	timeoutDuration := time.Duration(c.client.Config.RequestTimeout) * time.Second
	requestCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	// Строим промпт
	promptBuilder := ai.NewChatPromptBuilder()
	
	// Добавляем контекст если есть
	if req.Context != nil {
		for key, value := range req.Context {
			promptBuilder.WithContext(key, value)
		}
	}
	
	// Создаем сообщения для AI
	messages := make([]ai.ChatMessage, 0)
	
	// Добавляем системный промпт
	messages = append(messages, ai.ChatMessage{
		Role:    "system",
		Content: ai.SystemPrompts["chat"],
	})
	
	// Добавляем историю если есть
	if len(req.History) > 0 {
		messages = append(messages, req.History...)
	}
	
	// Добавляем текущее сообщение пользователя
	userPrompt := promptBuilder.BuildChatPrompt(req.Message, req.SampleData)
	messages = append(messages, ai.ChatMessage{
		Role:    "user",
		Content: userPrompt,
	})

	// Логируем построенный промпт
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_chat_prompt",
		"user_id": userID,
		"username": username,
		"prompt_length": len(userPrompt),
		"prompt_preview": func() string {
			if len(userPrompt) > 300 {
				return userPrompt[:300] + "..."
			}
			return userPrompt
		}(),
		"messages_count": len(messages),
	}).Info("AI Chat: Generated prompt for AI")

	// Отправляем запрос к AI
	response, err := c.client.Chat(requestCtx, messages, false)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_chat",
			"user_id": userID,
			"username": username,
			"error": "ai_request_failed",
			"details": err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Chat: AI request failed")
		
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка AI",
			"details": err.Error(),
		})
		return
	}

	// Логируем ответ от AI
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_chat_response",
		"user_id": userID,
		"username": username,
		"response_length": len(response),
		"response_preview": func() string {
			if len(response) > 300 {
				return response[:300] + "..."
			}
			return response
		}(),
		"duration": time.Since(startTime).String(),
	}).Info("AI Chat: Received response from AI")

	// Анализируем ответ и создаем предложения
	suggestions := c.createSuggestions(req.Message, req.SampleData, req.TargetAPI)
	
	// Логируем финальный результат
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_chat_success",
		"user_id": userID,
		"username": username,
		"suggestions_count": len(suggestions),
		"total_duration": time.Since(startTime).String(),
	}).Info("AI Chat: Chat request completed successfully")
	
	// Возвращаем ответ
	ctx.JSON(http.StatusOK, ai.ChatResponse{
		Response:    response,
		Suggestions: suggestions,
		NextSteps:   c.generateNextSteps(req.Message, req.TargetAPI),
		Confidence:  0.8, // Базовая уверенность
	})
}

// AnalyzeData анализирует структуру данных
func (c *AIController) AnalyzeData(ctx *gin.Context) {
	startTime := time.Now()
	
	// Получаем информацию о пользователе из сессии
	session := sessions.Default(ctx)
	userID := session.Get("user_id")
	username := session.Get("username")
	
	var req ai.DataAnalysisRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_analyze_data",
			"user_id": userID,
			"username": username,
			"error": "invalid_request_format",
			"details": err.Error(),
			"ip": ctx.ClientIP(),
		}).Error("AI Analyze Data: Invalid request format")
		
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
			"details": err.Error(),
		})
		return
	}

	// Логируем запрос на анализ данных
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_analyze_data_request",
		"user_id": userID,
		"username": username,
		"format": req.Format,
		"data_fields": len(req.Data),
		"data_preview": func() string {
			dataJSON, _ := json.Marshal(req.Data)
			if len(dataJSON) > 200 {
				return string(dataJSON[:200]) + "..."
			}
			return string(dataJSON)
		}(),
		"demo_mode": c.isDemoMode(),
		"ip": ctx.ClientIP(),
	}).Info("AI Analyze Data: Data analysis request received")

	// Проверяем демо режим
	if c.isDemoMode() {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_analyze_data_demo",
			"user_id": userID,
			"username": username,
			"format": req.Format,
			"data_fields": len(req.Data),
		}).Info("AI Analyze Data: Demo mode - returning demo analysis")
		
		// Возвращаем демо анализ
		demoAnalysis := c.demoResponses.GetDemoDataAnalysis(req.Data, req.Format)
		
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_analyze_data_demo_success",
			"user_id": userID,
			"username": username,
			"data_type": demoAnalysis.DataType,
			"fields_count": len(demoAnalysis.Fields),
			"duration": time.Since(startTime).String(),
		}).Info("AI Analyze Data: Demo analysis sent successfully")
		
		ctx.JSON(http.StatusOK, demoAnalysis)
		return
	}

	// Создаем контекст с таймаутом из конфигурации AI (только если не демо режим)
	timeoutDuration := time.Duration(c.client.Config.RequestTimeout) * time.Second
	requestCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	// Анализируем данные
	analysis, err := c.analyzer.AnalyzeDataStructure(requestCtx, req.Data, req.Format)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_analyze_data",
			"user_id": userID,
			"username": username,
			"error": "analysis_failed",
			"details": err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Analyze Data: Analysis failed")
		
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка анализа данных",
			"details": err.Error(),
		})
		return
	}

	// Логируем успешный результат анализа
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_analyze_data_success",
		"user_id": userID,
		"username": username,
		"data_type": analysis.DataType,
		"confidence": analysis.Confidence,
		"fields_count": len(analysis.Fields),
		"suggestions_count": len(analysis.Suggestions),
		"duration": time.Since(startTime).String(),
	}).Info("AI Analyze Data: Analysis completed successfully")

	ctx.JSON(http.StatusOK, analysis)
}

// GenerateMapping генерирует маппинг для интеграции
func (c *AIController) GenerateMapping(ctx *gin.Context) {
	startTime := time.Now()
	
	// Получаем информацию о пользователе из сессии
	session := sessions.Default(ctx)
	userID := session.Get("user_id")
	username := session.Get("username")
	
	var req ai.MappingGenerationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_generate_mapping",
			"user_id": userID,
			"username": username,
			"error": "invalid_request_format",
			"details": err.Error(),
			"ip": ctx.ClientIP(),
		}).Error("AI Generate Mapping: Invalid request format")
		
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
			"details": err.Error(),
		})
		return
	}

	// Логируем запрос на генерацию маппинга
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_generate_mapping_request",
		"user_id": userID,
		"username": username,
		"target_api": req.TargetAPI,
		"task": req.Task,
		"user_prompt": req.UserPrompt,
		"source_data_fields": len(req.SourceData),
		"source_data_preview": func() string {
			dataJSON, _ := json.Marshal(req.SourceData)
			if len(dataJSON) > 200 {
				return string(dataJSON[:200]) + "..."
			}
			return string(dataJSON)
		}(),
		"demo_mode": c.isDemoMode(),
		"ip": ctx.ClientIP(),
	}).Info("AI Generate Mapping: Mapping generation request received")

	// Проверяем демо режим
	if c.isDemoMode() {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_generate_mapping_demo",
			"user_id": userID,
			"username": username,
			"target_api": req.TargetAPI,
			"task": req.Task,
		}).Info("AI Generate Mapping: Demo mode - returning demo mapping")
		
		// Возвращаем демо маппинг
		demoMapping := c.demoResponses.GetDemoGeneratedMapping(&req)
		
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_generate_mapping_demo_success",
			"user_id": userID,
			"username": username,
			"mapping_type": demoMapping.Type,
			"target_url": demoMapping.TargetURL,
			"duration": time.Since(startTime).String(),
		}).Info("AI Generate Mapping: Demo mapping sent successfully")
		
		ctx.JSON(http.StatusOK, demoMapping)
		return
	}

	// Создаем контекст с таймаутом из конфигурации AI (только если не демо режим)
	timeoutDuration := time.Duration(c.client.Config.RequestTimeout) * time.Second
	requestCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	// Генерируем маппинг
	mapping, err := c.generator.GenerateMapping(requestCtx, &req)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_generate_mapping",
			"user_id": userID,
			"username": username,
			"error": "generation_failed",
			"details": err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Generate Mapping: Generation failed")
		
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка генерации маппинга",
			"details": err.Error(),
		})
		return
	}

	// Логируем успешный результат генерации
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_generate_mapping_success",
		"user_id": userID,
		"username": username,
		"mapping_type": mapping.Type,
		"target_url": mapping.TargetURL,
		"method": mapping.Method,
		"auth_type": mapping.AuthType,
		"template_length": len(mapping.GetTemplateString()),
		"template_preview": func() string {
			template := mapping.GetTemplateString()
			if len(template) > 150 {
				return template[:150] + "..."
			}
			return template
		}(),
		"duration": time.Since(startTime).String(),
	}).Info("AI Generate Mapping: Mapping generation completed successfully")

	ctx.JSON(http.StatusOK, mapping)
}

// ApplyMapping применяет сгенерированный маппинг к интеграции
func (c *AIController) ApplyMapping(ctx *gin.Context) {
	integrationIDStr := ctx.Param("id")
	integrationID, err := strconv.Atoi(integrationIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID интеграции",
		})
		return
	}

	var mapping ai.GeneratedMapping
	if err := ctx.ShouldBindJSON(&mapping); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат маппинга",
			"details": err.Error(),
		})
		return
	}

	// Получаем интеграцию из БД
	var integration models.Integration
	if err := database.DB.First(&integration, integrationID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Интеграция не найдена",
		})
		return
	}

	// Применяем маппинг к интеграции
	integration.TargetAPI = mapping.TargetURL
	integration.HTTPMethod = mapping.Method
	integration.OutputTemplate = mapping.GetTemplateString()
	// Note: Integration model doesn't have Description field, using Name instead
	if mapping.Description != "" {
		integration.Name = mapping.Description
	}

	// Настраиваем аутентификацию
	if mapping.AuthType != "none" {
		integration.AuthType = mapping.AuthType
		// Настраиваем конкретные поля аутентификации
		if mapping.AuthType == "bearer" {
			if token, ok := mapping.AuthConfig["token"].(string); ok {
				integration.BearerToken = token
			}
		} else if mapping.AuthType == "basic" {
			if username, ok := mapping.AuthConfig["username"].(string); ok {
				integration.BasicAuthUser = username
			}
			if password, ok := mapping.AuthConfig["password"].(string); ok {
				integration.BasicAuthPass = password
			}
		}
	}

	// Note: Integration model doesn't have Headers field
	// Headers would need to be handled differently or added to the model

	// Сохраняем изменения
	if err := database.DB.Save(&integration).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения интеграции",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Маппинг успешно применен к интеграции",
		"integration": integration,
	})
}

// GetStatus возвращает статус AI сервиса
func (c *AIController) GetStatus(ctx *gin.Context) {
	status := gin.H{
		"demo_mode":  c.isDemoMode(),
		"configured": c.client.IsConfigured(),
		"provider":   c.client.GetCurrentProvider(),
		"models":     c.client.GetAvailableModels(),
		"enabled":    c.client.Config.Enabled, // добавляем информацию о глобальном переключателе
	}

	// В демо режиме всегда показываем как доступный
	if c.isDemoMode() {
		status["available"] = true
		status["provider"] = "Demo Mode"
		status["demo_message"] = "AI функции работают в демонстрационном режиме"
		ctx.JSON(http.StatusOK, status)
		return
	}

	// Проверяем доступность AI (только если не демо режим)
	if c.client.IsConfigured() {
		// Используем настроенный таймаут из конфигурации AI вместо хардкода
		timeoutDuration := time.Duration(c.client.Config.RequestTimeout) * time.Second
		testCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()
		
		testMessages := []ai.ChatMessage{
			{Role: "user", Content: "test"},
		}
		
		_, err := c.client.Chat(testCtx, testMessages, false)
		status["available"] = err == nil
		if err != nil {
			status["error"] = err.Error()
		}
	} else {
		status["available"] = false
		provider := c.client.GetCurrentProvider()
		if provider == "Disabled (AI_ENABLED=false)" {
			status["error"] = "AI отключен в настройках сервера (AI_ENABLED=false)"
		} else {
			status["error"] = "AI не настроен - добавьте API ключи в разделе Настройки"
		}
	}

	ctx.JSON(http.StatusOK, status)
}

// GetQuickSuggestions возвращает быстрые предложения
func (c *AIController) GetQuickSuggestions(ctx *gin.Context) {
	suggestions := ai.GetQuickSuggestions()
	popularAPIs := ai.GetPopularAPIs()

	ctx.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
		"popular_apis": popularAPIs,
	})
}

// GetModels возвращает список доступных AI моделей
func (c *AIController) GetModels(ctx *gin.Context) {
	// Создаем контекст с таймаутом из конфигурации AI
	timeoutDuration := time.Duration(c.client.Config.RequestTimeout) * time.Second
	requestCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	result := gin.H{
		"recommended": c.client.GetRecommendedModels(),
		"static":      c.client.GetAvailableModels(),
	}

	// Пробуем получить актуальный список от OpenRouter (даже если AI не полностью настроен)
	// Для получения списка моделей достаточно иметь API ключ
	if c.client.Config.OpenRouterAPIKey != "" && c.client.Config.Enabled {
		if models, err := c.client.GetOpenRouterModels(requestCtx); err == nil {
			// Обрабатываем все модели и добавляем метаданные
			processedModels := make([]map[string]interface{}, 0)
			freeModels := make([]map[string]interface{}, 0)
			popularModels := make([]map[string]interface{}, 0)
			
			// Популярные БЕСПЛАТНЫЕ модели для быстрого доступа
			popularIDs := map[string]bool{
				"meta-llama/llama-3.2-3b-instruct:free":     true,
				"meta-llama/llama-3.1-8b-instruct:free":     true,
				"google/gemma-2-9b-it:free":                 true,
				"microsoft/phi-3-mini-128k-instruct:free":   true,
				"qwen/qwen-2-7b-instruct:free":              true,
				"mistralai/mistral-7b-instruct:free":        true,
				"huggingfaceh4/zephyr-7b-beta:free":         true,
			}

			for _, model := range models {
				// Определяем провайдера из ID модели
				provider := "Other"
				if strings.Contains(model.ID, "anthropic/") {
					provider = "Anthropic"
				} else if strings.Contains(model.ID, "openai/") {
					provider = "OpenAI"
				} else if strings.Contains(model.ID, "google/") {
					provider = "Google"
				} else if strings.Contains(model.ID, "meta-llama/") {
					provider = "Meta"
				} else if strings.Contains(model.ID, "mistralai/") {
					provider = "Mistral"
				} else if strings.Contains(model.ID, "microsoft/") {
					provider = "Microsoft"
				} else if strings.Contains(model.ID, "cohere/") {
					provider = "Cohere"
				}

				// Проверяем, бесплатная ли модель
				isFree := model.IsFree()

				// Создаем обработанную модель
				processedModel := map[string]interface{}{
					"id":          model.ID,
					"name":        model.Name,
					"description": model.Description,
					"provider":    provider,
					"is_free":     isFree,
					"context_length": model.ContextLength,
				}

				// Добавляем информацию о цене
				if model.Pricing != nil {
					processedModel["pricing"] = map[string]interface{}{
						"prompt":     model.GetPromptPrice(),
						"completion": model.GetCompletionPrice(),
					}
				}

				processedModels = append(processedModels, processedModel)

				// Добавляем в соответствующие категории
				if isFree {
					freeModels = append(freeModels, processedModel)
				}
				if popularIDs[model.ID] {
					popularModels = append(popularModels, processedModel)
				}
			}

			// Показываем только бесплатные модели для избежания проблем с кредитами
			result["openrouter_all"] = freeModels
			result["openrouter_free"] = freeModels
			result["openrouter_popular"] = freeModels // Только бесплатные популярные модели
			result["openrouter_available"] = true
		} else {
			result["openrouter_available"] = false
			result["openrouter_error"] = err.Error()
		}
	} else {
		result["openrouter_available"] = false
		result["openrouter_error"] = "OpenRouter API key not configured"
	}

	ctx.JSON(http.StatusOK, result)
}

// createSuggestions создает предложения на основе запроса
func (c *AIController) createSuggestions(message string, sampleData map[string]interface{}, targetAPI string) []ai.AISuggestion {
	suggestions := make([]ai.AISuggestion, 0)

	// Предложение анализа данных если есть образец
	if len(sampleData) > 0 {
		suggestions = append(suggestions, ai.AISuggestion{
			Type:        "data_analysis",
			Title:       "Анализировать данные",
			Description: "Проанализировать структуру входящих данных",
			Data: map[string]interface{}{
				"action": "analyze_data",
				"data":   sampleData,
			},
			Confidence: 0.9,
		})
	}

	// Предложение генерации маппинга если указан целевой API
	if targetAPI != "" {
		suggestions = append(suggestions, ai.AISuggestion{
			Type:        "mapping_generation",
			Title:       "Создать маппинг",
			Description: fmt.Sprintf("Создать маппинг для %s", targetAPI),
			Data: map[string]interface{}{
				"action":     "generate_mapping",
				"target_api": targetAPI,
				"task":       message,
			},
			Confidence: 0.8,
		})
	}

	// Предложения популярных API
	popularAPIs := ai.GetPopularAPIs()
	for _, api := range popularAPIs {
		suggestions = append(suggestions, ai.AISuggestion{
			Type:        "api_config",
			Title:       fmt.Sprintf("Настроить %s", api.Name),
			Description: api.Description,
			Data: map[string]interface{}{
				"action": "configure_api",
				"api":    api,
			},
			Confidence: 0.7,
		})
	}

	return suggestions
}

// CreateIntegration создает интеграцию с помощью AI
func (c *AIController) CreateIntegration(ctx *gin.Context) {
	startTime := time.Now()
	
	// Получаем пользователя из сессии
	session := sessions.Default(ctx)
	userIDInterface := session.Get("user_id")
	username := session.Get("username")
	
	if userIDInterface == nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_create_integration",
			"error":  "unauthorized_user",
			"ip":     ctx.ClientIP(),
		}).Error("AI Integration Creation: Unauthorized user")
		
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Пользователь не авторизован",
		})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_create_integration",
			"error":  "invalid_session",
			"ip":     ctx.ClientIP(),
		}).Error("AI Integration Creation: Invalid user session")
		
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Неверная сессия пользователя",
		})
		return
	}

	var req ai.CreateIntegrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action":   "ai_create_integration",
			"user_id":  userID,
			"username": username,
			"error":    "invalid_request_format",
			"details":  err.Error(),
			"ip":       ctx.ClientIP(),
		}).Error("AI Integration Creation: Invalid request format")
		
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
			"details": err.Error(),
		})
		return
	}

	// Логируем начало AI операции
	logger.Log.WithFields(map[string]interface{}{
		"action":      "ai_create_integration_start",
		"user_id":     userID,
		"username":    username,
		"project_id":  req.ProjectID,
		"description": req.Description,
		"has_sample":  req.SampleData != "",
		"sample_size": len(req.SampleData),
		"demo_mode":   c.isDemoMode(),
		"ip":          ctx.ClientIP(),
	}).Info("AI Integration Creation: Starting AI integration creation")

	// Проверяем демо режим
	if c.isDemoMode() {
		logger.Log.WithFields(map[string]interface{}{
			"action":      "ai_create_integration_demo",
			"user_id":     userID,
			"username":    username,
			"description": req.Description,
		}).Info("AI Integration Creation: Demo mode - returning demo integration")
		
		// Возвращаем демо интеграцию
		demoIntegration := c.demoResponses.GetDemoCreatedIntegration(&req)
		
		// Генерируем уникальный webhook токен для демо интеграции
		webhookToken, err := utils.GenerateToken(16)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка генерации токена",
				"details": err.Error(),
			})
			return
		}

		// Создаем модель интеграции для сохранения в БД (даже в демо режиме)
		dbIntegration := models.Integration{
			Name:           demoIntegration.Name,
			TargetAPI:      demoIntegration.TargetURL,
			HTTPMethod:     demoIntegration.Method,
			OutputTemplate: demoIntegration.Template,
			TemplateType:   demoIntegration.TemplateType,
			ProjectID:      uint(req.ProjectID),
			Mode:          "inactive", // Создаем в неактивном режиме
			SamplePayload: req.SampleData,
			WebhookToken:  webhookToken,
			CreatedByID:   userID,
		}

		// Настраиваем аутентификацию
		if demoIntegration.AuthType != "none" {
			dbIntegration.AuthType = demoIntegration.AuthType
			if demoIntegration.AuthType == "bearer" {
				if token, ok := demoIntegration.AuthConfig["token"].(string); ok {
					dbIntegration.BearerToken = token
				}
			}
		}

		// Сохраняем интеграцию в БД
		if err := database.DB.Create(&dbIntegration).Error; err != nil {
			logger.Log.WithFields(map[string]interface{}{
				"action":   "ai_create_integration_demo",
				"user_id":  userID,
				"username": username,
				"error":    "database_save_failed",
				"details":  err.Error(),
				"duration": time.Since(startTime).String(),
			}).Error("AI Integration Creation: Failed to save demo integration to database")
			
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка сохранения интеграции",
				"details": err.Error(),
			})
			return
		}

		logger.Log.WithFields(map[string]interface{}{
			"action":         "ai_create_integration_demo_success",
			"user_id":        userID,
			"username":       username,
			"integration_id": dbIntegration.ID,
			"integration_name": dbIntegration.Name,
			"webhook_token":  dbIntegration.WebhookToken,
			"duration":       time.Since(startTime).String(),
		}).Info("AI Integration Creation: Demo integration created successfully")

		// Возвращаем успешный ответ
		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
			"integration": gin.H{
				"id":           dbIntegration.ID,
				"name":         dbIntegration.Name,
				"target_api":   dbIntegration.TargetAPI,
				"method":       dbIntegration.HTTPMethod,
				"template":     dbIntegration.OutputTemplate,
				"template_type": dbIntegration.TemplateType,
				"mode":         dbIntegration.Mode,
			},
			"mapping":      demoIntegration.Mapping,
			"explanation":  demoIntegration.Explanation,
			"next_steps":   demoIntegration.NextSteps,
		})
		return
	}

	// Проверяем, настроен ли AI (только если не демо режим)
	if !c.client.IsConfigured() {
		provider := c.client.GetCurrentProvider()
		
		logger.Log.WithFields(map[string]interface{}{
			"action":   "ai_create_integration",
			"user_id":  userID,
			"username": username,
			"error":    "ai_not_configured",
			"provider": provider,
			"ip":       ctx.ClientIP(),
		}).Error("AI Integration Creation: AI not configured")
		
		if provider == "Disabled (AI_ENABLED=false)" {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "AI отключен",
				"details": "AI функциональность отключена в настройках сервера (AI_ENABLED=false)",
			})
		} else {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "AI не настроен",
				"details": "Необходимо настроить OPENROUTER_API_KEY или OPENAI_API_KEY в разделе Настройки",
			})
		}
		return
	}

	// Создаем контекст с увеличенным таймаутом для сложных операций создания интеграции
	timeoutDuration := time.Duration(c.client.Config.RequestTimeout*2) * time.Second
	requestCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	logger.Log.WithFields(map[string]interface{}{
		"action":   "ai_create_integration",
		"user_id":  userID,
		"username": username,
		"provider": c.client.GetCurrentProvider(),
		"timeout":  timeoutDuration.String(),
	}).Info("AI Integration Creation: Starting AI generation")

	// Генерируем интеграцию с помощью AI
	integration, err := c.generator.CreateIntegration(requestCtx, &req)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action":   "ai_create_integration",
			"user_id":  userID,
			"username": username,
			"error":    "ai_generation_failed",
			"details":  err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Integration Creation: AI generation failed")
		
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка создания интеграции",
			"details": err.Error(),
		})
		return
	}

	logger.Log.WithFields(map[string]interface{}{
		"action":        "ai_create_integration",
		"user_id":       userID,
		"username":      username,
		"generated_name": integration.Name,
		"target_url":    integration.TargetURL,
		"method":        integration.Method,
		"template_type": integration.TemplateType,
		"auth_type":     integration.AuthType,
		"duration":      time.Since(startTime).String(),
	}).Info("AI Integration Creation: AI generation completed successfully")

	// Генерируем уникальный webhook токен
	webhookToken, err := utils.GenerateToken(16)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка генерации токена",
			"details": err.Error(),
		})
		return
	}

	// Создаем модель интеграции для сохранения в БД
	dbIntegration := models.Integration{
		Name:           integration.Name,
		TargetAPI:      integration.TargetURL,
		HTTPMethod:     integration.Method,
		OutputTemplate: integration.Template,
		TemplateType:   integration.TemplateType,
		ProjectID:      uint(req.ProjectID),
		Mode:          "inactive", // Создаем в неактивном режиме для настройки
		SamplePayload: req.SampleData,
		WebhookToken:  webhookToken,
		CreatedByID:   userID,
	}

	// Настраиваем аутентификацию
	if integration.AuthType != "none" {
		dbIntegration.AuthType = integration.AuthType
		if integration.AuthType == "bearer" {
			if token, ok := integration.AuthConfig["token"].(string); ok {
				dbIntegration.BearerToken = token
			}
		} else if integration.AuthType == "basic" {
			if username, ok := integration.AuthConfig["username"].(string); ok {
				dbIntegration.BasicAuthUser = username
			}
			if password, ok := integration.AuthConfig["password"].(string); ok {
				dbIntegration.BasicAuthPass = password
			}
		}
	}

	// Сохраняем интеграцию в БД
	if err := database.DB.Create(&dbIntegration).Error; err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action":   "ai_create_integration",
			"user_id":  userID,
			"username": username,
			"error":    "database_save_failed",
			"details":  err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Integration Creation: Failed to save integration to database")
		
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения интеграции",
			"details": err.Error(),
		})
		return
	}

	// Логируем успешное завершение
	logger.Log.WithFields(map[string]interface{}{
		"action":         "ai_create_integration_success",
		"user_id":        userID,
		"username":       username,
		"integration_id": dbIntegration.ID,
		"integration_name": dbIntegration.Name,
		"webhook_token":  dbIntegration.WebhookToken,
		"target_api":     dbIntegration.TargetAPI,
		"method":         dbIntegration.HTTPMethod,
		"template_type":  dbIntegration.TemplateType,
		"auth_type":      dbIntegration.AuthType,
		"project_id":     dbIntegration.ProjectID,
		"total_duration": time.Since(startTime).String(),
		"template_preview": func() string {
			if len(integration.Template) > 100 {
				return integration.Template[:100] + "..."
			}
			return integration.Template
		}(), // Первые 100 символов шаблона
	}).Info("AI Integration Creation: Integration created successfully")

	// Возвращаем успешный ответ
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"integration": gin.H{
			"id":           dbIntegration.ID,
			"name":         dbIntegration.Name,
			"target_api":   dbIntegration.TargetAPI,
			"method":       dbIntegration.HTTPMethod,
			"template":     dbIntegration.OutputTemplate,
			"template_type": dbIntegration.TemplateType,
			"mode":         dbIntegration.Mode,
		},
		"mapping":      integration.Mapping,
		"explanation":  integration.Explanation,
		"next_steps":   integration.NextSteps,
	})
}

// generateNextSteps генерирует следующие шаги
func (c *AIController) generateNextSteps(message, targetAPI string) []string {
	steps := make([]string, 0)

	if targetAPI == "" {
		steps = append(steps, "Укажите целевой API (Slack, Telegram, Discord, etc.)")
	}

	steps = append(steps, "Предоставьте образец входящих данных")
	steps = append(steps, "Настройте аутентификацию")
	steps = append(steps, "Протестируйте интеграцию")

	return steps
}