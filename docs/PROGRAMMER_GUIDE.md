# 👨‍💻 Руководство программиста GIntegra

Этот документ для разработчиков, которые хотят развивать или модифицировать GIntegra.

## Настройка окружения разработки

### Требования
- **Go**: 1.20 или новее
- **Git**: Для управления версиями
- **IDE**: VS Code (с Go extension) или GoLand

### Установка

```bash
# Клонирование репозитория
git clone https://github.com/yourusername/gintegra.git
cd gintegra

# Установка зависимостей
go mod download
go mod tidy

# Создание .env файла
cp .env.example .env

# Запуск в режиме разработки
go run cmd/server/main.go
```

### Полезные команды

```bash
# Форматирование кода
go fmt ./...

# Проверка кода
go vet ./...

# Запуск тестов (когда будут добавлены)
go test ./...

# Обновление зависимостей
go get -u ./...
go mod tidy
```

## Структура проекта

### Архитектурные слои

```
┌─────────────────────────────────────┐
│         HTTP Layer (Gin)            │
│  ┌──────────────────────────────┐   │
│  │   Routes & Middleware        │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│      Controller Layer               │
│  ┌──────────────────────────────┐   │
│  │  HTTP Request Handlers       │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Service Layer                 │
│  ┌──────────────────────────────┐   │
│  │   Business Logic             │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Model Layer (GORM)            │
│  ┌──────────────────────────────┐   │
│  │   Database Models            │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

### Соглашения по коду

#### Именование
- **Файлы**: `snake_case.go` (например, `auth_controller.go`)
- **Функции**: `PascalCase` для экспортируемых, `camelCase` для внутренних
- **Переменные**: `camelCase`
- **Константы**: `UPPER_SNAKE_CASE`

#### Комментарии
```go
// LoginPost обрабатывает POST запрос на авторизацию.
// Проверяет учетные данные и создает сессию.
func LoginPost(c *gin.Context) {
    // Получение данных из формы
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    // ... остальной код
}
```

## Добавление нового функционала

### 1. Добавление новой модели

**Шаг 1**: Создайте файл в `internal/models/`

```go
// internal/models/notification.go
package models

import "gorm.io/gorm"

type Notification struct {
    gorm.Model
    UserID  uint   `gorm:"not null" json:"user_id"`
    Message string `gorm:"type:text;not null" json:"message"`
    IsRead  bool   `gorm:"default:false" json:"is_read"`
    
    // Связи
    User User `gorm:"foreignKey:UserID" json:"user"`
}
```

**Шаг 2**: Добавьте миграцию в `cmd/server/main.go`

```go
database.Migrate(
    &models.User{}, 
    &models.Integration{}, 
    &models.RequestLog{},
    &models.Notification{}, // Новая модель
)
```

### 2. Добавление нового контроллера

**Шаг 1**: Создайте файл в `internal/controllers/`

```go
// internal/controllers/notification_controller.go
package controllers

import (
    "gintegra/internal/database"
    "gintegra/internal/models"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

// NotificationList отображает список уведомлений
func NotificationList(c *gin.Context) {
    var notifications []models.Notification
    database.DB.Preload("User").Find(&notifications)
    
    c.HTML(http.StatusOK, "notifications.html", gin.H{
        "title":         "Уведомления",
        "notifications": notifications,
    })
}

// MarkAsRead помечает уведомление как прочитанное
func MarkAsRead(c *gin.Context) {
    id := c.Param("id")
    
    var notification models.Notification
    if err := database.DB.First(&notification, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
        return
    }
    
    notification.IsRead = true
    database.DB.Save(&notification)
    
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

**Шаг 2**: Зарегистрируйте маршруты в `internal/routes/routes.go`

```go
// В функции SetupRouter, внутри authorized группы
authorized.GET("/notifications", controllers.NotificationList)
authorized.POST("/notifications/:id/read", controllers.MarkAsRead)
```

### 3. Добавление нового сервиса

**Создайте файл в `internal/services/`**

```go
// internal/services/notification_service.go
package services

import (
    "gintegra/internal/database"
    "gintegra/internal/models"
)

// NotificationService управляет уведомлениями
type NotificationService struct{}

// SendNotification создает новое уведомление
func (s *NotificationService) SendNotification(userID uint, message string) error {
    notification := models.Notification{
        UserID:  userID,
        Message: message,
        IsRead:  false,
    }
    
    return database.DB.Create(&notification).Error
}

// GetUnreadCount возвращает количество непрочитанных уведомлений
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
    var count int64
    err := database.DB.Model(&models.Notification{}).
        Where("user_id = ? AND is_read = ?", userID, false).
        Count(&count).Error
    
    return count, err
}
```

### 4. Добавление нового шаблона

**Создайте файл в `templates/`**

```html
<!-- templates/notifications.html -->
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .title }}</title>
    <link rel="stylesheet" href="/static/css/modern.css">
</head>
<body>
    <!-- Sidebar (скопируйте из других шаблонов) -->
    <aside class="sidebar">
        <!-- ... -->
    </aside>

    <!-- Main Content -->
    <main class="main-content">
        <div class="page-header">
            <h1>Уведомления</h1>
        </div>

        <div class="card">
            {{ range .notifications }}
            <div class="notification {{ if .IsRead }}read{{ end }}">
                <p>{{ .Message }}</p>
                <small>{{ .CreatedAt.Format "2006-01-02 15:04:05" }}</small>
            </div>
            {{ end }}
        </div>
    </main>
</body>
</html>
```

## Работа с базой данных

### GORM основы

```go
// Создание записи
user := models.User{Username: "john", Password: "hash", Role: "specialist"}
database.DB.Create(&user)

// Поиск одной записи
var user models.User
database.DB.First(&user, 1) // По ID
database.DB.Where("username = ?", "john").First(&user)

// Поиск нескольких записей
var users []models.User
database.DB.Find(&users)
database.DB.Where("role = ?", "admin").Find(&users)

// Обновление
database.DB.Model(&user).Update("role", "admin")
database.DB.Model(&user).Updates(models.User{Role: "admin", Username: "john_admin"})

// Удаление
database.DB.Delete(&user, 1) // Soft delete (если есть DeletedAt)
database.DB.Unscoped().Delete(&user, 1) // Permanent delete
```

### Связи (Relations)

```go
// Один ко многим (One-to-Many)
type User struct {
    gorm.Model
    Integrations []Integration `gorm:"foreignKey:CreatedByID"`
}

type Integration struct {
    gorm.Model
    CreatedByID uint
    CreatedBy   User `gorm:"foreignKey:CreatedByID"`
}

// Загрузка со связями
var user models.User
database.DB.Preload("Integrations").First(&user, 1)
```

### Транзакции

```go
err := database.DB.Transaction(func(tx *gorm.DB) error {
    // Создание пользователя
    user := models.User{Username: "john"}
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    
    // Создание интеграции
    integration := models.Integration{CreatedByID: user.ID}
    if err := tx.Create(&integration).Error; err != nil {
        return err
    }
    
    return nil
})
```

## Работа с HTTP

### Получение данных из запроса

```go
// Query параметры (?key=value)
value := c.Query("key")
valueWithDefault := c.DefaultQuery("key", "default")

// URL параметры (/users/:id)
id := c.Param("id")

// Form данные (POST)
username := c.PostForm("username")

// JSON body
var payload map[string]interface{}
if err := c.BindJSON(&payload); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}

// Bind в структуру
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
var req LoginRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

### Ответы

```go
// HTML
c.HTML(http.StatusOK, "template.html", gin.H{
    "title": "Page Title",
    "data":  someData,
})

// JSON
c.JSON(http.StatusOK, gin.H{
    "status": "success",
    "data":   someData,
})

// Redirect
c.Redirect(http.StatusFound, "/path")

// File
c.File("./path/to/file.pdf")

// String
c.String(http.StatusOK, "Plain text response")
```

### Middleware

```go
// Создание middleware
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // До обработки запроса
        startTime := time.Now()
        
        // Обработка запроса
        c.Next()
        
        // После обработки запроса
        latency := time.Since(startTime)
        log.Printf("Request took %v", latency)
    }
}

// Использование
r.Use(LoggerMiddleware())
```

## Тестирование

### Структура тестов

```go
// internal/services/webhook_processor_test.go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestProcessWebhook(t *testing.T) {
    // Arrange
    payload := map[string]interface{}{
        "name": "John",
        "age":  30,
    }
    
    // Act
    result, err := ProcessWebhook(1, payload)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Запуск тестов

```bash
# Все тесты
go test ./...

# С покрытием
go test -cover ./...

# Конкретный пакет
go test ./internal/services

# Verbose режим
go test -v ./...
```

## Сборка и деплой

### Локальная сборка

```bash
# Windows
go build -o gintegra.exe cmd/server/main.go

# Linux/macOS
go build -o gintegra cmd/server/main.go

# С оптимизацией размера
go build -ldflags="-s -w" -o gintegra cmd/server/main.go
```

### Cross-compilation

```bash
# Для Linux (из Windows/macOS)
GOOS=linux GOARCH=amd64 go build -o gintegra-linux cmd/server/main.go

# Для Windows (из Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o gintegra.exe cmd/server/main.go

# Для macOS (из Windows/Linux)
GOOS=darwin GOARCH=amd64 go build -o gintegra-mac cmd/server/main.go
```

### Docker (будущее)

```dockerfile
# Dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o gintegra cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/gintegra .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./gintegra"]
```

## Отладка

### Логирование

```go
import "gintegra/internal/logger"

// Info
logger.Log.Info("Server started")

// С полями
logger.Log.WithFields(logrus.Fields{
    "user_id": 123,
    "action":  "login",
}).Info("User logged in")

// Error
logger.Log.Error("Failed to connect to database")

// Debug
logger.Log.Debug("Processing webhook", payload)
```

### Gin Debug Mode

```go
// В main.go
if os.Getenv("GIN_MODE") != "release" {
    gin.SetMode(gin.DebugMode)
}
```

### Delve (Go debugger)

```bash
# Установка
go install github.com/go-delve/delve/cmd/dlv@latest

# Запуск с отладчиком
dlv debug cmd/server/main.go

# В VS Code добавьте launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/server/main.go"
        }
    ]
}
```

## Best Practices

### 1. Обработка ошибок

```go
// ❌ Плохо
result, _ := someFunction()

// ✅ Хорошо
result, err := someFunction()
if err != nil {
    logger.Log.Error("Failed to execute function", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
    return
}
```

### 2. Валидация входных данных

```go
// Используйте binding tags
type CreateIntegrationRequest struct {
    Name      string `json:"name" binding:"required,min=3,max=100"`
    TargetAPI string `json:"target_api" binding:"required,url"`
}

if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

### 3. Безопасность

```go
// Всегда хешируйте пароли
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Проверяйте права доступа
if session.Get("role") != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
    return
}

// Экранируйте HTML в шаблонах (автоматически в Go templates)
```

### 4. Производительность

```go
// Используйте индексы в БД
type User struct {
    Username string `gorm:"uniqueIndex"`
    Email    string `gorm:"index"`
}

// Избегайте N+1 запросов
database.DB.Preload("Integrations").Find(&users)

// Используйте пагинацию
database.DB.Limit(10).Offset(20).Find(&items)
```

## Полезные ресурсы

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Контрибьюция

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

### Checklist перед PR

- [ ] Код отформатирован (`go fmt`)
- [ ] Нет ошибок линтера (`go vet`)
- [ ] Добавлены комментарии к публичным функциям
- [ ] Обновлена документация (если нужно)
- [ ] Тесты проходят (`go test`)
- [ ] Проверена работа в браузере
