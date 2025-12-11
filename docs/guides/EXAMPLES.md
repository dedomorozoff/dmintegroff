# Примеры использования dmIntegroff

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

4. **Настройте трансформацию в dmIntegroff**

 **Вариант A: Простой маппинг**
 
 | Поле источника | Поле назначения | Игнорировать |
 |----------------|-----------------|--------------|
 | lead_id | external_id | ☐ |
 | first_name | to_name | ☐ |
 | email_address | to_email | ☐ |
 | phone | - | |

 **Вариант B: Кастомный шаблон** (рекомендуется для сложных структур)
 ```json
 {
 "personalizations": [{
 "to": [{
 "email": "{{email_address}}",
 "name": "{{first_name}} {{last_name}}"
 }]
 }],
 "from": {
 "email": "noreply@mycompany.com",
 "name": "My Company"
 },
 "subject": "Welcome!",
 "content": [{
 "type": "text/plain",
 "value": "Hello {{first_name}}!"
 }],
 "custom_args": {
 "lead_id": "{{lead_id}}"
 }
 }
 ```

5. **Результат**
 
 dmIntegroff отправит на SendGrid (простой маппинг):
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
Форма на сайте → dmIntegroff #1 → CRM → dmIntegroff #2 → Email
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
# Пока не поддерживается напрямую в dmIntegroff
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

 **Примечание**: Сложные структуры пока не поддерживаются. Используйте промежуточный API.

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
7. Проверьте логи в dmIntegroff

### Использование webhook.site

1. Создайте интеграцию с Target API: `https://webhook.site/YOUR-UNIQUE-URL`
2. Отправьте данные на webhook dmIntegroff
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
- Посмотрите логи в dmIntegroff
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


---

## 🆕 Продвинутые сценарии с кастомными шаблонами

### Сценарий 7: Реструктуризация данных для внешнего API

**Задача**: Получать данные от одной системы и отправлять в другую с полностью измененной структурой.

**Входные данные от CRM:**
```json
{
 "contact": {
 "personal": {
 "first_name": "Мария",
 "last_name": "Иванова",
 "birth_date": "1990-05-15"
 },
 "communication": {
 "email": "maria@example.com",
 "phone": "+7 999 888-77-66",
 "preferred": "email"
 }
 },
 "lead": {
 "source": "website",
 "campaign": "summer_2024",
 "score": 85
 },
 "timestamp": "2024-11-30T10:00:00Z"
}
```

**Кастомный шаблон для Target API:**
```json
{
 "customer": {
 "fullName": "{{contact.personal.first_name}} {{contact.personal.last_name}}",
 "dateOfBirth": "{{contact.personal.birth_date}}",
 "contacts": {
 "primary": "{{contact.communication.email}}",
 "secondary": "{{contact.communication.phone}}",
 "preferredMethod": "{{contact.communication.preferred}}"
 }
 },
 "marketing": {
 "source": "{{lead.source}}",
 "campaign": "{{lead.campaign}}",
 "leadScore": {{lead.score}}
 },
 "metadata": {
 "receivedAt": "{{timestamp}}",
 "processedBy": "dmIntegroff",
 "version": "1.0"
 }
}
```

**Результат отправки:**
```json
{
 "customer": {
 "fullName": "Мария Иванова",
 "dateOfBirth": "1990-05-15",
 "contacts": {
 "primary": "maria@example.com",
 "secondary": "+7 999 888-77-66",
 "preferredMethod": "email"
 }
 },
 "marketing": {
 "source": "website",
 "campaign": "summer_2024",
 "leadScore": 85
 },
 "metadata": {
 "receivedAt": "2024-11-30T10:00:00Z",
 "processedBy": "dmIntegroff",
 "version": "1.0"
 }
}
```

**Преимущества:**
- Полный контроль над структурой
- Добавление статических полей
- Комбинирование данных из разных уровней
- Переименование и реорганизация

---

### Сценарий 8: Работа с массивами данных

**Задача**: Обработка заказа с несколькими товарами.

**Входные данные от интернет-магазина:**
```json
{
 "order": {
 "id": "ORD-12345",
 "customer": {
 "name": "Алексей Смирнов",
 "email": "alex@example.com"
 },
 "items": [
 {
 "product_id": "PROD-001",
 "name": "Ноутбук",
 "quantity": 1,
 "price": 50000
 },
 {
 "product_id": "PROD-002",
 "name": "Мышь",
 "quantity": 2,
 "price": 500
 }
 ],
 "total": 51000,
 "status": "paid"
 }
}
```

**Кастомный шаблон для системы учета:**
```json
{
 "orderId": "{{order.id}}",
 "customerName": "{{order.customer.name}}",
 "customerEmail": "{{order.customer.email}}",
 "orderStatus": "{{order.status}}",
 "totalAmount": {{order.total}},
 "firstItem": {
 "productId": "{{order.items[0].product_id}}",
 "productName": "{{order.items[0].name}}",
 "quantity": {{order.items[0].quantity}},
 "price": {{order.items[0].price}}
 },
 "secondItem": {
 "productId": "{{order.items[1].product_id}}",
 "productName": "{{order.items[1].name}}",
 "quantity": {{order.items[1].quantity}},
 "price": {{order.items[1].price}}
 },
 "integration": {
 "source": "online_store",
 "processor": "dmIntegroff"
 }
}
```

**Результат:**
```json
{
 "orderId": "ORD-12345",
 "customerName": "Алексей Смирнов",
 "customerEmail": "alex@example.com",
 "orderStatus": "paid",
 "totalAmount": 51000,
 "firstItem": {
 "productId": "PROD-001",
 "productName": "Ноутбук",
 "quantity": 1,
 "price": 50000
 },
 "secondItem": {
 "productId": "PROD-002",
 "productName": "Мышь",
 "quantity": 2,
 "price": 500
 },
 "integration": {
 "source": "online_store",
 "processor": "dmIntegroff"
 }
}
```

---

### Сценарий 9: Минималистичная трансформация

**Задача**: Извлечь только нужные данные из большого объекта.

**Входные данные (большой объект):**
```json
{
 "user": {
 "id": 123,
 "profile": {
 "name": "Иван",
 "email": "ivan@example.com",
 "phone": "+7 999 123-45-67",
 "address": {
 "city": "Москва",
 "street": "Ленина",
 "building": "10"
 },
 "preferences": {
 "language": "ru",
 "timezone": "Europe/Moscow"
 }
 },
 "metadata": {
 "created_at": "2024-01-01",
 "updated_at": "2024-11-30",
 "last_login": "2024-11-30T09:00:00Z"
 }
 }
}
```

**Кастомный шаблон (только нужное):**
```json
{
 "userId": {{user.id}},
 "name": "{{user.profile.name}}",
 "email": "{{user.profile.email}}",
 "city": "{{user.profile.address.city}}"
}
```

**Результат (компактный):**
```json
{
 "userId": 123,
 "name": "Иван",
 "email": "ivan@example.com",
 "city": "Москва"
}
```

---

## Сравнение методов трансформации

| Критерий | Простой маппинг | Кастомный шаблон |
|----------|----------------|------------------|
| **Сложность настройки** | Простая | Средняя |
| **Переименование полей** | | |
| **Изменение структуры** | | |
| **Статические значения** | | |
| **Вложенные объекты** | | |
| **Массивы** | | |
| **Комбинирование данных** | | |
| **Минимизация данных** | Частично | |
| **Время настройки** | 2-5 мин | 5-15 мин |
| **Гибкость** | Низкая | Высокая |

---

## Советы по выбору метода

### Используйте простой маппинг когда:
- Нужно только переименовать поля
- Структура данных остается той же
- Требуется быстрая настройка
- Нет необходимости в статических полях

### Используйте кастомный шаблон когда:
- Нужно изменить структуру данных
- Требуется добавить статические поля
- Нужно комбинировать данные из разных уровней
- Требуется минимизировать объем данных
- Целевой API требует специфичную структуру

---

## Дополнительные ресурсы

- **[Руководство по шаблонам](TEMPLATE_GUIDE.md)** - Подробная документация
- **[Техническая документация](TECHNICAL_DOCS.md)** - API и архитектура
- **[Руководство программиста](PROGRAMMER_GUIDE.md)** - Разработка и расширение

---

**Нужна помощь?** Создайте issue в репозитории или обратитесь к документации.
