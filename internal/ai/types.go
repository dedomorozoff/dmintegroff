package ai

import (
	"fmt"
	"time"
)

// ChatMessage представляет сообщение в чате с AI
type ChatMessage struct {
	Role    string `json:"role"`    // "user" или "assistant"
	Content string `json:"content"` // текст сообщения
}

// ChatRequest запрос к AI чату
type ChatRequest struct {
	Message     string                 `json:"message"`     // сообщение пользователя
	Context     map[string]interface{} `json:"context"`     // контекст (данные интеграции)
	History     []ChatMessage          `json:"history"`     // история чата
	SampleData  map[string]interface{} `json:"sample_data"` // образец данных для анализа
	TargetAPI   string                 `json:"target_api"`  // целевой API (slack, telegram, etc.)
}

// ChatResponse ответ от AI
type ChatResponse struct {
	Response    string                 `json:"response"`    // текстовый ответ
	Suggestions []AISuggestion         `json:"suggestions"` // предложения действий
	Mapping     *GeneratedMapping      `json:"mapping"`     // сгенерированный маппинг
	NextSteps   []string               `json:"next_steps"`  // следующие шаги
	Confidence  float64                `json:"confidence"`  // уверенность AI (0-1)
}

// AISuggestion предложение от AI
type AISuggestion struct {
	Type        string                 `json:"type"`        // "mapping", "api_config", "template"
	Title       string                 `json:"title"`       // заголовок предложения
	Description string                 `json:"description"` // описание
	Data        map[string]interface{} `json:"data"`        // данные предложения
	Confidence  float64                `json:"confidence"`  // уверенность (0-1)
}

// GeneratedMapping сгенерированный AI маппинг
type GeneratedMapping struct {
	Type         string                 `json:"type"`          // "json_template", "xml_template", "custom"
	Template     string                 `json:"template"`      // шаблон трансформации
	TargetURL    string                 `json:"target_url"`    // URL целевого API
	Method       string                 `json:"method"`        // HTTP метод
	Headers      map[string]string      `json:"headers"`       // заголовки
	AuthType     string                 `json:"auth_type"`     // тип аутентификации
	AuthConfig   map[string]interface{} `json:"auth_config"`   // настройки аутентификации
	Description  string                 `json:"description"`   // описание маппинга
	Reasoning    string                 `json:"reasoning"`     // объяснение логики AI
}

// DataAnalysisRequest запрос на анализ данных
type DataAnalysisRequest struct {
	Data   map[string]interface{} `json:"data"`   // данные для анализа
	Format string                 `json:"format"` // формат данных (json, xml, form)
}

// DataAnalysisResponse результат анализа данных
type DataAnalysisResponse struct {
	Fields      []FieldInfo    `json:"fields"`      // информация о полях
	Schema      string         `json:"schema"`      // схема данных
	Suggestions []FieldMapping `json:"suggestions"` // предложения маппинга
	DataType    string         `json:"data_type"`   // тип данных (order, user, event, etc.)
	Confidence  float64        `json:"confidence"`  // уверенность анализа
}

// FieldInfo информация о поле данных
type FieldInfo struct {
	Name        string   `json:"name"`        // имя поля
	Type        string   `json:"type"`        // тип (string, number, email, date, etc.)
	Required    bool     `json:"required"`    // обязательное поле
	Examples    []string `json:"examples"`    // примеры значений
	Description string   `json:"description"` // описание поля
	Pattern     string   `json:"pattern"`     // паттерн для валидации
}

// FieldMapping предложение маппинга поля
type FieldMapping struct {
	SourceField string  `json:"source_field"` // исходное поле
	TargetField string  `json:"target_field"` // целевое поле
	Transform   string  `json:"transform"`    // трансформация
	Confidence  float64 `json:"confidence"`   // уверенность предложения
	Reasoning   string  `json:"reasoning"`    // объяснение
}

// MappingGenerationRequest запрос на генерацию маппинга
type MappingGenerationRequest struct {
	SourceData map[string]interface{} `json:"source_data"` // исходные данные
	TargetAPI  string                 `json:"target_api"`  // целевой API
	Task       string                 `json:"task"`        // описание задачи
	UserPrompt string                 `json:"user_prompt"` // промпт пользователя
}

// CreateIntegrationRequest запрос на создание интеграции с помощью AI
type CreateIntegrationRequest struct {
	Description string `json:"description" binding:"required"` // описание интеграции
	SampleData  string `json:"sample_data"`                    // образец данных
	ProjectID   int    `json:"project_id"`                     // ID проекта
}

// CreatedIntegration результат создания интеграции
type CreatedIntegration struct {
	Name         string                 `json:"name"`          // название интеграции
	TargetURL    string                 `json:"target_url"`    // URL целевого API
	Method       string                 `json:"method"`        // HTTP метод
	Template     string                 `json:"template"`      // шаблон
	TemplateType string                 `json:"template_type"` // тип шаблона
	Mapping      map[string]string      `json:"mapping"`       // маппинг полей
	AuthType     string                 `json:"auth_type"`     // тип аутентификации
	AuthConfig   map[string]interface{} `json:"auth_config"`   // настройки аутентификации
	Headers      map[string]string      `json:"headers"`       // заголовки
	Explanation  string                 `json:"explanation"`   // объяснение
	NextSteps    []string               `json:"next_steps"`    // следующие шаги
}

// AIConfig конфигурация AI
type AIConfig struct {
	// Глобальные настройки
	Enabled bool `env:"AI_ENABLED" envDefault:"true"` // глобальное включение/выключение AI
	
	// OpenRouter настройки
	OpenRouterAPIKey string  `env:"OPENROUTER_API_KEY"`
	OpenRouterModel  string  `env:"OPENROUTER_MODEL" envDefault:"anthropic/claude-3.5-sonnet"`
	OpenRouterURL    string  `env:"OPENROUTER_URL" envDefault:"https://openrouter.ai/api/v1"`
	
	// Общие настройки
	MaxTokens       int     `env:"AI_MAX_TOKENS" envDefault:"4000"`
	Temperature     float32 `env:"AI_TEMPERATURE" envDefault:"0.3"`
	RequestTimeout  int     `env:"AI_REQUEST_TIMEOUT" envDefault:"30"` // секунды
	
	// Локальная LLM (опционально)
	LocalLLMEnabled bool   `env:"LOCAL_LLM_ENABLED" envDefault:"false"`
	LocalLLMURL     string `env:"LOCAL_LLM_URL" envDefault:"http://localhost:11434"`
	
	// Fallback на OpenAI (если OpenRouter недоступен)
	OpenAIAPIKey string `env:"OPENAI_API_KEY"`
	OpenAIModel  string `env:"OPENAI_MODEL" envDefault:"gpt-4-1106-preview"`
}

// AISession сессия чата с AI
type AISession struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Messages    []ChatMessage `json:"messages"`
	Context     map[string]interface{} `json:"context"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	IntegrationID *int        `json:"integration_id,omitempty"` // связанная интеграция
}

// PopularAPI информация о популярном API
type PopularAPI struct {
	Name         string            `json:"name"`          // название API
	BaseURL      string            `json:"base_url"`      // базовый URL
	AuthType     string            `json:"auth_type"`     // тип аутентификации
	CommonFields map[string]string `json:"common_fields"` // общие поля
	Templates    []APITemplate     `json:"templates"`     // шаблоны
	Description  string            `json:"description"`   // описание
}

// APITemplate шаблон для API
type APITemplate struct {
	Name         string                 `json:"name"`          // название шаблона
	Purpose      string                 `json:"purpose"`       // назначение
	Method       string                 `json:"method"`        // HTTP метод
	Path         string                 `json:"path"`          // путь API
	Headers      map[string]string      `json:"headers"`       // заголовки
	BodyTemplate string                 `json:"body_template"` // шаблон тела запроса
	RequiredFields []string             `json:"required_fields"` // обязательные поля
	Schema       map[string]interface{} `json:"schema"`        // схема данных
}

// IsConfigured проверяет, настроен ли AI
func (c *AIConfig) IsConfigured() bool {
	if !c.Enabled {
		return false
	}
	if c.LocalLLMEnabled && c.LocalLLMURL != "" {
		return true
	}
	return c.OpenRouterAPIKey != ""
}

// GetCurrentProvider возвращает текущего провайдера AI
func (c *AIConfig) GetCurrentProvider() string {
	if !c.Enabled {
		return "Disabled (AI_ENABLED=false)"
	}
	if c.LocalLLMEnabled && c.LocalLLMURL != "" {
		return "Local LLM"
	}
	if c.OpenRouterAPIKey != "" {
		return fmt.Sprintf("OpenRouter (%s)", c.OpenRouterModel)
	}
	return "Not configured"
}