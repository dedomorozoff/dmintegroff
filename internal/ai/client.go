package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"dmintegroff/internal/logger"
)

// Client клиент для работы с AI (OpenAI API)
type Client struct {
	Config     *AIConfig // публичное поле для доступа к конфигурации
	httpClient *http.Client
}

// NewClient создает новый AI клиент
func NewClient(config *AIConfig) *Client {
	return &Client{
		Config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.RequestTimeout) * time.Second,
		},
	}
}

// OpenAI API структуры
type openAIRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIMessage     `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float32             `json:"temperature"`
	ResponseFormat *responseFormat  `json:"response_format,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"` // "json_object"
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Code    interface{} `json:"code"` // может быть строкой или числом
	} `json:"error,omitempty"`
}

// Chat отправляет запрос в чат с AI
func (c *Client) Chat(ctx context.Context, messages []ChatMessage, jsonMode bool) (string, error) {
	// Проверяем глобальный переключатель AI
	if !c.Config.Enabled {
		return "", fmt.Errorf("AI functionality is disabled (AI_ENABLED=false)")
	}
	
	if c.Config.LocalLLMEnabled {
		return c.chatLocal(ctx, messages)
	}
	
	// Используем только OpenRouter
	if c.Config.OpenRouterAPIKey != "" {
		return c.chatOpenRouter(ctx, messages, jsonMode)
	}
	
	return "", fmt.Errorf("OpenRouter API key not configured")
}

// chatOpenRouter отправляет запрос в OpenRouter API
func (c *Client) chatOpenRouter(ctx context.Context, messages []ChatMessage, jsonMode bool) (string, error) {
	startTime := time.Now()
	
	// Конвертируем сообщения в формат OpenAI (OpenRouter использует тот же формат)
	openAIMessages := make([]openAIMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Создаем запрос
	request := openAIRequest{
		Model:       c.Config.OpenRouterModel,
		Messages:    openAIMessages,
		MaxTokens:   c.Config.MaxTokens,
		Temperature: c.Config.Temperature,
	}

	// Включаем JSON режим если нужно (не все модели поддерживают)
	if jsonMode && c.supportsJSONMode(c.Config.OpenRouterModel) {
		request.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	// Сериализуем запрос
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Логируем запрос к AI
	logger.Log.WithFields(map[string]interface{}{
		"action":       "ai_request",
		"provider":     "openrouter",
		"model":        c.Config.OpenRouterModel,
		"json_mode":    jsonMode,
		"max_tokens":   c.Config.MaxTokens,
		"temperature":  c.Config.Temperature,
		"message_count": len(messages),
		"request_size": len(requestBody),
		"request_preview": func() string {
			// Логируем только первые 500 символов запроса для безопасности
			if len(requestBody) > 500 {
				return string(requestBody[:500]) + "..."
			}
			return string(requestBody)
		}(),
	}).Info("AI Client: Sending request to OpenRouter")

	// Создаем HTTP запрос
	apiURL := c.Config.OpenRouterURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openrouter",
			"error": "http_request_creation_failed",
			"details": err.Error(),
		}).Error("AI Client: Failed to create HTTP request")
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки для OpenRouter
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.OpenRouterAPIKey)
	req.Header.Set("HTTP-Referer", "https://dmintegroff.com") // Для статистики OpenRouter
	req.Header.Set("X-Title", "dmIntegroff AI Assistant")

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openrouter",
			"error": "http_request_failed",
			"details": err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: HTTP request failed")
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openrouter",
			"error": "response_read_failed",
			"details": err.Error(),
			"status_code": resp.StatusCode,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Failed to read response")
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ
	var openAIResp openAIResponse
	if err := json.Unmarshal(responseBody, &openAIResp); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openrouter",
			"error": "response_parse_failed",
			"details": err.Error(),
			"status_code": resp.StatusCode,
			"response_size": len(responseBody),
			"response_preview": func() string {
				if len(responseBody) > 500 {
					return string(responseBody[:500]) + "..."
				}
				return string(responseBody)
			}(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Failed to parse response")
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Проверяем на ошибки
	if openAIResp.Error != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openrouter",
			"error": "api_error",
			"api_error_type": openAIResp.Error.Type,
			"api_error_message": openAIResp.Error.Message,
			"api_error_code": openAIResp.Error.Code,
			"status_code": resp.StatusCode,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: OpenRouter API returned error")
		return "", fmt.Errorf("OpenRouter API error: %s", openAIResp.Error.Message)
	}

	// Проверяем наличие ответа
	if len(openAIResp.Choices) == 0 {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openrouter",
			"error": "no_choices",
			"status_code": resp.StatusCode,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: No response choices from OpenRouter")
		return "", fmt.Errorf("no response from OpenRouter")
	}

	responseContent := openAIResp.Choices[0].Message.Content

	// Логируем успешный ответ
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_response",
		"provider": "openrouter",
		"model": c.Config.OpenRouterModel,
		"status_code": resp.StatusCode,
		"response_length": len(responseContent),
		"response_preview": func() string {
			if len(responseContent) > 300 {
				return responseContent[:300] + "..."
			}
			return responseContent
		}(),
		"usage": map[string]interface{}{
			"prompt_tokens": openAIResp.Usage.PromptTokens,
			"completion_tokens": openAIResp.Usage.CompletionTokens,
			"total_tokens": openAIResp.Usage.TotalTokens,
		},
		"finish_reason": openAIResp.Choices[0].FinishReason,
		"duration": time.Since(startTime).String(),
	}).Info("AI Client: Received successful response from OpenRouter")

	return responseContent, nil
}

// chatOpenAI отправляет запрос в OpenAI API (fallback)
func (c *Client) chatOpenAI(ctx context.Context, messages []ChatMessage, jsonMode bool) (string, error) {
	startTime := time.Now()
	
	// Конвертируем сообщения в формат OpenAI
	openAIMessages := make([]openAIMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Создаем запрос
	request := openAIRequest{
		Model:       c.Config.OpenAIModel,
		Messages:    openAIMessages,
		MaxTokens:   c.Config.MaxTokens,
		Temperature: c.Config.Temperature,
	}

	// Включаем JSON режим если нужно
	if jsonMode {
		request.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	// Сериализуем запрос
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Логируем запрос к AI
	logger.Log.WithFields(map[string]interface{}{
		"action":       "ai_request",
		"provider":     "openai",
		"model":        c.Config.OpenAIModel,
		"json_mode":    jsonMode,
		"max_tokens":   c.Config.MaxTokens,
		"temperature":  c.Config.Temperature,
		"message_count": len(messages),
		"request_size": len(requestBody),
		"request_preview": func() string {
			// Логируем только первые 500 символов запроса для безопасности
			if len(requestBody) > 500 {
				return string(requestBody[:500]) + "..."
			}
			return string(requestBody)
		}(),
	}).Info("AI Client: Sending request to OpenAI")

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openai",
			"error": "http_request_creation_failed",
			"details": err.Error(),
		}).Error("AI Client: Failed to create HTTP request")
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.OpenAIAPIKey)

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openai",
			"error": "http_request_failed",
			"details": err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: HTTP request failed")
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openai",
			"error": "response_read_failed",
			"details": err.Error(),
			"status_code": resp.StatusCode,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Failed to read response")
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ
	var openAIResp openAIResponse
	if err := json.Unmarshal(responseBody, &openAIResp); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openai",
			"error": "response_parse_failed",
			"details": err.Error(),
			"status_code": resp.StatusCode,
			"response_size": len(responseBody),
			"response_preview": func() string {
				if len(responseBody) > 500 {
					return string(responseBody[:500]) + "..."
				}
				return string(responseBody)
			}(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Failed to parse response")
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Проверяем на ошибки
	if openAIResp.Error != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openai",
			"error": "api_error",
			"api_error_type": openAIResp.Error.Type,
			"api_error_message": openAIResp.Error.Message,
			"api_error_code": openAIResp.Error.Code,
			"status_code": resp.StatusCode,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: OpenAI API returned error")
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	// Проверяем наличие ответа
	if len(openAIResp.Choices) == 0 {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_request",
			"provider": "openai",
			"error": "no_choices",
			"status_code": resp.StatusCode,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: No response choices from OpenAI")
		return "", fmt.Errorf("no response from OpenAI")
	}

	responseContent := openAIResp.Choices[0].Message.Content

	// Логируем успешный ответ
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_response",
		"provider": "openai",
		"model": c.Config.OpenAIModel,
		"status_code": resp.StatusCode,
		"response_length": len(responseContent),
		"response_preview": func() string {
			if len(responseContent) > 300 {
				return responseContent[:300] + "..."
			}
			return responseContent
		}(),
		"usage": map[string]interface{}{
			"prompt_tokens": openAIResp.Usage.PromptTokens,
			"completion_tokens": openAIResp.Usage.CompletionTokens,
			"total_tokens": openAIResp.Usage.TotalTokens,
		},
		"finish_reason": openAIResp.Choices[0].FinishReason,
		"duration": time.Since(startTime).String(),
	}).Info("AI Client: Received successful response from OpenAI")

	return responseContent, nil
}

// chatLocal отправляет запрос в локальную LLM (заглушка для будущего)
func (c *Client) chatLocal(ctx context.Context, messages []ChatMessage) (string, error) {
	// TODO: Реализовать поддержку локальных LLM (Ollama, etc.)
	return "", fmt.Errorf("local LLM not implemented yet")
}

// AnalyzeData анализирует структуру данных
func (c *Client) AnalyzeData(ctx context.Context, data map[string]interface{}, format string) (*DataAnalysisResponse, error) {
	startTime := time.Now()
	
	// Логируем начало анализа данных
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_analyze_data_start",
		"format": format,
		"data_fields": len(data),
		"data_preview": func() string {
			dataJSON, _ := json.Marshal(data)
			if len(dataJSON) > 200 {
				return string(dataJSON[:200]) + "..."
			}
			return string(dataJSON)
		}(),
	}).Info("AI Client: Starting data analysis")
	
	// Создаем промпт для анализа данных
	prompt := c.createDataAnalysisPrompt(data, format)
	
	messages := []ChatMessage{
		{Role: "system", Content: getSystemPrompt("data_analysis")},
		{Role: "user", Content: prompt},
	}

	// Логируем промпт
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_analyze_data_prompt",
		"prompt_length": len(prompt),
		"prompt_preview": func() string {
			if len(prompt) > 300 {
				return prompt[:300] + "..."
			}
			return prompt
		}(),
	}).Info("AI Client: Generated analysis prompt")

	// Отправляем запрос с JSON режимом
	response, err := c.Chat(ctx, messages, true)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_analyze_data",
			"error": "chat_request_failed",
			"details": err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Data analysis chat request failed")
		return nil, fmt.Errorf("failed to analyze data: %w", err)
	}

	// Логируем полученный ответ
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_analyze_data_response",
		"response_length": len(response),
		"response_preview": func() string {
			if len(response) > 300 {
				return response[:300] + "..."
			}
			return response
		}(),
	}).Info("AI Client: Received analysis response")

	// Парсим JSON ответ
	var analysis DataAnalysisResponse
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_analyze_data",
			"error": "json_parse_failed",
			"details": err.Error(),
			"response": response,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Failed to parse analysis response")
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	// Логируем успешный результат
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_analyze_data_success",
		"data_type": analysis.DataType,
		"confidence": analysis.Confidence,
		"fields_count": len(analysis.Fields),
		"suggestions_count": len(analysis.Suggestions),
		"duration": time.Since(startTime).String(),
	}).Info("AI Client: Data analysis completed successfully")

	return &analysis, nil
}

// GenerateMapping генерирует маппинг для интеграции
func (c *Client) GenerateMapping(ctx context.Context, req *MappingGenerationRequest) (*GeneratedMapping, error) {
	startTime := time.Now()
	
	// Логируем начало генерации маппинга
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_generate_mapping_start",
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
	}).Info("AI Client: Starting mapping generation")
	
	// Создаем промпт для генерации маппинга
	prompt := c.createMappingPrompt(req)
	
	messages := []ChatMessage{
		{Role: "system", Content: getSystemPrompt("mapping_generation")},
		{Role: "user", Content: prompt},
	}

	// Логируем промпт
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_generate_mapping_prompt",
		"prompt_length": len(prompt),
		"prompt_preview": func() string {
			if len(prompt) > 300 {
				return prompt[:300] + "..."
			}
			return prompt
		}(),
	}).Info("AI Client: Generated mapping prompt")

	// Отправляем запрос с JSON режимом
	response, err := c.Chat(ctx, messages, true)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_generate_mapping",
			"error": "chat_request_failed",
			"details": err.Error(),
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Mapping generation chat request failed")
		return nil, fmt.Errorf("failed to generate mapping: %w", err)
	}

	// Логируем полученный ответ
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_generate_mapping_response",
		"response_length": len(response),
		"response_preview": func() string {
			if len(response) > 300 {
				return response[:300] + "..."
			}
			return response
		}(),
	}).Info("AI Client: Received mapping response")

	// Парсим JSON ответ
	var mapping GeneratedMapping
	if err := json.Unmarshal([]byte(response), &mapping); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"action": "ai_generate_mapping",
			"error": "json_parse_failed",
			"details": err.Error(),
			"response": response,
			"duration": time.Since(startTime).String(),
		}).Error("AI Client: Failed to parse mapping response")
		return nil, fmt.Errorf("failed to parse mapping response: %w", err)
	}

	// Логируем успешный результат
	logger.Log.WithFields(map[string]interface{}{
		"action": "ai_generate_mapping_success",
		"mapping_type": mapping.Type,
		"target_url": mapping.TargetURL,
		"method": mapping.Method,
		"auth_type": mapping.AuthType,
		"template_length": len(mapping.Template),
		"template_preview": func() string {
			if len(mapping.Template) > 150 {
				return mapping.Template[:150] + "..."
			}
			return mapping.Template
		}(),
		"duration": time.Since(startTime).String(),
	}).Info("AI Client: Mapping generation completed successfully")

	return &mapping, nil
}

// createDataAnalysisPrompt создает промпт для анализа данных
func (c *Client) createDataAnalysisPrompt(data map[string]interface{}, format string) string {
	dataJSON, _ := json.MarshalIndent(data, "", "  ")
	
	return fmt.Sprintf(`Проанализируй структуру данных и верни результат в JSON формате.

Данные (%s формат):
%s

Верни JSON с полями:
- fields: массив объектов с информацией о каждом поле (name, type, required, examples, description)
- schema: строка с описанием схемы
- suggestions: массив предложений маппинга полей
- data_type: тип данных (order, user, event, notification, etc.)
- confidence: уверенность анализа (0-1)

Типы полей: string, number, boolean, email, phone, date, datetime, url, array, object`, format, string(dataJSON))
}

// createMappingPrompt создает промпт для генерации маппинга
func (c *Client) createMappingPrompt(req *MappingGenerationRequest) string {
	sourceJSON, _ := json.MarshalIndent(req.SourceData, "", "  ")
	
	return fmt.Sprintf(`Создай маппинг для интеграции webhook и верни результат в JSON формате.

Задача: %s
Целевой API: %s
Промпт пользователя: %s

Исходные данные:
%s

Верни JSON с полями:
- type: тип маппинга (json_template, xml_template, custom)
- template: шаблон трансформации с {{field}} плейсхолдерами
- target_url: URL целевого API (если известен)
- method: HTTP метод (POST, PUT, etc.)
- headers: объект с заголовками
- auth_type: тип аутентификации (bearer, oauth, basic, none)
- auth_config: настройки аутентификации
- description: описание маппинга
- reasoning: объяснение логики

Для популярных API (Slack, Telegram, Discord) используй правильные URL и форматы.`, 
		req.Task, req.TargetAPI, req.UserPrompt, string(sourceJSON))
}

// supportsJSONMode проверяет, поддерживает ли модель JSON режим
func (c *Client) supportsJSONMode(model string) bool {
	// Модели, которые поддерживают JSON режим
	jsonSupportedModels := []string{
		"openai/gpt-4-1106-preview",
		"openai/gpt-4-turbo-preview", 
		"openai/gpt-3.5-turbo-1106",
		"openai/gpt-4o",
		"openai/gpt-4o-mini",
	}
	
	for _, supportedModel := range jsonSupportedModels {
		if model == supportedModel {
			return true
		}
	}
	return false
}

// ModelInfo информация о модели
type ModelInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Pricing     *struct {
		Prompt     interface{} `json:"prompt"`     // может быть строкой или числом
		Completion interface{} `json:"completion"` // может быть строкой или числом
	} `json:"pricing,omitempty"`
	ContextLength int      `json:"context_length"`
	Architecture  *struct {
		Modality    string `json:"modality"`
		Tokenizer   string `json:"tokenizer"`
		InstructType string `json:"instruct_type"`
	} `json:"architecture,omitempty"`
	TopProvider *struct {
		MaxCompletionTokens int `json:"max_completion_tokens"`
	} `json:"top_provider,omitempty"`
}

// GetPromptPrice возвращает цену за prompt токены как float64
func (m *ModelInfo) GetPromptPrice() float64 {
	if m.Pricing == nil || m.Pricing.Prompt == nil {
		return 0
	}
	
	switch v := m.Pricing.Prompt.(type) {
	case float64:
		return v
	case string:
		if price, err := strconv.ParseFloat(v, 64); err == nil {
			return price
		}
	case int:
		return float64(v)
	}
	return 0
}

// GetCompletionPrice возвращает цену за completion токены как float64
func (m *ModelInfo) GetCompletionPrice() float64 {
	if m.Pricing == nil || m.Pricing.Completion == nil {
		return 0
	}
	
	switch v := m.Pricing.Completion.(type) {
	case float64:
		return v
	case string:
		if price, err := strconv.ParseFloat(v, 64); err == nil {
			return price
		}
	case int:
		return float64(v)
	}
	return 0
}

// IsFree проверяет, является ли модель бесплатной
func (m *ModelInfo) IsFree() bool {
	return m.GetPromptPrice() == 0 && m.GetCompletionPrice() == 0
}

// OpenRouterModelsResponse ответ от OpenRouter API с моделями
type OpenRouterModelsResponse struct {
	Data []ModelInfo `json:"data"`
}

// GetAvailableModels возвращает список рекомендуемых моделей (статический fallback)
func (c *Client) GetAvailableModels() []string {
	return []string{
		"anthropic/claude-3.5-sonnet",     // Лучший для сложных задач
		"openai/gpt-4o",                   // Быстрый и качественный
		"openai/gpt-4o-mini",              // Дешевый и быстрый
		"meta-llama/llama-3.1-70b-instruct", // Открытая модель
		"google/gemini-pro-1.5",           // Хорошо для анализа данных
		"anthropic/claude-3-haiku",        // Самый быстрый
	}
}

// GetOpenRouterModels получает список доступных моделей от OpenRouter API
func (c *Client) GetOpenRouterModels(ctx context.Context) ([]ModelInfo, error) {
	if !c.Config.Enabled {
		return nil, fmt.Errorf("AI functionality is disabled (AI_ENABLED=false)")
	}
	if c.Config.OpenRouterAPIKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	// Создаем HTTP запрос
	apiURL := c.Config.OpenRouterURL + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", "Bearer "+c.Config.OpenRouterAPIKey)
	req.Header.Set("HTTP-Referer", "https://dmintegroff.com")
	req.Header.Set("X-Title", "dmIntegroff AI Assistant")

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ
	var modelsResp OpenRouterModelsResponse
	if err := json.Unmarshal(responseBody, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return modelsResp.Data, nil
}

// GetRecommendedModels возвращает рекомендуемые модели с описаниями
func (c *Client) GetRecommendedModels() []ModelInfo {
	// Только бесплатные модели для избежания проблем с кредитами
	return []ModelInfo{
		{
			ID:          "meta-llama/llama-3.2-3b-instruct:free",
			Name:        "Llama 3.2 3B (Бесплатно)",
			Description: "Быстрая бесплатная модель для простых задач",
		},
		{
			ID:          "meta-llama/llama-3.1-8b-instruct:free",
			Name:        "Llama 3.1 8B (Бесплатно)",
			Description: "Хорошая бесплатная модель для программирования",
		},
		{
			ID:          "google/gemma-2-9b-it:free",
			Name:        "Gemma 2 9B (Бесплатно)",
			Description: "Бесплатная модель Google для общих задач",
		},
		{
			ID:          "microsoft/phi-3-mini-128k-instruct:free",
			Name:        "Phi-3 Mini (Бесплатно)",
			Description: "Компактная бесплатная модель Microsoft",
		},
		{
			ID:          "qwen/qwen-2-7b-instruct:free",
			Name:        "Qwen 2 7B (Бесплатно)",
			Description: "Бесплатная модель для многоязычных задач",
		},
	}
}

// IsConfigured проверяет, настроен ли AI клиент
func (c *Client) IsConfigured() bool {
	return c.Config.IsConfigured()
}

// GetCurrentProvider возвращает текущего провайдера AI
func (c *Client) GetCurrentProvider() string {
	return c.Config.GetCurrentProvider()
}