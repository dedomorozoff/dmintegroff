package ai

import (
	"fmt"
	"strings"
)

// SystemPrompts содержит системные промпты для разных задач
var SystemPrompts = map[string]string{
	"chat": `You are an AI assistant for creating webhook integrations in the dmIntegroff system.

Your tasks:
1. Understand integration requirements in Russian/English
2. Analyze incoming data structure (JSON, XML, form-data)
3. Create mappings for popular APIs (Slack, Telegram, Discord, CRM systems)
4. Suggest optimal authentication and configuration settings

Working principles:
- Answer concisely and to the point
- Always suggest concrete solutions
- Explain your logic in simple terms
- Ask clarifying questions if additional information is needed
- Suggest next steps to complete the setup

Popular scenarios:
- Notifications to Slack/Telegram/Discord
- CRM integration (creating leads, updating contacts)
- Email sending via SendGrid/Mailgun
- Webhooks for analytics and monitoring
- Data synchronization between systems`,

	"data_analysis": `You are an expert in data structure analysis for webhook integrations.

Your tasks:
1. Analyze incoming data structure
2. Determine field types and their purpose
3. Suggest optimal mappings for popular APIs
4. Identify data type (order, user, event, notification)

Field types to identify:
- string: regular text
- number: numeric values
- boolean: true/false
- email: email addresses
- phone: phone numbers
- date: dates (YYYY-MM-DD)
- datetime: date and time (ISO 8601)
- url: URL addresses
- array: data arrays
- object: nested objects

Always return result in JSON format with fields:
- fields: array of field information
- schema: data schema description
- suggestions: mapping suggestions
- data_type: data type
- confidence: analysis confidence (0-1)`,

	"mapping_generation": `You are an expert in creating mappings for webhook integrations.

Your tasks:
1. Create mapping between incoming data and target API
2. Configure proper authentication
3. Choose optimal data format (JSON, XML, form-data)
4. Suggest data transformations if needed

Popular APIs and their formats:

SLACK WEBHOOK:
- URL: https://hooks.slack.com/services/...
- Method: POST
- Format: JSON
- Fields: text (required), channel, username, attachments
- Example: {"text": "Message", "channel": "#general"}

TELEGRAM BOT:
- URL: https://api.telegram.org/bot{token}/sendMessage
- Method: POST
- Auth: Bearer token in URL
- Fields: chat_id (required), text (required), parse_mode
- Example: {"chat_id": "123456", "text": "Message"}

DISCORD WEBHOOK:
- URL: https://discord.com/api/webhooks/...
- Method: POST
- Format: JSON
- Fields: content (required), username, avatar_url
- Example: {"content": "Message", "username": "Bot"}

Always return result in JSON format with fields:
- type: mapping type
- template: template with {{field}} placeholders
- target_url: API URL
- method: HTTP method
- headers: headers
- auth_type: authentication type
- auth_config: authentication settings
- description: description
- reasoning: logic explanation`,
}

// getSystemPrompt возвращает системный промпт для задачи
func getSystemPrompt(task string) string {
	if prompt, exists := SystemPrompts[task]; exists {
		return prompt
	}
	return SystemPrompts["chat"] // fallback на общий промпт
}

// PromptTemplate шаблон промпта
type PromptTemplate struct {
	Name         string
	SystemPrompt string
	UserTemplate string
	Examples     []PromptExample
}

// PromptExample пример промпта
type PromptExample struct {
	Input  string
	Output string
}

// ChatPromptBuilder строитель промптов для чата
type ChatPromptBuilder struct {
	context map[string]interface{}
}

// NewChatPromptBuilder создает новый строитель промптов
func NewChatPromptBuilder() *ChatPromptBuilder {
	return &ChatPromptBuilder{
		context: make(map[string]interface{}),
	}
}

// WithContext добавляет контекст
func (b *ChatPromptBuilder) WithContext(key string, value interface{}) *ChatPromptBuilder {
	b.context[key] = value
	return b
}

// BuildChatPrompt создает промпт для чата
func (b *ChatPromptBuilder) BuildChatPrompt(userMessage string, sampleData map[string]interface{}) string {
	var parts []string
	
	// Добавляем сообщение пользователя
	parts = append(parts, fmt.Sprintf("User: %s", userMessage))
	
	// Добавляем образец данных если есть
	if len(sampleData) > 0 {
		parts = append(parts, "\nSample incoming data:")
		for key, value := range sampleData {
			parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
		}
	}
	
	// Добавляем контекст если есть
	if len(b.context) > 0 {
		parts = append(parts, "\nContext:")
		for key, value := range b.context {
			parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
		}
	}
	
	parts = append(parts, "\nPlease help create the integration. If you need additional information - ask questions.")
	
	return strings.Join(parts, "\n")
}

// BuildDataAnalysisPrompt создает промпт для анализа данных
func (b *ChatPromptBuilder) BuildDataAnalysisPrompt(data map[string]interface{}, format string) string {
	return fmt.Sprintf(`Analyze the data structure and return the result in JSON format.

Data (%s format):
%s

Return JSON with fields:
{
  "fields": [
    {
      "name": "field_name",
      "type": "field_type",
      "required": true/false,
      "examples": ["example1", "example2"],
      "description": "field description"
    }
  ],
  "schema": "data schema description",
  "suggestions": [
    {
      "source_field": "source_field",
      "target_field": "target_field", 
      "transform": "transformation",
      "confidence": 0.95,
      "reasoning": "explanation"
    }
  ],
  "data_type": "data type (order, user, event, etc.)",
  "confidence": 0.9
}

Field types: string, number, boolean, email, phone, date, datetime, url, array, object`, 
		format, formatDataForPrompt(data))
}

// BuildMappingPrompt создает промпт для генерации маппинга
func (b *ChatPromptBuilder) BuildMappingPrompt(sourceData map[string]interface{}, targetAPI, task, userPrompt string) string {
	return fmt.Sprintf(`Create a webhook integration mapping and return the result in JSON format.

Task: %s
Target API: %s
User prompt: %s

Source data:
%s

Return JSON with fields:
{
  "type": "json_template",
  "template": "template with {{field}} placeholders",
  "target_url": "target API URL",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "auth_type": "bearer|oauth|basic|none",
  "auth_config": {
    "token_field": "token field"
  },
  "description": "mapping description",
  "reasoning": "logic explanation"
}

For popular APIs use correct URLs and formats:
- Slack: hooks.slack.com/services/...
- Telegram: api.telegram.org/bot{token}/sendMessage  
- Discord: discord.com/api/webhooks/...`, 
		task, targetAPI, userPrompt, formatDataForPrompt(sourceData))
}

// formatDataForPrompt форматирует данные для промпта
func formatDataForPrompt(data map[string]interface{}) string {
	if len(data) == 0 {
		return "No data"
	}
	
	var parts []string
	for key, value := range data {
		parts = append(parts, fmt.Sprintf("  %s: %v", key, value))
	}
	
	return "{\n" + strings.Join(parts, ",\n") + "\n}"
}

