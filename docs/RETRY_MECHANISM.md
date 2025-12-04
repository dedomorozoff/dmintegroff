# Retry Mechanism with Exponential Backoff

## Обзор

Система автоматических повторных попыток (retry) с экспоненциальной задержкой обеспечивает надежность при взаимодействии с внешними API, автоматически повторяя неудачные запросы.

## Основные возможности

### 1. Экспоненциальная задержка (Exponential Backoff)

Задержка между попытками увеличивается экспоненциально:
- 1-я попытка: 1 секунда
- 2-я попытка: 2 секунды
- 3-я попытка: 4 секунды
- Максимальная задержка: 30 секунд

### 2. Автоматическое определение повторяемых ошибок

Retry автоматически срабатывает для следующих HTTP статусов:
- `429` - Too Many Requests (слишком много запросов)
- `500` - Internal Server Error (внутренняя ошибка сервера)
- `502` - Bad Gateway (плохой шлюз)
- `503` - Service Unavailable (сервис недоступен)
- `504` - Gateway Timeout (таймаут шлюза)

### 3. Применение

Retry механизм автоматически применяется к:
- **OAuth2 токен запросам** - получение access token
- **Webhook отправкам** - отправка данных в целевые API

## Конфигурация

### Параметры по умолчанию

```go
RetryConfig{
 MaxAttempts: 3, // Максимум 3 попытки
 InitialDelay: 1 * time.Second, // Начальная задержка 1 сек
 MaxDelay: 30 * time.Second, // Максимальная задержка 30 сек
 Multiplier: 2.0, // Множитель для экспоненты
 RetryableStatus: []int{429, 500, 502, 503, 504},
}
```

### Настройка retry для конкретных случаев

```go
// Пример: более агрессивный retry для критичных запросов
config := RetryConfig{
 MaxAttempts: 5,
 InitialDelay: 500 * time.Millisecond,
 MaxDelay: 60 * time.Second,
 Multiplier: 2.0,
 RetryableStatus: []int{429, 500, 502, 503, 504},
}

resp, err := retryWithBackoff(req, client, config)
```

## Примеры использования

### OAuth2 Token Request

```go
// Автоматически использует retry при получении токена
token, err := GetAccessToken(integration)
if err != nil {
 // Ошибка после всех попыток
 log.Error("Failed to get token after retries:", err)
}
```

### Webhook Delivery

```go
// Автоматически использует retry при отправке webhook
err := ProcessWebhook(integrationID, payload)
if err != nil {
 // Ошибка после всех попыток
 log.Error("Failed to deliver webhook after retries:", err)
}
```

## Логирование

Retry механизм логирует каждую попытку:

```
level=warning msg="Request returned retryable status, will retry" 
 attempt=1 max=3 status_code=503 url="https://api.example.com/token"

level=debug msg="Waiting before retry" delay_seconds=1.0

level=warning msg="Request returned retryable status, will retry" 
 attempt=2 max=3 status_code=503 url="https://api.example.com/token"

level=debug msg="Waiting before retry" delay_seconds=2.0

level=info msg="Successfully obtained OAuth2 access token" 
 integration_id=123 expires_in=3600
```

## Поведение при ошибках

### Временные ошибки (Retryable)
- Статусы 429, 500, 502, 503, 504
- Сетевые таймауты
- Временные проблемы с подключением

**Действие**: Автоматический retry с экспоненциальной задержкой

### Постоянные ошибки (Non-retryable)
- Статусы 400, 401, 403, 404
- Ошибки валидации
- Проблемы с аутентификацией

**Действие**: Немедленный возврат ошибки без retry

## Преимущества

1. **Надежность** - автоматическое восстановление после временных сбоев
2. **Защита от перегрузки** - экспоненциальная задержка предотвращает DDoS
3. **Прозрачность** - подробное логирование всех попыток
4. **Гибкость** - настраиваемые параметры для разных сценариев
5. **Эффективность** - не тратит ресурсы на повтор постоянных ошибок

## Лучшие практики

### 1. Мониторинг retry метрик
Следите за количеством retry попыток в логах:
```bash
# Поиск retry событий
grep "will retry" logs/app.log | wc -l
```

### 2. Настройка таймаутов
Убедитесь, что общий таймаут учитывает retry:
```go
// Для 3 попыток с задержками 1s, 2s
// Минимальное время: 3s (задержки) + 3*30s (таймауты) = 93s
client := &http.Client{
 Timeout: 30 * time.Second,
}
```

### 3. Обработка финальных ошибок
Всегда обрабатывайте ошибки после исчерпания retry:
```go
if err := ProcessWebhook(id, data); err != nil {
 // Отправить уведомление администратору
 // Сохранить в dead letter queue
 // Логировать для анализа
}
```

## Тестирование

Запуск тестов retry механизма:
```bash
cd dmintegroff
go test -v ./internal/services/retry_test.go ./internal/services/oauth_service.go
```

Тесты покрывают:
- Успешный запрос без retry
- Успех после нескольких retry
- Исчерпание всех попыток
- Не-повторяемые статусы
- Корректная обработка тела запроса при retry
- Расчет экспоненциальной задержки

## Troubleshooting

### Проблема: Слишком много retry попыток

**Решение**: Проверьте статус целевого API и увеличьте задержки:
```go
config.InitialDelay = 2 * time.Second
config.MaxDelay = 60 * time.Second
```

### Проблема: Retry не срабатывает

**Причина**: Статус код не в списке retryable

**Решение**: Добавьте нужный статус:
```go
config.RetryableStatus = append(config.RetryableStatus, 408) // Request Timeout
```

### Проблема: Долгое время ответа

**Причина**: Суммарное время всех retry попыток

**Решение**: Уменьшите количество попыток или задержки:
```go
config.MaxAttempts = 2
config.InitialDelay = 500 * time.Millisecond
```

## Roadmap

Планируемые улучшения:
- [ ] Jitter (случайная вариация задержки)
- [ ] Circuit breaker pattern
- [ ] Метрики retry в Prometheus
- [ ] Настройка retry через UI
- [ ] Retry для конкретных интеграций
