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
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
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
	
	// Пробуем OpenRouter сначала
	if c.Config.OpenRouterAPIKey != "" {
		response, err := c.chatOpenRouter(ctx, messages, jsonMode)
		if err == nil {
			return response, nil
		}
		// Логируем ошибку OpenRouter, но продолжаем с fallback
		fmt.Printf("OpenRouter failed, falling back to OpenAI: %v\n", err)
	}
	
	// Fallback на OpenAI
	if c.Config.OpenAIAPIKey != "" {
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

	// Создаем HTTP запрос
	apiURL := c.Config.OpenRouterURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
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

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.OpenAIAPIKey)

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
	return []ModelInfo{
		{
			ID:          "anthropic/claude-3.5-sonnet",
			Name:        "Claude 3.5 Sonnet",
			Description: "Лучший для сложных задач программирования и анализа",
		},
		{
			ID:          "openai/gpt-4o",
			Name:        "GPT-4o",
			Description: "Быстрый и качественный, отлично для чата",
		},
		{
			ID:          "openai/gpt-4o-mini",
			Name:        "GPT-4o Mini",
			Description: "Дешевый и быстрый, хорош для простых задач",
		},
		{
			ID:          "meta-llama/llama-3.1-70b-instruct",
			Name:        "Llama 3.1 70B",
			Description: "Открытая модель, хорошее качество",
		},
		{
			ID:          "google/gemini-pro-1.5",
			Name:        "Gemini Pro 1.5",
			Description: "Отлично для анализа данных и больших контекстов",
		},
		{
			ID:          "anthropic/claude-3-haiku",
			Name:        "Claude 3 Haiku",
			Description: "Самый быстрый, подходит для простых задач",
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