package models

import (
	"time"
	"gorm.io/gorm"
)

// AISettings настройки AI ассистента
type AISettings struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// OpenRouter настройки
	OpenRouterAPIKey string `json:"openrouter_api_key" gorm:"type:text"`
	OpenRouterModel  string `json:"openrouter_model" gorm:"type:varchar(255);default:'anthropic/claude-3.5-sonnet'"`
	OpenRouterURL    string `json:"openrouter_url" gorm:"type:varchar(255);default:'https://openrouter.ai/api/v1'"`
	
	// OpenAI настройки (fallback)
	OpenAIAPIKey string `json:"openai_api_key" gorm:"type:text"`
	OpenAIModel  string `json:"openai_model" gorm:"type:varchar(255);default:'gpt-4-1106-preview'"`
	
	// Общие настройки
	MaxTokens      int     `json:"max_tokens" gorm:"default:4000"`
	Temperature    float32 `json:"temperature" gorm:"default:0.3"`
	RequestTimeout int     `json:"request_timeout" gorm:"default:30"` // секунды
	
	// Локальная LLM (опционально)
	LocalLLMEnabled bool   `json:"local_llm_enabled" gorm:"default:false"`
	LocalLLMURL     string `json:"local_llm_url" gorm:"type:varchar(255);default:'http://localhost:11434'"`
	
	// Метаданные
	IsActive bool `json:"is_active" gorm:"default:true"` // только одна запись может быть активной
}

// TableName возвращает имя таблицы
func (AISettings) TableName() string {
	return "ai_settings"
}

// GetActiveAISettings получает активные настройки AI
func GetActiveAISettings(db *gorm.DB) (*AISettings, error) {
	var settings AISettings
	err := db.Where("is_active = ?", true).First(&settings).Error
	if err != nil {
		// Если настроек нет, создаем дефолтные
		if err == gorm.ErrRecordNotFound {
			settings = AISettings{
				OpenRouterModel:  "anthropic/claude-3.5-sonnet",
				OpenRouterURL:    "https://openrouter.ai/api/v1",
				OpenAIModel:      "gpt-4-1106-preview",
				MaxTokens:        4000,
				Temperature:      0.3,
				RequestTimeout:   30,
				LocalLLMEnabled:  false,
				LocalLLMURL:      "http://localhost:11434",
				IsActive:         true,
			}
			if createErr := db.Create(&settings).Error; createErr != nil {
				return nil, createErr
			}
			return &settings, nil
		}
		return nil, err
	}
	return &settings, nil
}

// SaveAISettings сохраняет настройки AI
func SaveAISettings(db *gorm.DB, settings *AISettings) error {
	// Деактивируем все существующие настройки
	if err := db.Model(&AISettings{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
		return err
	}
	
	// Устанавливаем новые настройки как активные
	settings.IsActive = true
	
	// Если ID не указан, создаем новую запись
	if settings.ID == 0 {
		return db.Create(settings).Error
	}
	
	// Иначе обновляем существующую
	return db.Save(settings).Error
}

// IsConfigured проверяет, настроен ли AI
func (s *AISettings) IsConfigured() bool {
	if s.LocalLLMEnabled && s.LocalLLMURL != "" {
		return true
	}
	return s.OpenRouterAPIKey != "" || s.OpenAIAPIKey != ""
}

// GetCurrentProvider возвращает текущего провайдера AI
func (s *AISettings) GetCurrentProvider() string {
	if s.LocalLLMEnabled && s.LocalLLMURL != "" {
		return "Local LLM"
	}
	if s.OpenRouterAPIKey != "" {
		return "OpenRouter (" + s.OpenRouterModel + ")"
	}
	if s.OpenAIAPIKey != "" {
		return "OpenAI (" + s.OpenAIModel + ")"
	}
	return "Not configured"
}