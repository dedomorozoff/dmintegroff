# Настройка с лендингом

## Быстрая настройка (2 минуты)

### 1️⃣ Запустите dmIntegroff

```bash
cd dmintegroff

# Настройте .env
echo "BASE_URL=http://localhost:3000" >> .env
echo "HOST=127.0.0.1" >> .env
echo "PORT=8080" >> .env

# Запустите
go run cmd/server/main.go
```

### 2️⃣ Запустите лендинг

```bash
cd dmintegroff-landing

# Настройте .env
echo "PORT=3000" >> .env
echo "TARGET_URL=http://localhost:8080" >> .env

# Запустите
go run main.go
```

### 3️⃣ Откройте браузер

```
http://localhost:3000
```

## ✅ Проверка

1. Откройте лендинг: `http://localhost:3000`
2. Нажмите "Попробовать демо"
3. Войдите в систему
4. Создайте тестовый вебхук
5. URL должен быть: `http://localhost:3000/webhook/test/...`

## 🎯 Как это работает

```
Пользователь → http://localhost:3000 (лендинг)
                    ↓
                Прокси
                    ↓
                http://localhost:8080 (dmIntegroff)
```

### Генерация URL

dmIntegroff использует `BASE_URL` из .env:

```go
baseURL := os.Getenv("BASE_URL")  // http://localhost:3000
webhookURL := fmt.Sprintf("%s/webhook/test/%s", baseURL, token)
// Результат: http://localhost:3000/webhook/test/abc123
```

### Обработка запросов

1. Запрос приходит на `http://localhost:3000/webhook/test/abc123`
2. Лендинг проксирует на `http://localhost:8080/webhook/test/abc123`
3. dmIntegroff обрабатывает запрос
4. Ответ возвращается через лендинг

## 🚀 Production настройка

### С доменом

```bash
# dmintegroff/.env
BASE_URL=https://yourdomain.com
HOST=127.0.0.1
PORT=8080
GIN_MODE=release

# dmintegroff-landing/.env
PORT=443
TARGET_URL=http://localhost:8080
SSL_CERT=/path/to/cert.pem
SSL_KEY=/path/to/key.pem
```

### С поддоменом

```bash
# dmintegroff/.env
BASE_URL=https://app.yourdomain.com
HOST=127.0.0.1
PORT=8080

# dmintegroff-landing/.env (на app.yourdomain.com)
PORT=443
TARGET_URL=http://localhost:8080
```

## ❓ Troubleshooting

### Вебхук возвращает 404

Проверьте, что лендинг проксирует все пути:

```go
// В main.go лендинга должно быть
proxy.ServeHTTP(w, r)  // Проксирует все запросы
```

### URL генерируется неправильно

Проверьте `BASE_URL` в dmIntegroff/.env:

```bash
# Должно быть
BASE_URL=http://localhost:3000  # Порт лендинга!

# А не
BASE_URL=http://localhost:8080  # Это неправильно
```

### Лендинг не запускается

Проверьте, что порты свободны:

```bash
# Проверить порт 3000
netstat -an | grep 3000

# Проверить порт 8080
netstat -an | grep 8080
```

## 📚 Подробная документация

- [Настройка прокси](docs/PROXY_SETUP.md)
- [Тестовые вебхуки](docs/WEBHOOK_TEST.md)
- [Демо режим](docs/DEMO_MODE.md)
