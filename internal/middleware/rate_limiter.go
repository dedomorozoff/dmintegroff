package middleware

import (
	"dmintegroff/internal/logger"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter представляет rate limiter для IP адресов
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // запросов в минуту
	burst    int           // максимальное количество запросов в burst
	cleanup  time.Duration // интервал очистки старых записей
}

// Visitor представляет посетителя с его лимитами
type Visitor struct {
	limiter  *TokenBucket
	lastSeen time.Time
}

// TokenBucket реализует алгоритм token bucket
type TokenBucket struct {
	tokens    int
	capacity  int
	refillRate int // токенов в минуту
	lastRefill time.Time
	mu        sync.Mutex
}

// NewTokenBucket создает новый token bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow проверяет, можно ли выполнить запрос
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Добавляем токены на основе прошедшего времени
	tokensToAdd := int(elapsed.Minutes() * float64(tb.refillRate))
	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	// Проверяем, есть ли доступные токены
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		burst:    burst,
		cleanup:  time.Minute * 10, // очистка каждые 10 минут
	}

	// Запускаем горутину для очистки старых записей
	go rl.cleanupVisitors()

	return rl
}

// getVisitor получает или создает visitor для IP
func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			limiter:  NewTokenBucket(rl.burst, rl.rate),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
	} else {
		visitor.lastSeen = time.Now()
	}

	return visitor
}

// Allow проверяет, разрешен ли запрос для данного IP
func (rl *RateLimiter) Allow(ip string) bool {
	visitor := rl.getVisitor(ip)
	return visitor.limiter.Allow()
}

// cleanupVisitors удаляет старые записи посетителей
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-time.Hour) // удаляем записи старше часа
			
			for ip, visitor := range rl.visitors {
				if visitor.lastSeen.Before(cutoff) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Глобальный rate limiter для webhook endpoints
var webhookRateLimiter *RateLimiter

// InitRateLimiter инициализирует глобальный rate limiter
func InitRateLimiter(rate, burst int) {
	webhookRateLimiter = NewRateLimiter(rate, burst)
	logger.Log.WithFields(map[string]interface{}{
		"rate":  rate,
		"burst": burst,
	}).Info("Rate limiter initialized")
}

// WebhookRateLimit middleware для ограничения частоты запросов к webhook endpoints
func WebhookRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if webhookRateLimiter == nil {
			// Если rate limiter не инициализирован, пропускаем
			c.Next()
			return
		}

		// Получаем IP адрес клиента
		clientIP := c.ClientIP()

		// Проверяем лимит
		if !webhookRateLimiter.Allow(clientIP) {
			logger.LogRateLimitExceeded(clientIP, c.Request.URL.Path, c.Request.Method)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetRateLimiterStats возвращает статистику rate limiter
func GetRateLimiterStats() map[string]interface{} {
	if webhookRateLimiter == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	webhookRateLimiter.mu.RLock()
	defer webhookRateLimiter.mu.RUnlock()

	return map[string]interface{}{
		"enabled":        true,
		"rate":           webhookRateLimiter.rate,
		"burst":          webhookRateLimiter.burst,
		"active_visitors": len(webhookRateLimiter.visitors),
	}
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}