# GraphQL для обогащения REST данных

## Концепция

Вы можете использовать GraphQL для **обогащения** данных из REST webhook перед отправкой в целевой REST API.

## Схема работы

```
┌─────────────┐ ┌──────────────┐ ┌─────────────┐ ┌──────────────┐
│ REST Webhook│ ───> │ dmIntegroff │ ───> │ GraphQL API │ ───> │ Target REST │
│ (JSON) │ │ Transform │ │ (Query) │ │ API (JSON) │
└─────────────┘ └──────────────┘ └─────────────┘ └──────────────┘
 Step 1 Step 2 Step 3 Step 4
```

### Пример сценария

**Задача**: Получаем webhook с `user_id`, нужно получить полные данные пользователя из GraphQL и отправить в CRM.

**Шаг 1**: Webhook приходит с минимальными данными
```json
{
 "user_id": "123",
 "action": "signup"
}
```

**Шаг 2**: dmIntegroff запрашивает данные из GraphQL
```graphql
query GetUser($id: ID!) {
 user(id: $id) {
 name
 email
 phone
 company
 }
}
```

**Шаг 3**: Получаем полные данные
```json
{
 "user": {
 "name": "John Doe",
 "email": "john@example.com",
 "phone": "+1234567890",
 "company": "Acme Corp"
 }
}
```

**Шаг 4**: Отправляем обогащенные данные в CRM
```json
{
 "contact": {
 "name": "John Doe",
 "email": "john@example.com",
 "phone": "+1234567890",
 "company": "Acme Corp",
 "source": "signup"
 }
}
```

## Как реализовать

### Вариант 1: Две интеграции (текущая реализация)

**Интеграция 1**: REST → GraphQL
- Получает webhook
- Запрашивает данные из GraphQL
- Сохраняет результат

**Интеграция 2**: GraphQL → REST
- Читает результат из интеграции 1
- Трансформирует данные
- Отправляет в REST API

**Недостаток**: Нужно две интеграции

### Вариант 2: Цепочка интеграций (будущая функция)

**Одна интеграция** с несколькими шагами:
1. Получить webhook (REST)
2. Обогатить данными из GraphQL
3. Трансформировать
4. Отправить в REST API

**Преимущество**: Всё в одной интеграции

### Вариант 3: GraphQL как источник данных (рекомендуется сейчас)

Используйте **внешний скрипт** для объединения:

```javascript
// webhook-handler.js
const axios = require('axios');

// 1. Получаем webhook
app.post('/webhook', async (req, res) => {
 const { user_id, action } = req.body;
 
 // 2. Запрашиваем данные из GraphQL через dmIntegroff
 const graphqlResponse = await axios.post(
 'http://localhost:8080/webhook/GRAPHQL_TOKEN',
 { user_id }
 );
 
 // 3. Объединяем данные
 const enrichedData = {
 ...graphqlResponse.data.user,
 action,
 source: 'webhook'
 };
 
 // 4. Отправляем в REST API через dmIntegroff
 await axios.post(
 'http://localhost:8080/webhook/REST_TOKEN',
 enrichedData
 );
 
 res.json({ success: true });
});
```

## Практические примеры

### Пример 1: Обогащение данных пользователя

**Сценарий**: Webhook с ID → GraphQL за деталями → CRM

**Webhook payload**:
```json
{
 "user_id": "123",
 "event": "purchase",
 "amount": 99.99
}
```

**GraphQL Query** (интеграция 1):
```graphql
query GetUser($id: ID!) {
 user(id: $id) {
 name
 email
 phone
 address {
 city
 country
 }
 }
}
```

**Результат GraphQL**:
```json
{
 "user": {
 "name": "John Doe",
 "email": "john@example.com",
 "phone": "+1234567890",
 "address": {
 "city": "New York",
 "country": "USA"
 }
 }
}
```

**Финальный payload в CRM** (интеграция 2):
```json
{
 "contact": {
 "name": "John Doe",
 "email": "john@example.com",
 "phone": "+1234567890",
 "city": "New York",
 "country": "USA"
 },
 "purchase": {
 "amount": 99.99,
 "date": "2024-12-04"
 }
}
```

### Пример 2: Проверка данных через GraphQL

**Сценарий**: Webhook → GraphQL проверка → REST API (только если валидно)

**Webhook payload**:
```json
{
 "order_id": "ORD-123",
 "customer_id": "CUST-456"
}
```

**GraphQL Query**:
```graphql
query ValidateOrder($orderId: ID!, $customerId: ID!) {
 order(id: $orderId) {
 id
 status
 customer {
 id
 verified
 }
 }
}
```

**Логика**:
- Если `order.customer.verified == true` → отправить в REST API
- Иначе → пропустить

### Пример 3: Агрегация данных из нескольких источников

**Сценарий**: Webhook → GraphQL (несколько запросов) → REST API

**Webhook payload**:
```json
{
 "user_id": "123"
}
```

**GraphQL Queries**:

Query 1 - Данные пользователя:
```graphql
query GetUser($id: ID!) {
 user(id: $id) {
 name
 email
 }
}
```

Query 2 - Заказы пользователя:
```graphql
query GetOrders($userId: ID!) {
 orders(userId: $userId) {
 id
 total
 date
 }
}
```

Query 3 - Статистика:
```graphql
query GetStats($userId: ID!) {
 userStats(userId: $userId) {
 totalSpent
 orderCount
 }
}
```

**Финальный payload**:
```json
{
 "user": {
 "name": "John Doe",
 "email": "john@example.com"
 },
 "orders": [...],
 "stats": {
 "totalSpent": 1234.56,
 "orderCount": 15
 }
}
```

## Реализация через Node.js/Python

### Node.js пример

```javascript
const express = require('express');
const axios = require('axios');

const app = express();
app.use(express.json());

// Webhook endpoint
app.post('/enrich-webhook', async (req, res) => {
 try {
 const payload = req.body;
 
 // 1. Запрос в GraphQL через dmIntegroff
 const graphqlResult = await axios.post(
 'http://localhost:8080/webhook/GRAPHQL_TOKEN',
 payload
 );
 
 // 2. Объединение данных
 const enrichedData = {
 ...payload,
 ...graphqlResult.data
 };
 
 // 3. Отправка в REST API через dmIntegroff
 await axios.post(
 'http://localhost:8080/webhook/REST_TOKEN',
 enrichedData
 );
 
 res.json({ success: true });
 } catch (error) {
 console.error(error);
 res.status(500).json({ error: error.message });
 }
});

app.listen(3000);
```

### Python пример

```python
from flask import Flask, request, jsonify
import requests

app = Flask(__name__)

@app.route('/enrich-webhook', methods=['POST'])
def enrich_webhook():
 payload = request.json
 
# 1. Запрос в GraphQL через dmIntegroff
 graphql_response = requests.post(
 'http://localhost:8080/webhook/GRAPHQL_TOKEN',
 json=payload
 )
 
# 2. Объединение данных
 enriched_data = {
 **payload,
 **graphql_response.json()
 }
 
# 3. Отправка в REST API через dmIntegroff
 requests.post(
 'http://localhost:8080/webhook/REST_TOKEN',
 json=enriched_data
 )
 
 return jsonify({'success': True})

if __name__ == '__main__':
 app.run(port=3000)
```

## Будущая функция: GraphQL Enrichment

В будущих версиях планируется добавить встроенную поддержку:

### Настройка в UI

```
┌─────────────────────────────────────────────┐
│ Обогащение данных │
├─────────────────────────────────────────────┤
│ │
│ Обогатить данные через GraphQL │
│ │
│ GraphQL Endpoint: │
│ [https://api.example.com/graphql] │
│ │
│ GraphQL Query: │
│ ┌─────────────────────────────────────┐ │
│ │ query GetUser($id: ID!) { │ │
│ │ user(id: $id) { │ │
│ │ name │ │
│ │ email │ │
│ │ } │ │
│ │ } │ │
│ └─────────────────────────────────────┘ │
│ │
│ Маппинг переменных: │
│ ┌─────────────────────────────────────┐ │
│ │ { │ │
│ │ "id": "user_id" │ │
│ │ } │ │
│ └─────────────────────────────────────┘ │
│ │
│ Объединить результат с исходными данными │
│ Да ☐ Заменить │
│ │
└─────────────────────────────────────────────┘
```

### Логика работы

```go
// Псевдокод будущей реализации
func ProcessWebhookWithEnrichment(integration *Integration, payload map[string]interface{}) error {
 // 1. Если включено обогащение
 if integration.EnrichmentEnabled {
 // 2. Выполнить GraphQL запрос
 graphqlResult, err := ExecuteGraphQLQuery(integration.EnrichmentQuery, payload)
 if err != nil {
 return err
 }
 
 // 3. Объединить данные
 if integration.EnrichmentMerge {
 payload = MergeData(payload, graphqlResult)
 } else {
 payload = graphqlResult
 }
 }
 
 // 4. Продолжить обычную обработку
 return ProcessWebhook(integration, payload)
}
```

## Резюме

### Сейчас доступно:

 **REST → GraphQL** - webhook преобразуется в GraphQL запрос 
 **GraphQL → REST** - через две интеграции или внешний скрипт 
 **Цепочка интеграций** - через внешний оркестратор

### Планируется:

 **Встроенное обогащение** - GraphQL как источник дополнительных данных 
 **Цепочка шагов** - несколько действий в одной интеграции 
 **Условная логика** - if/else для разных сценариев

### Рекомендации:

1. **Для простых случаев**: Используйте две интеграции
2. **Для сложных случаев**: Используйте внешний скрипт (Node.js/Python)
3. **Для будущего**: Ждите встроенную функцию обогащения

---

**Дата**: 2024-12-04 
**Версия**: 1.0 
**Статус**: Документация + Roadmap
