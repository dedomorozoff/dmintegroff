package routes

import (
	"dmintegroff/internal/controllers"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Настройка доверенных прокси (только localhost для разработки)
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	r.Use(logger.RequestLogger())
	r.Use(ErrorLogger())
	r.Use(PanicRecovery())

	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "secret"
	}
	store := cookie.NewStore([]byte(secret))
	r.Use(sessions.Sessions("mysession", store))

	// Serve static files
	r.Static("/static", "./static")

	r.LoadHTMLGlob("templates/*/*")

	// Get custom app path from env
	appPath := os.Getenv("APP_PATH")
	if appPath == "" {
		appPath = ""
	}

	r.GET(appPath+"/login", controllers.LoginPage)
	r.POST(appPath+"/login", controllers.LoginPost)
	r.GET(appPath+"/logout", controllers.Logout)

	// Protected routes
	authorized := r.Group(appPath + "/")
	authorized.Use(AuthRequired())
	authorized.Use(InjectUserData())
	{
		// Редирект с корня на dashboard
		authorized.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/dashboard")
		})
		
		authorized.GET("/dashboard", controllers.DashboardPage)
		authorized.GET("/api/activity", controllers.GetRecentActivity)
		authorized.GET("/api/stats", controllers.GetRequestStats)

		authorized.GET("/integrations", controllers.IntegrationList)
		authorized.GET("/api/integrations", controllers.IntegrationsListAPI)
		authorized.GET("/integrations/create", controllers.IntegrationCreate)
		authorized.POST("/integrations", controllers.IntegrationStore)
		authorized.GET("/integrations/:id/edit", controllers.IntegrationEdit)
		authorized.POST("/integrations/:id/update", controllers.IntegrationUpdate)
		authorized.POST("/integrations/:id/regenerate-token", controllers.IntegrationRegenerateToken)
		authorized.POST("/integrations/:id/delete", controllers.IntegrationDelete)
		authorized.POST("/integrations/:id/toggle", controllers.IntegrationToggle)
		authorized.POST("/integrations/:id/reconfigure", controllers.IntegrationReconfigure)
		authorized.POST("/integrations/:id/cancel-listening", controllers.IntegrationCancelListening)
		authorized.GET("/integrations/:id/configure", controllers.IntegrationConfigure)
		authorized.POST("/integrations/:id/configure", controllers.IntegrationSaveMapping)
		authorized.GET("/api/integrations/:id/check", controllers.IntegrationCheckUpdate)
		authorized.POST("/api/integrations/:id/test-oauth", controllers.IntegrationTestOAuth)

		// Projects
		authorized.GET("/projects", controllers.ProjectList)
		authorized.GET("/projects/create", controllers.ProjectCreate)
		authorized.POST("/projects", controllers.ProjectStore)
		authorized.GET("/projects/:id", controllers.ProjectView)
		authorized.GET("/projects/:id/edit", controllers.ProjectEdit)
		authorized.POST("/projects/:id/update", controllers.ProjectUpdate)
		authorized.POST("/projects/:id/delete", controllers.ProjectDelete)

		// Project integrations
		authorized.GET("/projects/:id/integrations/create", controllers.ProjectCreateIntegration)
		authorized.POST("/projects/:id/integrations", controllers.ProjectStoreIntegration)
		authorized.POST("/projects/:id/integrations/:integration_id/delete", controllers.ProjectDeleteIntegration)
		
		// Export/Import
		authorized.GET("/export/integrations", controllers.ExportIntegrationsHandler)
		authorized.GET("/export/projects/:id", controllers.ExportProjectIntegrationsHandler)
		authorized.POST("/import/integrations", controllers.ImportIntegrationsHandler)
		authorized.POST("/import/validate", controllers.ValidateImportHandler)
		
		// Git Export
		authorized.POST("/export/git/integrations", controllers.ExportToGitHandler)
		authorized.POST("/export/git/projects/:id", controllers.ExportProjectToGitHandler)

		// Logs
		authorized.GET("/logs", controllers.LogsPage)
		authorized.GET("/api/logs", controllers.LogsAPI)
		authorized.POST("/logs/clear", controllers.ClearLogs)
		authorized.POST("/logs/:id/delete", controllers.DeleteLog)

		// Help
		authorized.GET("/help", controllers.HelpPage)

		// Settings
		authorized.GET("/settings", controllers.SettingsPage)
		authorized.POST("/settings/change-password", controllers.ChangePassword)
	}

	// Public endpoints - принимаем все HTTP методы для webhook
	r.Any("/webhook/:token", controllers.WebhookHandler)
	r.Any("/webhook/test", controllers.TestEndpoint) // Test webhook endpoint

	// Error pages - должны быть в конце
	r.NoRoute(controllers.NotFoundPage)

	return r
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user_id")
		if user == nil {
			// Для AJAX/API запросов показываем страницу 401
			if c.GetHeader("X-Requested-With") == "XMLHttpRequest" || c.GetHeader("Accept") == "application/json" {
				c.HTML(http.StatusUnauthorized, "pages/401.html", gin.H{
					"title": "Доступ запрещён",
				})
			} else {
				// Для обычных запросов редиректим на логин
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}
		
		// Добавляем username и role в контекст для использования в шаблонах
		username := session.Get("username")
		role := session.Get("role")
		c.Set("username", username)
		c.Set("role", role)
		
		c.Next()
	}
}

// InjectUserData - middleware для автоматического добавления данных пользователя в шаблоны
func InjectUserData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем данные из контекста
		username, _ := c.Get("username")
		role, _ := c.Get("role")
		
		// Сохраняем в контексте для доступа в контроллерах
		c.Set("template_username", username)
		c.Set("template_role", role)
		
		c.Next()
	}
}

// ErrorLogger - middleware для логирования ошибок
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверяем статус код после выполнения запроса
		statusCode := c.Writer.Status()

		// Логируем только ошибки (4xx и 5xx)
		if statusCode >= 400 {
			errorMsg := ""

			// Пытаемся получить сообщение об ошибке из контекста
			if len(c.Errors) > 0 {
				errorMsg = c.Errors.String()
			} else {
				errorMsg = http.StatusText(statusCode)
			}

			// Логируем в файл через logrus
			logger.Log.WithFields(map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"status":     statusCode,
				"error":      errorMsg,
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			}).Error("HTTP Error")

			// Сохраняем в БД только для /webhook/test
			if c.Request.URL.Path == "/webhook/test" {
				log := models.RequestLog{
					Method:       c.Request.Method,
					URL:          c.Request.URL.Path,
					ErrorMessage: errorMsg,
					StatusCode:   statusCode,
					LogType:      "error",
				}
				controllers.CreateLogWithLimit(&log)
			}
		}
	}
}

// PanicRecovery - middleware для перехвата паник
func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Логируем панику в файл
				logger.Log.WithFields(map[string]interface{}{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"ip":     c.ClientIP(),
					"panic":  err,
				}).Error("PANIC recovered")

				// Сохраняем в БД только для /webhook/test
				if c.Request.URL.Path == "/webhook/test" {
					panicMsg := ""
					if str, ok := err.(string); ok {
						panicMsg = "PANIC: " + str
					} else {
						panicMsg = "PANIC: unknown error"
					}
					log := models.RequestLog{
						Method:       c.Request.Method,
						URL:          c.Request.URL.Path,
						ErrorMessage: panicMsg,
						StatusCode:   500,
						LogType:      "error",
					}
					controllers.CreateLogWithLimit(&log)
				}

				// Показываем красивую страницу 500 для HTML запросов
				if c.GetHeader("Accept") == "" || c.GetHeader("Accept") == "text/html" || c.Request.Header.Get("Accept") == "*/*" {
					controllers.InternalErrorPage(c)
				} else {
					// Для API запросов возвращаем JSON
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Internal Server Error",
					})
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
