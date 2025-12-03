# 📊 Метрики и мониторинг

## Обзор

DMIntegroff предоставляет полный набор метрик и health checks для мониторинга системы в production.

## Возможности

### 1. Prometheus метрики
- Счетчики webhook запросов
- Гистограммы времени выполнения
- Метрики OAuth токенов
- Метрики retry попыток
- Метрики webhook подписей
- Gauges для текущего состояния

### 2. Health Checks
- Liveness probe - система жива?
- Readiness probe - система готова?
- Детальная проверка компонентов
- Статистика базы данных

### 3. Structured Logging
- JSON формат логов
- Уровни логирования
- Контекстная информация
- Трассировка запросов

## Endpoints

### Prometheus метрики

```
GET /metrics
```

Возвращает метрики в формате Prometheus.

**Пример ответа:**
```
# HELP dmintegroff_webhook_total Total number of webhook requests
# TYPE dmintegroff_webhook_total counter
dmintegroff_webhook_total{integration_id="1",integration_name="GitHub",project_id="1"} 150

# HELP dmintegroff_webhook_success_total Total number of successful webhook requests
# TYPE dmintegroff_webhook_success_total counter
dmintegroff_webhook_success_total{integration_id="1",integration_name="GitHub",project_id="1"} 145

# HELP dmintegroff_webhook_duration_seconds Webhook request duration in seconds
# TYPE dmintegroff_webhook_duration_seconds histogram
dmintegroff_webhook_duration_seconds_bucket{integration_id="1",integration_name="GitHub",status="success",le="0.5"} 120
dmintegroff_webhook_duration_seconds_bucket{integration_id="1",integration_name="GitHub",status="success",le="1"} 140
```

### Health Check

```
GET /metrics/health
```

Полная проверка здоровья системы.

**Пример ответа:**
```json
{
  "status": "healthy",
  "timestamp": "2024-12-04T10:00:00Z",
  "uptime": "24h30m15s",
  "version": "1.0.0",
  "components": {
    "database": {
      "status": "healthy",
      "message": "Database connection OK",
      "last_checked": "2024-12-04T10:00:00Z",
      "details": {
        "response_time_ms": 5,
        "open_connections": 10,
        "in_use": 2,
        "idle": 8,
        "max_open": 25
      }
    },
    "disk": {
      "status": "healthy",
      "message": "Disk space OK",
      "last_checked": "2024-12-04T10:00:00Z"
    },
    "memory": {
      "status": "healthy",
      "message": "Memory usage OK",
      "last_checked": "2024-12-04T10:00:00Z"
    }
  }
}
```

**Статусы:**
- `healthy` - все компоненты работают нормально (HTTP 200)
- `degraded` - некоторые компоненты работают медленно (HTTP 200)
- `unhealthy` - критические компоненты не работают (HTTP 503)

### Liveness Probe

```
GET /metrics/health/live
```

Проверка, что приложение работает (для Kubernetes).

**Пример ответа:**
```json
{
  "status": "alive",
  "uptime": "24h30m15s"
}
```

### Readiness Probe

```
GET /metrics/health/ready
```

Проверка, что приложение готово принимать запросы (для Kubernetes).

**Пример ответа (готово):**
```json
{
  "status": "ready"
}
```

**Пример ответа (не готово):**
```json
{
  "status": "not_ready"
}
```

## Доступные метрики

### Webhook метрики

#### dmintegroff_webhook_total
Общее количество webhook запросов.

**Тип:** Counter  
**Labels:**
- `integration_id` - ID интеграции
- `integration_name` - название интеграции
- `project_id` - ID проекта

#### dmintegroff_webhook_success_total
Количество успешных webhook запросов.

**Тип:** Counter  
**Labels:** integration_id, integration_name, project_id

#### dmintegroff_webhook_failure_total
Количество неудачных webhook запросов.

**Тип:** Counter  
**Labels:**
- `integration_id`
- `integration_name`
- `project_id`
- `error_type` - тип ошибки (timeout, connection, auth, etc.)

#### dmintegroff_webhook_retries_total
Количество retry попыток.

**Тип:** Counter  
**Labels:**
- `integration_id`
- `integration_name`
- `attempt` - номер попытки (1, 2, 3...)

#### dmintegroff_webhook_duration_seconds
Время выполнения webhook запроса.

**Тип:** Histogram  
**Labels:**
- `integration_id`
- `integration_name`
- `status` - success или failure

**Buckets:** 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10

### OAuth метрики

#### dmintegroff_oauth_token_refresh_total
Количество обновлений OAuth токенов.

**Тип:** Counter  
**Labels:**
- `integration_id`
- `status` - success или failure

#### dmintegroff_oauth_token_cache_total
Попадания и промахи кэша OAuth токенов.

**Тип:** Counter  
**Labels:**
- `integration_id`
- `result` - hit или miss

#### dmintegroff_oauth_duration_seconds
Время получения OAuth токена.

**Тип:** Histogram  
**Labels:**
- `integration_id`
- `status` - success или failure

**Buckets:** 0.1, 0.25, 0.5, 1, 2.5, 5, 10

### Retry метрики

#### dmintegroff_retry_attempts
Количество retry попыток на запрос.

**Тип:** Histogram  
**Labels:**
- `integration_id`
- `final_status` - success или failure

**Buckets:** 1, 2, 3, 4, 5

### Signature метрики

#### dmintegroff_signature_generated_total
Количество сгенерированных подписей.

**Тип:** Counter  
**Labels:**
- `integration_id`
- `algorithm` - sha256, sha512, sha1

#### dmintegroff_signature_verified_total
Количество проверенных подписей.

**Тип:** Counter  
**Labels:**
- `integration_id`
- `result` - valid или invalid

### System метрики

#### dmintegroff_active_integrations
Количество активных интеграций.

**Тип:** Gauge

#### dmintegroff_queue_size
Текущий размер очереди webhook.

**Тип:** Gauge

## Интеграция с Prometheus

### Конфигурация Prometheus

```yaml
scrape_configs:
  - job_name: 'dmintegroff'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Примеры запросов PromQL

#### Успешность webhook запросов
```promql
rate(dmintegroff_webhook_success_total[5m]) / 
rate(dmintegroff_webhook_total[5m]) * 100
```

#### P95 время выполнения webhook
```promql
histogram_quantile(0.95, 
  rate(dmintegroff_webhook_duration_seconds_bucket[5m])
)
```

#### Частота retry
```promql
rate(dmintegroff_webhook_retries_total[5m])
```

#### Эффективность кэша OAuth
```promql
rate(dmintegroff_oauth_token_cache_total{result="hit"}[5m]) /
rate(dmintegroff_oauth_token_cache_total[5m]) * 100
```

## Интеграция с Grafana

### Пример дашборда

```json
{
  "dashboard": {
    "title": "DMIntegroff Monitoring",
    "panels": [
      {
        "title": "Webhook Success Rate",
        "targets": [
          {
            "expr": "rate(dmintegroff_webhook_success_total[5m]) / rate(dmintegroff_webhook_total[5m]) * 100"
          }
        ]
      },
      {
        "title": "Webhook Duration P95",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(dmintegroff_webhook_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Active Integrations",
        "targets": [
          {
            "expr": "dmintegroff_active_integrations"
          }
        ]
      }
    ]
  }
}
```

## Kubernetes Integration

### Deployment с health checks

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dmintegroff
spec:
  template:
    spec:
      containers:
      - name: dmintegroff
        image: dmintegroff:latest
        ports:
        - containerPort: 8080
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
```

### ServiceMonitor для Prometheus Operator

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: dmintegroff
spec:
  selector:
    matchLabels:
      app: dmintegroff
  endpoints:
  - port: http
    path: /metrics
    interval: 15s
```

## Алерты

### Примеры правил алертинга

```yaml
groups:
- name: dmintegroff
  rules:
  # Высокий процент ошибок
  - alert: HighWebhookFailureRate
    expr: |
      rate(dmintegroff_webhook_failure_total[5m]) / 
      rate(dmintegroff_webhook_total[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High webhook failure rate"
      description: "Webhook failure rate is {{ $value | humanizePercentage }}"

  # Медленные webhook
  - alert: SlowWebhookRequests
    expr: |
      histogram_quantile(0.95, 
        rate(dmintegroff_webhook_duration_seconds_bucket[5m])
      ) > 5
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "Slow webhook requests"
      description: "P95 webhook duration is {{ $value }}s"

  # База данных недоступна
  - alert: DatabaseUnhealthy
    expr: up{job="dmintegroff"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Database is unhealthy"
      description: "DMIntegroff cannot connect to database"

  # Много retry попыток
  - alert: HighRetryRate
    expr: rate(dmintegroff_webhook_retries_total[5m]) > 10
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High retry rate"
      description: "Retry rate is {{ $value }} per second"
```

## Best Practices

### 1. Мониторинг в production

- ✅ Настройте Prometheus для сбора метрик каждые 15-30 секунд
- ✅ Используйте Grafana для визуализации
- ✅ Настройте алерты для критичных метрик
- ✅ Мониторьте health checks в Kubernetes

### 2. Метрики для отслеживания

**Критичные:**
- Успешность webhook запросов (>95%)
- Время выполнения P95 (<5s)
- Доступность базы данных (100%)

**Важные:**
- Частота retry попыток
- Эффективность кэша OAuth (>80%)
- Размер очереди webhook

### 3. Алерты

**Critical (немедленное действие):**
- База данных недоступна
- Успешность webhook <80%
- Приложение не отвечает

**Warning (требует внимания):**
- Успешность webhook <95%
- Медленные запросы (P95 >5s)
- Высокая частота retry

### 4. Retention

- Метрики: храните минимум 30 дней
- Логи: храните минимум 7 дней
- Health checks: храните 24 часа

## Troubleshooting

### Метрики не собираются

**Проблема:** Prometheus не может подключиться к /metrics

**Решение:**
1. Проверьте, что приложение запущено
2. Проверьте firewall правила
3. Проверьте конфигурацию Prometheus
4. Проверьте логи приложения

### Health check возвращает unhealthy

**Проблема:** /metrics/health возвращает status: unhealthy

**Решение:**
1. Проверьте компоненты в ответе
2. Проверьте подключение к базе данных
3. Проверьте логи приложения
4. Проверьте ресурсы системы (CPU, память, диск)

### Высокая частота retry

**Проблема:** Много retry попыток

**Решение:**
1. Проверьте доступность целевого API
2. Проверьте сетевое подключение
3. Проверьте rate limits целевого API
4. Увеличьте timeout если нужно

## Примеры использования

### Запись метрик в коде

```go
// Создание сервиса метрик
metricsService := services.NewMetricsService()

// Запись webhook запроса
metricsService.RecordWebhookRequest("1", "GitHub", "project-1")

// Запись успешного webhook
start := time.Now()
// ... выполнение запроса ...
duration := time.Since(start)
metricsService.RecordWebhookSuccess("1", "GitHub", "project-1", duration)

// Запись неудачного webhook
metricsService.RecordWebhookFailure("1", "GitHub", "project-1", "timeout", duration)

// Запись retry попытки
metricsService.RecordWebhookRetry("1", "GitHub", 2)

// Запись OAuth метрик
metricsService.RecordOAuthTokenRefresh("1", true, 200*time.Millisecond)
metricsService.RecordOAuthCacheHit("1", true)

// Запись метрик подписей
metricsService.RecordSignatureGenerated("1", "sha256")
metricsService.RecordSignatureVerified("1", true)

// Обновление gauges
metricsService.SetActiveIntegrations(10)
metricsService.SetQueueSize(5)
```

### Проверка здоровья

```go
// Создание health service
healthService := services.NewHealthService(db, "1.0.0")

// Полная проверка
ctx := context.Background()
health := healthService.Check(ctx)
fmt.Printf("Status: %s\n", health.Status)

// Liveness проверка
if healthService.Liveness() {
    fmt.Println("Application is alive")
}

// Readiness проверка
if healthService.Readiness(ctx) {
    fmt.Println("Application is ready")
}

// Uptime
uptime := healthService.GetUptime()
fmt.Printf("Uptime: %s\n", uptime)
```

## См. также

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Kubernetes Health Checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
