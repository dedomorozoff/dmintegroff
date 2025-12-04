# Webhook подписи для безопасности

## Обзор

Webhook подписи (HMAC signatures) обеспечивают безопасность и аутентичность webhook запросов, позволяя получателю проверить, что запрос действительно пришел от доверенного источника и не был изменен в процессе передачи.

## Зачем нужны подписи?

### Проблемы без подписей
- Любой может отправить поддельный webhook на ваш endpoint
- Данные могут быть изменены злоумышленником (MITM атака)
- Невозможно проверить источник запроса
- Replay атаки (повторная отправка старых запросов)

### Преимущества с подписями
- Гарантия подлинности источника
- Защита от изменения данных
- Защита от поддельных запросов
- Соответствие стандартам безопасности

## Как это работает

### Процесс подписи (отправка)

```
1. Подготовка данных
 ┌─────────────────────┐
 │ {"user": "john"} │
 └──────────┬──────────┘
 │
2. Генерация HMAC
 ┌──────────▼──────────┐
 │ HMAC-SHA256( │
 │ data, │
 │ secret_key │
 │ ) │
 └──────────┬──────────┘
 │
3. Добавление заголовка
 ┌──────────▼──────────────────────────────┐
 │ X-Webhook-Signature: │
 │ sha256=abc123def456... │
 └─────────────────────────────────────────┘
 │
4. Отправка запроса
 ┌──────────▼──────────┐
 │ POST /webhook │
 │ Headers + Body │
 └─────────────────────┘
```

### Процесс верификации (получение)

```
1. Получение запроса
 ┌─────────────────────┐
 │ POST /webhook │
 │ + Signature header │
 └──────────┬──────────┘
 │
2. Извлечение подписи
 ┌──────────▼──────────┐
 │ X-Webhook-Signature │
 │ sha256=abc123... │
 └──────────┬──────────┘
 │
3. Вычисление ожидаемой подписи
 ┌──────────▼──────────┐
 │ HMAC-SHA256( │
 │ received_body, │
 │ secret_key │
 │ ) │
 └──────────┬──────────┘
 │
4. Сравнение
 ┌──────────▼──────────┐
 │ received_signature │
 │ == expected? │
 └──────────┬──────────┘
 │
 ┌────┴────┐
 │ │
 Да Нет
 Accept Reject
```

## Конфигурация

### Поля в модели Integration

```go
type Integration struct {
 // ... другие поля ...
 
 // Webhook Signature Configuration
 WebhookSignatureEnabled bool // Включить подписи
 WebhookSignatureSecret string // Секретный ключ
 WebhookSignatureHeader string // Имя заголовка (по умолчанию: X-Webhook-Signature)
 WebhookSignatureAlgorithm string // Алгоритм: sha256, sha512, sha1
}
```

### Параметры по умолчанию

```go
WebhookSignatureEnabled: false // Выключено по умолчанию
WebhookSignatureHeader: "X-Webhook-Signature" // Стандартный заголовок
WebhookSignatureAlgorithm: "sha256" // SHA-256 (рекомендуется)
```

## Поддерживаемые алгоритмы

### SHA-256 (рекомендуется)
- **Длина**: 256 бит (64 hex символа)
- **Безопасность**: Высокая
- **Производительность**: Отличная
- **Использование**: Большинство современных API (GitHub, Stripe, Slack)

```
sha256=5d41402abc4b2a76b9719d911017c592
```

### SHA-512 (максимальная безопасность)
- **Длина**: 512 бит (128 hex символов)
- **Безопасность**: Очень высокая
- **Производительность**: Хорошая
- **Использование**: Критичные системы

```
sha512=cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce
```

### SHA-1 (устаревший)
- **Длина**: 160 бит (40 hex символов)
- **Безопасность**: Низкая (не рекомендуется)
- **Производительность**: Отличная
- **Использование**: Только для совместимости со старыми системами

```
sha1=aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
```

## Примеры использования

### 1. Базовая настройка

```go
integration := &models.Integration{
 Name: "My Secure Webhook",
 TargetAPI: "https://api.example.com/webhook",
 WebhookSignatureEnabled: true,
 WebhookSignatureSecret: "my-super-secret-key-12345",
 WebhookSignatureHeader: "X-Webhook-Signature",
 WebhookSignatureAlgorithm: "sha256",
}
```

### 2. Генерация секретного ключа

```go
// Автоматическая генерация безопасного ключа
secret, err := GenerateRandomSecret(32) // 32 байта = 64 hex символа
if err != nil {
 log.Fatal(err)
}

integration.WebhookSignatureSecret = secret
// Результат: "a1b2c3d4e5f6...64 символа"
```

### 3. Отправка webhook с подписью

```go
// Автоматически добавляется в ProcessWebhook()
err := ProcessWebhook(integrationID, payload)

// Отправленный запрос будет содержать:
// POST /webhook HTTP/1.1
// Content-Type: application/json
// X-Webhook-Signature: sha256=5d41402abc4b2a76b9719d911017c592
//
// {"user":"john","action":"created"}
```

### 4. Проверка входящей подписи

```go
// В вашем webhook handler
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
 // Читаем тело запроса
 body, _ := io.ReadAll(r.Body)
 
 // Загружаем конфигурацию интеграции
 var integration models.Integration
 db.First(&integration, integrationID)
 
 // Проверяем подпись
 if err := VerifyIncomingSignature(r, body, &integration); err != nil {
 http.Error(w, "Invalid signature", http.StatusUnauthorized)
 return
 }
 
 // Подпись валидна, обрабатываем запрос
 processWebhook(body)
}
```

### 5. Кастомный заголовок (совместимость с GitHub)

```go
integration := &models.Integration{
 WebhookSignatureEnabled: true,
 WebhookSignatureSecret: "github-webhook-secret",
 WebhookSignatureHeader: "X-Hub-Signature-256", // GitHub стиль
 WebhookSignatureAlgorithm: "sha256",
}
```

### 6. Валидация конфигурации

```go
// Перед сохранением интеграции
if err := ValidateSignatureConfig(integration); err != nil {
 log.Error("Invalid signature config:", err)
 return err
}

// Проверяет:
// - Секрет не пустой (если подписи включены)
// - Секрет минимум 16 символов
// - Алгоритм поддерживается
```

## Интеграция с популярными сервисами

### GitHub Webhooks

```go
integration := &models.Integration{
 WebhookSignatureEnabled: true,
 WebhookSignatureSecret: "your-github-secret",
 WebhookSignatureHeader: "X-Hub-Signature-256",
 WebhookSignatureAlgorithm: "sha256",
}
```

### Stripe Webhooks

```go
integration := &models.Integration{
 WebhookSignatureEnabled: true,
 WebhookSignatureSecret: "whsec_...", // Stripe signing secret
 WebhookSignatureHeader: "Stripe-Signature",
 WebhookSignatureAlgorithm: "sha256",
}
```

### Slack Webhooks

```go
integration := &models.Integration{
 WebhookSignatureEnabled: true,
 WebhookSignatureSecret: "your-slack-signing-secret",
 WebhookSignatureHeader: "X-Slack-Signature",
 WebhookSignatureAlgorithm: "sha256",
}
```

### Custom API

```go
integration := &models.Integration{
 WebhookSignatureEnabled: true,
 WebhookSignatureSecret: "shared-secret-key",
 WebhookSignatureHeader: "X-Webhook-Signature",
 WebhookSignatureAlgorithm: "sha256",
}
```

## Формат подписи

### С префиксом алгоритма (рекомендуется)

```
X-Webhook-Signature: sha256=5d41402abc4b2a76b9719d911017c592
```

Преимущества:
- Явно указан алгоритм
- Поддержка нескольких алгоритмов
- Совместимость с GitHub, Stripe

### Без префикса

```
X-Webhook-Signature: 5d41402abc4b2a76b9719d911017c592
```

Используется алгоритм из конфигурации интеграции.

## Безопасность

### Лучшие практики

1. **Используйте длинные секреты**
 ```go
 // Плохо
 secret := "12345"
 
 // Хорошо
 secret, _ := GenerateRandomSecret(32) // 64 hex символа
 ```

2. **Используйте SHA-256 или SHA-512**
 ```go
 // Не рекомендуется
 algorithm := "sha1"
 
 // Рекомендуется
 algorithm := "sha256"
 ```

3. **Храните секреты безопасно**
 ```go
 // В переменных окружения
 secret := os.Getenv("WEBHOOK_SECRET")
 
 // В зашифрованной БД
 // В секретном хранилище (Vault, AWS Secrets Manager)
 ```

4. **Ротация секретов**
 ```go
 // Периодически обновляйте секреты
 newSecret, _ := GenerateRandomSecret(32)
 integration.WebhookSignatureSecret = newSecret
 db.Save(&integration)
 ```

5. **Проверяйте подписи на стороне получателя**
 ```go
 // Всегда проверяйте подпись перед обработкой
 if err := VerifyIncomingSignature(r, body, integration); err != nil {
 return http.StatusUnauthorized
 }
 ```

### Защита от атак

#### Timing Attack Protection
```go
// Используется constant-time сравнение
func VerifySignature(payload []byte, received string, secret string, algo SignatureAlgorithm) bool {
 expected, _ := GenerateSignature(payload, secret, algo)
 return hmac.Equal([]byte(received), []byte(expected)) // Безопасно
}
```

#### Replay Attack Protection
```go
// Добавьте timestamp в payload
payload := map[string]interface{}{
 "data": actualData,
 "timestamp": time.Now().Unix(),
}

// На стороне получателя проверяйте timestamp
if time.Now().Unix() - payload["timestamp"] > 300 { // 5 минут
 return errors.New("request too old")
}
```

## Логирование

### Успешная подпись

```
level=debug msg="Added webhook signature to request"
 integration_id=123
 header="X-Webhook-Signature"
 algorithm="sha256"
```

### Ошибка подписи

```
level=error msg="Failed to add webhook signature"
 integration_id=123
 error="signature secret is empty"
```

### Проверка подписи

```
level=debug msg="Webhook signature verified successfully"
 integration_id=123
 algorithm="sha256"
```

### Неверная подпись

```
level=warning msg="Webhook signature verification failed"
 integration_id=123
 algorithm="sha256"
```

## Troubleshooting

### Проблема: Подпись не совпадает

**Причины:**
1. Разные секретные ключи
2. Разные алгоритмы
3. Изменение тела запроса (пробелы, encoding)

**Решение:**
```go
// Убедитесь, что используется одинаковый секрет
log.Printf("Secret: %s", integration.WebhookSignatureSecret)

// Проверьте алгоритм
log.Printf("Algorithm: %s", integration.WebhookSignatureAlgorithm)

// Логируйте тело запроса
log.Printf("Body: %s", string(body))
```

### Проблема: Заголовок не найден

**Причина:** Неправильное имя заголовка

**Решение:**
```go
// Проверьте имя заголовка
log.Printf("Expected header: %s", integration.WebhookSignatureHeader)

// Логируйте все заголовки
for name, values := range r.Header {
 log.Printf("Header: %s = %v", name, values)
}
```

### Проблема: Секрет слишком короткий

**Причина:** Секрет меньше 16 символов

**Решение:**
```go
// Сгенерируйте новый безопасный секрет
secret, _ := GenerateRandomSecret(32)
integration.WebhookSignatureSecret = secret
```

## Тестирование

### Запуск тестов

```bash
cd dmintegroff
go test -v ./internal/services/webhook_signature_test.go ./internal/services/webhook_signature.go
```

### Покрытие тестами

- Генерация подписи (SHA256, SHA512, SHA1)
- Верификация подписи
- Добавление подписи в запрос
- Проверка входящей подписи
- Генерация случайного секрета
- Валидация конфигурации
- Обработка ошибок

## API Reference

### GenerateSignature

```go
func GenerateSignature(payload []byte, secret string, algorithm SignatureAlgorithm) (string, error)
```

Генерирует HMAC подпись для данных.

### VerifySignature

```go
func VerifySignature(payload []byte, receivedSignature string, secret string, algorithm SignatureAlgorithm) bool
```

Проверяет HMAC подпись (constant-time comparison).

### AddSignatureToRequest

```go
func AddSignatureToRequest(req *http.Request, payload []byte, integration *models.Integration) error
```

Добавляет подпись в исходящий HTTP запрос.

### VerifyIncomingSignature

```go
func VerifyIncomingSignature(req *http.Request, payload []byte, integration *models.Integration) error
```

Проверяет подпись входящего webhook запроса.

### GenerateRandomSecret

```go
func GenerateRandomSecret(length int) (string, error)
```

Генерирует криптографически безопасный случайный секрет.

### ValidateSignatureConfig

```go
func ValidateSignatureConfig(integration *models.Integration) error
```

Валидирует конфигурацию webhook подписей.

## Заключение

Webhook подписи - критически важный элемент безопасности для любой интеграции. Они обеспечивают:

- Аутентификацию источника
- Целостность данных
- Защиту от атак
- Соответствие стандартам

Всегда включайте подписи для production интеграций!
