# 📚 Примеры использования GIntegra

## Базовые сценарии

### Сценарий 1: Интеграция CRM → Email сервис

**Задача**: При создании нового лида в CRM автоматически отправлять welcome email.

**Шаги**:

1. **Создайте интеграцию**
   - Название: "CRM → SendGrid"
   - Target API: `https://api.sendgrid.com/v3/mail/send`
   - Source API: `https://mycrm.com/webhooks` (опционально)

2. **Получите webhook URL**
   ```
   http://localhost:8080/webhook/abc123def456
   ```

3. **Настройте CRM для отправки данных**
   
   CRM отправляет:
   ```json
   {
     "lead_id": 12345,
     "first_name": "Иван",
     "last_name": "Петров",
     "email_address": "ivan@example.com",
     "phone": "+7 999 123-45-67"
   }
   ```

4. **Настройте маппинг в GIntegra**
   
   | Поле источника | Поле назначения | Игнорировать |
   |----------------|-----------------|--------------|
   | lead_id | external_id | ☐ |
   | first_name | to_name | ☐ |
   | email_address | to_email | ☐ |
   | phone | - | ☑ |

5. **Результат**
   
   GIntegra отправит на SendGrid:
   ```json
   {
     "external_id": 12345,
     "to_name": "Иван",
     "to_email": "ivan@example.com"
   }
   ```

---

### Сценарий 2: Интеграция E-commerce → Склад

**Задача**: При оформлении заказа автоматически резервировать товары на складе.

**Пример данных от интернет-магазина**:
```json
{
  "order_number": "ORD-2024-001",
  "customer": {
    "name": "Мария Иванова",
    "email": "maria@example.com"
  },
  "items": [
    {
      "sku": "PROD-123",
      "quantity": 2
    }
  ],
  "total": 5000
}
```

**Настройка маппинга**:
```
order_number → order_id
customer.name → client_name
customer.email → client_email
items → products
total → amount
```

**Результат для склада**:
```json
{
  "order_id": "ORD-2024-001",
  "client_name": "Мария Иванова",
  "client_email": "maria@example.com",
  "products": [...],
  "amount": 5000
}
```

---

### Сценарий 3: Webhook тестирование

**Задача**: Протестировать интеграцию без реального API.

**Используйте встроенный тестовый endpoint**:

```bash
# Создайте интеграцию с Target API:
# http://localhost:8080/test

# Отправьте тестовые данные на webhook:
curl -X POST http://localhost:8080/webhook/YOUR_TOKEN \
  -H "Content-Type: application/json" \
  -d '{
    "test_field": "test_value",
    "number": 123
  }'

# Проверьте логи в интерфейсе (/logs)
# Вы увидите как входящие, так и исходящие данные
```

---

## Продвинутые сценарии

### Сценарий 4: Цепочка интеграций

**Задача**: Данные проходят через несколько систем.

```
Форма на сайте → GIntegra #1 → CRM → GIntegra #2 → Email
```

**Настройка**:

1. **Интеграция #1**: Форма → CRM
   - Webhook: `/webhook/token1`
   - Target: `https://crm.example.com/api/leads`

2. **Настройте CRM** для отправки webhook при создании лида:
   - Webhook URL: `http://your-server:8080/webhook/token2`

3. **Интеграция #2**: CRM → Email
   - Webhook: `/webhook/token2`
   - Target: `https://api.sendgrid.com/v3/mail/send`

---

### Сценарий 5: Обработка ошибок

**Проблема**: Target API недоступен или возвращает ошибку.

**Решение**:

1. Проверьте логи в интерфейсе `/logs`
2. Найдите запрос с ошибкой
3. Посмотрите статус код и тело ответа
4. Исправьте проблему (неверный URL, неправильный формат данных)
5. Повторно отправьте данные вручную через curl

```bash
# Скопируйте данные из лога и отправьте заново
curl -X POST http://localhost:8080/webhook/YOUR_TOKEN \
  -H "Content-Type: application/json" \
  -d '{"copied": "from logs"}'
```

---

## Примеры curl запросов

### Простой POST запрос
```bash
curl -X POST http://localhost:8080/webhook/abc123 \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}'
```

### С вложенными объектами
```bash
curl -X POST http://localhost:8080/webhook/abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "name": "John",
      "email": "john@example.com"
    },
    "order": {
      "id": 123,
      "total": 1000
    }
  }'
```

### С массивом
```bash
curl -X POST http://localhost:8080/webhook/abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"id": 1, "name": "Product 1"},
      {"id": 2, "name": "Product 2"}
    ]
  }'
```

### С авторизацией (для Target API)
```bash
# Пока не поддерживается напрямую в GIntegra
# Используйте промежуточный сервис или настройте Target API
# для приема без авторизации
```

---

## Интеграция с популярными сервисами

### SendGrid (Email)

**Target API**: `https://api.sendgrid.com/v3/mail/send`

**Требуемый формат**:
```json
{
  "personalizations": [{
    "to": [{"email": "recipient@example.com"}]
  }],
  "from": {"email": "sender@example.com"},
  "subject": "Hello",
  "content": [{
    "type": "text/plain",
    "value": "Email body"
  }]
}
```

**Маппинг**:
- `email` → `personalizations[0].to[0].email`
- `subject` → `subject`
- `message` → `content[0].value`

⚠️ **Примечание**: Сложные структуры пока не поддерживаются. Используйте промежуточный API.

---

### Slack (Notifications)

**Target API**: `https://hooks.slack.com/services/YOUR/WEBHOOK/URL`

**Требуемый формат**:
```json
{
  "text": "Message text",
  "username": "Bot Name",
  "icon_emoji": ":ghost:"
}
```

**Маппинг**:
- `message` → `text`
- `sender` → `username`

---

### Telegram Bot

**Target API**: `https://api.telegram.org/bot{TOKEN}/sendMessage`

**Требуемый формат**:
```json
{
  "chat_id": "123456789",
  "text": "Message text"
}
```

**Маппинг**:
- `user_id` → `chat_id`
- `message` → `text`

---

## Отладка и тестирование

### Использование Postman

1. Создайте новый запрос
2. Метод: POST
3. URL: `http://localhost:8080/webhook/YOUR_TOKEN`
4. Headers: `Content-Type: application/json`
5. Body (raw JSON):
   ```json
   {
     "test": "data"
   }
   ```
6. Отправьте запрос
7. Проверьте логи в GIntegra

### Использование webhook.site

1. Создайте интеграцию с Target API: `https://webhook.site/YOUR-UNIQUE-URL`
2. Отправьте данные на webhook GIntegra
3. Проверьте на webhook.site, что данные пришли корректно
4. Настройте маппинг при необходимости

### Локальное тестирование с ngrok

```bash
# Установите ngrok
# https://ngrok.com/download

# Запустите туннель
ngrok http 8080

# Используйте публичный URL для webhook
# https://abc123.ngrok.io/webhook/YOUR_TOKEN
```

---

## Частые проблемы и решения

### Проблема: "Integration not found"

**Причина**: Неверный webhook token

**Решение**: 
- Проверьте URL webhook в списке интеграций
- Убедитесь, что токен скопирован полностью

### Проблема: "Invalid JSON"

**Причина**: Неверный формат данных

**Решение**:
- Проверьте JSON на валидность (jsonlint.com)
- Убедитесь, что Content-Type: application/json

### Проблема: Target API возвращает ошибку

**Причина**: Неверный формат данных для целевого API

**Решение**:
- Проверьте документацию Target API
- Посмотрите логи в GIntegra
- Настройте маппинг правильно

### Проблема: Данные не трансформируются

**Причина**: Интеграция в режиме "listening"

**Решение**:
- Настройте маппинг полей
- Активируйте интеграцию

---

## Полезные советы

1. **Всегда тестируйте** с `/test` endpoint перед использованием реального API
2. **Проверяйте логи** после каждого изменения
3. **Используйте понятные названия** для интеграций
4. **Документируйте маппинг** в Source API поле
5. **Создавайте резервные копии** конфигураций (экспорт БД)

---

## Дополнительные ресурсы

- [Документация Postman](https://learning.postman.com/)
- [JSON Validator](https://jsonlint.com/)
- [Webhook.site](https://webhook.site/)
- [ngrok Documentation](https://ngrok.com/docs)
