# 🤖 AI MVP - Технические требования

Минимальные требования для первого этапа AI-ассистента в dmIntegroff.

---

## 🎯 Цель MVP

Создать базовый AI-чат, который может:
- Понимать задачи интеграции на естественном языке
- Анализировать структуру входящих данных
- Генерировать простые JSON маппинги
- Предлагать настройки для популярных сервисов (Slack, Telegram)

---

## 🏗️ Архитектура

```
Frontend (HTML/JS)     Backend (Go)           External
┌─────────────────┐   ┌─────────────────┐   ┌─────────────┐
│                 │   │                 │   │             │
│ AI Chat Widget  │◄─►│ AI Controller   │◄─►│ OpenAI API  │
│                 │   │                 │   │             │
│ • Input field   │   │ • Prompt system │   │ • GPT-4     │
│ • Chat history  │   │ • Data analysis │   │ • JSON mode │
│ • Preview       │   │ • Validation    │   │             │
└─────────────────┘   └─────────────────┘   └─────────────┘
```

---

## 📋 Новые файлы

### Backend (Go)

```
internal/ai/
├── client.go          // OpenAI API клиент
├── prompts.go         // Система промптов
├── analyzer.go        // Анализ входящих данных
├── generator.go       // Генерация маппингов
└── types.go          // Типы данных для AI

internal/controllers/
└── ai_controller.go   // HTTP контроллер для AI

internal/config/
└── ai_config.go      // Конфигурация AI (API ключи)
```

### Frontend

```
templates/components/
└── ai_chat.html      // Компонент чата

static/js/
└── ai-assistant.js   // JavaScript для чата

static/css/
└── ai-chat.css      // Стили для чата
```

---

## 🔧 API Endpoints

```
POST /api/ai/chat
- Основной чат с AI
- Input: { "message": "string", "context": {...} }
- Output: { "response": "string", "suggestions": [...] }

POST /api/ai/analyze-data
- Анализ структуры данных
- Input: { "data": {...}, "format": "json|xml|form" }
- Output: { "fields": [...], "suggestions": [...] }

POST /api/ai/generate-mapping
- Генерация маппинга
- Input: { "source_data": {...}, "target_api": "string", "task": "string" }
- Output: { "mapping": {...}, "template": "string" }
```

---

## 🤖 Система промптов

### Базовые промпты

**System Prompt:**
```
Ты AI-ассистент для создания webhook интеграций в системе dmIntegroff.

Твоя задача:
1. Понимать задачи интеграции на русском/английском языке
2. Анализировать структуру входящих данных (JSON, XML, form-data)
3. Создавать маппинги для популярных API (Slack, Telegram, CRM)
4. Предлагать оптимальные настройки

Отвечай кратко и по делу. Всегда предлагай конкретные решения.
```

**Примеры задач:**
```
User: "Хочу отправлять уведомления в Slack при новом заказе"
Data: {"order_id": 123, "customer": "John Doe", "amount": 250.00}

AI Response:
{
  "understanding": "Создать интеграцию для отправки уведомлений о заказах в Slack",
  "target_api": "slack_webhook",
  "mapping_suggestion": {
    "template_type": "json",
    "template": {
      "text": "🛒 Новый заказ #{{order_id}}\n👤 Клиент: {{customer}}\n💰 Сумма: ${{amount}}"
    }
  },
  "next_steps": [
    "Укажите URL вашего Slack webhook",
    "Выберите канал для уведомлений",
    "Протестируйте интеграцию"
  ]
}
```

---

## 🔑 Конфигурация

### Переменные окружения

```bash
# .env
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4-1106-preview
OPENAI_MAX_TOKENS=2000
OPENAI_TEMPERATURE=0.3

# Опционально для локальной LLM
LOCAL_LLM_ENABLED=false
LOCAL_LLM_URL=http://localhost:11434
```

### Конфигурация в коде

```go
type AIConfig struct {
    OpenAIAPIKey    string `env:"OPENAI_API_KEY"`
    OpenAIModel     string `env:"OPENAI_MODEL" envDefault:"gpt-4-1106-preview"`
    MaxTokens       int    `env:"OPENAI_MAX_TOKENS" envDefault:"2000"`
    Temperature     float32 `env:"OPENAI_TEMPERATURE" envDefault:"0.3"`
    LocalLLMEnabled bool   `env:"LOCAL_LLM_ENABLED" envDefault:"false"`
    LocalLLMURL     string `env:"LOCAL_LLM_URL" envDefault:"http://localhost:11434"`
}
```

---

## 📱 UI/UX

### Размещение чата

**Вариант 1: Модальное окно**
- Кнопка "🤖 AI Помощник" на странице создания интеграции
- Открывается модальное окно с чатом
- Результат применяется к форме создания

**Вариант 2: Боковая панель**
- Выдвижная панель справа
- Всегда доступна при настройке интеграций
- Контекстные подсказки

**Рекомендация для MVP: Модальное окно** (проще реализовать)

### Элементы интерфейса

```html
<!-- Кнопка запуска -->
<button class="btn btn-primary" onclick="openAIAssistant()">
  🤖 AI Помощник
</button>

<!-- Модальное окно -->
<div class="modal" id="aiAssistantModal">
  <div class="modal-content">
    <div class="chat-header">
      <h3>🤖 AI Ассистент dmIntegroff</h3>
    </div>
    
    <div class="chat-messages" id="chatMessages">
      <!-- Сообщения чата -->
    </div>
    
    <div class="chat-input">
      <input type="text" placeholder="Опишите что хотите сделать..." />
      <button>Отправить</button>
    </div>
    
    <div class="chat-suggestions">
      <!-- Быстрые действия -->
      <button>Slack уведомления</button>
      <button>Telegram бот</button>
      <button>CRM интеграция</button>
    </div>
  </div>
</div>
```

---

## 🧪 Тестирование

### Тестовые сценарии

**Сценарий 1: Slack уведомления**
```
Input: "Хочу получать уведомления в Slack о новых заказах"
Sample Data: {"order_id": 123, "customer": "John", "amount": 100}
Expected: Генерация Slack webhook маппинга
```

**Сценарий 2: Telegram бот**
```
Input: "Отправлять сообщения в Telegram при ошибках"
Sample Data: {"error": "Database connection failed", "timestamp": "2024-12-12T10:30:00Z"}
Expected: Генерация Telegram Bot API маппинга
```

**Сценарий 3: Анализ данных**
```
Input: Загрузка JSON с 10+ полями
Expected: Автоматическое определение типов полей и предложения маппинга
```

### Метрики качества

- **Точность понимания задач:** >80% правильных интерпретаций
- **Качество маппингов:** >70% работающих без правок
- **Время ответа:** <5 секунд для простых задач
- **Удовлетворенность пользователей:** >4/5 в опросах

---

## 🚀 План реализации MVP

### Неделя 1: Инфраструктура
- [ ] Настройка OpenAI API клиента
- [ ] Базовая система промптов
- [ ] API endpoints для AI
- [ ] Простейший UI чата

### Неделя 2: Основная логика
- [ ] Анализ структуры данных
- [ ] Генерация базовых маппингов
- [ ] Интеграция с формой создания интеграции
- [ ] Тестирование на простых сценариях

### Неделя 3: Полировка
- [ ] Улучшение промптов
- [ ] Обработка ошибок
- [ ] Валидация результатов
- [ ] Документация и примеры

---

## 💰 Стоимость MVP

### OpenAI API
- **GPT-4 Turbo:** $0.01 за 1K input tokens, $0.03 за 1K output tokens
- **Средний запрос:** ~500 input + 300 output tokens = ~$0.014
- **100 запросов в день:** ~$1.40/день = $42/месяц
- **Бюджет на тестирование:** $100-200/месяц

### Альтернативы
- **GPT-3.5 Turbo:** в 10 раз дешевле, но хуже качество
- **Локальная LLM:** бесплатно, но нужны ресурсы сервера

---

## 🔒 Безопасность

### Защита данных
- [ ] Не отправлять чувствительные данные в OpenAI
- [ ] Маскировать персональные данные в примерах
- [ ] Логирование всех AI-запросов
- [ ] Rate limiting для AI endpoints

### Валидация результатов
- [ ] Проверка сгенерированного JSON на корректность
- [ ] Валидация URL и API endpoints
- [ ] Sandbox для тестирования маппингов
- [ ] Ограничения на выполнение кода

---

## 📈 Метрики успеха MVP

### Технические метрики
- Время ответа AI: <5 секунд
- Успешность генерации маппингов: >70%
- Uptime AI сервиса: >99%

### Пользовательские метрики
- Использование AI при создании интеграций: >30%
- Принятие AI-предложений без правок: >50%
- NPS для AI-функций: >6/10

### Бизнес-метрики
- Сокращение времени создания интеграции: >50%
- Увеличение конверсии новых пользователей: >20%
- Снижение обращений в поддержку: >30%

---

**Документ создан:** 12 декабря 2024  
**Статус:** Готов к разработке  
**Приоритет:** 🔥 Критический (главная фича v2.0)