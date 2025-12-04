# Примеры метрик и мониторинга

## Базовое использование

### 1. Инициализация сервисов

```go
package main

import (
 "database/sql"
 "github.com/gin-gonic/gin"
 "dmintegroff/internal/services"
 "dmintegroff/internal/controllers"
)

func main() {
 // Подключение к БД
 db, _ := sql.Open("postgres", "...")
 
 // Создание сервисов
 metricsService := services.NewMetricsService()
 healthService := services.NewHealthService(db, "1.0.0")
 
 // Создание контроллера
 metricsController := controllers.NewMetricsController(healthService)
 
 // Регистрация маршрутов
 router := gin.Default()
 metricsController.RegisterRoutes(router)
 
 router.Run(":8080")
}
```

### 2. Запись метрик webhook

```go
func ProcessWebhook(integrationID, integrationName, projectID string, payload []byte) error {
 // Записываем начало запроса
 metricsService.RecordWebhookRequest(integrationID, integrationName, projectID)
 
 start := time.Now()
 
 // Выполняем запрос
 err := sendWebhook(payload)
 duration := time.Since(start)
 
 if err != nil {
 // Записываем ошибку
 errorType := classifyError(err)
 metricsService.RecordWebhookFailure(
 integrationID, 
 integrationName, 
 projectID, 
 errorType, 
 duration,
 )
 return err
 }
 
 // Записываем успех
 metricsService.RecordWebhookSuccess(
 integrationID, 
 integrationName, 
 projectID, 
 duration,
 )
 
 return nil
}

func classifyError(err error) string {
 switch {
 case strings.Contains(err.Error(), "timeout"):
 return "timeout"
 case strings.Contains(err.Error(), "connection"):
 return "connection"
 case strings.Contains(err.Error(), "401"):
 return "auth"
 case strings.Contains(err.Error(), "429"):
 return "rate_limit"
 default:
 return "unknown"
 }
}
```

### 3. Метрики с retry

```go
func ProcessWebhookWithRetry(integrationID, integrationName, projectID string, payload []byte) error {
 maxAttempts := 3
 
 for attempt := 1; attempt <= maxAttempts; attempt++ {
 err := ProcessWebhook(integrationID, integrationName, projectID, payload)
 
 if err == nil {
 // Успех - записываем количество попыток
 metricsService.RecordRetryAttempts(integrationID, attempt, true)
 return nil
 }
 
 if attempt < maxAttempts {
 // Записываем retry попытку
 metricsService.RecordWebhookRetry(integrationID, integrationName, attempt)
 
 // Ждем перед следующей попыткой
 delay := time.Duration(attempt) * time.Second
 time.Sleep(delay)
 }
 }
 
 // Все попытки исчерпаны
 metricsService.RecordRetryAttempts(integrationID, maxAttempts, false)
 return fmt.Errorf("failed after %d attempts", maxAttempts)
}
```

### 4. OAuth метрики

```go
func GetOAuthToken(integrationID string) (string, error) {
 // Проверяем кэш
 if token, found := tokenCache.Get(integrationID); found {
 metricsService.RecordOAuthCacheHit(integrationID, true)
 return token, nil
 }
 
 metricsService.RecordOAuthCacheHit(integrationID, false)
 
 // Получаем новый токен
 start := time.Now()
 token, err := requestNewToken(integrationID)
 duration := time.Since(start)
 
 success := err == nil
 metricsService.RecordOAuthTokenRefresh(integrationID, success, duration)
 
 if success {
 tokenCache.Set(integrationID, token)
 }
 
 return token, err
}
```

### 5. Метрики подписей

```go
func SignWebhook(integrationID string, payload []byte, algorithm string) (string, error) {
 signature, err := generateSignature(payload, algorithm)
 
 if err == nil {
 metricsService.RecordSignatureGenerated(integrationID, algorithm)
 }
 
 return signature, err
}

func VerifyWebhook(integrationID string, payload []byte, signature string) bool {
 valid := verifySignature(payload, signature)
 metricsService.RecordSignatureVerified(integrationID, valid)
 return valid
}
```

## Prometheus запросы

### Успешность webhook по интеграциям

```promql
# Процент успешных запросов за последние 5 минут
sum(rate(dmintegroff_webhook_success_total[5m])) by (integration_name) /
sum(rate(dmintegroff_webhook_total[5m])) by (integration_name) * 100
```

### Топ медленных интеграций

```promql
# P95 время выполнения по интеграциям
topk(5, 
 histogram_quantile(0.95, 
 sum(rate(dmintegroff_webhook_duration_seconds_bucket[5m])) by (integration_name, le)
 )
)
```

### Частота ошибок по типам

```promql
# Количество ошибок по типам за последний час
sum(increase(dmintegroff_webhook_failure_total[1h])) by (error_type)
```

### Эффективность retry

```promql
# Процент запросов, успешных после retry
sum(rate(dmintegroff_retry_attempts{final_status="success"}[5m])) /
sum(rate(dmintegroff_retry_attempts[5m])) * 100
```

### OAuth кэш hit rate

```promql
# Процент попаданий в кэш OAuth токенов
sum(rate(dmintegroff_oauth_token_cache_total{result="hit"}[5m])) /
sum(rate(dmintegroff_oauth_token_cache_total[5m])) * 100
```

## Grafana дашборды

### Панель: Webhook Overview

```json
{
 "title": "Webhook Overview",
 "panels": [
 {
 "title": "Total Requests",
 "type": "stat",
 "targets": [{
 "expr": "sum(rate(dmintegroff_webhook_total[5m]))"
 }]
 },
 {
 "title": "Success Rate",
 "type": "gauge",
 "targets": [{
 "expr": "sum(rate(dmintegroff_webhook_success_total[5m])) / sum(rate(dmintegroff_webhook_total[5m])) * 100"
 }],
 "fieldConfig": {
 "min": 0,
 "max": 100,
 "thresholds": {
 "steps": [
 {"value": 0, "color": "red"},
 {"value": 90, "color": "yellow"},
 {"value": 95, "color": "green"}
 ]
 }
 }
 },
 {
 "title": "Request Duration P95",
 "type": "graph",
 "targets": [{
 "expr": "histogram_quantile(0.95, sum(rate(dmintegroff_webhook_duration_seconds_bucket[5m])) by (le))"
 }]
 }
 ]
}
```

### Панель: Integration Performance

```json
{
 "title": "Integration Performance",
 "panels": [
 {
 "title": "Requests by Integration",
 "type": "graph",
 "targets": [{
 "expr": "sum(rate(dmintegroff_webhook_total[5m])) by (integration_name)"
 }],
 "legend": {
 "show": true
 }
 },
 {
 "title": "Success Rate by Integration",
 "type": "table",
 "targets": [{
 "expr": "sum(rate(dmintegroff_webhook_success_total[5m])) by (integration_name) / sum(rate(dmintegroff_webhook_total[5m])) by (integration_name) * 100",
 "format": "table"
 }]
 }
 ]
}
```

### Панель: OAuth Metrics

```json
{
 "title": "OAuth Metrics",
 "panels": [
 {
 "title": "Token Refresh Rate",
 "type": "graph",
 "targets": [{
 "expr": "sum(rate(dmintegroff_oauth_token_refresh_total[5m])) by (status)"
 }]
 },
 {
 "title": "Cache Hit Rate",
 "type": "gauge",
 "targets": [{
 "expr": "sum(rate(dmintegroff_oauth_token_cache_total{result=\"hit\"}[5m])) / sum(rate(dmintegroff_oauth_token_cache_total[5m])) * 100"
 }],
 "fieldConfig": {
 "min": 0,
 "max": 100,
 "thresholds": {
 "steps": [
 {"value": 0, "color": "red"},
 {"value": 70, "color": "yellow"},
 {"value": 80, "color": "green"}
 ]
 }
 }
 }
 ]
}
```

## Kubernetes примеры

### Deployment с метриками

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
 name: dmintegroff
 labels:
 app: dmintegroff
spec:
 replicas: 3
 selector:
 matchLabels:
 app: dmintegroff
 template:
 metadata:
 labels:
 app: dmintegroff
 annotations:
 prometheus.io/scrape: "true"
 prometheus.io/port: "8080"
 prometheus.io/path: "/metrics"
 spec:
 containers:
- name: dmintegroff
 image: dmintegroff:1.0.0
 ports:
- name: http
 containerPort: 8080
 env:
- name: DATABASE_URL
 valueFrom:
 secretKeyRef:
 name: dmintegroff-secrets
 key: database-url
 livenessProbe:
 httpGet:
 path: /metrics/health/live
 port: 8080
 initialDelaySeconds: 30
 periodSeconds: 10
 timeoutSeconds: 5
 failureThreshold: 3
 readinessProbe:
 httpGet:
 path: /metrics/health/ready
 port: 8080
 initialDelaySeconds: 10
 periodSeconds: 5
 timeoutSeconds: 3
 failureThreshold: 3
 resources:
 requests:
 memory: "128Mi"
 cpu: "100m"
 limits:
 memory: "512Mi"
 cpu: "500m"
```

### Service для метрик

```yaml
apiVersion: v1
kind: Service
metadata:
 name: dmintegroff
 labels:
 app: dmintegroff
spec:
 type: ClusterIP
 ports:
- port: 8080
 targetPort: 8080
 protocol: TCP
 name: http
 selector:
 app: dmintegroff
```

### ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
 name: dmintegroff
 labels:
 app: dmintegroff
spec:
 selector:
 matchLabels:
 app: dmintegroff
 endpoints:
- port: http
 path: /metrics
 interval: 15s
 scrapeTimeout: 10s
```

## Алерты

### PrometheusRule

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
 name: dmintegroff-alerts
spec:
 groups:
- name: dmintegroff
 interval: 30s
 rules:
# Высокий процент ошибок
- alert: HighWebhookFailureRate
 expr: |
 (
 sum(rate(dmintegroff_webhook_failure_total[5m])) /
 sum(rate(dmintegroff_webhook_total[5m]))
 ) > 0.05
 for: 5m
 labels:
 severity: warning
 component: webhook
 annotations:
 summary: "High webhook failure rate"
 description: "Webhook failure rate is {{ $value | humanizePercentage }} (threshold: 5%)"
 
# Критический процент ошибок
- alert: CriticalWebhookFailureRate
 expr: |
 (
 sum(rate(dmintegroff_webhook_failure_total[5m])) /
 sum(rate(dmintegroff_webhook_total[5m]))
 ) > 0.20
 for: 2m
 labels:
 severity: critical
 component: webhook
 annotations:
 summary: "Critical webhook failure rate"
 description: "Webhook failure rate is {{ $value | humanizePercentage }} (threshold: 20%)"
 
# Медленные запросы
- alert: SlowWebhookRequests
 expr: |
 histogram_quantile(0.95,
 sum(rate(dmintegroff_webhook_duration_seconds_bucket[5m])) by (le)
 ) > 5
 for: 10m
 labels:
 severity: warning
 component: webhook
 annotations:
 summary: "Slow webhook requests"
 description: "P95 webhook duration is {{ $value }}s (threshold: 5s)"
 
# Много retry
- alert: HighRetryRate
 expr: sum(rate(dmintegroff_webhook_retries_total[5m])) > 10
 for: 5m
 labels:
 severity: warning
 component: retry
 annotations:
 summary: "High retry rate"
 description: "Retry rate is {{ $value }} per second (threshold: 10/s)"
 
# OAuth проблемы
- alert: LowOAuthCacheHitRate
 expr: |
 (
 sum(rate(dmintegroff_oauth_token_cache_total{result="hit"}[5m])) /
 sum(rate(dmintegroff_oauth_token_cache_total[5m]))
 ) < 0.70
 for: 10m
 labels:
 severity: warning
 component: oauth
 annotations:
 summary: "Low OAuth cache hit rate"
 description: "OAuth cache hit rate is {{ $value | humanizePercentage }} (threshold: 70%)"
 
# База данных
- alert: DatabaseUnhealthy
 expr: up{job="dmintegroff"} == 0
 for: 1m
 labels:
 severity: critical
 component: database
 annotations:
 summary: "Database is unhealthy"
 description: "DMIntegroff cannot connect to database"
```

## Логирование

### Structured logging

```go
import (
 "github.com/sirupsen/logrus"
)

func setupLogging() {
 // JSON формат для production
 logrus.SetFormatter(&logrus.JSONFormatter{})
 
 // Уровень логирования
 logrus.SetLevel(logrus.InfoLevel)
}

func logWebhookRequest(integrationID, integrationName string, duration time.Duration, err error) {
 fields := logrus.Fields{
 "integration_id": integrationID,
 "integration_name": integrationName,
 "duration_ms": duration.Milliseconds(),
 }
 
 if err != nil {
 fields["error"] = err.Error()
 logrus.WithFields(fields).Error("Webhook request failed")
 } else {
 logrus.WithFields(fields).Info("Webhook request successful")
 }
}
```

### Пример лога

```json
{
 "level": "info",
 "msg": "Webhook request successful",
 "integration_id": "1",
 "integration_name": "GitHub",
 "duration_ms": 245,
 "time": "2024-12-04T10:00:00Z"
}
```

## См. также

- [METRICS_MONITORING.md](METRICS_MONITORING.md) - Полная документация
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
