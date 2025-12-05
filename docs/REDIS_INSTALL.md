# Установка Redis поддержки

## Шаг 1: Установите зависимость

```bash
go get github.com/redis/go-redis/v9
go get github.com/google/uuid
go mod tidy
```

## Шаг 2: Установите Redis сервер

### Вариант 1: Docker (рекомендуется)

```bash
docker run -d --name redis -p 6379:6379 redis:alpine
```

### Вариант 2: Локальная установка

**Windows:**
```powershell
choco install redis-64
```

**Linux:**
```bash
sudo apt install redis-server
sudo systemctl start redis
```

**macOS:**
```bash
brew install redis
brew services start redis
```

## Шаг 3: Настройте .env

```bash
# Добавьте в .env
REDIS_URL=redis://localhost:6379/0
```

## Шаг 4: Перезапустите сервер

```bash
go run cmd/server/main.go
```

## Проверка

Вы должны увидеть в логах:
```
Redis connected successfully
```

Если Redis не настроен:
```
Redis not configured, using database for webhook tests
```

## Тестирование

1. Создайте тестовый вебхук через UI
2. Проверьте Redis:

```bash
redis-cli keys "webhook_*"
redis-cli get "webhook_test:YOUR_TOKEN"
```

## Отключение Redis

Просто закомментируйте или удалите `REDIS_URL` из .env:

```bash
#REDIS_URL=redis://localhost:6379/0
```

Система автоматически вернётся к использованию базы данных.

## Преимущества Redis

- ✅ Автоматическое истечение (TTL)
- ✅ Быстрее чем БД (в 5-10 раз)
- ✅ Меньше нагрузки на основную БД
- ✅ Простое масштабирование
- ✅ Встроенные структуры данных

## Без Redis

Если вы не хотите использовать Redis, ничего делать не нужно!  
Система будет работать с базой данных (SQLite/MySQL) как раньше.
