# 🚀 Production Deployment Guide

## ⚠️ Обязательные изменения для production

### 1. Переменные окружения (.env)

```env
# ОБЯЗАТЕЛЬНО измените!
SESSION_SECRET=ваш-очень-длинный-случайный-секретный-ключ-минимум-32-символа

# Режим Gin
GIN_MODE=release

# База данных (рекомендуется MySQL)
DB_TYPE=mysql
DB_DSN=user:password@tcp(localhost:3306)/dmIntegroff?charset=utf8mb4&parseTime=True&loc=Local

# Логирование
LOG_LEVEL=info
LOG_FILE=/var/log/dmIntegroff/dmIntegroff.log

# Порт (опционально)
PORT=8080
```

### 2. Генерация SESSION_SECRET

```bash
# Linux/macOS
openssl rand -base64 32

# Windows PowerShell
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))

# Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

### 3. Миграция базы данных

**Первоначальная настройка:**

```bash
# Сборка CLI инструмента
go build -o dmIntegroff-admin cmd/admin/main.go

# Выполнение миграции
./dmIntegroff-admin --migrate

# Или интерактивно
./dmIntegroff-admin
# Выберите: 1. Выполнить миграцию базы данных
```

**Что делает миграция:**
- ✅ Создает все необходимые таблицы
- ✅ Создает индексы для производительности
- ✅ Создает администратора по умолчанию (admin/admin)

### 4. База данных

**SQLite (текущая)** - НЕ рекомендуется для production:
- ❌ Ограничения по производительности
- ❌ Проблемы с конкурентным доступом
- ❌ Нет репликации

**MySQL/PostgreSQL** - рекомендуется:
```bash
# Установка MySQL
sudo apt-get install mysql-server

# Создание базы данных
mysql -u root -p
CREATE DATABASE dmIntegroff CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'dmIntegroff'@'localhost' IDENTIFIED BY 'strong_password';
GRANT ALL PRIVILEGES ON dmIntegroff.* TO 'dmIntegroff'@'localhost';
FLUSH PRIVILEGES;
```

### 4. HTTPS/TLS

**Вариант 1: Nginx reverse proxy (рекомендуется)**

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Вариант 2: Let's Encrypt + Certbot**
```bash
sudo apt-get install certbot
sudo certbot certonly --standalone -d your-domain.com
```

### 5. Systemd Service

Создайте `/etc/systemd/system/dmIntegroff.service`:

```ini
[Unit]
Description=dmIntegroff Integration Service
After=network.target mysql.service

[Service]
Type=simple
User=dmIntegroff
Group=dmIntegroff
WorkingDirectory=/opt/dmIntegroff
Environment="GIN_MODE=release"
EnvironmentFile=/opt/dmIntegroff/.env
ExecStart=/opt/dmIntegroff/dmIntegroff
Restart=always
RestartSec=10

# Безопасность
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/dmIntegroff /opt/dmIntegroff

[Install]
WantedBy=multi-user.target
```

Активация:
```bash
sudo systemctl daemon-reload
sudo systemctl enable dmIntegroff
sudo systemctl start dmIntegroff
sudo systemctl status dmIntegroff
```

### 6. Логирование

```bash
# Создайте директорию для логов
sudo mkdir -p /var/log/dmIntegroff
sudo chown dmIntegroff:dmIntegroff /var/log/dmIntegroff

# Настройте logrotate
sudo nano /etc/logrotate.d/dmIntegroff
```

Содержимое `/etc/logrotate.d/dmIntegroff`:
```
/var/log/dmIntegroff/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 dmIntegroff dmIntegroff
    sharedscripts
    postrotate
        systemctl reload dmIntegroff > /dev/null 2>&1 || true
    endscript
}
```

### 7. Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# Если используете прямой доступ к приложению
sudo ufw allow 8080/tcp
```

### 8. Мониторинг

**Проверка здоровья приложения:**
```bash
# Добавьте health check endpoint
curl http://localhost:8080/health
```

**Мониторинг логов:**
```bash
# Просмотр логов в реальном времени
sudo journalctl -u dmIntegroff -f

# Или файловые логи
tail -f /var/log/dmIntegroff/dmIntegroff.log
```

### 9. Резервное копирование

**База данных:**
```bash
# MySQL
mysqldump -u dmIntegroff -p dmIntegroff > backup_$(date +%Y%m%d).sql

# Автоматическое резервное копирование (cron)
0 2 * * * /usr/bin/mysqldump -u dmIntegroff -p'password' dmIntegroff | gzip > /backup/dmIntegroff_$(date +\%Y\%m\%d).sql.gz
```

**Конфигурация:**
```bash
# Бэкап .env и других конфигов
tar -czf config_backup_$(date +%Y%m%d).tar.gz /opt/dmIntegroff/.env
```

### 10. Безопасность

**Обновите код для production:**

1. **Измените доверенные прокси** (если используете load balancer):
   ```go
   // В internal/routes/routes.go
   r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12"})
   ```

2. **Добавьте rate limiting** (защита от DDoS)

3. **Настройте CORS** (если нужно)

4. **Используйте webhook подписи** (HMAC)

5. **Регулярно обновляйте зависимости:**
   ```bash
   go get -u ./...
   go mod tidy
   ```

### 11. Производительность

**Оптимизация сборки:**
```bash
# Сборка с оптимизацией
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o dmIntegroff cmd/server/main.go

# Дополнительное сжатие (опционально)
upx --best --lzma dmIntegroff
```

**Настройки MySQL:**
```ini
# /etc/mysql/my.cnf
[mysqld]
max_connections = 200
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
```

### 12. Мониторинг и алерты

**Prometheus + Grafana** (опционально):
- Добавьте метрики в приложение
- Настройте алерты на ошибки
- Мониторьте производительность

**Простой мониторинг:**
```bash
# Скрипт проверки доступности
#!/bin/bash
if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "dmIntegroff is down!" | mail -s "Alert: dmIntegroff Down" admin@example.com
    systemctl restart dmIntegroff
fi
```

## 📋 Чеклист перед деплоем

- [ ] Изменен SESSION_SECRET
- [ ] GIN_MODE=release
- [ ] Настроена MySQL/PostgreSQL
- [ ] Настроен HTTPS (Nginx + Let's Encrypt)
- [ ] Создан systemd service
- [ ] Настроен logrotate
- [ ] Настроен firewall
- [ ] Настроено резервное копирование
- [ ] Изменен пароль admin по умолчанию
- [ ] Проверены права доступа к файлам
- [ ] Настроен мониторинг
- [ ] Протестирована работа приложения

## 🔒 Безопасность после установки

1. **Смените пароль admin:**
   - Войдите как admin/admin
   - Создайте нового администратора
   - Удалите или измените пароль дефолтного admin

2. **Ограничьте доступ:**
   - Используйте VPN для доступа к админке
   - Настройте IP whitelist в Nginx
   - Используйте 2FA (если добавите)

3. **Регулярные обновления:**
   ```bash
   # Обновление системы
   sudo apt-get update && sudo apt-get upgrade
   
   # Обновление приложения
   cd /opt/dmIntegroff
   git pull
   go build -o dmIntegroff cmd/server/main.go
   sudo systemctl restart dmIntegroff
   ```

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи: `sudo journalctl -u dmIntegroff -n 100`
2. Проверьте статус: `sudo systemctl status dmIntegroff`
3. Проверьте подключение к БД
4. Проверьте права доступа к файлам

---

**Важно:** Это минимальный набор для production. Для критичных систем рекомендуется:
- Load balancing (несколько инстансов)
- Database replication
- Automated backups
- Professional monitoring (Datadog, New Relic)
- Security audit
