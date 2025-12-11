package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client клиент для работы с AI (OpenAI API)
type Client struct {
	config     *AIConfig
	httpClient *http.Client
}

// NewClient создает новый AI клиент
func NewClient(config *AIConfig) *Client {
	return &Client{
		config: config,
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
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Chat отправляет запрос в чат с AI
func (c *Client) Chat(ctx context.Context, messages []ChatMessage, jsonMode bool) (string, error) {
	if c.config.LocalLLMEnabled {
		return c.chatLocal(ctx, messages)
	}
	
	// Пробуем OpenRouter сначала
	if c.config.OpenRouterAPIKey != "" {
		response, err := c.chatOpenRouter(ctx, messages, jsonMode)
		if err == nil {
			return response, nil
		}
		// Логируем ошибку OpenRouter, но продолжаем с fallback
		fmt.Printf("OpenRouter failed, falling back to OpenAI: %v\n", err)
	}
	
	// Fallback на OpenAI
	if c.config.OpenAIAPIKey != "" {
		return c.chatOpenAI(ctx, messages, jsonMode)
	}
	
	return "", fmt.Errorf("no AI provider configured")
}

// chatOpenRouter отправляет запрос в OpenRouter API
func (c *Client) chatOpenRouter(ctx context.Context, messages []ChatMessage, jsonMode bool) (string, error) {
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
		Model:       c.config.OpenRouterModel,
		Messages:    openAIMessages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	// Включаем JSON режим если нужно (не все модели поддерживают)
	if jsonMode && c.supportsJSONMode(c.config.OpenRouterModel) {
		request.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	// Сериализуем запрос
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем HTTP запрос
	apiURL := c.config.OpenRouterURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки для OpenRouter
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.OpenRouterAPIKey)
	req.Header.Set("HTTP-Referer", "https://dmintegroff.com") // Для статистики OpenRouter
	req.Header.Set("X-Title", "dmIntegroff AI Assistant")

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ
	var openAIResp openAIResponse
	if err := json.Unmarshal(responseBody, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Проверяем на ошибки
	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenRouter API error: %s", openAIResp.Error.Message)
	}

	// Проверяем наличие ответа
	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenRouter")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// chatOpenAI отправляет запрос в OpenAI API (fallback)
func (c *Client) chatOpenAI(ctx context.Context, messages []ChatMessage, jsonMode bool) (string, error) {
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
		Model:       c.config.OpenAIModel,
		Messages:    openAIMessages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
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

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.OpenAIAPIKey)

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ
	var openAIResp openAIResponse
	if err := json.Unmarshal(responseBody, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Проверяем на ошибки
	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	// Проверяем наличие ответа
	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// chatLocal отправляет запрос в локальную LLM (заглушка для будущего)
func (c *Client) chatLocal(ctx context.Context, messages []ChatMessage) (string, error) {
	// TODO: Реализовать поддержку локальных LLM (Ollama, etc.)
	return "", fmt.Errorf("local LLM not implemented yet")
}

// AnalyzeData анализирует структуру данных
func (c *Client) AnalyzeData(ctx context.Context, data map[string]interface{}, format string) (*DataAnalysisResponse, error) {
	// Создаем промпт для анализа данных
	prompt := c.createDataAnalysisPrompt(data, format)
	
	messages := []ChatMessage{
		{Role: "system", Content: getSystemPrompt("data_analysis")},
		{Role: "user", Content: prompt},
	}

	// Отправляем запрос с JSON режимом
	response, err := c.Chat(ctx, messages, true)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze data: %w", err)
	}

	// Парсим JSON ответ
	var analysis DataAnalysisResponse
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	return &analysis, nil
}

// GenerateMapping генерирует маппинг для интеграции
func (c *Client) GenerateMapping(ctx context.Context, req *MappingGenerationRequest) (*GeneratedMapping, error) {
	// Создаем промпт для генерации маппинга
	prompt := c.createMappingPrompt(req)
	
	messages := []ChatMessage{
		{Role: "system", Content: getSystemPrompt("mapping_generation")},
		{Role: "user", Content: prompt},
	}

	// Отправляем запрос с JSON режимом
	response, err := c.Chat(ctx, messages, true)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mapping: %w", err)
	}

	// Парсим JSON ответ
	var mapping GeneratedMapping
	if err := json.Unmarshal([]byte(response), &mapping); err != nil {
		return nil, fmt.Errorf("failed to parse mapping response: %w", err)
	}

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

// GetAvailableModels возвращает список рекомендуемых моделей
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

// IsConfigured проверяет, настроен ли AI клиент
func (c *Client) IsConfigured() bool {
	if c.config.LocalLLMEnabled {
		return c.config.LocalLLMURL != ""
	}
	
	// Проверяем OpenRouter или OpenAI
	return c.config.OpenRouterAPIKey != "" || c.config.OpenAIAPIKey != ""
}

// GetCurrentProvider возвращает текущего провайдера AI
func (c *Client) GetCurrentProvider() string {
	if c.config.LocalLLMEnabled && c.config.LocalLLMURL != "" {
		return "Local LLM"
	}
	if c.config.OpenRouterAPIKey != "" {
		return fmt.Sprintf("OpenRouter (%s)", c.config.OpenRouterModel)
	}
	if c.config.OpenAIAPIKey != "" {
		return fmt.Sprintf("OpenAI (%s)", c.config.OpenAIModel)
	}
	return "Not configured"
}