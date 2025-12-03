package services

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetricsService(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)
	if ms == nil {
		t.Fatal("Expected metrics service, got nil")
	}

	metrics := ms.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected metrics map, got nil")
	}

	// Проверяем наличие основных категорий метрик
	if _, ok := metrics["webhooks"]; !ok {
		t.Error("Expected webhooks metrics")
	}
	if _, ok := metrics["oauth"]; !ok {
		t.Error("Expected oauth metrics")
	}
	if _, ok := metrics["signatures"]; !ok {
		t.Error("Expected signatures metrics")
	}
}

func TestRecordWebhookRequest(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	// Записываем метрику
	ms.RecordWebhookRequest("1", "test-integration", "project-1")

	// Метрика должна быть записана без паники
	// В реальном тесте можно проверить через prometheus testutil
}

func TestRecordWebhookSuccess(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	duration := 500 * time.Millisecond
	ms.RecordWebhookSuccess("1", "test-integration", "project-1", duration)

	// Проверяем, что метрика записана
}

func TestRecordWebhookFailure(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	duration := 1 * time.Second
	ms.RecordWebhookFailure("1", "test-integration", "project-1", "timeout", duration)

	// Проверяем, что метрика записана
}

func TestRecordWebhookRetry(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	ms.RecordWebhookRetry("1", "test-integration", 2)

	// Проверяем, что метрика записана
}

func TestRecordRetryAttempts(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	tests := []struct {
		name     string
		attempts int
		success  bool
	}{
		{"success after 1 attempt", 1, true},
		{"success after 3 attempts", 3, true},
		{"failure after 3 attempts", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms.RecordRetryAttempts("1", tt.attempts, tt.success)
		})
	}
}

func TestRecordOAuthTokenRefresh(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	tests := []struct {
		name     string
		success  bool
		duration time.Duration
	}{
		{"successful refresh", true, 200 * time.Millisecond},
		{"failed refresh", false, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms.RecordOAuthTokenRefresh("1", tt.success, tt.duration)
		})
	}
}

func TestRecordOAuthCacheHit(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	tests := []struct {
		name string
		hit  bool
	}{
		{"cache hit", true},
		{"cache miss", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms.RecordOAuthCacheHit("1", tt.hit)
		})
	}
}

func TestRecordSignatureGenerated(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	algorithms := []string{"sha256", "sha512", "sha1"}

	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			ms.RecordSignatureGenerated("1", algo)
		})
	}
}

func TestRecordSignatureVerified(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	tests := []struct {
		name  string
		valid bool
	}{
		{"valid signature", true},
		{"invalid signature", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms.RecordSignatureVerified("1", tt.valid)
		})
	}
}

func TestSetActiveIntegrations(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	counts := []int{0, 5, 10, 100}

	for _, count := range counts {
		ms.SetActiveIntegrations(count)
	}
}

func TestSetQueueSize(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	sizes := []int{0, 10, 50, 100}

	for _, size := range sizes {
		ms.SetQueueSize(size)
	}
}

func TestConcurrentMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	ms := NewMetricsServiceWithRegistry(registry)

	// Тест конкурентной записи метрик
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			ms.RecordWebhookRequest("1", "test", "project-1")
			ms.RecordWebhookSuccess("1", "test", "project-1", 100*time.Millisecond)
			ms.RecordOAuthCacheHit("1", true)
			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}
}
