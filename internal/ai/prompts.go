package ai

import (
	"fmt"
	"strings"
)

// SystemPrompts содержит системные промпты для разных задач
var SystemPrompts = map[string]string{
	"chat": `Ты AI-ассистент для создания webhook интеграций в системе dmIntegroff.

Твоя задача:
1. Понимать задачи интеграции на русском/английском языке
2. Анализировать структуру входящих данных (JSON, XML, form-data)
3. Создавать маппинги для популярных API (Slack, Telegram, Discord, CRM системы)
4. Предлагать оптимальные настройки аутентификации и конфигурации

Принципы работы:
- Отвечай кратко и по делу
- Всегда предлагай конкретные решения
- Объясняй свою логику простыми словами
- Если нужна дополнительная информация - задавай уточняющие вопросы
- Предлагай следующие шаги для завершения настройки

Популярные сценарии:
- Уведомления в Slack/Telegram/Discord
- Интеграция с CRM (создание лидов, обновление контактов)
- Отправка email через SendGrid/Mailgun
- Webhook'и для аналитики и мониторинга
- Синхронизация данных между системами`,

	"data_analysis": `Ты эксперт по анализу структуры данных для webhook интеграций.

Твоя задача:
1. Проанализировать структуру входящих данных
2. Определить типы полей и их назначение
3. Предложить оптимальные маппинги для популярных API
4. Определить тип данных (заказ, пользователь, событие, уведомление)

Типы полей для определения:
- string: обычный текст
- number: числовые значения
- boolean: true/false
- email: email адреса
- phone: телефонные номера
- date: даты (YYYY-MM-DD)
- datetime: дата и время (ISO 8601)
- url: URL адреса
- array: массивы данных
- object: вложенные объекты

Всегда возвращай результат в JSON формате с полями:
- fields: массив информации о полях
- schema: описание схемы данных
- suggestions: предложения маппинга
- data_type: тип данных
- confidence: уверенность анализа (0-1)`,

	"mapping_generation": `Ты эксперт по созданию маппингов для webhook интеграций.

Твоя задача:
1. Создать маппинг между входящими данными и целевым API
2. Настроить правильную аутентификацию
3. Выбрать оптимальный формат данных (JSON, XML, form-data)
4. Предложить трансформации данных если нужно

Популярные API и их форматы:

SLACK WEBHOOK:
- URL: https://hooks.slack.com/services/...
- Method: POST
- Format: JSON
- Fields: text (обязательно), channel, username, attachments
- Example: {"text": "Сообщение", "channel": "#general"}

TELEGRAM BOT:
- URL: https://api.telegram.org/bot{token}/sendMessage
- Method: POST
- Auth: Bearer token в URL
- Fields: chat_id (обязательно), text (обязательно), parse_mode
- Example: {"chat_id": "123456", "text": "Сообщение"}

DISCORD WEBHOOK:
- URL: https://discord.com/api/webhooks/...
- Method: POST
- Format: JSON
- Fields: content (обязательно), username, avatar_url
- Example: {"content": "Сообщение", "username": "Bot"}

Всегда возвращай результат в JSON формате с полями:
- type: тип маппинга
- template: шаблон с {{field}} плейсхолдерами
- target_url: URL API
- method: HTTP метод
- headers: заголовки
- auth_type: тип аутентификации
- auth_config: настройки аутентификации
- description: описание
- reasoning: объяснение логики`,
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
	parts = append(parts, fmt.Sprintf("Пользователь: %s", userMessage))
	
	// Добавляем образец данных если есть
	if len(sampleData) > 0 {
		parts = append(parts, "\nОбразец входящих данных:")
		for key, value := range sampleData {
			parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
		}
	}
	
	// Добавляем контекст если есть
	if len(b.context) > 0 {
		parts = append(parts, "\nКонтекст:")
		for key, value := range b.context {
			parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
		}
	}
	
	parts = append(parts, "\nПожалуйста, помоги создать интеграцию. Если нужна дополнительная информация - задай вопросы.")
	
	return strings.Join(parts, "\n")
}

// BuildDataAnalysisPrompt создает промпт для анализа данных
func (b *ChatPromptBuilder) BuildDataAnalysisPrompt(data map[string]interface{}, format string) string {
	return fmt.Sprintf(`Проанализируй структуру данных и верни результат в JSON формате.

Данные (%s формат):
%s

Верни JSON с полями:
{
  "fields": [
    {
      "name": "имя_поля",
      "type": "тип_поля",
      "required": true/false,
      "examples": ["пример1", "пример2"],
      "description": "описание поля"
    }
  ],
  "schema": "описание схемы данных",
  "suggestions": [
    {
      "source_field": "исходное_поле",
      "target_field": "целевое_поле", 
      "transform": "трансформация",
      "confidence": 0.95,
      "reasoning": "объяснение"
    }
  ],
  "data_type": "тип данных (order, user, event, etc.)",
  "confidence": 0.9
}

Типы полей: string, number, boolean, email, phone, date, datetime, url, array, object`, 
		format, formatDataForPrompt(data))
}

// BuildMappingPrompt создает промпт для генерации маппинга
func (b *ChatPromptBuilder) BuildMappingPrompt(sourceData map[string]interface{}, targetAPI, task, userPrompt string) string {
	return fmt.Sprintf(`Создай маппинг для интеграции webhook и верни результат в JSON формате.

Задача: %s
Целевой API: %s
Промпт пользователя: %s

Исходные данные:
%s

Верни JSON с полями:
{
  "type": "json_template",
  "template": "шаблон с {{field}} плейсхолдерами",
  "target_url": "URL целевого API",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "auth_type": "bearer|oauth|basic|none",
  "auth_config": {
    "token_field": "поле для токена"
  },
  "description": "описание маппинга",
  "reasoning": "объяснение логики"
}

Для популярных API используй правильные URL и форматы:
- Slack: hooks.slack.com/services/...
- Telegram: api.telegram.org/bot{token}/sendMessage  
- Discord: discord.com/api/webhooks/...`, 
		task, targetAPI, userPrompt, formatDataForPrompt(sourceData))
}

// formatDataForPrompt форматирует данные для промпта
func formatDataForPrompt(data map[string]interface{}) string {
	if len(data) == 0 {
		return "Нет данных"
	}
	
	var parts []string
	for key, value := range data {
		parts = append(parts, fmt.Sprintf("  %s: %v", key, value))
	}
	
	return "{\n" + strings.Join(parts, ",\n") + "\n}"
}

