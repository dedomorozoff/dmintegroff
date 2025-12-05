# Тестирование вебхуков

## Описание

Функционал тестирования вебхуков позволяет создавать временные вебхуки для отладки и тестирования интеграций, аналогично сервису webhook.site.

## Возможности

- **Быстрое создание**: Создание тестового вебхука в один клик с дашборда
- **Автоматическое копирование**: URL вебхука автоматически копируется в буфер обмена
- **Просмотр в реальном времени**: Все запросы отображаются на странице в реальном времени (обновление каждые 2 секунды)
- **Детальная информация**: Просмотр заголовков, тела запроса, query параметров и IP адреса
- **Временные вебхуки**: Автоматическое истечение через 24 часа
- **Тестовые запросы**: Возможность отправить тестовый запрос прямо со страницы

## Использование

### Создание тестового вебхука

1. Перейдите на дашборд
2. Нажмите кнопку "Протестировать вебхук" в разделе "Быстрые действия"
3. URL вебхука будет автоматически скопирован в буфер обмена
4. Откроется страница с тестовым вебхуком

### Просмотр запросов

На странице тестового вебхука вы увидите:

- **URL вебхука**: Можно скопировать в буфер обмена
- **Срок действия**: Когда истечёт срок действия вебхука
- **Счётчик запросов**: Количество полученных запросов
- **Список запросов**: Все запросы в реальном времени

Для каждого запроса доступна следующая информация:

- HTTP метод (GET, POST, PUT, DELETE и т.д.)
- URL с query параметрами
- Заголовки запроса
- Тело запроса
- IP адрес отправителя
- Время получения запроса

### Отправка тестового запроса

Нажмите кнопку "Отправить тестовый запрос" на странице вебхука. Будет отправлен POST запрос с тестовыми данными:

```json
{
  "test": true,
  "message": "Тестовый запрос",
  "timestamp": "2024-12-06T12:00:00.000Z"
}
```

### Удаление тестового вебхука

Нажмите кнопку "Удалить" на странице вебхука. Все данные будут удалены, и вебхук перестанет принимать запросы.

## API Endpoints

### Создание тестового вебхука

```
POST /api/webhook-test/create
```

**Ответ:**
```json
{
  "id": 1,
  "token": "abc123...",
  "webhook_url": "http://localhost:8080/webhook/test/abc123...",
  "expires_at": "2024-12-07T12:00:00Z"
}
```

### Получение запросов

```
GET /api/webhook-test/:token/requests
```

**Ответ:**
```json
{
  "requests": [
    {
      "ID": 1,
      "Method": "POST",
      "URL": "/webhook/test/abc123",
      "Headers": "{\"Content-Type\":\"application/json\"}",
      "Body": "{\"test\":true}",
      "QueryParams": "{}",
      "ClientIP": "127.0.0.1",
      "CreatedAt": "2024-12-06T12:00:00Z"
    }
  ]
}
```

### Удаление тестового вебхука

```
DELETE /api/webhook-test/:token
```

**Ответ:**
```json
{
  "status": "success"
}
```

### Отправка запроса на тестовый вебхук

```
ANY /webhook/test/:token
```

Принимает любой HTTP метод (GET, POST, PUT, DELETE и т.д.)

**Ответ:**
```json
{
  "status": "success",
  "message": "Request received and logged",
  "id": 1
}
```

## Технические детали

### Модели данных

#### WebhookTest

```go
type WebhookTest struct {
    gorm.Model
    Token      string    // Уникальный токен
    UserID     uint      // ID пользователя
    ExpiresAt  time.Time // Время истечения
    IsActive   bool      // Активен ли вебхук
}
```

#### WebhookTestRequest

```go
type WebhookTestRequest struct {
    gorm.Model
    WebhookTestID uint   // ID тестового вебхука
    Method        string // HTTP метод
    URL           string // URL запроса
    Headers       string // JSON с заголовками
    Body          string // Тело запроса
    QueryParams   string // JSON с query параметрами
    ClientIP      string // IP адрес клиента
}
```

### Безопасность

- Каждый тестовый вебхук привязан к пользователю
- Доступ к вебхуку возможен только для владельца
- Автоматическое удаление через 24 часа
- Ограничение на количество сохраняемых запросов (100 последних)

### Производительность

- Polling каждые 2 секунды для обновления списка запросов
- Анимация для новых запросов
- Автоматическая остановка polling при уходе со страницы

## Примеры использования

### Тестирование с curl

```bash
# POST запрос с JSON
curl -X POST http://localhost:8080/webhook/test/abc123 \
  -H "Content-Type: application/json" \
  -d '{"test": true, "data": "example"}'

# GET запрос с параметрами
curl "http://localhost:8080/webhook/test/abc123?param1=value1&param2=value2"

# PUT запрос с заголовками
curl -X PUT http://localhost:8080/webhook/test/abc123 \
  -H "X-Custom-Header: CustomValue" \
  -d "test data"
```

### Тестирование с JavaScript

```javascript
// Отправка POST запроса
fetch('http://localhost:8080/webhook/test/abc123', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Custom-Header': 'CustomValue'
  },
  body: JSON.stringify({
    test: true,
    data: 'example'
  })
})
.then(response => response.json())
.then(data => console.log(data));
```

## Миграции

Для создания таблиц выполните миграцию:

- SQLite: `migrations/013_add_webhook_test_sqlite.sql`
- MySQL/PostgreSQL: `migrations/013_add_webhook_test.sql`

Миграция выполняется автоматически при запуске сервера через GORM AutoMigrate.
