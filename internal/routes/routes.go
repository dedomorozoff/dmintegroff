package routes

import (
	"gintegra/internal/controllers"
	"gintegra/internal/database"
	"gintegra/internal/logger"
	"gintegra/internal/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(logger.RequestLogger())

	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "secret"
	}
	store := cookie.NewStore([]byte(secret))
	r.Use(sessions.Sessions("mysession", store))

	// Serve static files
	r.Static("/static", "./static")

	r.LoadHTMLGlob("templates/*")

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
	{
		authorized.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			role := session.Get("role")
			
			var totalIntegrations int64
			var activeIntegrations int64
			database.DB.Model(&models.Integration{}).Count(&totalIntegrations)
			database.DB.Model(&models.Integration{}).Where("status = ?", "active").Count(&activeIntegrations)
			
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"title":               "Главная",
				"role":                role,
				"total_integrations":  totalIntegrations,
				"active_integrations": activeIntegrations,
			})
		})

		authorized.GET("/integrations", controllers.IntegrationList)
		authorized.GET("/integrations/create", controllers.IntegrationCreate)
		authorized.POST("/integrations", controllers.IntegrationStore)
		authorized.GET("/integrations/:id/edit", controllers.IntegrationEdit)
		authorized.POST("/integrations/:id/update", controllers.IntegrationUpdate)
		authorized.POST("/integrations/:id/delete", controllers.IntegrationDelete)
		authorized.GET("/integrations/:id/configure", controllers.IntegrationConfigure)
		authorized.POST("/integrations/:id/configure", controllers.IntegrationSaveMapping)
		
		// Logs
		authorized.GET("/logs", controllers.LogsPage)
		authorized.POST("/logs/clear", controllers.ClearLogs)
		authorized.POST("/logs/:id/delete", controllers.DeleteLog)
	}

	// Public endpoints
	r.POST("/webhook/:token", controllers.WebhookHandler)
	r.POST("/test", controllers.TestEndpoint) // Test endpoint for debugging

	return r
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user_id")
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}
