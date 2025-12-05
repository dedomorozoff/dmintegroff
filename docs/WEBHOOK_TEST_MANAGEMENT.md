# Управление тестовыми вебхуками

## Хранение данных

### База данных

Тестовые вебхуки хранятся в двух таблицах:

#### 1. `webhook_tests`
Основная таблица с информацией о вебхуках:

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INTEGER/BIGINT | Уникальный ID |
| `token` | VARCHAR(255) | Уникальный токен (32 символа hex) |
| `user_id` | INTEGER/BIGINT | ID владельца |
| `expires_at` | TIMESTAMP | Время истечения (24 часа) |
| `is_active` | BOOLEAN | Статус активности |
| `created_at` | TIMESTAMP | Время создания |
| `updated_at` | TIMESTAMP | Время обновления |
| `deleted_at` | TIMESTAMP | Время удаления (soft delete) |

**Индексы:**
- `token` - UNIQUE INDEX (быстрый поиск по токену)
- `user_id` - INDEX (фильтрация по пользователю)

#### 2. `webhook_test_requests`
Таблица с запросами к вебхукам:

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INTEGER/BIGINT | Уникальный ID |
| `webhook_test_id` | INTEGER/BIGINT | ID вебхука |
| `method` | VARCHAR(10) | HTTP метод |
| `url` | TEXT | URL запроса |
| `headers` | TEXT | JSON с заголовками |
| `body` | TEXT/LONGTEXT | Тело запроса |
| `query_params` | TEXT | JSON с query параметрами |
| `client_ip` | VARCHAR(45) | IP адрес клиента |
| `created_at` | TIMESTAMP | Время получения |
| `updated_at` | TIMESTAMP | Время обновления |
| `deleted_at` | TIMESTAMP | Время удаления |

**Индексы:**
- `webhook_test_id` - INDEX (быстрый поиск запросов)

**Каскадное удаление:**
При удалении вебхука автоматически удаляются все его запросы (`ON DELETE CASCADE`).

## Генерация токенов

### Алгоритм

```go
func generateWebhookTestToken() string {
    b := make([]byte, 16)      // 16 байт = 128 бит
    rand.Read(b)                // crypto/rand - криптографически безопасный
    return hex.EncodeToString(b) // 32 символа hex (0-9, a-f)
}
```

### Характеристики

- **Длина:** 32 символа (16 байт в hex)
- **Энтропия:** 128 бит
- **Уникальность:** Гарантируется индексом UNIQUE в БД
- **Безопасность:** Использует `crypto/rand` (не `math/rand`)
- **Коллизии:** Практически невозможны (2^128 вариантов)

### Пример токена
```
fe1c92bc3f27031bf00ec9c4f3057f7e
```

### Частота генерации

Токены генерируются **только при создании** нового тестового вебхука:
- ✅ Пользователь нажимает "Протестировать вебхук"
- ✅ Генерируется новый уникальный токен
- ✅ Создаётся запись в БД
- ✅ Токен не меняется в течение жизни вебхука

## Жизненный цикл

### 1. Создание
```
POST /api/webhook-test/create
→ Генерация токена
→ Сохранение в БД (expires_at = now + 24h)
→ Возврат URL: http://localhost:8080/webhook/test/{token}
```

### 2. Использование
```
ANY /webhook/test/{token}
→ Проверка существования и срока действия
→ Сохранение запроса в webhook_test_requests
→ Возврат успешного ответа
```

### 3. Истечение
```
Через 24 часа:
→ expires_at < now
→ Вебхук перестаёт принимать запросы
→ Возвращает 410 Gone
```

### 4. Удаление
```
DELETE /api/webhook-test/{token}
→ Удаление всех запросов (каскадно)
→ Удаление вебхука
```

### 5. Автоматическая очистка
```
Каждый час:
→ Поиск истекших вебхуков (expires_at < now)
→ Удаление вебхуков и их запросов
→ Логирование количества удалённых
```

## Автоматическая очистка

### Настройка

Очистка запускается автоматически:
- **При старте сервера** - удаляет все истекшие вебхуки
- **Каждый час** - периодическая очистка

### Код

```go
// В cmd/server/main.go
controllers.CleanupExpiredWebhooks()

go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        controllers.CleanupExpiredWebhooks()
    }
}()
```

### Логи

```
Cleaned up 5 expired test webhooks
```

### Ручная очистка

Можно вызвать очистку вручную через CLI:

```bash
# Подключиться к БД и выполнить
DELETE FROM webhook_tests WHERE expires_at < datetime('now');
```

Или через Go:
```go
controllers.CleanupExpiredWebhooks()
```

## Ограничения и квоты

### Текущие ограничения

- **Срок жизни:** 24 часа (фиксированный)
- **Запросов на вебхук:** 100 последних (в UI)
- **Запросов в БД:** Без ограничений (очищаются при удалении вебхука)
- **Вебхуков на пользователя:** Без ограничений

### Рекомендуемые квоты (для production)

```go
// Ограничение на количество активных вебхуков на пользователя
const MaxWebhooksPerUser = 5

// Ограничение на количество запросов на вебхук
const MaxRequestsPerWebhook = 1000

// Ограничение на размер тела запроса
const MaxRequestBodySize = 1 * 1024 * 1024 // 1 MB
```

## Мониторинг

### Метрики для отслеживания

1. **Количество активных вебхуков**
```sql
SELECT COUNT(*) FROM webhook_tests 
WHERE is_active = 1 AND expires_at > datetime('now');
```

2. **Количество запросов за последний час**
```sql
SELECT COUNT(*) FROM webhook_test_requests 
WHERE created_at > datetime('now', '-1 hour');
```

3. **Средний размер запросов**
```sql
SELECT AVG(LENGTH(body)) FROM webhook_test_requests;
```

4. **Топ пользователей по количеству вебхуков**
```sql
SELECT user_id, COUNT(*) as count 
FROM webhook_tests 
GROUP BY user_id 
ORDER BY count DESC 
LIMIT 10;
```

## Безопасность

### Защита от злоупотреблений

1. **Аутентификация:** Только авторизованные пользователи могут создавать вебхуки
2. **Изоляция:** Пользователь видит только свои вебхуки
3. **Автоматическое истечение:** Вебхуки удаляются через 24 часа
4. **Уникальные токены:** Невозможно угадать токен другого пользователя
5. **Каскадное удаление:** При удалении пользователя удаляются его вебхуки

### Рекомендации для production

1. **Rate limiting:** Ограничить количество создаваемых вебхуков в час
2. **Размер запросов:** Ограничить максимальный размер тела запроса
3. **Квоты:** Ограничить количество активных вебхуков на пользователя
4. **Мониторинг:** Отслеживать подозрительную активность
5. **HTTPS:** Использовать только HTTPS для production

## Производительность

### Оптимизация запросов

1. **Индексы:** Используются индексы на `token` и `webhook_test_id`
2. **Лимиты:** В UI показываются только последние 100 запросов
3. **Polling:** Интервал 5 секунд снижает нагрузку
4. **Каскадное удаление:** Эффективное удаление через БД

### Оценка нагрузки

При 100 активных пользователях:
- **Запросов в секунду:** ~20 (100 пользователей / 5 секунд)
- **Запросов в час:** ~72,000
- **Размер БД:** ~10-50 MB (зависит от размера запросов)

## Резервное копирование

### Экспорт данных

```bash
# SQLite
sqlite3 dmIntegroff.db ".dump webhook_tests" > webhook_tests_backup.sql
sqlite3 dmIntegroff.db ".dump webhook_test_requests" > webhook_test_requests_backup.sql

# MySQL
mysqldump -u user -p database webhook_tests > webhook_tests_backup.sql
mysqldump -u user -p database webhook_test_requests > webhook_test_requests_backup.sql
```

### Восстановление

```bash
# SQLite
sqlite3 dmIntegroff.db < webhook_tests_backup.sql

# MySQL
mysql -u user -p database < webhook_tests_backup.sql
```

## Troubleshooting

### Проблема: Вебхуки не удаляются автоматически

**Решение:**
1. Проверьте логи сервера на наличие ошибок
2. Убедитесь, что горутина очистки запущена
3. Проверьте права доступа к БД
4. Запустите очистку вручную

### Проблема: Слишком много запросов в БД

**Решение:**
1. Уменьшите срок жизни вебхуков
2. Добавьте ограничение на количество запросов
3. Увеличьте частоту очистки
4. Добавьте автоматическое удаление старых запросов

### Проблема: Медленные запросы

**Решение:**
1. Проверьте наличие индексов
2. Добавьте LIMIT в запросы
3. Оптимизируйте размер хранимых данных
4. Используйте кэширование

## API для управления

### Получить список всех вебхуков пользователя

```go
// Добавить в routes.go
authorized.GET("/api/webhook-tests", controllers.GetUserWebhookTests)

// Добавить в webhook_test_controller.go
func GetUserWebhookTests(c *gin.Context) {
    session := sessions.Default(c)
    userID := session.Get("user_id")
    
    var webhooks []models.WebhookTest
    database.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
        Order("created_at desc").
        Find(&webhooks)
    
    c.JSON(http.StatusOK, gin.H{"webhooks": webhooks})
}
```

### Получить статистику

```go
authorized.GET("/api/webhook-tests/stats", controllers.GetWebhookTestStats)

func GetWebhookTestStats(c *gin.Context) {
    session := sessions.Default(c)
    userID := session.Get("user_id")
    
    var activeCount int64
    database.DB.Model(&models.WebhookTest{}).
        Where("user_id = ? AND is_active = ? AND expires_at > ?", 
            userID, true, time.Now()).
        Count(&activeCount)
    
    var totalRequests int64
    database.DB.Model(&models.WebhookTestRequest{}).
        Joins("JOIN webhook_tests ON webhook_tests.id = webhook_test_requests.webhook_test_id").
        Where("webhook_tests.user_id = ?", userID).
        Count(&totalRequests)
    
    c.JSON(http.StatusOK, gin.H{
        "active_webhooks": activeCount,
        "total_requests": totalRequests,
    })
}
```
