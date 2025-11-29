package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// SettingsPage - страница настроек
func SettingsPage(c *gin.Context) {
	session := sessions.Default(c)
	role := session.Get("role")
	userID := session.Get("user_id")

	// Получаем информацию о пользователе
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"role":        role,
			"CurrentPage": "settings",
			"error":       "Не удалось загрузить информацию о пользователе",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/settings.html", gin.H{
		"title":       "Настройки",
		"role":        role,
		"CurrentPage": "settings",
		"username":    user.Username,
	})
}

// ChangePassword - смена пароля пользователя
func ChangePassword(c *gin.Context) {
	session := sessions.Default(c)
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
			"role":        role,
			"CurrentPage": "settings",
			"username":    "",
			"error":       "Пользователь не найден",
		})
		return
	}

	// Проверяем текущий пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"role":        role,
			"CurrentPage": "settings",
			"username":    user.Username,
			"error":       "Неверный текущий пароль",
		})
		return
	}

	// Проверяем совпадение нового пароля
	if newPassword != confirmPassword {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"role":        role,
			"CurrentPage": "settings",
			"username":    user.Username,
			"error":       "Новые пароли не совпадают",
		})
		return
	}

	// Проверяем минимальную длину пароля
	if len(newPassword) < 6 {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"role":        role,
			"CurrentPage": "settings",
			"username":    user.Username,
			"error":       "Пароль должен содержать минимум 6 символов",
		})
		return
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"role":        role,
			"CurrentPage": "settings",
			"username":    user.Username,
			"error":       "Ошибка при обработке пароля",
		})
		return
	}

	// Обновляем пароль в БД
	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"title":       "Настройки",
			"role":        role,
			"CurrentPage": "settings",
			"username":    user.Username,
			"error":       "Ошибка при сохранении пароля",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/settings.html", gin.H{
		"title":       "Настройки",
		"role":        role,
		"CurrentPage": "settings",
		"username":    user.Username,
		"success":     "Пароль успешно изменен",
	})
}
