# 📘 Техническая документация GIntegra

## Архитектура

GIntegra построен по архитектуре MVC с использованием фреймворка Gin и следует принципам чистой архитектуры.

### Структура проекта

```
gintegra/
├── cmd/
│   └── server/           # Точка входа приложения
│       └── main.go       # Инициализация и запуск сервера
├── internal/
│   ├── controllers/      # HTTP обработчики (Controller layer)
│   │   ├── auth_controller.go
│   │   ├── integration_controller.go
│   │   └── logs_controller.go
│   ├── database/         # Работа с БД
│   │   └── database.go   # Подключение, миграции, seed
│   ├── logger/           # Логирование
│   │   └── logger.go     # Настройка logrus
│   ├── models/           # Модели данных (Model layer)
│   │   ├── user.go
│   │   ├── integration.go
│   │   └── request_log.go
│   ├── routes/           # Маршруты и middleware
│   │   └── routes.go
│   ├── services/         # Бизнес-логика (Service layer)
│   │   └── webhook_processor.go
│   └── utils/            # Утилиты
│       └── token.go
├── static/
│   ├── css/
│   │   ├── common.css    # Старые стили (legacy)
│   │   └── modern.css    # Новый дизайн
│   └── js/
│       └── common.js     # Общие JS функции
├── templates/            # HTML шаблоны (View layer)
│   ├── dashboard.html
│   ├── integrations.html
│   ├── integration_create.html
│   ├── integration_edit.html
│   ├── integration_configure.html
│   ├── logs.html
│   └── login.html
└── docs/                 # Документация
```

## База данных

### ORM и драйверы
- **ORM**: GORM v2
- **SQLite**: По умолчанию (gorm.io/driver/sqlite)
- **MySQL**: Опционально (gorm.io/driver/mysql)

### Модели данных

#### User (Пользователи)
```go
type User struct {
    gorm.Model              // ID, CreatedAt, UpdatedAt, DeletedAt
    Username string         // Уникальное имя пользователя
    Password string         // Bcrypt хеш пароля
    Role     string         // "admin" или "specialist"
}
```

**Таблица**: `users`

**Индексы**:
- `username` - UNIQUE INDEX

**Seed данные**:
- Username: `admin`
- Password: `admin` (bcrypt hash)
- Role: `admin`

#### Integration (Интеграции)
```go
type Integration struct {
    gorm.Model
    Name          string  // Название интеграции
    WebhookToken  string  // Уникальный токен для webhook URL
    SourceAPI     string  // URL источника (опционально, для документации)
    TargetAPI     string  // URL назначения (куда отправлять данные)
    MappingConfig string  // JSON конфигурация маппинга полей
    SamplePayload string  // Пример полученных данных (для настройки)
    Mode          string  // "listening", "active", "inactive"
    Status        string  // Статус интеграции (deprecated, используйте Mode)
    CreatedByID   uint    // Foreign Key на users.ID
}
```

**Таблица**: `integrations`

**Режимы работы**:
- `listening` - Режим прослушивания, захватывает структуру данных
- `active` - Активная интеграция, обрабатывает и пересылает данные
- `inactive` - Неактивная интеграция

#### RequestLog (Логи запросов)
```go
type RequestLog struct {
    gorm.Model
    Method       string  // HTTP метод
    URL          string  // URL запроса
    RequestBody  string  // Тело запроса (JSON)
    ResponseBody string  // Тело ответа (опционально)
    StatusCode   int     // HTTP статус код
}
```

**Таблица**: `request_logs`

## Логика работы интеграций

### 1. Создание интеграции

**Endpoint**: `POST /integrations`

**Процесс**:
1. Пользователь заполняет форму (название, Target API, Source API)
2. Генерируется уникальный webhook токен (16 байт, hex)
3. Интеграция создается в режиме `listening`
4. Возвращается webhook URL: `/webhook/{token}`

**Код**:
```go
token, _ := utils.GenerateToken(16)
integration := models.Integration{
    Name:         name,
    WebhookToken: token,
    TargetAPI:    targetAPI,
    Mode:         "listening",
    CreatedByID:  userID,
}
```

### 2. Режим прослушивания (Listening Mode)

**Endpoint**: `POST /webhook/{token}`

**Процесс**:
1. Внешняя система отправляет тестовый запрос
2. Система сохраняет `SamplePayload` (структуру данных)
3. Возвращает статус "captured"
4. Пользователь может настроить маппинг

**Пример запроса**:
```bash
curl -X POST http://localhost:8080/webhook/abc123 \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "name": "John", "email": "john@example.com"}'
```

**Ответ**:
```json
{
  "status": "captured",
  "message": "Sample data captured. Configure field mapping to activate integration."
}
```

### 3. Настройка маппинга

**Endpoint**: `GET /integrations/{id}/configure`

**Процесс**:
1. Отображается `SamplePayload` в читаемом виде
2. Для каждого поля можно указать:
   - Новое имя поля для Target API
   - Флаг "игнорировать" (не отправлять это поле)
3. Маппинг сохраняется в JSON формате

**Пример MappingConfig**:
```json
{
  "user_id": "external_id",
  "name": "client_name",
  "email": "client_email"
}
```

**Сохранение**: `POST /integrations/{id}/configure`
- Сохраняет `MappingConfig`
- Переводит интеграцию в режим `active`

### 4. Активная обработка (Active Mode)

**Endpoint**: `POST /webhook/{token}`

**Процесс**:
1. Получение входящих данных
2. Применение маппинга полей
3. Отправка на Target API
4. Логирование результата

**Код обработки** (упрощенно):
```go
func ProcessWebhook(integrationID uint, payload map[string]interface{}) error {
    // 1. Загрузка интеграции
    var integration models.Integration
    db.First(&integration, integrationID)
    
    // 2. Применение маппинга
    var mapping map[string]string
    json.Unmarshal([]byte(integration.MappingConfig), &mapping)
    
    transformed := make(map[string]interface{})
    for targetField, sourceField := range mapping {
        if value, exists := payload[sourceField]; exists {
            transformed[targetField] = value
        }
    }
    
    // 3. Отправка на Target API
    jsonData, _ := json.Marshal(transformed)
    resp, err := http.Post(integration.TargetAPI, "application/json", bytes.NewBuffer(jsonData))
    
    // 4. Логирование
    log := models.RequestLog{
        Method:      "POST",
        URL:         integration.TargetAPI,
        RequestBody: string(jsonData),
        StatusCode:  resp.StatusCode,
    }
    db.Create(&log)
    
    return err
}
```

## API Endpoints

### Публичные (без авторизации)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/login` | Страница входа |
| POST | `/login` | Авторизация пользователя |
| GET | `/logout` | Выход из системы |
| POST | `/webhook/:token` | Прием данных для интеграции |
| POST | `/test` | Тестовый endpoint для отладки |

### Защищенные (требуется авторизация)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/` | Главная страница (дашборд) |
| GET | `/integrations` | Список всех интеграций |
| GET | `/integrations/create` | Форма создания интеграции |
| POST | `/integrations` | Сохранение новой интеграции |
| GET | `/integrations/:id/edit` | Форма редактирования |
| POST | `/integrations/:id/update` | Обновление интеграции |
| POST | `/integrations/:id/delete` | Удаление интеграции |
| GET | `/integrations/:id/configure` | Настройка маппинга полей |
| POST | `/integrations/:id/configure` | Сохранение маппинга |
| GET | `/logs` | Просмотр логов запросов |
| POST | `/logs/clear` | Очистка всех логов |
| POST | `/logs/:id/delete` | Удаление конкретного лога |

## Безопасность

### Аутентификация
- **Механизм**: Cookie-based sessions
- **Библиотека**: `github.com/gin-contrib/sessions`
- **Хранилище**: Cookie store с секретным ключом

**Настройка**:
```go
secret := os.Getenv("SESSION_SECRET")
if secret == "" {
    secret = "secret" // Не используйте в production!
}
store := cookie.NewStore([]byte(secret))
r.Use(sessions.Sessions("mysession", store))
```

### Хеширование паролей
- **Алгоритм**: bcrypt
- **Библиотека**: `golang.org/x/crypto/bcrypt`
- **Cost**: DefaultCost (10)

**Пример**:
```go
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
```

### Middleware авторизации
```go
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        user := session.Get("user_id")
        if user == nil {
            c.Redirect(http.StatusFound, "/login")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### Генерация токенов
- **Длина**: 16 байт (32 символа hex)
- **Источник**: `crypto/rand`
- **Формат**: Hexadecimal

## Логирование

### Библиотека
- **Logrus**: Структурированное логирование
- **Формат**: JSON (можно настроить)

### Уровни логирования
- `Info` - Обычные операции
- `Error` - Ошибки выполнения
- `Debug` - Отладочная информация

### HTTP Request Logger
Middleware логирует каждый HTTP запрос:
```go
logger.Log.WithFields(logrus.Fields{
    "method":  c.Request.Method,
    "path":    c.Request.URL.Path,
    "status":  c.Writer.Status(),
    "latency": latency,
    "ip":      c.ClientIP(),
}).Info("HTTP Request")
```

## Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SESSION_SECRET` | Секретный ключ для сессий | `secret` |
| `DB_TYPE` | Тип БД (sqlite/mysql) | `sqlite` |
| `DB_DSN` | Строка подключения к БД | `gintegra.db` |
| `APP_PATH` | Префикс пути приложения | `` |
| `PORT` | Порт сервера | `8080` |

### Пример .env файла
```env
SESSION_SECRET=your-super-secret-key-change-this-in-production
DB_TYPE=sqlite
DB_DSN=gintegra.db
APP_PATH=
PORT=8080
```

## Зависимости

### Основные
```go
require (
    github.com/gin-gonic/gin v1.11.0
    github.com/gin-contrib/sessions v1.0.1
    gorm.io/gorm v1.25.12
    gorm.io/driver/sqlite v1.5.6
    golang.org/x/crypto v0.28.0
    github.com/sirupsen/logrus v1.9.3
    github.com/joho/godotenv v1.5.1
)
```

### Установка
```bash
go mod download
go mod tidy
```

## Производительность

### Рекомендации
- Используйте connection pooling для БД
- Настройте индексы для часто запрашиваемых полей
- Используйте кеширование для статических ресурсов
- Ограничьте размер логов (автоочистка старых записей)

### Масштабирование
- Для высоких нагрузок используйте MySQL/PostgreSQL вместо SQLite
- Рассмотрите использование очередей (Redis, RabbitMQ) для обработки webhook
- Используйте reverse proxy (nginx) для статических файлов

## Отладка

### Включение debug режима Gin
```go
gin.SetMode(gin.DebugMode) // или gin.ReleaseMode
```

### Просмотр SQL запросов
```go
db, _ := gorm.Open(sqlite.Open("gintegra.db"), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

### Тестовый endpoint
```bash
# Отправка тестовых данных
curl -X POST http://localhost:8080/test \
  -H "Content-Type: application/json" \
  -d '{"test": "data", "value": 123}'

# Просмотр в интерфейсе: /logs
```

## Известные ограничения

1. **Нет аутентификации для Target API** - Пока не поддерживается OAuth, API keys
2. **Простой маппинг** - Только переименование полей, нет трансформаций
3. **Нет retry механизма** - При ошибке отправки данные теряются
4. **SQLite ограничения** - Не подходит для высоких нагрузок
5. **Нет rate limiting** - Webhook может быть перегружен

## Планы развития

- [ ] Поддержка OAuth 2.0 для Target API
- [ ] Сложные трансформации данных (JSONPath, шаблоны)
- [ ] Retry механизм с экспоненциальной задержкой
- [ ] Webhook подписи (HMAC)
- [ ] Rate limiting
- [ ] Метрики и мониторинг (Prometheus)
- [ ] WebSocket для real-time обновлений
- [ ] Поддержка GraphQL
