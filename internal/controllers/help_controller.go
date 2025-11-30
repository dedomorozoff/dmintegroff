package controllers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// HelpPage отображает страницу помощи
func HelpPage(c *gin.Context) {
	session := sessions.Default(c)
	c.HTML(http.StatusOK, "pages/help.html", gin.H{
		"title":       "Помощь - dmIntegroff",
		"CurrentPage": "help",
		"username":    session.Get("username"),
		"role":        session.Get("role"),
	})
}
