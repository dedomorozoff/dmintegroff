# Настройка за прокси (Nginx, Apache, Landing)

## Проблема

Когда dmIntegroff работает за прокси (например, через лендинг), генерируемые URL вебхуков должны использовать внешний адрес, а не внутренний.

## Решение

Используйте переменную окружения `BASE_URL` для указания внешнего адреса.

## Конфигурация

### 1. Для лендинга

В `.env` основного приложения:

```bash
# Внешний URL (как видят пользователи)
BASE_URL=https://yourdomain.com

# Внутренний адрес для прокси
HOST=127.0.0.1
PORT=8080
```

В лендинге прокси настроен на `http://localhost:8080`, но пользователи видят `https://yourdomain.com`.

### 2. Для Nginx

**nginx.conf:**
```nginx
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**dmIntegroff .env:**
```bash
BASE_URL=https://yourdomain.com
HOST=127.0.0.1
PORT=8080
```

### 3. Для Apache

**apache.conf:**
```apache
<VirtualHost *:80>
    ServerName yourdomain.com
    
    ProxyPreserveHost On
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/
    
    RequestHeader set X-Forwarded-Proto "https"
    RequestHeader set X-Forwarded-Port "443"
</VirtualHost>
```

**dmIntegroff .env:**
```bash
BASE_URL=https://yourdomain.com
HOST=127.0.0.1
PORT=8080
```

### 4. Для Caddy

**Caddyfile:**
```
yourdomain.com {
    reverse_proxy localhost:8080
}
```

**dmIntegroff .env:**
```bash
BASE_URL=https://yourdomain.com
HOST=127.0.0.1
PORT=8080
```

## Как работают вебхуки за прокси

### Схема работы

```
Пользователь → https://yourdomain.com/webhook/test/abc123
                ↓
            Прокси (Nginx/Landing)
                ↓
            http://localhost:8080/webhook/test/abc123
                ↓
            dmIntegroff обрабатывает запрос
```

### Генерация URL

Когда создаётся тестовый вебхук:

```go
baseURL := os.Getenv("BASE_URL")
if baseURL == "" {
    baseURL = "http://localhost:8080"  // fallback
}
webhookURL := fmt.Sprintf("%s/webhook/test/%s", baseURL, token)
```

**Результат:**
- С `BASE_URL=https://yourdomain.com` → `https://yourdomain.com/webhook/test/abc123`
- Без `BASE_URL` → `http://localhost:8080/webhook/test/abc123`

## Проверка настройки

### 1. Проверьте BASE_URL

```bash
# В .env должно быть
BASE_URL=https://yourdomain.com
```

### 2. Перезапустите сервер

```bash
go run cmd/server/main.go
```

### 3. Создайте тестовый вебхук

Нажмите "Протестировать вебхук" на дашборде.

### 4. Проверьте URL

URL должен начинаться с `https://yourdomain.com`, а не `http://localhost:8080`.

### 5. Отправьте тестовый запрос

```bash
curl -X POST https://yourdomain.com/webhook/test/YOUR_TOKEN \
  -H "Content-Type: application/json" \
  -d '{"test": true}'
```

## Получение реального IP за прокси

### Проблема

За прокси `c.ClientIP()` возвращает IP прокси (127.0.0.1), а не реальный IP клиента.

### Решение

Gin автоматически использует заголовки `X-Real-IP` и `X-Forwarded-For` если настроены доверенные прокси.

**В main.go уже настроено:**
```go
r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
```

### Для production

Добавьте IP вашего прокси:

```go
r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.1"})
```

Или отключите проверку (не рекомендуется):
```go
r.SetTrustedProxies(nil)
```

## SSL/TLS (HTTPS)

### Вариант 1: SSL на прокси (рекомендуется)

Прокси (Nginx/Caddy) обрабатывает SSL, dmIntegroff работает по HTTP.

**Преимущества:**
- ✅ Проще настройка
- ✅ Централизованное управление сертификатами
- ✅ Можно использовать Let's Encrypt

**Настройка:**
```bash
# dmIntegroff работает по HTTP
BASE_URL=https://yourdomain.com  # Внешний HTTPS
HOST=127.0.0.1
PORT=8080
```

### Вариант 2: SSL в dmIntegroff

dmIntegroff сам обрабатывает HTTPS.

**Настройка:**
```bash
BASE_URL=https://yourdomain.com
HOST=0.0.0.0
PORT=443
SSL_CERT=/path/to/cert.pem
SSL_KEY=/path/to/key.pem
```

**Код (нужно добавить в main.go):**
```go
certFile := os.Getenv("SSL_CERT")
keyFile := os.Getenv("SSL_KEY")

if certFile != "" && keyFile != "" {
    r.RunTLS(addr, certFile, keyFile)
} else {
    r.Run(addr)
}
```

## Пример с лендингом

### Структура

```
dmintegroff-landing (порт 3000)
    ↓ проксирует на
dmintegroff (порт 8080)
```

### Конфигурация лендинга

```bash
# dmintegroff-landing/.env
PORT=3000
TARGET_URL=http://localhost:8080
APP_URL=/
```

### Конфигурация dmIntegroff

```bash
# dmintegroff/.env
BASE_URL=http://localhost:3000  # URL лендинга
HOST=127.0.0.1
PORT=8080
```

### Результат

Пользователь открывает `http://localhost:3000` (лендинг), нажимает "Попробовать демо", попадает в dmIntegroff, создаёт вебхук с URL `http://localhost:3000/webhook/test/abc123`.

## Troubleshooting

### Вебхук возвращает 404

**Причина:** Прокси не проксирует путь `/webhook/test/*`

**Решение:** Проверьте конфигурацию прокси, должно быть `proxy_pass http://localhost:8080;` без дополнительных путей.

### URL генерируется с localhost

**Причина:** `BASE_URL` не установлен или неправильный

**Решение:** 
```bash
# В .env
BASE_URL=https://yourdomain.com
```

### Неправильный IP клиента

**Причина:** Прокси не передаёт заголовки `X-Real-IP` или `X-Forwarded-For`

**Решение:** Добавьте в конфигурацию прокси:
```nginx
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
```

### Вебхук работает, но UI не открывается

**Причина:** Прокси не проксирует статические файлы

**Решение:** Убедитесь, что прокси проксирует все пути:
```nginx
location / {
    proxy_pass http://localhost:8080;
}
```

## Рекомендации для production

1. **Используйте HTTPS** - всегда в production
2. **SSL на прокси** - проще управлять сертификатами
3. **Правильный BASE_URL** - внешний адрес, который видят пользователи
4. **Доверенные прокси** - настройте `SetTrustedProxies`
5. **Мониторинг** - следите за логами прокси и dmIntegroff
6. **Rate limiting** - настройте на уровне прокси
7. **Firewall** - dmIntegroff должен быть доступен только для прокси

## Примеры конфигураций

### Development (без прокси)

```bash
BASE_URL=http://localhost:8080
HOST=0.0.0.0
PORT=8080
```

### Staging (с Nginx)

```bash
BASE_URL=https://staging.yourdomain.com
HOST=127.0.0.1
PORT=8080
```

### Production (с Nginx + SSL)

```bash
BASE_URL=https://yourdomain.com
HOST=127.0.0.1
PORT=8080
GIN_MODE=release
```

### Production (с лендингом)

```bash
# Лендинг на порту 443 (HTTPS)
BASE_URL=https://yourdomain.com
HOST=127.0.0.1
PORT=8080
GIN_MODE=release
```
