package services

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsService предоставляет метрики для мониторинга
type MetricsService struct {
	// Счетчики webhook
	webhookTotal       *prometheus.CounterVec
	webhookSuccess     *prometheus.CounterVec
	webhookFailure     *prometheus.CounterVec
	webhookRetries     *prometheus.CounterVec
	
	// Гистограммы времени выполнения
	webhookDuration    *prometheus.HistogramVec
	oauthDuration      *prometheus.HistogramVec
	
	// Gauges для текущего состояния
	activeIntegrations prometheus.Gauge
	queueSize          prometheus.Gauge
	
	// OAuth метрики
	oauthTokenRefresh  *prometheus.CounterVec
	oauthTokenCache    *prometheus.CounterVec
	
	// Retry метрики
	retryAttempts      *prometheus.HistogramVec
	
	// Signature метрики
	signatureGenerated *prometheus.CounterVec
	signatureVerified  *prometheus.CounterVec
	
	registry prometheus.Registerer
	mu       sync.RWMutex
}

// NewMetricsService создает новый сервис метрик
func NewMetricsService() *MetricsService {
	return NewMetricsServiceWithRegistry(prometheus.DefaultRegisterer)
}

// NewMetricsServiceWithRegistry создает новый сервис метрик с кастомным registry
func NewMetricsServiceWithRegistry(registry prometheus.Registerer) *MetricsService {
	factory := promauto.With(registry)
	
	return &MetricsService{
		registry: registry,
		// Webhook счетчики
		webhookTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_webhook_total",
				Help: "Total number of webhook requests",
			},
			[]string{"integration_id", "integration_name", "project_id"},
		),
		webhookSuccess: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_webhook_success_total",
				Help: "Total number of successful webhook requests",
			},
			[]string{"integration_id", "integration_name", "project_id"},
		),
		webhookFailure: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_webhook_failure_total",
				Help: "Total number of failed webhook requests",
			},
			[]string{"integration_id", "integration_name", "project_id", "error_type"},
		),
		webhookRetries: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_webhook_retries_total",
				Help: "Total number of webhook retry attempts",
			},
			[]string{"integration_id", "integration_name", "attempt"},
		),
		
		// Гистограммы времени
		webhookDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dmintegroff_webhook_duration_seconds",
				Help:    "Webhook request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"integration_id", "integration_name", "status"},
		),
		oauthDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dmintegroff_oauth_duration_seconds",
				Help:    "OAuth token request duration in seconds",
				Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"integration_id", "status"},
		),
		
		// Gauges
		activeIntegrations: factory.NewGauge(
			prometheus.GaugeOpts{
				Name: "dmintegroff_active_integrations",
				Help: "Number of active integrations",
			},
		),
		queueSize: factory.NewGauge(
			prometheus.GaugeOpts{
				Name: "dmintegroff_queue_size",
				Help: "Current webhook queue size",
			},
		),
		
		// OAuth метрики
		oauthTokenRefresh: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_oauth_token_refresh_total",
				Help: "Total number of OAuth token refreshes",
			},
			[]string{"integration_id", "status"},
		),
		oauthTokenCache: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_oauth_token_cache_total",
				Help: "OAuth token cache hits and misses",
			},
			[]string{"integration_id", "result"},
		),
		
		// Retry метрики
		retryAttempts: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dmintegroff_retry_attempts",
				Help:    "Number of retry attempts per request",
				Buckets: []float64{1, 2, 3, 4, 5},
			},
			[]string{"integration_id", "final_status"},
		),
		
		// Signature метрики
		signatureGenerated: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_signature_generated_total",
				Help: "Total number of webhook signatures generated",
			},
			[]string{"integration_id", "algorithm"},
		),
		signatureVerified: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "dmintegroff_signature_verified_total",
				Help: "Total number of webhook signatures verified",
			},
			[]string{"integration_id", "result"},
		),
	}
}

// RecordWebhookRequest записывает метрику webhook запроса
func (m *MetricsService) RecordWebhookRequest(integrationID, integrationName, projectID string) {
	m.webhookTotal.WithLabelValues(integrationID, integrationName, projectID).Inc()
}

// RecordWebhookSuccess записывает успешный webhook
func (m *MetricsService) RecordWebhookSuccess(integrationID, integrationName, projectID string, duration time.Duration) {
	m.webhookSuccess.WithLabelValues(integrationID, integrationName, projectID).Inc()
	m.webhookDuration.WithLabelValues(integrationID, integrationName, "success").Observe(duration.Seconds())
}

// RecordWebhookFailure записывает неудачный webhook
func (m *MetricsService) RecordWebhookFailure(integrationID, integrationName, projectID, errorType string, duration time.Duration) {
	m.webhookFailure.WithLabelValues(integrationID, integrationName, projectID, errorType).Inc()
	m.webhookDuration.WithLabelValues(integrationID, integrationName, "failure").Observe(duration.Seconds())
}

// RecordWebhookRetry записывает попытку retry
func (m *MetricsService) RecordWebhookRetry(integrationID, integrationName string, attempt int) {
	m.webhookRetries.WithLabelValues(integrationID, integrationName, string(rune(attempt))).Inc()
}

// RecordRetryAttempts записывает общее количество попыток
func (m *MetricsService) RecordRetryAttempts(integrationID string, attempts int, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	m.retryAttempts.WithLabelValues(integrationID, status).Observe(float64(attempts))
}

// RecordOAuthTokenRefresh записывает обновление OAuth токена
func (m *MetricsService) RecordOAuthTokenRefresh(integrationID string, success bool, duration time.Duration) {
	status := "failure"
	if success {
		status = "success"
	}
	m.oauthTokenRefresh.WithLabelValues(integrationID, status).Inc()
	m.oauthDuration.WithLabelValues(integrationID, status).Observe(duration.Seconds())
}

// RecordOAuthCacheHit записывает попадание в кэш OAuth токена
func (m *MetricsService) RecordOAuthCacheHit(integrationID string, hit bool) {
	result := "miss"
	if hit {
		result = "hit"
	}
	m.oauthTokenCache.WithLabelValues(integrationID, result).Inc()
}

// RecordSignatureGenerated записывает генерацию подписи
func (m *MetricsService) RecordSignatureGenerated(integrationID, algorithm string) {
	m.signatureGenerated.WithLabelValues(integrationID, algorithm).Inc()
}

// RecordSignatureVerified записывает проверку подписи
func (m *MetricsService) RecordSignatureVerified(integrationID string, valid bool) {
	result := "invalid"
	if valid {
		result = "valid"
	}
	m.signatureVerified.WithLabelValues(integrationID, result).Inc()
}

// SetActiveIntegrations устанавливает количество активных интеграций
func (m *MetricsService) SetActiveIntegrations(count int) {
	m.activeIntegrations.Set(float64(count))
}

// SetQueueSize устанавливает размер очереди
func (m *MetricsService) SetQueueSize(size int) {
	m.queueSize.Set(float64(size))
}

// GetMetrics возвращает текущие метрики (для тестирования)
func (m *MetricsService) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]interface{}{
		"webhooks": map[string]interface{}{
			"total":   "counter",
			"success": "counter",
			"failure": "counter",
			"retries": "counter",
		},
		"oauth": map[string]interface{}{
			"token_refresh": "counter",
			"cache":         "counter",
		},
		"signatures": map[string]interface{}{
			"generated": "counter",
			"verified":  "counter",
		},
	}
}
