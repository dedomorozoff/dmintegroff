# 🛠️ Разработка dmIntegroff

## Быстрый старт для разработчиков

```bash
# 1. Клонировать репозиторий
git clone https://github.com/dedomorozoff/dmintegroff.git
cd dmintegroff

# 2. Установить зависимости
go mod download

# 3. Настроить окружение
cp .env.example .env

# 4. Применить миграции
go build -o dmIntegroff-admin.exe cmd/admin/main.go
.\dmIntegroff-admin.exe --migrate

# 5. Запустить сервер
go run cmd/server/main.go
```

Приложение доступно на http://localhost:8080
- Логин: `admin`
- Пароль: `admin`

## Структура проекта

```
dmintegroff/
├── cmd/
│   ├── admin/          # CLI инструмент администрирования
│   └── server/         # Основное приложение
├── internal/
│   ├── controllers/    # HTTP обработчики
│   ├── database/       # Подключение к БД и сидеры
│   ├── logger/         # Логирование
│   ├── models/         # Модели данных (GORM)
│   ├── routes/         # Маршруты и middleware
│   └── utils/          # Утилиты (JWT, хеширование)
├── migrations/         # SQL миграции
├── static/            # CSS, JS, изображения
├── templates/         # HTML шаблоны (Go templates)
└── docs/              # Документация
```

## Основные команды

### Разработка
```bash
# Запуск с hot-reload (требует air)
air

# Обычный запуск
go run cmd/server/main.go

# Сборка
go build -o dmintegroff.exe cmd/server/main.go
```

### Миграции
```bash
# Применить все миграции
.\dmIntegroff-admin.exe --migrate

# Создать новую миграцию
# Добавьте SQL файл в migrations/ с номером версии
# Например: 004_add_new_feature.sql
```

### Тестирование
```bash
# Тест логирования
.\test_logs.ps1

# Тест проектов
.\test_projects.sh

# Отправка тестового запроса
curl -X POST http://localhost:8080/test \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

## Добавление новой функциональности

### 1. Создание модели

```go
// internal/models/new_model.go
package models

import "gorm.io/gorm"

type NewModel struct {
    gorm.Model
    Name string `json:"name"`
}
```

### 2. Добавление в миграции

```go
// cmd/admin/main.go
database.DB.AutoMigrate(
    &models.User{},
    &models.Project{},
    &models.Integration{},
    &models.RequestLog{},
    &models.NewModel{}, // Добавить здесь
)
```

### 3. Создание контроллера

```go
// internal/controllers/new_controller.go
package controllers

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func NewList(c *gin.Context) {
    // Ваш код
}
```

### 4. Добавление маршрутов

```go
// internal/routes/routes.go
authorized.GET("/new", controllers.NewList)
authorized.POST("/new", controllers.NewCreate)
```

### 5. Создание шаблона

```html
<!-- templates/new.html -->
<!DOCTYPE html>
<html lang="ru">
<head>
    <link rel="stylesheet" href="/static/css/modern.css">
</head>
<body>
    <!-- Ваш HTML -->
</body>
</html>
```

## Работа с базой данных

### SQLite (по умолчанию)
```bash
# Просмотр данных
sqlite3 dmintegroff.db "SELECT * FROM users;"

# Бэкап
copy dmintegroff.db dmintegroff.db.backup
```

### MySQL
```env
# .env
DB_TYPE=mysql
DB_DSN=user:password@tcp(localhost:3306)/dmintegroff?charset=utf8mb4&parseTime=True&loc=Local
```

## Логирование

### Файловое логирование
```env
# .env
LOG_LEVEL=debug
LOG_FILE=dmintegroff.log
```

### Просмотр логов
```bash
# Windows
Get-Content dmintegroff.log -Tail 50 -Wait

# Linux/Mac
tail -f dmintegroff.log
```

## Отладка

### Включить debug режим
```env
# .env
GIN_MODE=debug
LOG_LEVEL=debug
```

### Проверка ошибок
```bash
# Проверить синтаксис
go vet ./...

# Форматирование
go fmt ./...

# Линтер (требует golangci-lint)
golangci-lint run
```

## Полезные ссылки

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Go Templates](https://pkg.go.dev/html/template)

## Документация проекта

- [README.md](README.md) - Общая информация
- [QUICKSTART.md](QUICKSTART.md) - Быстрый старт
- [UPDATE.md](UPDATE.md) - Обновление проекта
- [CLI.md](CLI.md) - CLI инструмент
- [PRODUCTION.md](PRODUCTION.md) - Деплой в продакшн
- [CHANGELOG.md](CHANGELOG.md) - История изменений
- [docs/PROJECTS.md](docs/PROJECTS.md) - Работа с проектами
