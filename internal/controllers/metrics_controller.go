package controllers

import (
	"dmintegroff/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsController обрабатывает запросы метрик и здоровья
type MetricsController struct {
	healthService *services.HealthService
}

// NewMetricsController создает новый контроллер метрик
func NewMetricsController(healthService *services.HealthService) *MetricsController {
	return &MetricsController{
		healthService: healthService,
	}
}

// RegisterRoutes регистрирует маршруты метрик
func (mc *MetricsController) RegisterRoutes(router *gin.Engine) {
	metrics := router.Group("/metrics")
	{
		// Prometheus метрики
		metrics.GET("", gin.WrapH(promhttp.Handler()))
		
		// Health checks
		metrics.GET("/health", mc.Health)
		metrics.GET("/health/live", mc.Liveness)
		metrics.GET("/health/ready", mc.Readiness)
		
		// Dashboard
		metrics.GET("/dashboard", mc.Dashboard)
	}
}

// Dashboard отображает веб-дашборд с метриками
// @Summary Веб-дашборд метрик
// @Description Отображает красивый веб-интерфейс с метриками и статусом системы
// @Tags Monitoring
// @Produce html
// @Success 200 {string} html "HTML страница дашборда"
// @Router /metrics/dashboard [get]
func (mc *MetricsController) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/metrics_dashboard.html", gin.H{
		"title": "Метрики и мониторинг",
	})
}

// Health возвращает полную информацию о здоровье системы
// @Summary Полная проверка здоровья
// @Description Возвращает детальную информацию о состоянии всех компонентов системы
// @Tags Monitoring
// @Produce json
// @Success 200 {object} services.HealthCheck "Система здорова"
// @Success 503 {object} services.HealthCheck "Система нездорова"
// @Router /metrics/health [get]
func (mc *MetricsController) Health(c *gin.Context) {
	ctx := c.Request.Context()
	health := mc.healthService.Check(ctx)
	
	// Определяем HTTP статус на основе здоровья
	statusCode := http.StatusOK
	if health.Status != services.HealthStatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, health)
}

// Liveness проверяет, жива ли система
// @Summary Liveness проверка
// @Description Проверяет, работает ли приложение (для Kubernetes liveness probe)
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{} "Система жива"
// @Router /metrics/health/live [get]
func (mc *MetricsController) Liveness(c *gin.Context) {
	if mc.healthService.Liveness() {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
			"uptime": mc.healthService.GetUptime().String(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "dead",
		})
	}
}

// Readiness проверяет, готова ли система принимать запросы
// @Summary Readiness проверка
// @Description Проверяет, готова ли система обрабатывать запросы (для Kubernetes readiness probe)
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{} "Система готова"
// @Success 503 {object} map[string]interface{} "Система не готова"
// @Router /metrics/health/ready [get]
func (mc *MetricsController) Readiness(c *gin.Context) {
	ctx := c.Request.Context()
	
	if mc.healthService.Readiness(ctx) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
		})
	}
}
