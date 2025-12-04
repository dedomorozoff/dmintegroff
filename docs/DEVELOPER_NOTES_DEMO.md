# 👨‍ Заметки для разработчиков - Демо-режим

## Архитектура решения

### Принципы проектирования
1. **Минимальная инвазивность** - изменения затрагивают только необходимые части кода
2. **Обратная совместимость** - существующий функционал не нарушен
3. **Отключаемость** - функционал полностью отключается через `DEMO_MODE=false`
4. **Безопасность по умолчанию** - демо-режим выключен по умолчанию

### Компоненты

#### 1. Модель данных (`internal/models/user.go`)
```go
type User struct {
 // ... существующие поля
 IsDemo bool `gorm:"default:false"`
 ExpiresAt *time.Time `gorm:"index"`
}
```

**Решения:**
- `IsDemo` - флаг для быстрой идентификации
- `ExpiresAt` - nullable для обычных пользователей
- Индекс на `expires_at` для быстрой очистки

#### 2. Контроллер авторизации (`internal/controllers/auth_controller.go`)

**Функции:**
- `LoginPage()` - точка входа, обработка GET-параметров
- `RegisterDemoUser()` - создание/обновление демо-пользователя
- `CleanupExpiredDemoUsers()` - очистка истекших

**Логика:**
```go
if DEMO_MODE && demo_key == DEMO_SECRET {
 RegisterDemoUser(username, password)
 // Автозаполнение формы
}
```

#### 3. Сервис интеграций (`internal/services/integration_service.go`)

**Функции:**
- `ValidateDemoMode()` - проверка ограничений
- `ProcessWebhook()` - вызов валидации перед отправкой

**Логика:**
```go
if DEMO_MODE && user.IsDemo {
 if !strings.HasPrefix(target_api, DEMO_TARGET_URL) {
 return error
 }
}
```

## Потоки данных

### Регистрация
```
GET /login?demo_key=X&username=Y&password=Z
 ↓
LoginPage()
 ↓
CleanupExpiredDemoUsers() // Очистка старых
 ↓
Проверка DEMO_MODE && demo_key
 ↓
RegisterDemoUser()
 ↓
 ├─ Пользователь существует?
 │ ├─ Да → Обновить expires_at
 │ └─ Нет → Создать нового
 ↓
Рендер страницы с автозаполнением
```

### Отправка вебхука
```
POST /webhook/{token}
 ↓
WebhookHandler()
 ↓
ProcessWebhook()
 ↓
Preload("CreatedBy") // Загрузка пользователя
 ↓
ValidateDemoMode()
 ↓
 ├─ DEMO_MODE выключен → OK
 ├─ Пользователь не демо → OK
 └─ Пользователь демо
 ↓
 Проверка target_api
 ↓
 ├─ Начинается с DEMO_TARGET_URL → OK
 └─ Нет → ERROR
```

## Оптимизации

### 1. Индекс на expires_at
```sql
CREATE INDEX idx_users_expires_at ON users(expires_at);
```
**Причина:** Быстрая очистка истекших пользователей

### 2. Очистка при входе на /login
**Причина:** Не требует cron или фоновых задач

### 3. Preload("CreatedBy")
```go
database.DB.Preload("CreatedBy").First(&integration, integrationID)
```
**Причина:** Один запрос вместо двух (N+1 problem)

### 4. Nullable expires_at
**Причина:** Обычные пользователи не имеют срока действия

## Безопасность

### Защита от атак

#### 1. Brute-force регистрации
**Защита:** Требуется секретный ключ (`DEMO_SECRET`)

#### 2. Повышение привилегий
**Защита:** Роль всегда `specialist`, нельзя изменить

#### 3. Отправка на произвольные URL
**Защита:** Валидация `target_api` в `ValidateDemoMode()`

#### 4. Переполнение БД
**Защита:** Автоматическая очистка через 24 часа

### Рекомендации

1. **Сложный DEMO_SECRET**
 ```env
 DEMO_SECRET=$(openssl rand -hex 32)
 ```

2. **HTTPS в продакшене**
- GET-параметры видны в логах
- Используйте HTTPS для защиты

3. **Ограничение rate limit**
- Добавьте middleware для ограничения запросов
- Защита от спама регистраций

4. **Мониторинг**
- Логируйте создание демо-пользователей
- Отслеживайте количество активных демо-пользователей

## Расширения

### Возможные улучшения

#### 1. Настраиваемый срок действия
```go
// В .env
DEMO_EXPIRATION_HOURS=48

// В коде
expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)
```

#### 2. Множественные разрешенные URL
```go
// В .env
DEMO_TARGET_URLS=https://webhook.site/,https://example.com/

// В коде
allowedURLs := strings.Split(os.Getenv("DEMO_TARGET_URLS"), ",")
for _, url := range allowedURLs {
 if strings.HasPrefix(target_api, url) {
 return nil
 }
}
```

#### 3. Автоматическое продление
```go
// При каждом действии демо-пользователя
if user.IsDemo {
 user.ExpiresAt = time.Now().Add(24 * time.Hour)
 database.DB.Save(&user)
}
```

#### 4. Статистика использования
```go
type DemoStats struct {
 TotalCreated int
 ActiveNow int
 TotalWebhooks int
 AverageLifetime time.Duration
}
```

#### 5. Email уведомления
```go
// За час до истечения
if time.Until(*user.ExpiresAt) < time.Hour {
 SendExpirationEmail(user.Email)
}
```

## Тестирование

### Unit тесты

```go
func TestRegisterDemoUser(t *testing.T) {
 // Setup
 os.Setenv("DEMO_MODE", "true")
 os.Setenv("DEMO_SECRET", "test-secret")
 
 // Test
 err := RegisterDemoUser("test", "pass")
 assert.NoError(t, err)
 
 // Verify
 var user models.User
 database.DB.Where("username = ?", "test").First(&user)
 assert.True(t, user.IsDemo)
 assert.NotNil(t, user.ExpiresAt)
}

func TestValidateDemoMode(t *testing.T) {
 // Setup
 os.Setenv("DEMO_MODE", "true")
 os.Setenv("DEMO_TARGET_URL", "https://webhook.site/")
 
 integration := &models.Integration{
 TargetAPI: "https://webhook.site/abc-123",
 CreatedBy: models.User{IsDemo: true},
 }
 
 // Test
 err := ValidateDemoMode(integration)
 assert.NoError(t, err)
 
 // Test invalid URL
 integration.TargetAPI = "https://evil.com/"
 err = ValidateDemoMode(integration)
 assert.Error(t, err)
}
```

### Integration тесты

```go
func TestDemoUserFlow(t *testing.T) {
 // 1. Регистрация
 resp := httptest.NewRequest("GET", "/login?demo_key=secret&username=test&password=pass", nil)
 // Assert: форма заполнена
 
 // 2. Логин
 resp = httptest.NewRequest("POST", "/login", loginForm)
 // Assert: сессия создана
 
 // 3. Создание интеграции
 resp = httptest.NewRequest("POST", "/integrations", integrationForm)
 // Assert: интеграция создана
 
 // 4. Отправка вебхука
 resp = httptest.NewRequest("POST", "/webhook/token", webhookPayload)
 // Assert: вебхук отправлен
 
 // 5. Очистка
 time.Sleep(25 * time.Hour) // Симуляция
 CleanupExpiredDemoUsers()
 // Assert: пользователь удален
}
```

## Отладка

### Логирование

Добавьте логи для отладки:

```go
logger.Log.WithFields(map[string]interface{}{
 "username": username,
 "is_demo": true,
 "expires_at": expiresAt,
}).Info("Demo user registered")

logger.Log.WithFields(map[string]interface{}{
 "user_id": user.ID,
 "target_api": integration.TargetAPI,
 "demo_target": os.Getenv("DEMO_TARGET_URL"),
}).Warn("Demo mode validation failed")
```

### SQL запросы

Проверка демо-пользователей:
```sql
-- Активные демо-пользователи
SELECT username, expires_at, 
 TIMESTAMPDIFF(HOUR, NOW(), expires_at) as hours_left
FROM users 
WHERE is_demo = 1 AND expires_at > NOW();

-- Истекшие демо-пользователи
SELECT username, expires_at
FROM users 
WHERE is_demo = 1 AND expires_at < NOW();

-- Статистика
SELECT 
 COUNT(*) as total_demo_users,
 COUNT(CASE WHEN expires_at > NOW() THEN 1 END) as active,
 COUNT(CASE WHEN expires_at < NOW() THEN 1 END) as expired
FROM users 
WHERE is_demo = 1;
```

## Производительность

### Метрики

- **Время регистрации**: ~50ms (включая bcrypt)
- **Время очистки**: ~10ms (с индексом)
- **Время валидации**: ~1ms (проверка строки)

### Нагрузочное тестирование

```bash
# 100 одновременных регистраций
ab -n 100 -c 10 "http://localhost:8080/login?demo_key=secret&username=test&password=pass"

# 1000 вебхуков
ab -n 1000 -c 50 -p webhook.json -T application/json "http://localhost:8080/webhook/token"
```

## Миграция

### Применение

```bash
# Автоматически (рекомендуется)
go run cmd/admin/main.go

# Вручную
sqlite3 dmintegroff.db < migrations/006_add_demo_users_sqlite.sql
```

### Откат

```sql
-- SQLite
ALTER TABLE users DROP COLUMN is_demo;
ALTER TABLE users DROP COLUMN expires_at;
DROP INDEX idx_users_expires_at;

-- MySQL
ALTER TABLE users DROP COLUMN is_demo;
ALTER TABLE users DROP COLUMN expires_at;
DROP INDEX idx_users_expires_at ON users;
```

## Чеклист для code review

- [ ] Проверена безопасность (секретный ключ, валидация URL)
- [ ] Добавлены индексы для производительности
- [ ] Обратная совместимость сохранена
- [ ] Документация обновлена
- [ ] Миграции созданы для всех БД
- [ ] Логирование добавлено
- [ ] Тесты написаны
- [ ] Код следует стилю проекта
- [ ] Нет hardcoded значений
- [ ] Ошибки обрабатываются корректно

## Контакты

При возникновении вопросов или проблем:
- Создайте issue в репозитории
- Проверьте FAQ: `docs/DEMO_MODE_FAQ.md`
- Изучите схемы: `docs/DEMO_MODE_FLOW.md`
