package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// generateDemoSecret генерирует HMAC-подпись для проверки демо-доступа
// Должна совпадать с функцией на лендинге
func generateDemoSecret(username, password string, timestamp int64, sharedSecret string) string {
	// Формируем данные для подписи (тот же формат, что и на лендинге)
	data := fmt.Sprintf("%s:%s:%d", username, password, timestamp)
	
	// Создаем HMAC с SHA256
	h := hmac.New(sha256.New, []byte(sharedSecret))
	h.Write([]byte(data))
	
	// Возвращаем hex-представление подписи
	return hex.EncodeToString(h.Sum(nil))
}

func LoginPage(c *gin.Context) {
	// Очищаем старых демо-пользователей при каждом заходе на страницу логина
	CleanupExpiredDemoUsers()

	// Проверяем демо-режим
	demoMode := os.Getenv("DEMO_MODE") == "true"
	
	username := ""
	password := ""
	autoRegister := false

	// Если демо-режим включен, проверяем заголовки с HMAC-проверкой
	if demoMode {
		// Получаем общий секретный ключ для HMAC
		sharedSecret := os.Getenv("DEMO_SHARED_SECRET")
		
		// Проверяем заголовки (приоритет)
		headerSecret := c.GetHeader("X-Demo-Secret")
		headerUsername := c.GetHeader("X-Demo-Username")
		headerPassword := c.GetHeader("X-Demo-Password")
		headerTimestamp := c.GetHeader("X-Demo-Timestamp")

		secret := headerSecret
		user := headerUsername
		pass := headerPassword
		timestampStr := headerTimestamp

		if secret != "" && user != "" && pass != "" && timestampStr != "" && sharedSecret != "" {
			// HMAC-проверка для заголовков
			
			// Парсим timestamp
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				logger.Log.Warn("Invalid demo timestamp: %v", err)
			} else {
				// Проверяем, что timestamp не старше 5 минут (защита от replay-атак)
				currentTime := time.Now().Unix()
				timeDiff := math.Abs(float64(currentTime - timestamp))
				
				if timeDiff > 300 { // 5 минут = 300 секунд
					logger.Log.Warn("Demo timestamp too old: %d seconds", int(timeDiff))
				} else {
					// Генерируем ожидаемую HMAC-подпись
					expectedSecret := generateDemoSecret(user, pass, timestamp, sharedSecret)
					
					// Сравниваем подписи (constant-time comparison для защиты от timing-атак)
					if hmac.Equal([]byte(secret), []byte(expectedSecret)) {
						logger.Log.Info("Valid HMAC demo access for user: %s", user)
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
					} else {
						logger.Log.Warn("Invalid HMAC signature for demo access")
					}
				}
			}
		} else if demoMode && sharedSecret == "" {
			logger.Log.Error("DEMO_MODE is enabled but DEMO_SHARED_SECRET is not set!")
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

	c.Redirect(http.StatusFound, "/dashboard")
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
