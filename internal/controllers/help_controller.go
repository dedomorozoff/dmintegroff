package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelpPage отображает страницу помощи
func HelpPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/help.html", gin.H{
		"title":       "Помощь - dmIntegroff",
		"CurrentPage": "help",
	})
}
