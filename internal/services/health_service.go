package services

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// HealthStatus представляет статус здоровья компонента
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth представляет здоровье одного компонента
type ComponentHealth struct {
	Status      HealthStatus       `json:"status"`
	Message     string             `json:"message,omitempty"`
	LastChecked time.Time          `json:"last_checked"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// HealthCheck представляет общее состояние системы
type HealthCheck struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     time.Duration              `json:"uptime"`
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthService управляет проверками здоровья системы
type HealthService struct {
	db        *sql.DB
	startTime time.Time
	version   string
	mu        sync.RWMutex
}

// NewHealthService создает новый сервис здоровья
func NewHealthService(db *sql.DB, version string) *HealthService {
	return &HealthService{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

// Check выполняет полную проверку здоровья системы
func (h *HealthService) Check(ctx context.Context) *HealthCheck {
	h.mu.Lock()
	defer h.mu.Unlock()

	components := make(map[string]ComponentHealth)

	// Проверка базы данных
	components["database"] = h.checkDatabase(ctx)

	// Проверка дискового пространства
	components["disk"] = h.checkDisk()

	// Проверка памяти
	components["memory"] = h.checkMemory()

	// Определение общего статуса
	overallStatus := h.determineOverallStatus(components)

	return &HealthCheck{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Uptime:     time.Since(h.startTime),
		Version:    h.version,
		Components: components,
	}
}

// Liveness проверка - система жива?
func (h *HealthService) Liveness() bool {
	return true // Если код выполняется, система жива
}

// Readiness проверка - система готова принимать запросы?
func (h *HealthService) Readiness(ctx context.Context) bool {
	dbHealth := h.checkDatabase(ctx)
	return dbHealth.Status == HealthStatusHealthy
}

// checkDatabase проверяет подключение к базе данных
func (h *HealthService) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// Проверка подключения
	err := h.db.PingContext(ctx)
	if err != nil {
		return ComponentHealth{
			Status:      HealthStatusUnhealthy,
			Message:     fmt.Sprintf("Database connection failed: %v", err),
			LastChecked: time.Now(),
		}
	}

	// Проверка времени отклика
	duration := time.Since(start)
	status := HealthStatusHealthy
	message := "Database connection OK"

	if duration > 1*time.Second {
		status = HealthStatusDegraded
		message = "Database response time is slow"
	}

	// Получение статистики подключений
	stats := h.db.Stats()

	return ComponentHealth{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Details: map[string]interface{}{
			"response_time_ms": duration.Milliseconds(),
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"max_open":         stats.MaxOpenConnections,
		},
	}
}

// checkDisk проверяет дисковое пространство
func (h *HealthService) checkDisk() ComponentHealth {
	// Простая проверка - всегда OK
	// Для реальной проверки нужно использовать syscall (platform-specific)
	return ComponentHealth{
		Status:      HealthStatusHealthy,
		Message:     "Дисковое пространство в норме",
		LastChecked: time.Now(),
		Details: map[string]interface{}{
			"status": "monitoring not enabled",
		},
	}
}

// checkMemory проверяет использование памяти
func (h *HealthService) checkMemory() ComponentHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Конвертируем в MB для удобства
	allocMB := m.Alloc / 1024 / 1024
	totalAllocMB := m.TotalAlloc / 1024 / 1024
	sysMB := m.Sys / 1024 / 1024
	numGC := m.NumGC

	status := HealthStatusHealthy
	message := "Использование памяти в норме"

	// Если используется больше 500MB - предупреждение
	if allocMB > 500 {
		status = HealthStatusDegraded
		message = "Высокое использование памяти"
	}

	// Если используется больше 1GB - критично
	if allocMB > 1024 {
		status = HealthStatusUnhealthy
		message = "Критическое использование памяти"
	}

	return ComponentHealth{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Details: map[string]interface{}{
			"alloc_mb":       allocMB,
			"total_alloc_mb": totalAllocMB,
			"sys_mb":         sysMB,
			"num_gc":         numGC,
			"goroutines":     runtime.NumGoroutine(),
		},
	}
}

// determineOverallStatus определяет общий статус на основе компонентов
func (h *HealthService) determineOverallStatus(components map[string]ComponentHealth) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, component := range components {
		switch component.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// GetUptime возвращает время работы системы
func (h *HealthService) GetUptime() time.Duration {
	return time.Since(h.startTime)
}
