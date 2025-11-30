package utils

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// MergeTemplateData объединяет данные с информацией о пользователе из сессии
func MergeTemplateData(c *gin.Context, data gin.H) gin.H {
	session := sessions.Default(c)
	
	// Добавляем username и role если их еще нет
	if _, exists := data["username"]; !exists {
		if username := session.Get("username"); username != nil {
			data["username"] = username
		}
	}
	
	if _, exists := data["role"]; !exists {
		if role := session.Get("role"); role != nil {
			data["role"] = role
		}
	}
	
	return data
}
