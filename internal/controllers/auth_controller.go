package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func LoginPage(c *gin.Context) {
	// Очищаем старых демо-пользователей при каждом заходе на страницу логина
	CleanupExpiredDemoUsers()

	// Проверяем демо-режим
	demoMode := os.Getenv("DEMO_MODE") == "true"
	demoSecret := os.Getenv("DEMO_SECRET")
	
	username := ""
	password := ""
	autoRegister := false

	// Если демо-режим включен, проверяем GET-параметры
	if demoMode && demoSecret != "" {
		querySecret := c.Query("demo_key")
		queryUsername := c.Query("username")
		queryPassword := c.Query("password")

		// Если ключ совпадает и есть логин/пароль
		if querySecret == demoSecret && queryUsername != "" && queryPassword != "" {
			username = queryUsername
			password = queryPassword
			autoRegister = true

			// Автоматически регистрируем демо-пользователя
			RegisterDemoUser(queryUsername, queryPassword)
		}
	}

	c.HTML(http.StatusOK, "pages/login.html", gin.H{
		"title":        "Вход в систему",
		"username":     username,
		"password":     password,
		"autoRegister": autoRegister,
	})
}

// RegisterDemoUser - автоматическая регистрация демо-пользователя
func RegisterDemoUser(username, password string) error {
	// Проверяем, существует ли уже такой пользователь
	var existingUser models.User
	if err := database.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		// Пользователь существует, обновляем срок действия
		expiresAt := time.Now().Add(24 * time.Hour)
		existingUser.ExpiresAt = &expiresAt
		existingUser.IsDemo = true
		database.DB.Save(&existingUser)
		return nil
	}

	// Создаем нового демо-пользователя
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	user := models.User{
		Username:  username,
		Password:  string(hashedPassword),
		Role:      "specialist",
		IsDemo:    true,
		ExpiresAt: &expiresAt,
	}

	return database.DB.Create(&user).Error
}

// CleanupExpiredDemoUsers - удаление истекших демо-пользователей
func CleanupExpiredDemoUsers() {
	now := time.Now()
	database.DB.Where("is_demo = ? AND expires_at < ?", true, now).Delete(&models.User{})
}

func LoginPost(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "Неверное имя пользователя или пароль",
			"title": "Вход в систему",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "Неверное имя пользователя или пароль",
			"title": "Вход в систему",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
