# ⚡ Быстрый старт dmIntegroff

Это руководство поможет вам запустить dmIntegroff за 5 минут.

## 📦 Установка

```bash
# 1. Клонируйте репозиторий
git clone https://github.com/dedomorozoff/dmintegroff.git
cd dmIntegroff

# 2. Установите зависимости
go mod download

# 3. Выполните миграцию базы данных (первый запуск)
go run cmd/admin/main.go --migrate

# 4. Запустите сервер
go run cmd/server/main.go
```

## 🔐 Первый вход

Откройте браузер: **http://localhost:8080**

```
Логин: admin
Пароль: admin
```

⚠️ **Важно**: Смените пароль после первого входа!

## 🚀 Создание первой интеграции

### Шаг 1: Создайте интеграцию

1. Перейдите в **"Интеграции"** → **"Создать"**
2. Заполните форму:
   - **Название**: "Тестовая интеграция"
   - **Target API**: `http://localhost:8080/test`
3. Нажмите **"Создать интеграцию"**

### Шаг 2: Получите webhook URL

Скопируйте webhook URL из списка интеграций:
```
http://localhost:8080/webhook/abc123def456
```

### Шаг 3: Отправьте тестовые данные

```bash
curl -X POST http://localhost:8080/webhook/abc123def456 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Иван Петров",
    "email": "ivan@example.com",
    "phone": "+7 999 123-45-67"
  }'
```

**Ответ**:
```json
{
  "status": "captured",
  "message": "Sample data captured. Configure field mapping to activate integration."
}
```

### Шаг 4: Настройте маппинг

1. Вернитесь в интерфейс
2. Нажмите **"⚙️ Настроить"** рядом с интеграцией
3. Настройте маппинг полей:

| Поле источника | Поле назначения | Игнорировать |
|----------------|-----------------|--------------|
| name | client_name | ☐ |
| email | client_email | ☐ |
| phone | - | ☑ |

4. Нажмите **"✅ Сохранить и активировать"**

### Шаг 5: Проверьте работу

Отправьте данные еще раз:

```bash
curl -X POST http://localhost:8080/webhook/abc123def456 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Мария Иванова",
    "email": "maria@example.com",
    "phone": "+7 999 987-65-43"
  }'
```

**Ответ**:
```json
{
  "status": "success"
}
```

### Шаг 6: Проверьте логи

1. Перейдите в **"Логи"**
2. Вы увидите:
   - Входящий запрос на webhook
   - Исходящий запрос на Target API с трансформированными данными

## 📚 Что дальше?

- 📖 Прочитайте [полную документацию](README.md)
- 💡 Изучите [примеры использования](docs/EXAMPLES.md)
- 🛠 Узнайте о [разработке](docs/PROGRAMMER_GUIDE.md)
- 🔧 Настройте [конфигурацию](.env.example)

## 🆘 Проблемы?

### Порт 8080 занят

```bash
# Измените порт в .env
PORT=3000

# Или через переменную окружения
PORT=3000 go run cmd/server/main.go
```

### Ошибка "command not found: go"

Установите Go: https://golang.org/dl/

### База данных не создается

Проверьте права на запись в текущей директории:
```bash
ls -la dmIntegroff.db
```

### Не могу войти

Проверьте логи:
```bash
tail -f dmIntegroff.log
```

## 🎯 Полезные команды

```bash
# Просмотр логов в реальном времени
tail -f dmIntegroff.log

# Очистка базы данных
rm dmIntegroff.db

# Форматирование кода
go fmt ./...

# Проверка кода
go vet ./...

# Сборка исполняемого файла
go build -o dmIntegroff cmd/server/main.go
```

## 🌟 Следующие шаги

1. **Настройте реальную интеграцию**
   - Используйте реальный Target API
   - Настройте webhook в вашей системе

2. **Изучите продвинутые функции**
   - Цепочки интеграций
   - Обработка ошибок
   - Мониторинг логов

3. **Настройте production**
   - Измените SESSION_SECRET
   - Используйте MySQL вместо SQLite
   - Настройте HTTPS
   - Добавьте резервное копирование

## 📞 Поддержка

- 🐛 [Сообщить о баге](https://github.com/dedomorozoff/dmintegroff/issues)
- 💡 [Предложить улучшение](https://github.com/dedomorozoff/dmintegroff/issues)
- 📖 [Документация](docs/)
- 💬 [Обсуждения](https://github.com/dedomorozoff/dmintegroff/discussions)

---

**Готово!** Теперь вы можете создавать интеграции между любыми API! 🎉
