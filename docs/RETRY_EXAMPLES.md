# Примеры использования Retry механизма

## Базовое использование

### 1. OAuth2 токен запрос (автоматический retry)

```go
// Retry автоматически применяется при получении токена
token, err := GetAccessToken(integration)
if err != nil {
 // Ошибка после всех попыток retry
 log.Error("Failed to get OAuth token:", err)
 return err
}

// Используем токен
req.Header.Set("Authorization", "Bearer "+token)
```

### 2. Webhook отправка (автоматический retry)

```go
// Retry автоматически применяется при отправке webhook
err := ProcessWebhook(integrationID, payload)
if err != nil {
 // Ошибка после всех попыток retry
 log.Error("Failed to deliver webhook:", err)
 // Можно добавить в очередь для повторной обработки
 addToDeadLetterQueue(integrationID, payload)
}
```

## Кастомная конфигурация

### 3. Более агрессивный retry для критичных запросов

```go
// Создаем кастомную конфигурацию
config := RetryConfig{
 MaxAttempts: 5, // 5 попыток вместо 3
 InitialDelay: 500 * time.Millisecond, // Быстрее начинаем
 MaxDelay: 60 * time.Second, // Больше максимум
 Multiplier: 2.0,
 RetryableStatus: []int{429, 500, 502, 503, 504},
}

// Используем кастомную конфигурацию
client := &http.Client{Timeout: 30 * time.Second}
resp, err := retryWithBackoff(req, client, config)
```

### 4. Консервативный retry для некритичных запросов

```go
// Меньше попыток, больше задержки
config := RetryConfig{
 MaxAttempts: 2, // Только 2 попытки
 InitialDelay: 2 * time.Second, // Начинаем с 2 секунд
 MaxDelay: 10 * time.Second, // Максимум 10 секунд
 Multiplier: 2.0,
 RetryableStatus: []int{503, 504}, // Только для timeout
}

resp, err := retryWithBackoff(req, client, config)
```

### 5. Добавление дополнительных retryable статусов

```go
config := DefaultRetryConfig()

// Добавляем 408 Request Timeout
config.RetryableStatus = append(config.RetryableStatus, 408)

// Добавляем 423 Locked (для WebDAV)
config.RetryableStatus = append(config.RetryableStatus, 423)

resp, err := retryWithBackoff(req, client, config)
```

## Продвинутые сценарии

### 6. Retry с логированием метрик

```go
func ProcessWebhookWithMetrics(integrationID uint, payload map[string]interface{}) error {
 startTime := time.Now()
 
 err := ProcessWebhook(integrationID, payload)
 
 duration := time.Since(startTime)
 
 if err != nil {
 // Логируем неудачу с метриками
 metrics.RecordWebhookFailure(integrationID, duration)
 log.WithFields(map[string]interface{}{
 "integration_id": integrationID,
 "duration_ms": duration.Milliseconds(),
 "error": err.Error(),
 }).Error("Webhook failed after retries")
 } else {
 // Логируем успех
 metrics.RecordWebhookSuccess(integrationID, duration)
 log.WithFields(map[string]interface{}{
 "integration_id": integrationID,
 "duration_ms": duration.Milliseconds(),
 }).Info("Webhook delivered successfully")
 }
 
 return err
}
```

### 7. Retry с fallback стратегией

```go
func SendWebhookWithFallback(integration *models.Integration, payload map[string]interface{}) error {
 // Пробуем основной endpoint
 err := ProcessWebhook(integration.ID, payload)
 if err == nil {
 return nil // Успех
 }
 
 // Если есть fallback URL, пробуем его
 if integration.FallbackURL != "" {
 log.Warn("Primary endpoint failed, trying fallback")
 
 originalURL := integration.TargetAPI
 integration.TargetAPI = integration.FallbackURL
 
 err = ProcessWebhook(integration.ID, payload)
 
 integration.TargetAPI = originalURL // Восстанавливаем
 
 if err == nil {
 return nil // Fallback сработал
 }
 }
 
 // Оба endpoint не сработали
 return fmt.Errorf("both primary and fallback endpoints failed: %w", err)
}
```

### 8. Retry с circuit breaker

```go
type CircuitBreaker struct {
 failures int
 lastFailure time.Time
 threshold int
 timeout time.Duration
}

func (cb *CircuitBreaker) IsOpen() bool {
 if cb.failures >= cb.threshold {
 if time.Since(cb.lastFailure) < cb.timeout {
 return true // Circuit открыт
 }
 // Timeout прошел, сбрасываем
 cb.failures = 0
 }
 return false
}

func ProcessWebhookWithCircuitBreaker(
 integrationID uint, 
 payload map[string]interface{},
 cb *CircuitBreaker,
) error {
 // Проверяем circuit breaker
 if cb.IsOpen() {
 return errors.New("circuit breaker is open, skipping request")
 }
 
 // Пробуем отправить
 err := ProcessWebhook(integrationID, payload)
 
 if err != nil {
 cb.failures++
 cb.lastFailure = time.Now()
 return err
 }
 
 // Успех - сбрасываем счетчик
 cb.failures = 0
 return nil
}
```

### 9. Batch retry для множественных webhook

```go
func ProcessWebhookBatch(webhooks []WebhookRequest) []error {
 errors := make([]error, len(webhooks))
 
 // Используем goroutines для параллельной обработки
 var wg sync.WaitGroup
 for i, webhook := range webhooks {
 wg.Add(1)
 go func(idx int, wh WebhookRequest) {
 defer wg.Done()
 
 err := ProcessWebhook(wh.IntegrationID, wh.Payload)
 if err != nil {
 errors[idx] = err
 log.WithFields(map[string]interface{}{
 "integration_id": wh.IntegrationID,
 "batch_index": idx,
 }).Error("Webhook in batch failed")
 }
 }(i, webhook)
 }
 
 wg.Wait()
 return errors
}
```

### 10. Retry с rate limiting

```go
type RateLimiter struct {
 tokens int
 maxTokens int
 refillRate time.Duration
 lastRefill time.Time
 mu sync.Mutex
}

func (rl *RateLimiter) Allow() bool {
 rl.mu.Lock()
 defer rl.mu.Unlock()
 
 // Пополняем токены
 now := time.Now()
 if now.Sub(rl.lastRefill) >= rl.refillRate {
 rl.tokens = rl.maxTokens
 rl.lastRefill = now
 }
 
 // Проверяем доступность
 if rl.tokens > 0 {
 rl.tokens--
 return true
 }
 
 return false
}

func ProcessWebhookWithRateLimit(
 integrationID uint,
 payload map[string]interface{},
 limiter *RateLimiter,
) error {
 // Ждем доступности токена
 for !limiter.Allow() {
 time.Sleep(100 * time.Millisecond)
 }
 
 // Отправляем с retry
 return ProcessWebhook(integrationID, payload)
}
```

## Мониторинг и отладка

### 11. Подсчет retry попыток

```go
type RetryStats struct {
 TotalRequests int
 SuccessFirst int
 SuccessRetry int
 Failed int
 TotalRetries int
}

var stats RetryStats

func TrackRetryStats(err error, attempts int) {
 stats.TotalRequests++
 
 if err == nil {
 if attempts == 1 {
 stats.SuccessFirst++
 } else {
 stats.SuccessRetry++
 stats.TotalRetries += (attempts - 1)
 }
 } else {
 stats.Failed++
 stats.TotalRetries += (attempts - 1)
 }
}

// Использование
err := ProcessWebhook(id, payload)
TrackRetryStats(err, 3) // Предполагаем 3 попытки
```

### 12. Анализ логов retry

```bash
# Подсчет retry событий
grep "will retry" logs/app.log | wc -l

# Группировка по URL
grep "will retry" logs/app.log | grep -oP 'url="[^"]*"' | sort | uniq -c

# Анализ статус кодов
grep "will retry" logs/app.log | grep -oP 'status_code=\d+' | sort | uniq -c

# Средняя задержка
grep "delay_seconds" logs/app.log | grep -oP 'delay_seconds=[\d.]+' | \
 awk -F= '{sum+=$2; count++} END {print sum/count}'
```

## Best Practices

### Рекомендуется

1. **Использовать retry для временных ошибок** (5xx, 429)
2. **Логировать все retry попытки** для анализа
3. **Мониторить метрики retry** (количество, успешность)
4. **Настраивать таймауты** с учетом retry
5. **Использовать circuit breaker** для защиты от каскадных сбоев

### Не рекомендуется

1. **Retry для 4xx ошибок** (кроме 429) - это постоянные ошибки
2. **Слишком много попыток** - может перегрузить систему
3. **Слишком короткие задержки** - не дает системе восстановиться
4. **Игнорировать финальные ошибки** - нужна обработка
5. **Retry без логирования** - теряется информация для отладки

## Заключение

Retry механизм с экспоненциальной задержкой - мощный инструмент для повышения надежности интеграций. Используйте его разумно, мониторьте метрики и адаптируйте под свои нужды.
