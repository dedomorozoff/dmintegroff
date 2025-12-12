package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SettingsPage - страница настроек
func SettingsPage(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	userID := session.Get("user_id")

	// Получаем информацию о пользователе
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"username":    username,
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Не удалось загрузить информацию о пользователе",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/settings.html", gin.H{
		"title":       "Настройки",
		"username":    username,
		"role":        role,
		"CurrentPage": "settings",
	})
}

// ChangePassword - смена пароля пользователя
func ChangePassword(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	userID := session.Get("user_id")

	// Получаем данные из формы
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	// Получаем пользователя из БД
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"username":    username,
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Пользователь не найден",
		})
		return
	}

	// Проверяем текущий пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"username":    username,
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Неверный текущий пароль",
		})
		return
	}

	// Проверяем совпадение нового пароля
	if newPassword != confirmPassword {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"username":    username,
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Новые пароли не совпадают",
		})
		return
	}

	// Проверяем минимальную длину пароля
	if len(newPassword) < 6 {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"username":    username,
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Пароль должен содержать минимум 6 символов",
		})
		return
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"username":    username,
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Ошибка при обработке пароля",
		})
		return
	}

	// Обновляем пароль в БД
	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"username":    username,
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Ошибка при сохранении пароля",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/settings.html", gin.H{
		"title":       "Настройки",
		"username":    username,
		"role":        role,
		"CurrentPage": "settings",
		"success":     "Пароль успешно изменен",
	})
}

// SaveAISettings сохраняет настройки AI
func SaveAISettings(c *gin.Context) {
	session := sessions.Default(c)
	role := session.Get("role")

	// Проверяем права администратора
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Недостаточно прав для изменения настроек AI",
		})
		return
	}

	// Получаем данные из формы
	var req struct {
		OpenRouterAPIKey *string `json:"openrouter_api_key"` // указатель для обработки null
		OpenRouterModel  string  `json:"openrouter_model"`
		OpenAIAPIKey     *string `json:"openai_api_key"`     // указатель для обработки null
		OpenAIModel      string  `json:"openai_model"`
		MaxTokens        int     `json:"max_tokens"`
		Temperature      float32 `json:"temperature"`
		RequestTimeout   int     `json:"request_timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
			"details": err.Error(),
		})
		return
	}

	// Валидация
	if req.MaxTokens < 100 || req.MaxTokens > 8000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Максимальное количество токенов должно быть от 100 до 8000",
		})
		return
	}

	if req.Temperature < 0 || req.Temperature > 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Температура должна быть от 0 до 2",
		})
		return
	}

	if req.RequestTimeout < 5 || req.RequestTimeout > 120 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Таймаут должен быть от 5 до 120 секунд",
		})
		return
	}

	// Получаем существующие настройки для сохранения API ключей
	existingSettings, err := models.GetActiveAISettings(database.DB)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка получения существующих настроек",
			"details": err.Error(),
		})
		return
	}

	// Создаем объект настроек
	aiSettings := &models.AISettings{
		OpenRouterModel:  req.OpenRouterModel,
		OpenRouterURL:    "https://openrouter.ai/api/v1", // фиксированный URL
		OpenAIModel:      req.OpenAIModel,
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		RequestTimeout:   req.RequestTimeout,
		LocalLLMEnabled:  false, // пока не поддерживается в UI
		LocalLLMURL:      "http://localhost:11434",
	}

	// Обрабатываем API ключи - если null, сохраняем существующие
	if req.OpenRouterAPIKey != nil {
		aiSettings.OpenRouterAPIKey = *req.OpenRouterAPIKey
	} else if existingSettings != nil {
		aiSettings.OpenRouterAPIKey = existingSettings.OpenRouterAPIKey
	}

	if req.OpenAIAPIKey != nil {
		aiSettings.OpenAIAPIKey = *req.OpenAIAPIKey
	} else if existingSettings != nil {
		aiSettings.OpenAIAPIKey = existingSettings.OpenAIAPIKey
	}

	// Сохраняем в базу данных
	if err := models.SaveAISettings(database.DB, aiSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения настроек в базу данных",
			"details": err.Error(),
		})
		return
	}

	// Перезагружаем AI контроллер с новыми настройками
	reloadAIController()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Настройки AI успешно сохранены и применены!",
	})
}

// GetAISettings получает текущие настройки AI
func GetAISettings(c *gin.Context) {
	session := sessions.Default(c)
	role := session.Get("role")

	// Проверяем права администратора
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Недостаточно прав для просмотра настроек AI",
		})
		return
	}

	// Получаем настройки из базы данных
	settings, err := models.GetActiveAISettings(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка получения настроек AI",
			"details": err.Error(),
		})
		return
	}

	// Маскируем API ключи для безопасности
	response := gin.H{
		"openrouter_api_key": maskAPIKey(settings.OpenRouterAPIKey),
		"openrouter_model":   settings.OpenRouterModel,
		"openai_api_key":     maskAPIKey(settings.OpenAIAPIKey),
		"openai_model":       settings.OpenAIModel,
		"max_tokens":         settings.MaxTokens,
		"temperature":        settings.Temperature,
		"request_timeout":    settings.RequestTimeout,
		"configured":         settings.IsConfigured(),
		"provider":           settings.GetCurrentProvider(),
	}

	c.JSON(http.StatusOK, response)
}

// maskAPIKey маскирует API ключ для безопасности
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
