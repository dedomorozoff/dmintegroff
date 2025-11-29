# ⚙️ Руководство по настройке dmIntegroff

## Содержание

1. [Конфигурационный файл .env](#конфигурационный-файл-env)
2. [Настройки сервера](#настройки-сервера)
3. [Настройки базы данных](#настройки-базы-данных)
4. [Настройки безопасности](#настройки-безопасности)
5. [Настройки логирования](#настройки-логирования)
6. [Настройки разработки](#настройки-разработки)
7. [Настройки пользователя](#настройки-пользователя)
8. [Production конфигурация](#production-конфигурация)

---

## Конфигурационный файл .env

Все настройки системы хранятся в файле `.env` в корне проекта. Если файл отсутствует, создайте его на основе `.env.example`:

```bash
cp .env.example .env
```

⚠️ **Важно**: После изменения настроек в `.env` необходимо перезапустить сервер!

---

## Настройки сервера

### PORT
Порт, на котором будет работать веб-сервер.

```env
PORT=8080
```

**Значения:**
- По умолчанию: `8080`
- Диапазон: `1-65535`
- Рекомендуется: `8080` для разработки, `80` или `443` для production

**Примеры:**
```env
PORT=8080          # Разработка
PORT=3000          # Альтернативный порт
PORT=80            # Production (требует root/admin)
```

### GIN_MODE
Режим работы веб-фреймворка Gin.

```env
GIN_MODE=debug
```

**Значения:**
- `debug` - режим отладки (подробные логи, stack traces)
- `release` - продакшн режим (минимальные логи, оптимизация)

**Рекомендации:**
- Используйте `debug` для разработки
- Используйте `release` для production

### APP_PATH
Базовый путь для всех маршрутов приложения.

```env
APP_PATH=
```

**Значения:**
- Пустая строка - приложение доступно в корне (по умолчанию)
- `/api/v1` - приложение доступно по пути /api/v1
- `/dmintegroff` - приложение доступно по пути /dmintegroff

**Примеры:**
```env
APP_PATH=              # http://localhost:8080/
APP_PATH=/api/v1       # http://localhost:8080/api/v1/
APP_PATH=/dmintegroff  # http://localhost:8080/dmintegroff/
```

⚠️ **Важно**: При изменении APP_PATH обновите webhook URL в интеграциях!

---

## Настройки базы данных

### DB_TYPE
Тип используемой базы данных.

```env
DB_TYPE=sqlite
```

**Значения:**
- `sqlite` - SQLite (по умолчанию, файловая БД)
- `mysql` - MySQL/MariaDB (требует отдельного сервера)

**Рекомендации:**
- SQLite для разработки и небольших проектов
- MySQL для production и высоких нагрузок

### DB_DSN
Строка подключения к базе данных (Data Source Name).

**Для SQLite:**
```env
DB_DSN=dmintegroff.db
```

Указывается путь к файлу базы данных:
- `dmintegroff.db` - файл в корне проекта
- `/var/lib/dmintegroff/data.db` - абсолютный путь
- `./data/dmintegroff.db` - относительный путь

**Для MySQL:**
```env
DB_DSN=user:password@tcp(localhost:3306)/dmintegroff?charset=utf8mb4&parseTime=True&loc=Local
```

Формат: `username:password@tcp(host:port)/database?parameters`

**Примеры MySQL:**
```env
# Локальный MySQL
DB_DSN=root:password@tcp(localhost:3306)/dmintegroff?charset=utf8mb4&parseTime=True&loc=Local

# Удаленный MySQL
DB_DSN=dmuser:secret@tcp(192.168.1.100:3306)/dmintegroff?charset=utf8mb4&parseTime=True&loc=Local

# MySQL с SSL
DB_DSN=user:pass@tcp(host:3306)/db?tls=true&charset=utf8mb4&parseTime=True&loc=Local
```

---

## Настройки безопасности

### SESSION_SECRET
Секретный ключ для шифрования сессий пользователей.

```env
SESSION_SECRET=your-super-secret-key-change-this-in-production
```

⚠️ **КРИТИЧЕСКИ ВАЖНО**: Обязательно измените это значение в production!

**Требования:**
- Минимум 32 символа
- Случайная строка
- Уникальная для каждой установки

**Генерация безопасного ключа:**

```bash
# Linux/Mac
openssl rand -base64 32

# Windows PowerShell
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))

# Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

**Пример:**
```env
SESSION_SECRET=Kx9mP2vN8qR5tY7wZ3aB6cD1eF4gH0jL
```

---

## Настройки логирования

### LOG_LEVEL
Минимальный уровень логирования.

```env
LOG_LEVEL=info
```

**Значения (от наименее к наиболее подробному):**
- `error` - только ошибки
- `warn` - предупреждения и ошибки
- `info` - информационные сообщения (рекомендуется)
- `debug` - отладочная информация (очень подробно)

**Рекомендации:**
- `info` для production
- `debug` для разработки и отладки
- `error` для минимального логирования

### LOG_FILE
Путь к файлу логов.

```env
LOG_FILE=dmintegroff.log
```

**Значения:**
- Имя файла - создается в корне проекта
- Абсолютный путь - `/var/log/dmintegroff/app.log`
- Пустая строка - логи только в консоль

**Примеры:**
```env
LOG_FILE=dmintegroff.log                    # Файл в корне
LOG_FILE=./logs/app.log                     # Относительный путь
LOG_FILE=/var/log/dmintegroff/app.log       # Абсолютный путь
LOG_FILE=                                   # Только консоль
```

**Ротация логов:**

Для автоматической ротации логов используйте logrotate (Linux):

```bash
# /etc/logrotate.d/dmintegroff
/var/log/dmintegroff/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 dmintegroff dmintegroff
}
```

---

## Настройки разработки

### DEBUG
Включение режима отладки.

```env
DEBUG=true
```

**Значения:**
- `true` - режим отладки включен
- `false` - режим отладки выключен

**Что включает режим отладки:**
- Подробные сообщения об ошибках
- Stack traces в логах
- Дополнительная информация в консоли
- Отключение некоторых оптимизаций

**Рекомендации:**
- `true` для разработки
- `false` для production

---

## Настройки пользователя

Настройки пользователя управляются через веб-интерфейс в разделе **"Настройки"**.

### Смена пароля

1. Перейдите в **Настройки** → **Смена пароля**
2. Введите текущий пароль
3. Введите новый пароль (минимум 6 символов)
4. Подтвердите новый пароль
5. Нажмите **"Изменить пароль"**

**Требования к паролю:**
- Минимум 6 символов
- Рекомендуется использовать буквы, цифры и спецсимволы
- Пароли должны совпадать

### Сброс пароля через CLI

Если вы забыли пароль, используйте CLI инструмент:

```bash
# Сброс пароля пользователя
go run cmd/cli/main.go reset-password admin newpassword123
```

---

## Production конфигурация

### Минимальная безопасная конфигурация

```env
# Server
PORT=8080
GIN_MODE=release
APP_PATH=

# Database
DB_TYPE=mysql
DB_DSN=dmuser:SecurePassword123@tcp(localhost:3306)/dmintegroff?charset=utf8mb4&parseTime=True&loc=Local

# Security
SESSION_SECRET=Kx9mP2vN8qR5tY7wZ3aB6cD1eF4gH0jL9mN8pQ2rS5tV7wX0yZ3aB6cD1eF4gH

# Logging
LOG_LEVEL=info
LOG_FILE=/var/log/dmintegroff/app.log

# Development
DEBUG=false
```

### Рекомендации для production

1. **Безопасность:**
   - Используйте сложный SESSION_SECRET (минимум 32 символа)
   - Используйте MySQL вместо SQLite
   - Создайте отдельного пользователя БД с ограниченными правами
   - Используйте HTTPS (настройте reverse proxy)

2. **Производительность:**
   - Установите GIN_MODE=release
   - Установите DEBUG=false
   - Используйте LOG_LEVEL=info или warn
   - Настройте ротацию логов

3. **Мониторинг:**
   - Настройте логирование в файл
   - Используйте logrotate для ротации логов
   - Мониторьте размер базы данных
   - Настройте бэкапы БД

4. **Сеть:**
   - Используйте reverse proxy (nginx, Apache)
   - Настройте SSL/TLS сертификаты
   - Ограничьте доступ к порту через firewall
   - Используйте CDN для статических файлов (опционально)

### Пример nginx конфигурации

```nginx
server {
    listen 80;
    server_name dmintegroff.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name dmintegroff.example.com;

    ssl_certificate /etc/ssl/certs/dmintegroff.crt;
    ssl_certificate_key /etc/ssl/private/dmintegroff.key;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Systemd service

Создайте файл `/etc/systemd/system/dmintegroff.service`:

```ini
[Unit]
Description=dmIntegroff Integration Service
After=network.target mysql.service

[Service]
Type=simple
User=dmintegroff
Group=dmintegroff
WorkingDirectory=/opt/dmintegroff
ExecStart=/opt/dmintegroff/dmintegroff
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Активация:
```bash
sudo systemctl daemon-reload
sudo systemctl enable dmintegroff
sudo systemctl start dmintegroff
```

---

## Проверка конфигурации

### Проверка подключения к БД

```bash
# SQLite
ls -lh dmintegroff.db

# MySQL
mysql -u dmuser -p -h localhost dmintegroff -e "SELECT 1"
```

### Проверка логов

```bash
# Просмотр последних логов
tail -f dmintegroff.log

# Поиск ошибок
grep ERROR dmintegroff.log

# Статистика по уровням
grep -c INFO dmintegroff.log
grep -c ERROR dmintegroff.log
```

### Проверка порта

```bash
# Linux/Mac
netstat -tuln | grep 8080
lsof -i :8080

# Windows
netstat -ano | findstr :8080
```

---

## Устранение проблем

### Сервер не запускается

1. Проверьте, что порт свободен
2. Проверьте права доступа к файлам
3. Проверьте подключение к БД
4. Проверьте логи на наличие ошибок

### Ошибки подключения к БД

1. Проверьте DB_DSN
2. Убедитесь, что БД существует
3. Проверьте права пользователя БД
4. Для MySQL: проверьте, что сервер запущен

### Проблемы с сессиями

1. Проверьте SESSION_SECRET
2. Очистите cookies в браузере
3. Перезапустите сервер

---

## Дополнительные ресурсы

- [Документация Gin](https://gin-gonic.com/docs/)
- [Документация GORM](https://gorm.io/docs/)
- [Документация SQLite](https://www.sqlite.org/docs.html)
- [Документация MySQL](https://dev.mysql.com/doc/)

