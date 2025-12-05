# Примеры шаблонов трансформации

## Тестовые данные

Для всех примеров используем следующие входные данные:

```json
{
  "user": {
    "id": 12345,
    "name": "Иван Петров",
    "email": "ivan@example.com",
    "phone": "+7 (999) 123-45-67"
  },
  "order": {
    "id": "ORD-2024-001",
    "items": [
      {
        "name": "Товар 1",
        "price": 1500,
        "quantity": 2
      },
      {
        "name": "Товар 2",
        "price": 2500,
        "quantity": 1
      }
    ],
    "total": 5500,
    "currency": "RUB"
  },
  "timestamp": "2024-12-06T15:30:00Z"
}
```

## 1. JSON шаблон

### Простой JSON
```json
{
  "customer_id": "{{user.id}}",
  "customer_name": "{{user.name}}",
  "customer_email": "{{user.email}}",
  "order_number": "{{order.id}}",
  "order_total": {{order.total}},
  "created_at": "{{timestamp}}"
}
```

**Результат:**
```json
{
  "customer_id": "12345",
  "customer_name": "Иван Петров",
  "customer_email": "ivan@example.com",
  "order_number": "ORD-2024-001",
  "order_total": 5500,
  "created_at": "2024-12-06T15:30:00Z"
}
```

**Content-Type:** `application/json`

### JSON с массивом
```json
{
  "order": {
    "id": "{{order.id}}",
    "customer": "{{user.name}}",
    "items": {{order.items.*}},
    "total": {{order.total}}
  }
}
```

**Результат:**
```json
{
  "order": {
    "id": "ORD-2024-001",
    "customer": "Иван Петров",
    "items": [
      {
        "name": "Товар 1",
        "price": 1500,
        "quantity": 2
      },
      {
        "name": "Товар 2",
        "price": 2500,
        "quantity": 1
      }
    ],
    "total": 5500
  }
}
```

## 2. XML шаблон

### Простой XML
```xml
<?xml version="1.0" encoding="UTF-8"?>
<order>
  <customer>
    <id>{{user.id}}</id>
    <name>{{user.name}}</name>
    <email>{{user.email}}</email>
    <phone>{{user.phone}}</phone>
  </customer>
  <orderInfo>
    <orderId>{{order.id}}</orderId>
    <total>{{order.total}}</total>
    <currency>{{order.currency}}</currency>
  </orderInfo>
  <timestamp>{{timestamp}}</timestamp>
</order>
```

**Результат:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<order>
  <customer>
    <id>12345</id>
    <name>Иван Петров</name>
    <email>ivan@example.com</email>
    <phone>+7 (999) 123-45-67</phone>
  </customer>
  <orderInfo>
    <orderId>ORD-2024-001</orderId>
    <total>5500</total>
    <currency>RUB</currency>
  </orderInfo>
  <timestamp>2024-12-06T15:30:00Z</timestamp>
</order>
```

**Content-Type:** `application/xml`

### XML с атрибутами
```xml
<?xml version="1.0" encoding="UTF-8"?>
<order id="{{order.id}}" total="{{order.total}}" currency="{{order.currency}}">
  <customer id="{{user.id}}">
    <name>{{user.name}}</name>
    <email>{{user.email}}</email>
  </customer>
  <created>{{timestamp}}</created>
</order>
```

## 3. Plain Text шаблон

### Email уведомление
```
Здравствуйте, {{user.name}}!

Ваш заказ №{{order.id}} успешно оформлен.

Детали заказа:
- Сумма: {{order.total}} {{order.currency}}
- Email: {{user.email}}
- Телефон: {{user.phone}}

Дата и время: {{timestamp}}

С уважением,
Команда интернет-магазина
```

**Результат:**
```
Здравствуйте, Иван Петров!

Ваш заказ №ORD-2024-001 успешно оформлен.

Детали заказа:
- Сумма: 5500 RUB
- Email: ivan@example.com
- Телефон: +7 (999) 123-45-67

Дата и время: 2024-12-06T15:30:00Z

С уважением,
Команда интернет-магазина
```

**Content-Type:** `text/plain`

### Лог-формат
```
[{{timestamp}}] ORDER_CREATED order_id={{order.id}} customer_id={{user.id}} customer_name="{{user.name}}" total={{order.total}} currency={{order.currency}}
```

**Результат:**
```
[2024-12-06T15:30:00Z] ORDER_CREATED order_id=ORD-2024-001 customer_id=12345 customer_name="Иван Петров" total=5500 currency=RUB
```

## 4. Custom шаблоны

### CSV формат
```
{{user.id}},{{user.name}},{{user.email}},{{order.id}},{{order.total}},{{timestamp}}
```

**Результат:**
```
12345,Иван Петров,ivan@example.com,ORD-2024-001,5500,2024-12-06T15:30:00Z
```

**Content-Type:** `text/plain`

**Примечание:** Для CSV можно добавить заголовок `Content-Type: text/csv` через Custom Headers.

### TSV формат
```
{{user.id}}	{{user.name}}	{{user.email}}	{{order.id}}	{{order.total}}	{{timestamp}}
```

**Результат:**
```
12345	Иван Петров	ivan@example.com	ORD-2024-001	5500	2024-12-06T15:30:00Z
```

### URL-encoded формат
```
user_id={{user.id}}&user_name={{user.name}}&user_email={{user.email}}&order_id={{order.id}}&order_total={{order.total}}&timestamp={{timestamp}}
```

**Результат:**
```
user_id=12345&user_name=Иван Петров&user_email=ivan@example.com&order_id=ORD-2024-001&order_total=5500&timestamp=2024-12-06T15:30:00Z
```

**Примечание:** Для URL-encoded можно добавить заголовок `Content-Type: application/x-www-form-urlencoded` через Custom Headers.

### INI/Config формат
```
[customer]
id = {{user.id}}
name = {{user.name}}
email = {{user.email}}
phone = {{user.phone}}

[order]
id = {{order.id}}
total = {{order.total}}
currency = {{order.currency}}

[metadata]
timestamp = {{timestamp}}
```

**Результат:**
```
[customer]
id = 12345
name = Иван Петров
email = ivan@example.com
phone = +7 (999) 123-45-67

[order]
id = ORD-2024-001
total = 5500
currency = RUB

[metadata]
timestamp = 2024-12-06T15:30:00Z
```

### Custom Protocol
```
CMD:ORDER_CREATE|USER:{{user.id}}|NAME:{{user.name}}|EMAIL:{{user.email}}|ORDER:{{order.id}}|TOTAL:{{order.total}}|TIME:{{timestamp}}
```

**Результат:**
```
CMD:ORDER_CREATE|USER:12345|NAME:Иван Петров|EMAIL:ivan@example.com|ORDER:ORD-2024-001|TOTAL:5500|TIME:2024-12-06T15:30:00Z
```

## Переопределение Content-Type

Для Custom шаблонов можно переопределить Content-Type через Custom Headers:

1. Перейдите в настройки интеграции
2. В разделе "Произвольные заголовки" добавьте:
   - **Имя:** `Content-Type`
   - **Значение:** нужный тип (например, `text/csv`, `application/x-www-form-urlencoded`)

## Проверка в логах

После отправки данных проверьте в разделе "Тесты интеграций":
- Content-Type отображается в виде цветного бейджа
- JSON данные подсвечиваются синтаксически
- Не-JSON данные отображаются как plain text
- В заголовках запроса Content-Type выделен жирным

## Советы по отладке

1. **Используйте тестовый URL** для проверки шаблонов
2. **Проверяйте логи** после каждого изменения шаблона
3. **Начните с простого** шаблона и постепенно усложняйте
4. **Используйте кнопку "Проверить шаблон"** перед сохранением
5. **Для JSON используйте "Автоотступы"** для форматирования

## Частые ошибки

### JSON шаблон
❌ **Неправильно:**
```json
{
  "name": {{user.name}}  // Строка без кавычек
}
```

✅ **Правильно:**
```json
{
  "name": "{{user.name}}"  // Строка в кавычках
}
```

### XML шаблон
❌ **Неправильно:**
```xml
<name>{{user.name}}</name>  // Без XML декларации
```

✅ **Правильно:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<name>{{user.name}}</name>
```

### Массивы
❌ **Неправильно:**
```json
{
  "items": "{{order.items.*}}"  // Массив в кавычках
}
```

✅ **Правильно:**
```json
{
  "items": {{order.items.*}}  // Массив без кавычек
}
```
