package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFoundPage renders the 404 error page
func NotFoundPage(c *gin.Context) {
	c.HTML(http.StatusNotFound, "pages/404.html", gin.H{
		"title": "404 - Страница не найдена",
	})
}

// InternalErrorPage renders the 500 error page
func InternalErrorPage(c *gin.Context) {
	c.HTML(http.StatusInternalServerError, "pages/500.html", gin.H{
		"title": "500 - Внутренняя ошибка сервера",
	})
}
