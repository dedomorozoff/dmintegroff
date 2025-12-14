package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"dmintegroff/internal/ai"
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
)

// AIController контроллер для AI функций
type AIController struct {
	client    *ai.Client
	generator *ai.Generator
	analyzer  *ai.Analyzer
}

// NewAIController создает новый AI контроллер
func NewAIController(config *ai.AIConfig) *AIController {
	client := ai.NewClient(config)
	return &AIController{
		client:    client,
		generator: ai.NewGenerator(client),
		analyzer:  ai.NewAnalyzer(client),
	}
}

// Chat обрабатывает запросы к AI чату
func (c *AIController) Chat(ctx *gin.Context) {
	var req ai.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
			"details": err.Error(),
		})
		return
	}

	// Проверяем, настроен ли AI
	if !c.client.IsConfigured() {
		provider := c.client.GetCurrentProvider()
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

	// Создаем контекст с таймаутом
	requestCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	// Отправляем запрос к AI
	response, err := c.client.Chat(requestCtx, messages, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка AI",
			"details": err.Error(),
		})
		return
	}

	// Анализируем ответ и создаем предложения
	suggestions := c.createSuggestions(req.Message, req.SampleData, req.TargetAPI)
	
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
	var req ai.DataAnalysisRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
			"details": err.Error(),
		})
		return
	}

	// Создаем контекст с таймаутом
	requestCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Анализируем данные
	analysis, err := c.analyzer.AnalyzeDataStructure(requestCtx, req.Data, req.Format)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка анализа данных",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, analysis)
}

// GenerateMapping генерирует маппинг для интеграции
func (c *AIController) GenerateMapping(ctx *gin.Context) {
	var req ai.MappingGenerationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат запроса",
			"details": err.Error(),
		})
		return
	}

	// Создаем контекст с таймаутом
	requestCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Генерируем маппинг
	mapping, err := c.generator.GenerateMapping(requestCtx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка генерации маппинга",
			"details": err.Error(),
		})
		return
	}

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
	integration.OutputTemplate = mapping.Template
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
		"configured": c.client.IsConfigured(),
		"provider":   c.client.GetCurrentProvider(),
		"models":     c.client.GetAvailableModels(),
		"enabled":    c.client.Config.Enabled, // добавляем информацию о глобальном переключателе
	}

	// Проверяем доступность AI
	if c.client.IsConfigured() {
		testCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	// Создаем контекст с таймаутом
	requestCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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