# 🔀 Множественные маппинги на один webhook

Дизайн-документ для функции множественных выходов (outputs) на один входящий webhook.

---

## 📋 Обзор

Функция позволяет настроить несколько маппингов для одного входящего webhook с отправкой трансформированных данных на разные Target API.

### Проблема

**Текущая ситуация:**
- Один webhook = одна интеграция = один Target API
- Для отправки данных в несколько систем нужно создавать несколько интеграций
- Дублирование webhook URL и настроек
- Сложность управления связанными интеграциями

**Пример:**
```
Webhook: новый заказ от Shopify

Нужно отправить в:
├─ CRM (Salesforce) - создать лид
├─ Analytics (Google Analytics) - событие покупки
├─ Email (SendGrid) - подтверждение заказа
└─ Warehouse - задача на отгрузку

Сейчас: 4 отдельные интеграции с 4 разными webhook URL
Нужно: 1 интеграция с 4 выходами на один webhook URL
```

---

## 🎯 Цели

1. **Упрощение архитектуры** - один webhook для множественных назначений
2. **Централизованное управление** - все выходы в одном месте
3. **Гибкость** - разные маппинги для разных систем
4. **Надежность** - независимая обработка ошибок
5. **Производительность** - параллельная обработка выходов

---

## 🏗️ Архитектура

### Модель данных

```
Integration (основная интеграция)
├─ id: 1
├─ name: "Shopify Orders"
├─ webhook_token: "abc123"
├─ sample_payload: "{...}"
├─ mode: "active"
└─ outputs: [...]
    │
    ├─ IntegrationOutput #1 (CRM)
    │  ├─ id: 1
    │  ├─ name: "Salesforce Lead"
    │  ├─ target_api: "https://api.salesforce.com/leads"
    │  ├─ mapping_config: "{...}"
    │  ├─ enabled: true
    │  └─ priority: 1
    │
    ├─ IntegrationOutput #2 (Analytics)
    │  ├─ id: 2
    │  ├─ name: "Google Analytics Event"
    │  ├─ target_api: "https://www.google-analytics.com/collect"
    │  ├─ output_template: "{...}"
    │  ├─ enabled: true
    │  └─ priority: 2
    │
    └─ IntegrationOutput #3 (Email)
       ├─ id: 3
       ├─ name: "SendGrid Confirmation"
       ├─ target_api: "https://api.sendgrid.com/v3/mail/send"
       ├─ condition: "{{order.status}} == 'paid'"
       ├─ enabled: true
       └─ priority: 3
```

### Структура таблиц

**Таблица: `integrations`** (существующая)
- Остается без изменений
- Добавляется связь `has_many :outputs`

**Таблица: `integration_outputs`** (новая)
```sql
CREATE TABLE integration_outputs (
    id SERIAL PRIMARY KEY,
    integration_id INTEGER NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Target API
    target_api VARCHAR(500) NOT NULL,
    http_method VARCHAR(10) DEFAULT 'POST',
    
    -- Трансформация данных
    mapping_config TEXT,
    output_template TEXT,
    template_type VARCHAR(50) DEFAULT 'json',
    
    -- Условия выполнения
    condition TEXT,
    
    -- Настройки выполнения
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    timeout INTEGER DEFAULT 30,
    
    -- Аутентификация
    auth_type VARCHAR(50),
    oauth2_token_url VARCHAR(500),
    oauth2_client_id VARCHAR(255),
    oauth2_client_secret VARCHAR(255),
    oauth2_scope VARCHAR(255),
    bearer_token VARCHAR(500),
    basic_auth_user VARCHAR(255),
    basic_auth_pass VARCHAR(255),
    custom_headers TEXT,
    
    -- Retry настройки
    retry_enabled BOOLEAN DEFAULT true,
    retry_max_attempts INTEGER DEFAULT 3,
    retry_delay INTEGER DEFAULT 5,
    
    -- Метаданные
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_integration_id (integration_id),
    INDEX idx_enabled (enabled),
    INDEX idx_priority (priority)
);
```

---

## 🔄 Процесс обработки

### Диаграмма потока

```
┌─────────────────────────────────────────────────────────────┐
│                    Входящий Webhook                          │
│                  POST /webhook/abc123                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              1. Получение Integration                        │
│              - Поиск по webhook_token                        │
│              - Проверка mode == "active"                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              2. Загрузка всех Outputs                        │
│              - WHERE enabled = true                          │
│              - ORDER BY priority ASC                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              3. Обработка каждого Output                     │
│              (параллельно или последовательно)               │
└────────────────────────┬────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Output #1   │  │  Output #2   │  │  Output #3   │
│              │  │              │  │              │
│ 1. Проверка  │  │ 1. Проверка  │  │ 1. Проверка  │
│    условия   │  │    условия   │  │    условия   │
│              │  │              │  │              │
│ 2. Трансфор- │  │ 2. Трансфор- │  │ 2. Трансфор- │
│    мация     │  │    мация     │  │    мация     │
│              │  │              │  │              │
│ 3. Отправка  │  │ 3. Отправка  │  │ 3. Отправка  │
│    на Target │  │    на Target │  │    на Target │
│              │  │              │  │              │
│ 4. Логиро-   │  │ 4. Логиро-   │  │ 4. Логиро-   │
│    вание     │  │    вание     │  │    вание     │
└──────────────┘  └──────────────┘  └──────────────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              4. Возврат результата                           │
│              - Успешно обработано: X из Y                    │
│              - Ошибки: [...]                                 │
└─────────────────────────────────────────────────────────────┘
```

### Псевдокод обработки

```go
func ProcessWebhookWithOutputs(integrationID uint, payload map[string]interface{}) error {
    // 1. Загрузить интеграцию
    integration := GetIntegration(integrationID)
    
    // 2. Загрузить все активные выходы
    outputs := GetActiveOutputs(integrationID)
    
    // 3. Обработать каждый выход
    results := make([]OutputResult, len(outputs))
    
    // Параллельная обработка
    var wg sync.WaitGroup
    for i, output := range outputs {
        wg.Add(1)
        go func(idx int, out IntegrationOutput) {
            defer wg.Done()
            results[idx] = ProcessOutput(out, payload)
        }(i, output)
    }
    wg.Wait()
    
    // 4. Логирование результатов
    LogOutputResults(integrationID, results)
    
    // 5. Возврат результата
    return AggregateResults(results)
}

func ProcessOutput(output IntegrationOutput, payload map[string]interface{}) OutputResult {
    // 1. Проверка условия
    if output.Condition != "" {
        if !EvaluateCondition(output.Condition, payload) {
            return OutputResult{Skipped: true, Reason: "Condition not met"}
        }
    }
    
    // 2. Трансформация данных
    var transformedData interface{}
    if output.OutputTemplate != "" {
        transformedData = ApplyTemplate(output.OutputTemplate, payload)
    } else {
        transformedData = ApplyMapping(output.MappingConfig, payload)
    }
    
    // 3. Отправка на Target API
    response, err := SendToTargetAPI(output, transformedData)
    
    // 4. Retry при ошибке
    if err != nil && output.RetryEnabled {
        response, err = RetryWithBackoff(output, transformedData)
    }
    
    return OutputResult{
        OutputID:   output.ID,
        Success:    err == nil,
        Error:      err,
        Response:   response,
        Duration:   time.Since(start),
    }
}
```

---

## 🎨 UI/UX Дизайн

### Страница интеграции с выходами

```
┌─────────────────────────────────────────────────────────────┐
│ Интеграция: Shopify Orders                    [⚙️ Настройки] │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│ 📊 Статистика за последние 24 часа                           │
│ ┌─────────────┬─────────────┬─────────────┬─────────────┐   │
│ │ Получено    │ Обработано  │ Ошибки      │ Пропущено   │   │
│ │ 1,234       │ 1,180       │ 12          │ 42          │   │
│ └─────────────┴─────────────┴─────────────┴─────────────┘   │
│                                                               │
│ 🔀 Выходы (Outputs)                      [➕ Добавить выход] │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ ✅ Salesforce Lead                          [⚙️] [🗑️]  │   │
│ │ https://api.salesforce.com/leads                      │   │
│ │ Успешно: 1,180 | Ошибки: 8 | Среднее время: 245ms    │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ ✅ Google Analytics Event                   [⚙️] [🗑️]  │   │
│ │ https://www.google-analytics.com/collect              │   │
│ │ Успешно: 1,180 | Ошибки: 0 | Среднее время: 89ms     │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ ⚠️ SendGrid Confirmation                    [⚙️] [🗑️]  │   │
│ │ https://api.sendgrid.com/v3/mail/send                 │   │
│ │ Условие: {{order.status}} == 'paid'                   │   │
│ │ Успешно: 1,138 | Ошибки: 4 | Пропущено: 42           │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ ❌ Warehouse System                         [⚙️] [🗑️]  │   │
│ │ https://warehouse.example.com/api/orders              │   │
│ │ Отключен                                              │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### Форма создания выхода

```
┌─────────────────────────────────────────────────────────────┐
│ Создание выхода для интеграции: Shopify Orders               │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│ Основные настройки                                            │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ Название выхода *                                     │   │
│ │ [Salesforce Lead Creation                          ]  │   │
│ │                                                       │   │
│ │ Описание (опционально)                                │   │
│ │ [Создание лида в Salesforce при новом заказе       ]  │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ Target API                                                    │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ URL *                                                 │   │
│ │ [https://api.salesforce.com/services/data/v52.0/   ]  │   │
│ │                                                       │   │
│ │ HTTP метод                                            │   │
│ │ [POST ▼]                                              │   │
│ │                                                       │   │
│ │ Timeout (секунды)                                     │   │
│ │ [30                                                ]  │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ Аутентификация                                                │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ Тип аутентификации                                    │   │
│ │ [OAuth 2.0 ▼]                                         │   │
│ │                                                       │   │
│ │ Token URL                                             │   │
│ │ [https://login.salesforce.com/services/oauth2/token]  │   │
│ │                                                       │   │
│ │ Client ID                                             │   │
│ │ [••••••••••••••••••••••••••••••••••••••••••••••]     │   │
│ │                                                       │   │
│ │ Client Secret                                         │   │
│ │ [••••••••••••••••••••••••••••••••••••••••••••••]     │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ Условия выполнения (опционально)                             │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ Выполнять только при условии:                         │   │
│ │ [{{order.status}} == 'paid'                        ]  │   │
│ │                                                       │   │
│ │ 💡 Примеры:                                           │   │
│ │    {{amount}} > 1000                                  │   │
│ │    {{event_type}} == "user.created"                   │   │
│ │    {{status}} in ["pending", "processing"]            │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ Настройки повторных попыток                                  │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ ☑️ Включить retry механизм                            │   │
│ │                                                       │   │
│ │ Максимум попыток: [3  ]                               │   │
│ │ Задержка (секунды): [5  ]                             │   │
│ └───────────────────────────────────────────────────────┘   │
│                                                               │
│ [Создать и настроить маппинг]  [Отмена]                      │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Технические детали

### API Endpoints

```
# Управление выходами
GET    /integrations/:id/outputs              # Список выходов
GET    /integrations/:id/outputs/create       # Форма создания
POST   /integrations/:id/outputs              # Создать выход
GET    /integrations/:id/outputs/:output_id/edit  # Форма редактирования
PUT    /integrations/:id/outputs/:output_id   # Обновить выход
DELETE /integrations/:id/outputs/:output_id   # Удалить выход

# Настройка маппинга
GET    /integrations/:id/outputs/:output_id/configure  # Страница настройки маппинга
POST   /integrations/:id/outputs/:output_id/configure  # Сохранить маппинг

# Управление состоянием
POST   /integrations/:id/outputs/:output_id/toggle     # Включить/выключить
POST   /integrations/:id/outputs/:output_id/test       # Протестировать

# Статистика
GET    /api/integrations/:id/outputs/:output_id/stats  # Статистика выхода
```

### Логирование

Расширить таблицу `request_logs`:

```sql
ALTER TABLE request_logs ADD COLUMN output_id INTEGER REFERENCES integration_outputs(id);
ALTER TABLE request_logs ADD COLUMN output_name VARCHAR(255);
ALTER TABLE request_logs ADD COLUMN condition_result BOOLEAN;
```

Это позволит отслеживать:
- Какой выход обработал запрос
- Был ли выход пропущен из-за условия
- Ошибки для каждого выхода отдельно

---

## 📊 Примеры использования

### Пример 1: E-commerce интеграция

**Сценарий:** Новый заказ в Shopify нужно обработать в 4 системах

**Настройка:**

1. **Создать интеграцию** "Shopify Orders"
2. **Добавить выход #1:** Salesforce
   - Target API: `https://api.salesforce.com/leads`
   - Маппинг: создать лид с данными клиента
   - Условие: нет (всегда выполнять)

3. **Добавить выход #2:** Google Analytics
   - Target API: `https://www.google-analytics.com/collect`
   - Маппинг: событие покупки
   - Условие: `{{order.financial_status}} == 'paid'`

4. **Добавить выход #3:** SendGrid
   - Target API: `https://api.sendgrid.com/v3/mail/send`
   - Маппинг: email подтверждение
   - Условие: `{{order.email}} != ''`

5. **Добавить выход #4:** Warehouse
   - Target API: `https://warehouse.example.com/api/orders`
   - Маппинг: задача на отгрузку
   - Условие: `{{order.fulfillment_status}} == 'unfulfilled'`

**Результат:**
- Один webhook URL для Shopify
- Автоматическая обработка в 4 системах
- Условное выполнение для каждой системы
- Независимое логирование и retry

---

## ⚠️ Ограничения и соображения

### Производительность

1. **Параллельная обработка:**
   - По умолчанию все выходы обрабатываются параллельно
   - Можно настроить последовательную обработку (по priority)
   - Timeout для каждого выхода отдельно

2. **Ограничения:**
   - Максимум 10 выходов на одну интеграцию (настраивается)
   - Общий timeout для всех выходов: 60 секунд
   - При превышении - логирование ошибки

### Обработка ошибок

1. **Независимость:**
   - Ошибка в одном выходе не влияет на другие
   - Каждый выход имеет свой retry механизм
   - Логирование ошибок для каждого выхода

2. **Возврат результата:**
   - HTTP 200: если хотя бы один выход успешен
   - HTTP 500: если все выходы упали
   - Детали в response body

### Безопасность

1. **Аутентификация:**
   - Каждый выход имеет свои credentials
   - Шифрование sensitive данных
   - Валидация Target API URL

2. **Условия:**
   - Безопасное выполнение условий (sandbox)
   - Защита от injection атак
   - Timeout для вычисления условий

---

## 🚀 План реализации

### Фаза 1: Базовая функциональность (v2.0.0)
- [ ] Создание модели `IntegrationOutput`
- [ ] Миграция БД
- [ ] CRUD операции для выходов
- [ ] Базовая обработка множественных выходов
- [ ] UI для управления выходами
- [ ] Логирование результатов

### Фаза 2: Расширенные возможности (v2.1.0)
- [ ] Условное выполнение
- [ ] Параллельная обработка
- [ ] Приоритеты выполнения
- [ ] Статистика по выходам
- [ ] Тестирование выходов

### Фаза 3: Оптимизация (v2.2.0)
- [ ] Кэширование OAuth токенов
- [ ] Оптимизация параллельной обработки
- [ ] Circuit breaker для выходов
- [ ] Webhook replay для выходов
- [ ] Расширенная аналитика

---

**Дата создания:** 8 декабря 2024  
**Статус:** Планируется для v2.0.0
