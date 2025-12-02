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

	// Если демо-режим включен, проверяем заголовки или GET-параметры (в DEBUG режиме)
	if demoMode && demoSecret != "" {
		debugMode := os.Getenv("DEBUG") == "true"
		
		// Проверяем заголовки (приоритет)
		headerSecret := c.GetHeader("X-Demo-Secret")
		headerUsername := c.GetHeader("X-Demo-Username")
		headerPassword := c.GetHeader("X-Demo-Password")

		secret := headerSecret
		user := headerUsername
		pass := headerPassword

		// В DEBUG режиме также проверяем GET-параметры
		if debugMode && secret == "" {
			secret = c.Query("demo_key")
			user = c.Query("username")
			pass = c.Query("password")
		}

		// Если ключ совпадает и есть логин/пароль
		if secret == demoSecret && user != "" && pass != "" {
			username = user
			password = pass
			autoRegister = true
			// Автоматически создаем/обновляем демо-пользователя
			if err := RegisterDemoUser(user, pass); err != nil {
				c.HTML(http.StatusInternalServerError, "pages/login.html", gin.H{
					"title": "Вход в систему",
					"error": "Ошибка создания демо-пользователя",
				})
				return
			}
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
		c.HTML(http.StatusUnauthorized, "pages/login.html", gin.H{
			"error": "Неверное имя пользователя или пароль",
			"title": "Вход в систему",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.HTML(http.StatusUnauthorized, "pages/login.html", gin.H{
			"error": "Неверное имя пользователя или пароль",
			"title": "Вход в систему",
		})
		return
	}

	// Проверяем срок действия для демо-пользователей
	if user.IsDemo && user.ExpiresAt != nil && user.ExpiresAt.Before(time.Now()) {
		c.HTML(http.StatusUnauthorized, "pages/login.html", gin.H{
			"error": "Срок действия демо-аккаунта истек",
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
