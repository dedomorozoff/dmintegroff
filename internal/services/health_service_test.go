package services

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewHealthService(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	hs := NewHealthService(db, "1.0.0")
	if hs == nil {
		t.Fatal("Expected health service, got nil")
	}

	if hs.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", hs.version)
	}
}

func TestHealthCheck_Healthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	// Ожидаем успешный ping
	mock.ExpectPing()

	hs := NewHealthService(db, "1.0.0")
	ctx := context.Background()

	health := hs.Check(ctx)

	if health.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", health.Status)
	}

	if health.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", health.Version)
	}

	// Проверяем компоненты
	if dbHealth, ok := health.Components["database"]; ok {
		if dbHealth.Status != HealthStatusHealthy {
			t.Errorf("Expected database healthy, got %s", dbHealth.Status)
		}
	} else {
		t.Error("Expected database component in health check")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestHealthCheck_UnhealthyDatabase(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	// Ожидаем неудачный ping
	mock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)

	hs := NewHealthService(db, "1.0.0")
	ctx := context.Background()

	health := hs.Check(ctx)

	if health.Status != HealthStatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", health.Status)
	}

	// Проверяем компонент базы данных
	if dbHealth, ok := health.Components["database"]; ok {
		if dbHealth.Status != HealthStatusUnhealthy {
			t.Errorf("Expected database unhealthy, got %s", dbHealth.Status)
		}
	} else {
		t.Error("Expected database component in health check")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestLiveness(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	hs := NewHealthService(db, "1.0.0")

	if !hs.Liveness() {
		t.Error("Expected liveness to be true")
	}
}

func TestReadiness_Ready(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	mock.ExpectPing()

	hs := NewHealthService(db, "1.0.0")
	ctx := context.Background()

	if !hs.Readiness(ctx) {
		t.Error("Expected readiness to be true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestReadiness_NotReady(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	mock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)

	hs := NewHealthService(db, "1.0.0")
	ctx := context.Background()

	if hs.Readiness(ctx) {
		t.Error("Expected readiness to be false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetUptime(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	hs := NewHealthService(db, "1.0.0")

	// Ждем немного
	time.Sleep(10 * time.Millisecond)

	uptime := hs.GetUptime()
	if uptime < 10*time.Millisecond {
		t.Errorf("Expected uptime >= 10ms, got %v", uptime)
	}
}

func TestDetermineOverallStatus(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	hs := NewHealthService(db, "1.0.0")

	tests := []struct {
		name       string
		components map[string]ComponentHealth
		expected   HealthStatus
	}{
		{
			name: "all healthy",
			components: map[string]ComponentHealth{
				"db":   {Status: HealthStatusHealthy},
				"disk": {Status: HealthStatusHealthy},
			},
			expected: HealthStatusHealthy,
		},
		{
			name: "one degraded",
			components: map[string]ComponentHealth{
				"db":   {Status: HealthStatusHealthy},
				"disk": {Status: HealthStatusDegraded},
			},
			expected: HealthStatusDegraded,
		},
		{
			name: "one unhealthy",
			components: map[string]ComponentHealth{
				"db":   {Status: HealthStatusHealthy},
				"disk": {Status: HealthStatusUnhealthy},
			},
			expected: HealthStatusUnhealthy,
		},
		{
			name: "degraded and unhealthy",
			components: map[string]ComponentHealth{
				"db":     {Status: HealthStatusDegraded},
				"disk":   {Status: HealthStatusUnhealthy},
				"memory": {Status: HealthStatusHealthy},
			},
			expected: HealthStatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := hs.determineOverallStatus(tt.components)
			if status != tt.expected {
				t.Errorf("Expected status %s, got %s", tt.expected, status)
			}
		})
	}
}

func TestCheckDatabase_Details(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	defer db.Close()

	mock.ExpectPing()

	hs := NewHealthService(db, "1.0.0")
	ctx := context.Background()

	health := hs.checkDatabase(ctx)

	if health.Details == nil {
		t.Fatal("Expected details in database health")
	}

	// Проверяем наличие ключевых метрик
	if _, ok := health.Details["response_time_ms"]; !ok {
		t.Error("Expected response_time_ms in details")
	}
	if _, ok := health.Details["open_connections"]; !ok {
		t.Error("Expected open_connections in details")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
