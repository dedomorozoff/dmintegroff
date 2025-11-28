# Техническая документация GIntegra

## 1. Архитектура

Проект построен по архитектуре MVC (Model-View-Controller) с использованием фреймворка Gin.

### Структура проекта

```
gintegra/
├── cmd/
│   └── server/       # Точка входа в приложение (main.go)
├── internal/
│   ├── controllers/  # Обработчики HTTP-запросов
│   ├── database/     # Инициализация и подключение к БД
│   ├── models/       # Структуры данных (GORM модели)
│   ├── routes/       # Определение маршрутов и middleware
│   └── services/     # Бизнес-логика (обработка интеграций)
├── templates/        # HTML-шаблоны
└── go.mod            # Зависимости Go
```

## 2. База данных

Используется ORM GORM. Поддерживаются SQLite и MySQL.

### Модели

#### User (Пользователи)
Таблица: `users`
*   `ID`: uint, Primary Key
*   `Username`: string, Unique
*   `Password`: string (хеш bcrypt)
*   `Role`: string (`admin` или `specialist`)

#### Integration (Интеграции)
Таблица: `integrations`
*   `ID`: uint, Primary Key
*   `Name`: string (Название интеграции)
*   `SourceAPI`: string (URL источника - информационное поле)
*   `TargetAPI`: string (URL назначения, куда отправлять данные)
*   `MappingConfig`: string (JSON конфигурация маппинга полей)
*   `Status`: string (`active` / `inactive`)
*   `CreatedByID`: uint (Foreign Key на `users`)

## 3. Логика работы интеграций

### Создание интеграции
Специалист создает интеграцию, указывая `TargetAPI` и `MappingConfig`.
Пример `MappingConfig`:
```json
{
  "external_id": "id",
  "client_name": "name",
  "order_total": "amount"
}
```
Где ключ — это поле в целевом API, а значение — поле во входящем JSON.

### Обработка вебхука
1.  Внешняя система отправляет POST запрос на `/webhook/:id`.
2.  Система находит интеграцию по `:id`.
3.  Парсится входящий JSON.
4.  Применяется маппинг: создается новый JSON объект, где поля соответствуют ключам из `MappingConfig`, а значения берутся из входящего JSON по соответствующим путям.
5.  Если маппинг пустой, данные передаются "как есть".
6.  Сформированный JSON отправляется POST запросом на `TargetAPI`.

## 4. API Endpoints

### Публичные
*   `GET /login` - Страница входа
*   `POST /login` - Авторизация
*   `GET /logout` - Выход
*   `POST /webhook/:id` - Прием данных для интеграции

### Защищенные (требуется авторизация)
*   `GET /` - Дашборд
*   `GET /integrations` - Список интеграций
*   `GET /integrations/create` - Форма создания
*   `POST /integrations` - Сохранение интеграции

## 5. Безопасность
*   Пароли хранятся в зашифрованном виде (bcrypt).
*   Сессии управляются через `github.com/gin-contrib/sessions` (cookie-based).
*   Доступ к управлению интеграциями закрыт middleware `AuthRequired`.
