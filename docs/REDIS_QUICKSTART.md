# Redis Quick Start для dmIntegroff

## Зачем Redis?

Redis ускоряет работу тестовых вебхуков в **5-10 раз** и автоматически удаляет истекшие данные.

## Быстрая установка (3 минуты)

### 1️⃣ Запустите Redis

**Docker (рекомендуется):**
```bash
docker run -d --name redis -p 6379:6379 redis:alpine
```

**Или локально:**
```bash
# Windows
choco install redis-64

# Linux
sudo apt install redis-server && sudo systemctl start redis

# macOS
brew install redis && brew services start redis
```

### 2️⃣ Установите зависимости

```bash
cd dmintegroff
go get github.com/redis/go-redis/v9
go get github.com/google/uuid
go mod tidy
```

### 3️⃣ Настройте .env

```bash
# Добавьте эту строку в .env
REDIS_URL=redis://localhost:6379/0
```

### 4️⃣ Перезапустите сервер

```bash
go run cmd/server/main.go
```

## ✅ Проверка

В логах должно быть:
```
Redis connected successfully
```

Проверьте Redis:
```bash
redis-cli ping
# Ответ: PONG
```

## 🎉 Готово!

Теперь тестовые вебхуки используют Redis:
- ⚡ Быстрее в 5-10 раз
- 🔄 Автоматическое удаление через 24 часа
- 📊 Меньше нагрузки на БД

## 🔄 Отключение Redis

Просто закомментируйте в .env:
```bash
#REDIS_URL=redis://localhost:6379/0
```

Система автоматически вернётся к использованию БД.

## 📚 Подробная документация

- [Полная настройка Redis](REDIS_SETUP.md)
- [Установка и конфигурация](REDIS_INSTALL.md)
- [Управление вебхуками](WEBHOOK_TEST_MANAGEMENT.md)

## ❓ Проблемы?

**Redis не подключается:**
```bash
# Проверьте, запущен ли Redis
redis-cli ping

# Проверьте порт
netstat -an | grep 6379
```

**Ошибка "Connection refused":**
- Убедитесь, что Redis запущен
- Проверьте REDIS_URL в .env
- Проверьте firewall

## 💡 Без Redis

Если не хотите использовать Redis - ничего не делайте!  
Всё будет работать через базу данных как раньше.
