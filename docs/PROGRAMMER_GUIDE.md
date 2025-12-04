# 👨‍ Руководство программиста dmIntegroff

Этот документ для разработчиков, которые хотят развивать или модифицировать dmIntegroff.

## Настройка окружения разработки

### Требования
- **Go**: 1.20 или новее
- **Git**: Для управления версиями
- **IDE**: VS Code (с Go extension) или GoLand

### Установка

```bash
# Клонирование репозитория
git clone https://github.com/dedomorozoff/dmintegroff.git
cd dmIntegroff

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
│ HTTP Layer (Gin) │
│ ┌──────────────────────────────┐ │
│ │ Routes & Middleware │ │
│ └──────────────────────────────┘ │
└─────────────────────────────────────┘
 ↓
┌─────────────────────────────────────┐
│ Controller Layer │
│ ┌──────────────────────────────┐ │
│ │ HTTP Request Handlers │ │
│ └──────────────────────────────┘ │
└─────────────────────────────────────┘
 ↓
┌─────────────────────────────────────┐
│ Service Layer │
│ ┌──────────────────────────────┐ │
│ │ Business Logic │ │
│ └──────────────────────────────┘ │
└─────────────────────────────────────┘
 ↓
┌─────────────────────────────────────┐
│ Model Layer (GORM) │
│ ┌──────────────────────────────┐ │
│ │ Database Models │ │
│ └──────────────────────────────┘ │
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

## Архитектура системы

### Основные компоненты

```
dmIntegroff
├── Веб-интерфейс (Gin + HTML Templates)
│ ├── Дашборд с статистикой
│ ├── Управление интеграциями
│ ├── Настройки пользователя
│ └── Мониторинг и логи
│
├── API Layer
│ ├── REST endpoints для интеграций
│ ├── Webhook endpoints
│ ├── API для статистики
│ └── API для активности
│
├── Business Logic
│ ├── Обработка webhook запросов
│ ├── Трансформация данных (маппинг)
│ ├── Отправка на Target API
│ └── Логирование с ограничением
│
└── Data Layer
 ├── SQLite/MySQL
 ├── Модели GORM
 └── Автоматическая очистка логов
```

### Поток обработки webhook

```
1. Webhook запрос → /webhook/:token
2. Проверка токена интеграции
3. Валидация JSON
4. Логирование входящего запроса (тип: webhook)
5. Проверка режима интеграции:
- listening: сохранить sample payload
- active: обработать и отправить
- inactive: вернуть статус
6. Трансформация данных по маппингу
7. Отправка на Target API
8. Логирование исходящего запроса (тип: webhook)
9. Автоочистка логов (лимит 50 записей)
```

### Типы логов

Система использует унифицированные типы логов:

- **webhook** - входящие и исходящие webhook запросы (учитываются в статистике)
- **test** - тестовые запросы через /webhook/test (не учитываются в статистике)
- **error** - ошибки системы

## Добавление нового функционала

### 1. Добавление новой модели

**Шаг 1**: Создайте файл в `internal/models/`

```go
// internal/models/notification.go
package models

import "gorm.io/gorm"

type Notification struct {
 gorm.Model
 UserID uint `gorm:"not null" json:"user_id"`
 Message string `gorm:"type:text;not null" json:"message"`
 IsRead bool `gorm:"default:false" json:"is_read"`
 
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
 "dmIntegroff/internal/database"
 "dmIntegroff/internal/models"
 "net/http"
 
 "github.com/gin-gonic/gin"
)

// NotificationList отображает список уведомлений
func NotificationList(c *gin.Context) {
 var notifications []models.Notification
 database.DB.Preload("User").Find(&notifications)
 
 c.HTML(http.StatusOK, "notifications.html", gin.H{
 "title": "Уведомления",
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
 "dmIntegroff/internal/database"
 "dmIntegroff/internal/models"
)

// NotificationService управляет уведомлениями
type NotificationService struct{}

// SendNotification создает новое уведомление
func (s *NotificationService) SendNotification(userID uint, message string) error {
 notification := models.Notification{
 UserID: userID,
 Message: message,
 IsRead: false,
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

## Система логирования

### Автоматическое ограничение логов

Система автоматически ограничивает количество логов до 50 записей:

```go
// internal/services/integration_service.go
func CreateLogWithLimit(log *models.RequestLog) {
 database.DB.Create(log)

 // Подсчитываем количество логов
 var count int64
 database.DB.Model(&models.RequestLog{}).Count(&count)

 // Если больше 50, удаляем самые старые
 if count > 50 {
 database.DB.Exec("DELETE FROM request_logs WHERE id IN (SELECT id FROM request_logs ORDER BY created_at ASC LIMIT ?)", count-50)
 }
}
```

**Важно**: Всегда используйте `CreateLogWithLimit` вместо прямого `database.DB.Create` для логов!

### Правильное логирование

```go
// Правильно - с ограничением
log := models.RequestLog{
 IntegrationID: integrationID,
 Method: "POST",
 URL: targetURL,
 RequestBody: string(jsonData),
 StatusCode: 200,
 LogType: "webhook",
}
CreateLogWithLimit(&log)

// Неправильно - без ограничения
database.DB.Create(&log)
```

### Типы логов в коде

```go
// Webhook запросы (входящие и исходящие)
log.LogType = "webhook" // Учитываются в статистике

// Тестовые запросы
log.LogType = "test" // Не учитываются в статистике

// Ошибки
log.LogType = "error" // Системные ошибки
```

## Статистика и мониторинг

### API для статистики

```go
// internal/controllers/dashboard_controller.go
func GetRequestStats(c *gin.Context) {
 type DayStats struct {
 Date string `json:"date"`
 Count int `json:"count"`
 }

 var stats []DayStats
 
 // SQL запрос для группировки по дням
 query := `
 SELECT 
 DATE(created_at) as date,
 COUNT(*) as count
 FROM request_logs
 WHERE log_type = 'webhook'
 AND created_at >= datetime('now', '-30 days')
 GROUP BY DATE(created_at)
 ORDER BY date ASC
 `
 
 rows, err := database.DB.Raw(query).Rows()
 // ... обработка результатов
}
```

### Добавление нового API endpoint

```go
// 1. Создайте функцию в контроллере
func GetCustomStats(c *gin.Context) {
 // Ваша логика
 c.JSON(http.StatusOK, gin.H{
 "data": result,
 })
}

// 2. Зарегистрируйте маршрут
// internal/routes/routes.go
authorized.GET("/api/custom-stats", controllers.GetCustomStats)
```

### Интеграция с Chart.js

```javascript
// templates/pages/dashboard.html
fetch('/api/stats')
 .then(response => response.json())
 .then(data => {
 const labels = data.stats.map(s => s.date);
 const counts = data.stats.map(s => s.count);
 
 new Chart(ctx, {
 type: 'line',
 data: {
 labels: labels,
 datasets: [{
 label: 'Количество запросов',
 data: counts,
 borderColor: 'rgb(139, 142, 255)',
 backgroundColor: 'rgba(139, 142, 255, 0.3)',
 borderWidth: 4
 }]
 }
 });
 });
```

## Работа с настройками пользователя

### Смена пароля

```go
// internal/controllers/settings_controller.go
func ChangePassword(c *gin.Context) {
 session := sessions.Default(c)
 userID := session.Get("user_id")
 
 // Получаем данные из формы
 currentPassword := c.PostForm("current_password")
 newPassword := c.PostForm("new_password")
 confirmPassword := c.PostForm("confirm_password")
 
 // Получаем пользователя
 var user models.User
 database.DB.First(&user, userID)
 
 // Проверяем текущий пароль
 if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
 // Неверный пароль
 return
 }
 
 // Проверяем совпадение нового пароля
 if newPassword != confirmPassword {
 // Пароли не совпадают
 return
 }
 
 // Хешируем новый пароль
 hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
 
 // Обновляем в БД
 user.Password = string(hashedPassword)
 database.DB.Save(&user)
}
```

### Добавление новых настроек

1. Добавьте поля в форму (`templates/pages/settings.html`)
2. Обработайте в контроллере
3. Сохраните в модель User или создайте новую модель Settings

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
 CreatedBy User `gorm:"foreignKey:CreatedByID"`
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
 "data": someData,
})

// JSON
c.JSON(http.StatusOK, gin.H{
 "status": "success",
 "data": someData,
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
 "age": 30,
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
go build -o dmIntegroff.exe cmd/server/main.go

# Linux/macOS
go build -o dmIntegroff cmd/server/main.go

# С оптимизацией размера
go build -ldflags="-s -w" -o dmIntegroff cmd/server/main.go
```

### Cross-compilation

```bash
# Для Linux (из Windows/macOS)
GOOS=linux GOARCH=amd64 go build -o dmIntegroff-linux cmd/server/main.go

# Для Windows (из Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o dmIntegroff.exe cmd/server/main.go

# Для macOS (из Windows/Linux)
GOOS=darwin GOARCH=amd64 go build -o dmIntegroff-mac cmd/server/main.go
```

### Docker (будущее)

```dockerfile
# Dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o dmIntegroff cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dmIntegroff .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./dmIntegroff"]
```

## Отладка

### Логирование

```go
import "dmIntegroff/internal/logger"

// Info
logger.Log.Info("Server started")

// С полями
logger.Log.WithFields(logrus.Fields{
 "user_id": 123,
 "action": "login",
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

## Работа с UI и темной темой

### CSS переменные

Система использует CSS переменные для темной темы:

```css
/* static/css/modern.css */
:root {
--background: oklch(0.13 0.01 260);
--foreground: oklch(0.97 0 0);
--card: oklch(0.17 0.01 260);
--primary: oklch(0.65 0.15 280);
--success: oklch(0.65 0.17 145);
--destructive: oklch(0.55 0.2 25);
--border: oklch(0.28 0.01 260);
}
```

### Добавление новых стилей

```css
/* Используйте существующие переменные */
.my-component {
 background: var(--card);
 color: var(--foreground);
 border: 1px solid var(--border);
}

/* Для интерактивных элементов */
.my-button {
 background: var(--primary);
 color: var(--primary-foreground);
}

.my-button:hover {
 opacity: 0.9;
}
```

### Модальные окна

```html
<!-- HTML структура -->
<div id="myModal" class="modal" style="display: none;">
 <div class="modal-content">
 <div class="modal-header">
 <h2>Заголовок</h2>
 <button class="modal-close" onclick="closeModal()">&times;</button>
 </div>
 <div class="modal-body">
 <!-- Контент -->
 </div>
 </div>
</div>

<script>
function showModal() {
 document.getElementById('myModal').style.display = 'flex';
}

function closeModal() {
 document.getElementById('myModal').style.display = 'none';
}

// Закрытие по клику вне окна
window.onclick = function(event) {
 const modal = document.getElementById('myModal');
 if (event.target === modal) {
 closeModal();
 }
}
</script>
```

### Интеграция с Lucide Icons

```html
<!-- Используйте data-lucide атрибут -->
<i data-lucide="settings"></i>
<i data-lucide="check-circle-2"></i>
<i data-lucide="alert-triangle"></i>

<!-- Инициализация в scripts.html -->
<script>
 lucide.createIcons();
</script>
```

## Оптимизация производительности

### SQL запросы

```go
// Плохо - N+1 запросов
var integrations []models.Integration
database.DB.Find(&integrations)
for _, integration := range integrations {
 var project models.Project
 database.DB.First(&project, integration.ProjectID)
}

// Хорошо - один запрос с JOIN
var integrations []models.Integration
database.DB.Preload("Project").Find(&integrations)
```

### Кэширование

```go
// Пример простого кэша в памяти
var statsCache struct {
 data []DayStats
 timestamp time.Time
 mu sync.RWMutex
}

func GetRequestStats(c *gin.Context) {
 statsCache.mu.RLock()
 // Проверяем кэш (5 минут)
 if time.Since(statsCache.timestamp) < 5*time.Minute {
 c.JSON(http.StatusOK, gin.H{"stats": statsCache.data})
 statsCache.mu.RUnlock()
 return
 }
 statsCache.mu.RUnlock()
 
 // Получаем свежие данные
 stats := fetchStatsFromDB()
 
 // Обновляем кэш
 statsCache.mu.Lock()
 statsCache.data = stats
 statsCache.timestamp = time.Now()
 statsCache.mu.Unlock()
 
 c.JSON(http.StatusOK, gin.H{"stats": stats})
}
```

### Пагинация

```go
func GetLogs(c *gin.Context) {
 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
 pageSize := 20
 offset := (page - 1) * pageSize
 
 var logs []models.RequestLog
 var total int64
 
 database.DB.Model(&models.RequestLog{}).Count(&total)
 database.DB.Limit(pageSize).Offset(offset).Find(&logs)
 
 c.JSON(http.StatusOK, gin.H{
 "logs": logs,
 "total": total,
 "page": page,
 "page_size": pageSize,
 "total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
 })
}
```

## Безопасность

### Защита от SQL инъекций

```go
// Правильно - параметризованные запросы
database.DB.Where("username = ?", username).First(&user)

// Неправильно - конкатенация строк
database.DB.Where("username = '" + username + "'").First(&user)
```

### Защита от XSS

```html
<!-- Go templates автоматически экранируют HTML -->
<p>{{ .userInput }}</p> <!-- Безопасно -->

<!-- Для вывода HTML используйте template.HTML осторожно -->
<div>{{ .trustedHTML }}</div>
```

### Валидация webhook токенов

```go
func WebhookHandler(c *gin.Context) {
 token := c.Param("token")
 
 var integration models.Integration
 if err := database.DB.Where("webhook_token = ?", token).First(&integration).Error; err != nil {
 c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
 return
 }
 
 // Продолжаем обработку
}
```

### Проверка прав доступа

```go
func AdminOnly() gin.HandlerFunc {
 return func(c *gin.Context) {
 session := sessions.Default(c)
 role := session.Get("role")
 
 if role != "admin" {
 c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
 c.Abort()
 return
 }
 
 c.Next()
 }
}

// Использование
authorized.GET("/admin/users", AdminOnly(), controllers.UserList)
```

## Отладка и профилирование

### Логирование запросов

```go
// Middleware для логирования всех запросов
func RequestLogger() gin.HandlerFunc {
 return func(c *gin.Context) {
 start := time.Now()
 path := c.Request.URL.Path
 
 c.Next()
 
 latency := time.Since(start)
 statusCode := c.Writer.Status()
 
 logger.Log.WithFields(map[string]interface{}{
 "method": c.Request.Method,
 "path": path,
 "status": statusCode,
 "latency": latency,
 "ip": c.ClientIP(),
 }).Info("Request processed")
 }
}
```

### Профилирование

```go
import _ "net/http/pprof"

// В main.go для режима разработки
if os.Getenv("DEBUG") == "true" {
 go func() {
 log.Println(http.ListenAndServe("localhost:6060", nil))
 }()
}

// Доступ к профилировщику:
// http://localhost:6060/debug/pprof/
```

### Мониторинг памяти

```go
import "runtime"

func GetMemStats(c *gin.Context) {
 var m runtime.MemStats
 runtime.ReadMemStats(&m)
 
 c.JSON(http.StatusOK, gin.H{
 "alloc_mb": m.Alloc / 1024 / 1024,
 "total_alloc_mb": m.TotalAlloc / 1024 / 1024,
 "sys_mb": m.Sys / 1024 / 1024,
 "num_gc": m.NumGC,
 })
}
```

## Best Practices

### 1. Обработка ошибок

```go
// Плохо
result, _ := someFunction()

// Хорошо
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
 Name string `json:"name" binding:"required,min=3,max=100"`
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
 Email string `gorm:"index"`
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
- [ ] Используется `CreateLogWithLimit` для логов
- [ ] Правильные типы логов (webhook/test/error)
- [ ] Проверена работа в темной теме
- [ ] Нет SQL инъекций и XSS уязвимостей

## Частые задачи

### Добавление нового типа статистики

1. Создайте SQL запрос для получения данных
2. Добавьте API endpoint в контроллере
3. Зарегистрируйте маршрут
4. Создайте визуализацию на фронтенде

```go
// 1. Контроллер
func GetIntegrationStats(c *gin.Context) {
 type IntegrationStat struct {
 Name string `json:"name"`
 Count int `json:"count"`
 }
 
 var stats []IntegrationStat
 database.DB.Raw(`
 SELECT i.name, COUNT(rl.id) as count
 FROM integrations i
 LEFT JOIN request_logs rl ON rl.integration_id = i.id
 WHERE rl.log_type = 'webhook'
 GROUP BY i.id, i.name
 ORDER BY count DESC
 LIMIT 10
 `).Scan(&stats)
 
 c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// 2. Маршрут
authorized.GET("/api/integration-stats", controllers.GetIntegrationStats)

// 3. Фронтенд
fetch('/api/integration-stats')
 .then(response => response.json())
 .then(data => {
 // Отобразить данные
 });
```

### Добавление нового фильтра в логи

```go
// Контроллер
func LogsPage(c *gin.Context) {
 query := database.DB.Preload("Integration").Order("created_at desc")
 
 // Существующие фильтры
 if integrationID := c.Query("integration_id"); integrationID != "" {
 query = query.Where("integration_id = ?", integrationID)
 }
 
 // Новый фильтр по дате
 if dateFrom := c.Query("date_from"); dateFrom != "" {
 query = query.Where("created_at >= ?", dateFrom)
 }
 if dateTo := c.Query("date_to"); dateTo != "" {
 query = query.Where("created_at <= ?", dateTo)
 }
 
 var logs []models.RequestLog
 query.Limit(100).Find(&logs)
 
 c.HTML(http.StatusOK, "pages/logs.html", gin.H{
 "logs": logs,
 "filter_date_from": c.Query("date_from"),
 "filter_date_to": c.Query("date_to"),
 })
}
```

### Добавление уведомлений

```go
// 1. Модель
type Notification struct {
 gorm.Model
 UserID uint `gorm:"not null"`
 Message string `gorm:"type:text;not null"`
 Type string `gorm:"default:'info'"` // info, success, warning, error
 IsRead bool `gorm:"default:false"`
}

// 2. Сервис
func SendNotification(userID uint, message, notifType string) {
 notification := models.Notification{
 UserID: userID,
 Message: message,
 Type: notifType,
 }
 database.DB.Create(&notification)
}

// 3. Использование
SendNotification(userID, "Интеграция успешно создана", "success")
```

### Экспорт данных в CSV

```go
func ExportLogsCSV(c *gin.Context) {
 var logs []models.RequestLog
 database.DB.Preload("Integration").Find(&logs)
 
 c.Header("Content-Type", "text/csv")
 c.Header("Content-Disposition", "attachment; filename=logs.csv")
 
 writer := csv.NewWriter(c.Writer)
 defer writer.Flush()
 
 // Заголовки
 writer.Write([]string{"ID", "Integration", "Method", "URL", "Status", "Created At"})
 
 // Данные
 for _, log := range logs {
 integrationName := ""
 if log.Integration.ID > 0 {
 integrationName = log.Integration.Name
 }
 
 writer.Write([]string{
 fmt.Sprintf("%d", log.ID),
 integrationName,
 log.Method,
 log.URL,
 fmt.Sprintf("%d", log.StatusCode),
 log.CreatedAt.Format("2006-01-02 15:04:05"),
 })
 }
}
```

## Миграция данных

### Добавление нового поля в существующую таблицу

```go
// 1. Обновите модель
type Integration struct {
 gorm.Model
 Name string
 WebhookToken string
 // Новое поле
 Description string `gorm:"type:text"`
}

// 2. GORM автоматически добавит поле при следующем запуске
// Или создайте миграцию вручную:
database.DB.Exec("ALTER TABLE integrations ADD COLUMN description TEXT")
```

### Изменение типа поля

```go
// SQLite не поддерживает ALTER COLUMN, нужно пересоздать таблицу
// Для MySQL:
database.DB.Exec("ALTER TABLE integrations MODIFY COLUMN name VARCHAR(255)")
```

### Миграция данных

```go
func MigrateOldLogs() {
 // Обновляем старые логи с типом "incoming" на "webhook"
 database.DB.Exec("UPDATE request_logs SET log_type = 'webhook' WHERE log_type = 'incoming'")
 database.DB.Exec("UPDATE request_logs SET log_type = 'webhook' WHERE log_type = 'outgoing'")
 database.DB.Exec("UPDATE request_logs SET log_type = 'test' WHERE log_type = 'request'")
}
```

## Troubleshooting

### Проблема: Логи не отображаются

**Решение:**
1. Проверьте тип лога: `log.LogType = "webhook"`
2. Убедитесь, что используется `CreateLogWithLimit`
3. Проверьте фильтры в UI

### Проблема: График не отображается

**Решение:**
1. Проверьте консоль браузера на ошибки JavaScript
2. Убедитесь, что API возвращает данные: `/api/stats`
3. Проверьте SQL запрос для вашей БД (SQLite vs MySQL)
4. Убедитесь, что Chart.js загружен

### Проблема: Высокое потребление памяти

**Решение:**
1. Проверьте количество логов в БД
2. Убедитесь, что работает автоочистка
3. Используйте пагинацию для больших списков
4. Добавьте индексы в БД

### Проблема: Медленные запросы

**Решение:**
1. Добавьте индексы на часто используемые поля
2. Используйте `Preload` для связей
3. Ограничьте количество возвращаемых записей
4. Используйте кэширование для статистики

## Дополнительные ресурсы

### Документация проекта
- [README.md](../README.md) - Общая информация
- [USER_GUIDE.md](USER_GUIDE.md) - Руководство пользователя
- [SETTINGS_GUIDE.md](SETTINGS_GUIDE.md) - Настройка системы
- [CHANGELOG.md](CHANGELOG.md) - История изменений

### Внешние ресурсы
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Chart.js Documentation](https://www.chartjs.org/docs/)
- [Lucide Icons](https://lucide.dev/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go)
