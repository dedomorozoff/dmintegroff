# Redis Setup для тестовых вебхуков

## Зачем Redis?

Redis идеально подходит для хранения временных данных тестовых вебхуков:

✅ **Автоматическое истечение (TTL)** - не нужна ручная очистка  
✅ **Высокая производительность** - быстрее чем БД  
✅ **Меньше нагрузки на БД** - разгружает основную базу  
✅ **Простое масштабирование** - легко добавить Redis кластер  
✅ **Встроенные структуры данных** - списки для запросов  

## Установка Redis

### Windows

```powershell
# Через Chocolatey
choco install redis-64

# Или скачать с GitHub
# https://github.com/microsoftarchive/redis/releases
```

### Linux (Ubuntu/Debian)

```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

### macOS

```bash
brew install redis
brew services start redis
```

### Docker

```bash
docker run -d --name redis -p 6379:6379 redis:alpine
```

### Docker Compose

```yaml
version: '3.8'
services:
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

volumes:
  redis_data:
```

## Конфигурация

### 1. Добавьте в .env

```bash
# Local Redis
REDIS_URL=redis://localhost:6379/0

# С паролем
REDIS_URL=redis://:your_password@localhost:6379/0

# Remote Redis
REDIS_URL=redis://user:pass@redis.example.com:6379/0

# Redis Cloud
REDIS_URL=redis://default:password@redis-12345.cloud.redislabs.com:12345
```

### 2. Перезапустите сервер

```bash
go run cmd/server/main.go
```

### 3. Проверьте логи

```
Redis connected successfully
```

Или если Redis не настроен:
```
Redis not configured, using database for webhook tests
```

## Проверка работы

### 1. Проверка подключения

```bash
redis-cli ping
# Ответ: PONG
```

### 2. Мониторинг в реальном времени

```bash
redis-cli monitor
```

### 3. Просмотр ключей

```bash
redis-cli keys "webhook_*"
```

### 4. Просмотр данных вебхука

```bash
redis-cli get "webhook_test:YOUR_TOKEN"
```

### 5. Просмотр запросов

```bash
redis-cli lrange "webhook_requests:YOUR_TOKEN" 0 -1
```

## Структура данных в Redis

### Ключи

```
webhook_test:{token}           - Данные вебхука (String, TTL 24h)
webhook_requests:{token}       - Список запросов (List, TTL 24h)
```

### Пример данных вебхука

```json
{
  "token": "fe1c92bc3f27031bf00ec9c4f3057f7e",
  "user_id": 1,
  "expires_at": "2024-12-07T12:00:00Z",
  "is_active": true,
  "created_at": "2024-12-06T12:00:00Z"
}
```

### Пример данных запроса

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "url": "/webhook/test/fe1c92bc3f27031bf00ec9c4f3057f7e",
  "headers": "{\"Content-Type\":\"application/json\"}",
  "body": "{\"test\":true}",
  "query_params": "{}",
  "client_ip": "127.0.0.1",
  "created_at": "2024-12-06T12:00:00Z"
}
```

## Fallback на БД

Если Redis недоступен, система автоматически использует базу данных:

```go
// Попытка использовать Redis
if cache.IsRedisAvailable() {
    // Сохранить в Redis
} else {
    // Fallback на БД
}
```

Это обеспечивает:
- ✅ Работу без Redis
- ✅ Плавную деградацию
- ✅ Отсутствие ошибок при отключении Redis

## Производительность

### Сравнение Redis vs БД

| Операция | Redis | SQLite | MySQL |
|----------|-------|--------|-------|
| Создание вебхука | ~1ms | ~5ms | ~3ms |
| Сохранение запроса | ~1ms | ~10ms | ~5ms |
| Получение запросов | ~2ms | ~20ms | ~10ms |
| Удаление | ~1ms | ~15ms | ~8ms |

### Нагрузка

При 100 активных пользователях (polling каждые 5 сек):

**Redis:**
- Запросов в секунду: ~20
- Использование памяти: ~10-50 MB
- CPU: <1%

**БД:**
- Запросов в секунду: ~20
- Использование диска: ~10-50 MB
- CPU: ~5-10%

## Мониторинг Redis

### Redis CLI

```bash
# Информация о сервере
redis-cli info

# Статистика памяти
redis-cli info memory

# Количество ключей
redis-cli dbsize

# Мониторинг команд
redis-cli monitor
```

### Redis Commander (GUI)

```bash
npm install -g redis-commander
redis-commander
# Открыть http://localhost:8081
```

### RedisInsight (официальный GUI)

Скачать с https://redis.com/redis-enterprise/redis-insight/

## Безопасность

### 1. Установите пароль

```bash
# В redis.conf
requirepass your_strong_password

# Или через CLI
redis-cli config set requirepass your_strong_password
```

### 2. Ограничьте доступ

```bash
# В redis.conf
bind 127.0.0.1 ::1  # Только localhost
```

### 3. Отключите опасные команды

```bash
# В redis.conf
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG ""
```

### 4. Используйте TLS

```bash
# В .env
REDIS_URL=rediss://user:pass@host:6379/0  # rediss:// для TLS
```

## Резервное копирование

### RDB (снимки)

```bash
# В redis.conf
save 900 1      # Сохранять каждые 15 минут если был хотя бы 1 ключ
save 300 10     # Сохранять каждые 5 минут если было 10 изменений
save 60 10000   # Сохранять каждую минуту если было 10000 изменений
```

### AOF (журнал)

```bash
# В redis.conf
appendonly yes
appendfsync everysec
```

### Ручное сохранение

```bash
redis-cli save      # Синхронное
redis-cli bgsave    # Асинхронное
```

## Troubleshooting

### Redis не подключается

```bash
# Проверьте, запущен ли Redis
redis-cli ping

# Проверьте порт
netstat -an | grep 6379

# Проверьте логи
tail -f /var/log/redis/redis-server.log
```

### Ошибка "Connection refused"

1. Проверьте, что Redis запущен
2. Проверьте REDIS_URL в .env
3. Проверьте firewall
4. Проверьте bind в redis.conf

### Ошибка "NOAUTH Authentication required"

Добавьте пароль в REDIS_URL:
```bash
REDIS_URL=redis://:your_password@localhost:6379/0
```

### Медленная работа

1. Проверьте память: `redis-cli info memory`
2. Проверьте медленные команды: `redis-cli slowlog get 10`
3. Увеличьте maxmemory в redis.conf
4. Настройте eviction policy

## Production рекомендации

### 1. Используйте Redis Sentinel

Для высокой доступности:

```bash
# sentinel.conf
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 10000
```

### 2. Используйте Redis Cluster

Для масштабирования:

```bash
redis-cli --cluster create \
  127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 \
  --cluster-replicas 1
```

### 3. Настройте мониторинг

- Prometheus + Redis Exporter
- Grafana дашборды
- Alerting на критичные метрики

### 4. Настройте лимиты

```bash
# В redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
maxclients 10000
```

## Миграция с БД на Redis

### 1. Включите Redis

```bash
REDIS_URL=redis://localhost:6379/0
```

### 2. Перезапустите сервер

Новые вебхуки будут создаваться в Redis.

### 3. Старые вебхуки

Старые вебхуки в БД продолжат работать до истечения (24 часа).

### 4. Очистка БД (опционально)

После 24 часов можно очистить старые данные:

```sql
DELETE FROM webhook_tests WHERE expires_at < datetime('now');
DELETE FROM webhook_test_requests WHERE webhook_test_id NOT IN (SELECT id FROM webhook_tests);
```

## Альтернативы Redis

Если Redis недоступен, можно использовать:

1. **Memcached** - похож на Redis, но проще
2. **KeyDB** - форк Redis с улучшенной производительностью
3. **DragonflyDB** - современная альтернатива Redis
4. **База данных** - fallback уже реализован

## Полезные ссылки

- [Redis Documentation](https://redis.io/documentation)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)
- [Redis Security](https://redis.io/docs/manual/security/)
- [Redis Persistence](https://redis.io/docs/manual/persistence/)
- [go-redis Documentation](https://redis.uptrace.dev/)
