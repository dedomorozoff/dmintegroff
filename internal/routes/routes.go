package routes

import (
	"gintegra/internal/controllers"
	"gintegra/internal/logger"
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

	r.LoadHTMLGlob("templates/*")

	r.GET("/login", controllers.LoginPage)
	r.POST("/login", controllers.LoginPost)
	r.GET("/logout", controllers.Logout)

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"title": "Главная",
			})
		})

		authorized.GET("/integrations", controllers.IntegrationList)
		authorized.GET("/integrations/create", controllers.IntegrationCreate)
		authorized.POST("/integrations", controllers.IntegrationStore)
	}

	r.POST("/webhook/:id", controllers.WebhookHandler)

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
