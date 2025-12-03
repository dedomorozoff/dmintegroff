# Примеры использования Webhook подписей

## Базовые примеры

### 1. Включение подписей для новой интеграции

```go
package main

import (
    "dmintegroff/internal/models"
    "dmintegroff/internal/services"
)

func CreateSecureIntegration() {
    // Генерируем безопасный секрет
    secret, err := services.GenerateRandomSecret(32)
    if err != nil {
        log.Fatal("Failed to generate secret:", err)
    }
    
    integration := &models.Integration{
        Name:      "Secure Webhook Integration",
        TargetAPI: "https://api.example.com/webhook",
        
        // Включаем подписи
        WebhookSignatureEnabled:   true,
        WebhookSignatureSecret:    secret,
        WebhookSignatureHeader:    "X-Webhook-Signature",
        WebhookSignatureAlgorithm: "sha256",
    }
    
    // Валидируем конфигурацию
    if err := services.ValidateSignatureConfig(integration); err != nil {
        log.Fatal("Invalid signature config:", err)
    }
    
    // Сохраняем
    db.Create(integration)
    
    log.Printf("Created integration with signature secret: %s", secret)
}
```

### 2. Отправка webhook с подписью

```go
func SendSecureWebhook() {
    integrationID := uint(123)
    payload := map[string]interface{}{
        "event": "user.created",
        "user": map[string]interface{}{
            "id":    456,
            "name":  "John Doe",
            "email": "john@example.com",
        },
        "timestamp": time.Now().Unix(),
    }
    
    // Подпись автоматически добавляется в ProcessWebhook
    err := services.ProcessWebhook(integrationID, payload)
    if err != nil {
        log.Error("Failed to send webhook:", err)
        return
    }
    
    log.Info("Webhook sent successfully with signature")
}
```

### 3. Проверка входящего webhook

```go
func HandleIncomingWebhook(w http.ResponseWriter, r *http.Request) {
    // Читаем тело запроса
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // Загружаем конфигурацию интеграции
    var integration models.Integration
    if err := db.First(&integration, integrationID).Error; err != nil {
        http.Error(w, "Integration not found", http.StatusNotFound)
        return
    }
    
    // Проверяем подпись
    if err := services.VerifyIncomingSignature(r, body, &integration); err != nil {
        log.Warn("Invalid webhook signature:", err)
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Подпись валидна - обрабатываем webhook
    var payload map[string]interface{}
    if err := json.Unmarshal(body, &payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    processWebhookData(payload)
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"success"}`))
}
```

## Интеграция с популярными сервисами

### GitHub Webhooks

```go
func SetupGitHubWebhook() {
    integration := &models.Integration{
        Name:      "GitHub Repository Events",
        TargetAPI: "https://api.example.com/github-webhook",
        
        // GitHub использует X-Hub-Signature-256
        WebhookSignatureEnabled:   true,
        WebhookSignatureSecret:    "your-github-webhook-secret",
        WebhookSignatureHeader:    "X-Hub-Signature-256",
        WebhookSignatureAlgorithm: "sha256",
    }
    
    db.Create(integration)
}

// Проверка GitHub webhook
func VerifyGitHubWebhook(r *http.Request, body []byte) error {
    signature := r.Header.Get("X-Hub-Signature-256")
    if signature == "" {
        return errors.New("missing X-Hub-Signature-256 header")
    }
    
    // GitHub формат: sha256=<signature>
    if !strings.HasPrefix(signature, "sha256=") {
        return errors.New("invalid signature format")
    }
    
    signatureValue := strings.TrimPrefix(signature, "sha256=")
    
    secret := "your-github-webhook-secret"
    if !services.VerifySignature(body, signatureValue, secret, services.AlgorithmSHA256) {
        return errors.New("signature verification failed")
    }
    
    return nil
}
```

### Stripe Webhooks

```go
func SetupStripeWebhook() {
    integration := &models.Integration{
        Name:      "Stripe Payment Events",
        TargetAPI: "https://api.example.com/stripe-webhook",
        
        // Stripe использует Stripe-Signature
        WebhookSignatureEnabled:   true,
        WebhookSignatureSecret:    "whsec_...", // Stripe signing secret
        WebhookSignatureHeader:    "Stripe-Signature",
        WebhookSignatureAlgorithm: "sha256",
    }
    
    db.Create(integration)
}

// Stripe использует более сложный формат с timestamp
func VerifyStripeWebhook(r *http.Request, body []byte) error {
    signature := r.Header.Get("Stripe-Signature")
    if signature == "" {
        return errors.New("missing Stripe-Signature header")
    }
    
    // Stripe формат: t=timestamp,v1=signature
    // Для простоты используем базовую проверку
    parts := strings.Split(signature, ",")
    var signatureValue string
    
    for _, part := range parts {
        if strings.HasPrefix(part, "v1=") {
            signatureValue = strings.TrimPrefix(part, "v1=")
            break
        }
    }
    
    if signatureValue == "" {
        return errors.New("signature not found in header")
    }
    
    secret := "whsec_..."
    if !services.VerifySignature(body, signatureValue, secret, services.AlgorithmSHA256) {
        return errors.New("signature verification failed")
    }
    
    return nil
}
```

### Slack Webhooks

```go
func SetupSlackWebhook() {
    integration := &models.Integration{
        Name:      "Slack Events",
        TargetAPI: "https://api.example.com/slack-webhook",
        
        // Slack использует X-Slack-Signature
        WebhookSignatureEnabled:   true,
        WebhookSignatureSecret:    "your-slack-signing-secret",
        WebhookSignatureHeader:    "X-Slack-Signature",
        WebhookSignatureAlgorithm: "sha256",
    }
    
    db.Create(integration)
}

// Slack включает timestamp в подпись
func VerifySlackWebhook(r *http.Request, body []byte) error {
    signature := r.Header.Get("X-Slack-Signature")
    timestamp := r.Header.Get("X-Slack-Request-Timestamp")
    
    if signature == "" || timestamp == "" {
        return errors.New("missing Slack signature headers")
    }
    
    // Проверяем timestamp (защита от replay атак)
    ts, _ := strconv.ParseInt(timestamp, 10, 64)
    if time.Now().Unix()-ts > 300 { // 5 минут
        return errors.New("request timestamp too old")
    }
    
    // Slack формат: v0=<signature>
    signatureValue := strings.TrimPrefix(signature, "v0=")
    
    // Slack подписывает: v0:timestamp:body
    signedData := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
    
    secret := "your-slack-signing-secret"
    if !services.VerifySignature([]byte(signedData), signatureValue, secret, services.AlgorithmSHA256) {
        return errors.New("signature verification failed")
    }
    
    return nil
}
```

## Продвинутые сценарии

### 4. Ротация секретов

```go
func RotateWebhookSecret(integrationID uint) error {
    var integration models.Integration
    if err := db.First(&integration, integrationID).Error; err != nil {
        return err
    }
    
    // Сохраняем старый секрет для переходного периода
    oldSecret := integration.WebhookSignatureSecret
    
    // Генерируем новый секрет
    newSecret, err := services.GenerateRandomSecret(32)
    if err != nil {
        return err
    }
    
    // Обновляем секрет
    integration.WebhookSignatureSecret = newSecret
    if err := db.Save(&integration).Error; err != nil {
        return err
    }
    
    log.Printf("Rotated secret for integration %d", integrationID)
    log.Printf("Old secret: %s", oldSecret)
    log.Printf("New secret: %s", newSecret)
    
    // Уведомляем администратора о необходимости обновить секрет на стороне получателя
    notifySecretRotation(integrationID, newSecret)
    
    return nil
}
```

### 5. Поддержка нескольких секретов (graceful rotation)

```go
type IntegrationWithMultipleSecrets struct {
    models.Integration
    SecondarySecret string `json:"secondary_secret"`
}

func VerifyWithMultipleSecrets(r *http.Request, body []byte, integration *IntegrationWithMultipleSecrets) error {
    signature := r.Header.Get(integration.WebhookSignatureHeader)
    if signature == "" {
        return errors.New("signature header missing")
    }
    
    // Извлекаем значение подписи
    parts := strings.SplitN(signature, "=", 2)
    algorithm := services.SignatureAlgorithm(parts[0])
    signatureValue := parts[1]
    
    // Пробуем основной секрет
    if services.VerifySignature(body, signatureValue, integration.WebhookSignatureSecret, algorithm) {
        return nil
    }
    
    // Пробуем вторичный секрет (для переходного периода)
    if integration.SecondarySecret != "" {
        if services.VerifySignature(body, signatureValue, integration.SecondarySecret, algorithm) {
            log.Warn("Webhook verified with secondary secret - update to primary")
            return nil
        }
    }
    
    return errors.New("signature verification failed with all secrets")
}
```

### 6. Защита от replay атак с timestamp

```go
func SendWebhookWithTimestamp(integrationID uint, data map[string]interface{}) error {
    // Добавляем timestamp в payload
    payload := map[string]interface{}{
        "data":      data,
        "timestamp": time.Now().Unix(),
        "nonce":     generateNonce(), // Дополнительная защита
    }
    
    return services.ProcessWebhook(integrationID, payload)
}

func VerifyWebhookWithTimestamp(r *http.Request, body []byte, integration *models.Integration) error {
    // Сначала проверяем подпись
    if err := services.VerifyIncomingSignature(r, body, integration); err != nil {
        return err
    }
    
    // Парсим payload
    var payload map[string]interface{}
    if err := json.Unmarshal(body, &payload); err != nil {
        return err
    }
    
    // Проверяем timestamp
    timestamp, ok := payload["timestamp"].(float64)
    if !ok {
        return errors.New("missing or invalid timestamp")
    }
    
    age := time.Now().Unix() - int64(timestamp)
    if age > 300 { // 5 минут
        return fmt.Errorf("webhook too old: %d seconds", age)
    }
    
    if age < -60 { // Не более 1 минуты в будущем
        return errors.New("webhook timestamp in future")
    }
    
    return nil
}
```

### 7. Логирование и мониторинг подписей

```go
type SignatureMetrics struct {
    TotalVerifications   int64
    SuccessfulVerifications int64
    FailedVerifications  int64
    AverageVerificationTime time.Duration
}

var metrics SignatureMetrics

func VerifyWithMetrics(r *http.Request, body []byte, integration *models.Integration) error {
    start := time.Now()
    
    metrics.TotalVerifications++
    
    err := services.VerifyIncomingSignature(r, body, integration)
    
    duration := time.Since(start)
    metrics.AverageVerificationTime = (metrics.AverageVerificationTime + duration) / 2
    
    if err != nil {
        metrics.FailedVerifications++
        log.WithFields(map[string]interface{}{
            "integration_id": integration.ID,
            "error":          err.Error(),
            "duration_ms":    duration.Milliseconds(),
        }).Warn("Signature verification failed")
    } else {
        metrics.SuccessfulVerifications++
        log.WithFields(map[string]interface{}{
            "integration_id": integration.ID,
            "duration_ms":    duration.Milliseconds(),
        }).Debug("Signature verified successfully")
    }
    
    return err
}

func GetSignatureMetrics() SignatureMetrics {
    return metrics
}
```

### 8. Тестирование подписей

```go
func TestWebhookSignature(t *testing.T) {
    // Создаем тестовую интеграцию
    integration := &models.Integration{
        WebhookSignatureEnabled:   true,
        WebhookSignatureSecret:    "test-secret-key",
        WebhookSignatureHeader:    "X-Webhook-Signature",
        WebhookSignatureAlgorithm: "sha256",
    }
    
    // Тестовые данные
    payload := []byte(`{"test":"data"}`)
    
    // Генерируем подпись
    signature, err := services.GenerateSignature(
        payload,
        integration.WebhookSignatureSecret,
        services.AlgorithmSHA256,
    )
    if err != nil {
        t.Fatal(err)
    }
    
    // Создаем тестовый запрос
    req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
    req.Header.Set("X-Webhook-Signature", "sha256="+signature)
    
    // Проверяем подпись
    if err := services.VerifyIncomingSignature(req, payload, integration); err != nil {
        t.Errorf("Signature verification failed: %v", err)
    }
}
```

### 9. CLI инструмент для генерации подписей

```go
// cmd/signature-tool/main.go
package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    
    "dmintegroff/internal/services"
)

func main() {
    var (
        payloadFile = flag.String("payload", "", "Path to payload file")
        secret      = flag.String("secret", "", "Secret key")
        algorithm   = flag.String("algorithm", "sha256", "Algorithm (sha256, sha512, sha1)")
    )
    flag.Parse()
    
    if *payloadFile == "" || *secret == "" {
        fmt.Println("Usage: signature-tool -payload <file> -secret <key> [-algorithm sha256]")
        os.Exit(1)
    }
    
    // Читаем payload
    payload, err := ioutil.ReadFile(*payloadFile)
    if err != nil {
        fmt.Printf("Error reading payload: %v\n", err)
        os.Exit(1)
    }
    
    // Генерируем подпись
    algo := services.SignatureAlgorithm(*algorithm)
    signature, err := services.GenerateSignature(payload, *secret, algo)
    if err != nil {
        fmt.Printf("Error generating signature: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Signature: %s=%s\n", *algorithm, signature)
}
```

### 10. Webhook proxy с проверкой подписей

```go
func WebhookProxy(w http.ResponseWriter, r *http.Request) {
    // Читаем тело
    body, _ := io.ReadAll(r.Body)
    r.Body = io.NopCloser(bytes.NewReader(body))
    
    // Загружаем конфигурацию
    var integration models.Integration
    db.First(&integration, integrationID)
    
    // Проверяем входящую подпись
    if err := services.VerifyIncomingSignature(r, body, &integration); err != nil {
        log.Warn("Invalid incoming signature:", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Пересылаем на целевой endpoint
    targetReq, _ := http.NewRequest(r.Method, integration.TargetAPI, bytes.NewReader(body))
    
    // Копируем заголовки
    for name, values := range r.Header {
        for _, value := range values {
            targetReq.Header.Add(name, value)
        }
    }
    
    // Добавляем новую подпись для целевого API
    if err := services.AddSignatureToRequest(targetReq, body, &integration); err != nil {
        log.Error("Failed to add signature:", err)
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    
    // Отправляем
    client := &http.Client{}
    resp, err := client.Do(targetReq)
    if err != nil {
        log.Error("Failed to forward webhook:", err)
        http.Error(w, "Gateway error", http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()
    
    // Возвращаем ответ
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}
```

## Заключение

Эти примеры демонстрируют различные способы использования webhook подписей:

- ✅ Базовая настройка и использование
- ✅ Интеграция с популярными сервисами
- ✅ Ротация секретов
- ✅ Защита от replay атак
- ✅ Мониторинг и метрики
- ✅ Тестирование
- ✅ CLI инструменты
- ✅ Webhook proxy

Выбирайте подходящий пример для вашего use case!
