package ai

// GetQuickSuggestions возвращает быстрые предложения для пользователя
func GetQuickSuggestions() []AISuggestion {
	return []AISuggestion{
		{
			Type:        "integration_setup",
			Title:       "Настроить Slack уведомления",
			Description: "Отправлять уведомления о новых заказах в Slack канал",
			Data: map[string]interface{}{
				"target_api": "slack",
				"template":   "webhook_notification",
			},
			Confidence: 0.9,
		},
		{
			Type:        "integration_setup",
			Title:       "Интеграция с Telegram",
			Description: "Отправлять сообщения в Telegram бот или канал",
			Data: map[string]interface{}{
				"target_api": "telegram",
				"template":   "bot_message",
			},
			Confidence: 0.9,
		},
		{
			Type:        "integration_setup",
			Title:       "Webhook в Discord",
			Description: "Отправлять уведомления в Discord канал",
			Data: map[string]interface{}{
				"target_api": "discord",
				"template":   "webhook_message",
			},
			Confidence: 0.8,
		},
		{
			Type:        "data_analysis",
			Title:       "Анализировать входящие данные",
			Description: "Проанализировать структуру и типы полей в данных",
			Data: map[string]interface{}{
				"action": "analyze_structure",
			},
			Confidence: 0.8,
		},
		{
			Type:        "mapping_generation",
			Title:       "Создать JSON маппинг",
			Description: "Автоматически создать маппинг для трансформации данных",
			Data: map[string]interface{}{
				"action": "generate_json_mapping",
			},
			Confidence: 0.7,
		},
		{
			Type:        "integration_setup",
			Title:       "REST API интеграция",
			Description: "Настроить отправку данных в любой REST API",
			Data: map[string]interface{}{
				"target_api": "rest",
				"template":   "generic_rest",
			},
			Confidence: 0.7,
		},
	}
}

// GetPopularAPIs возвращает список популярных API для интеграции
func GetPopularAPIs() []PopularAPI {
	return []PopularAPI{
		{
			Name:        "Slack",
			BaseURL:     "https://hooks.slack.com/services",
			AuthType:    "webhook",
			Description: "Отправка уведомлений в Slack каналы",
			CommonFields: map[string]string{
				"text":     "Основной текст сообщения",
				"username": "Имя бота отправителя",
				"channel":  "Канал для отправки",
			},
			Templates: []APITemplate{
				{
					Name:    "Простое уведомление",
					Purpose: "Отправка текстового уведомления",
					Method:  "POST",
					Path:    "/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					BodyTemplate: `{
  "text": "{{.message}}",
  "username": "dmIntegroff",
  "icon_emoji": ":robot_face:"
}`,
					RequiredFields: []string{"message"},
				},
				{
					Name:    "Богатое уведомление",
					Purpose: "Уведомление с дополнительными полями",
					Method:  "POST",
					Path:    "/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					BodyTemplate: `{
  "text": "{{.title}}",
  "attachments": [
    {
      "color": "good",
      "fields": [
        {
          "title": "Детали",
          "value": "{{.details}}",
          "short": false
        }
      ]
    }
  ]
}`,
					RequiredFields: []string{"title", "details"},
				},
			},
		},
		{
			Name:        "Telegram",
			BaseURL:     "https://api.telegram.org/bot",
			AuthType:    "bearer",
			Description: "Отправка сообщений через Telegram Bot API",
			CommonFields: map[string]string{
				"chat_id": "ID чата или канала",
				"text":    "Текст сообщения",
			},
			Templates: []APITemplate{
				{
					Name:    "Текстовое сообщение",
					Purpose: "Отправка простого текстового сообщения",
					Method:  "POST",
					Path:    "YOUR_BOT_TOKEN/sendMessage",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					BodyTemplate: `{
  "chat_id": "{{.chat_id}}",
  "text": "{{.message}}",
  "parse_mode": "HTML"
}`,
					RequiredFields: []string{"chat_id", "message"},
				},
			},
		},
		{
			Name:        "Discord",
			BaseURL:     "https://discord.com/api/webhooks",
			AuthType:    "webhook",
			Description: "Отправка сообщений в Discord каналы",
			CommonFields: map[string]string{
				"content":  "Текст сообщения",
				"username": "Имя бота",
			},
			Templates: []APITemplate{
				{
					Name:    "Простое сообщение",
					Purpose: "Отправка текстового сообщения",
					Method:  "POST",
					Path:    "/WEBHOOK_ID/WEBHOOK_TOKEN",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					BodyTemplate: `{
  "content": "{{.message}}",
  "username": "dmIntegroff"
}`,
					RequiredFields: []string{"message"},
				},
				{
					Name:    "Embed сообщение",
					Purpose: "Сообщение с богатым форматированием",
					Method:  "POST",
					Path:    "/WEBHOOK_ID/WEBHOOK_TOKEN",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					BodyTemplate: `{
  "embeds": [
    {
      "title": "{{.title}}",
      "description": "{{.description}}",
      "color": 5814783,
      "timestamp": "{{.timestamp}}"
    }
  ]
}`,
					RequiredFields: []string{"title", "description"},
				},
			},
		},
		{
			Name:        "Microsoft Teams",
			BaseURL:     "https://outlook.office.com/webhook",
			AuthType:    "webhook",
			Description: "Отправка уведомлений в Microsoft Teams",
			CommonFields: map[string]string{
				"text":    "Текст сообщения",
				"title":   "Заголовок",
				"summary": "Краткое описание",
			},
			Templates: []APITemplate{
				{
					Name:    "Простое уведомление",
					Purpose: "Отправка уведомления в Teams канал",
					Method:  "POST",
					Path:    "/WEBHOOK_URL",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					BodyTemplate: `{
  "@type": "MessageCard",
  "@context": "http://schema.org/extensions",
  "summary": "{{.summary}}",
  "themeColor": "0076D7",
  "sections": [
    {
      "activityTitle": "{{.title}}",
      "activitySubtitle": "{{.subtitle}}",
      "text": "{{.message}}"
    }
  ]
}`,
					RequiredFields: []string{"summary", "title", "message"},
				},
			},
		},
		{
			Name:        "AmoCRM",
			BaseURL:     "https://SUBDOMAIN.amocrm.ru/api/v4",
			AuthType:    "bearer",
			Description: "Интеграция с AmoCRM для создания лидов и контактов",
			CommonFields: map[string]string{
				"name":  "Имя контакта",
				"phone": "Телефон",
				"email": "Email адрес",
			},
			Templates: []APITemplate{
				{
					Name:    "Создание контакта",
					Purpose: "Создание нового контакта в AmoCRM",
					Method:  "POST",
					Path:    "/contacts",
					Headers: map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer YOUR_ACCESS_TOKEN",
					},
					BodyTemplate: `[
  {
    "name": "{{.name}}",
    "custom_fields_values": [
      {
        "field_id": 123456,
        "values": [
          {
            "value": "{{.phone}}",
            "enum_code": "WORK"
          }
        ]
      },
      {
        "field_id": 123457,
        "values": [
          {
            "value": "{{.email}}",
            "enum_code": "WORK"
          }
        ]
      }
    ]
  }
]`,
					RequiredFields: []string{"name"},
				},
				{
					Name:    "Создание лида",
					Purpose: "Создание нового лида в AmoCRM",
					Method:  "POST",
					Path:    "/leads",
					Headers: map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer YOUR_ACCESS_TOKEN",
					},
					BodyTemplate: `[
  {
    "name": "Лид из {{.source}}",
    "price": {{.amount}},
    "custom_fields_values": [
      {
        "field_id": 123458,
        "values": [
          {
            "value": "{{.description}}"
          }
        ]
      }
    ],
    "_embedded": {
      "contacts": [
        {
          "name": "{{.client_name}}",
          "custom_fields_values": [
            {
              "field_id": 123456,
              "values": [
                {
                  "value": "{{.phone}}",
                  "enum_code": "WORK"
                }
              ]
            }
          ]
        }
      ]
    }
  }
]`,
					RequiredFields: []string{"client_name", "source"},
				},
			},
		},
		{
			Name:        "Generic REST API",
			BaseURL:     "https://api.example.com",
			AuthType:    "bearer",
			Description: "Универсальная интеграция с любым REST API",
			CommonFields: map[string]string{
				"data": "Данные для отправки",
			},
			Templates: []APITemplate{
				{
					Name:    "POST запрос",
					Purpose: "Отправка данных методом POST",
					Method:  "POST",
					Path:    "/api/endpoint",
					Headers: map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer YOUR_TOKEN",
					},
					BodyTemplate: `{{.data}}`,
					RequiredFields: []string{"data"},
				},
				{
					Name:    "PUT запрос",
					Purpose: "Обновление данных методом PUT",
					Method:  "PUT",
					Path:    "/api/endpoint/{{.id}}",
					Headers: map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer YOUR_TOKEN",
					},
					BodyTemplate: `{{.data}}`,
					RequiredFields: []string{"id", "data"},
				},
			},
		},
	}
}