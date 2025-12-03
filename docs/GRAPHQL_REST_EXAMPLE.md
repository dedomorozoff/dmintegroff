# 🔄 Пример: REST + GraphQL интеграция

## 🎯 Задача

Получаем webhook с `user_id`, нужно:
1. Получить полные данные пользователя из GraphQL API
2. Отправить обогащенные данные в CRM (REST API)

## 📊 Архитектура

```
Webhook (user_id) → dmIntegroff → GraphQL API → dmIntegroff → CRM REST API
```

## 🚀 Решение 1: Две интеграции

### Интеграция 1: REST → GraphQL

**Название**: "Get User Data from GraphQL"

**Настройки**:
- Тип API: **GraphQL API**
- GraphQL Endpoint: `https://api.example.com/graphql`
- Auth: Bearer Token

**GraphQL Query**:
```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    name
    email
    phone
    company
    address {
      city
      country
    }
  }
}
```

**Variable Mapping**:
```json
{
  "id": "user_id"
}
```

**Webhook URL**: `http://localhost:8080/webhook/TOKEN1`

### Интеграция 2: Webhook → CRM

**Название**: "Send to CRM"

**Настройки**:
- Тип API: **REST API**
- Target API: `https://crm.example.com/api/contacts`
- HTTP Method: POST

**Output Template**:
```json
{
  "contact": {
    "name": "{{user.name}}",
    "email": "{{user.email}}",
    "phone": "{{user.phone}}",
    "company": "{{user.company}}",
    "city": "{{user.address.city}}",
    "country": "{{user.address.country}}"
  },
  "source": "webhook",
  "created_at": "{{timestamp}}"
}
```

**Webhook URL**: `http://localhost:8080/webhook/TOKEN2`

### Использование

**Шаг 1**: Отправьте webhook в интеграцию 1
```bash
curl -X POST http://localhost:8080/webhook/TOKEN1 \
  -H "Content-Type: application/json" \
  -d '{"user_id": "123"}'
```

**Шаг 2**: Скопируйте результат из логов интеграции 1

**Шаг 3**: Отправьте результат в интеграцию 2
```bash
curl -X POST http://localhost:8080/webhook/TOKEN2 \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "id": "123",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+1234567890",
      "company": "Acme Corp",
      "address": {
        "city": "New York",
        "country": "USA"
      }
    }
  }'
```

**Недостаток**: Ручное копирование данных между интеграциями

## 🚀 Решение 2: Внешний оркестратор (Node.js)

### Создайте файл `orchestrator.js`

```javascript
const express = require('express');
const axios = require('axios');

const app = express();
app.use(express.json());

// Конфигурация
const DMINTEGROFF_URL = 'http://localhost:8080';
const GRAPHQL_TOKEN = 'TOKEN1'; // Токен GraphQL интеграции
const CRM_TOKEN = 'TOKEN2';     // Токен CRM интеграции

// Webhook endpoint
app.post('/webhook/enrich', async (req, res) => {
  try {
    console.log('📥 Received webhook:', req.body);
    
    // Шаг 1: Получить данные из GraphQL через dmIntegroff
    console.log('🔍 Fetching user data from GraphQL...');
    const graphqlResponse = await axios.post(
      `${DMINTEGROFF_URL}/webhook/${GRAPHQL_TOKEN}`,
      req.body
    );
    
    console.log('✅ GraphQL response:', graphqlResponse.data);
    
    // Шаг 2: Отправить обогащенные данные в CRM через dmIntegroff
    console.log('📤 Sending to CRM...');
    const crmResponse = await axios.post(
      `${DMINTEGROFF_URL}/webhook/${CRM_TOKEN}`,
      graphqlResponse.data
    );
    
    console.log('✅ CRM response:', crmResponse.data);
    
    res.json({
      success: true,
      message: 'Data enriched and sent to CRM',
      graphql_data: graphqlResponse.data,
      crm_response: crmResponse.data
    });
    
  } catch (error) {
    console.error('❌ Error:', error.message);
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

const PORT = 3000;
app.listen(PORT, () => {
  console.log(`🚀 Orchestrator running on http://localhost:${PORT}`);
  console.log(`📥 Webhook endpoint: http://localhost:${PORT}/webhook/enrich`);
});
```

### Установите зависимости

```bash
npm init -y
npm install express axios
```

### Запустите оркестратор

```bash
node orchestrator.js
```

### Отправьте webhook

```bash
curl -X POST http://localhost:3000/webhook/enrich \
  -H "Content-Type: application/json" \
  -d '{"user_id": "123"}'
```

### Результат

```json
{
  "success": true,
  "message": "Data enriched and sent to CRM",
  "graphql_data": {
    "user": {
      "id": "123",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+1234567890",
      "company": "Acme Corp",
      "address": {
        "city": "New York",
        "country": "USA"
      }
    }
  },
  "crm_response": {
    "id": "contact_456",
    "created": true
  }
}
```

## 🚀 Решение 3: Python оркестратор

### Создайте файл `orchestrator.py`

```python
from flask import Flask, request, jsonify
import requests
import logging

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)

# Конфигурация
DMINTEGROFF_URL = 'http://localhost:8080'
GRAPHQL_TOKEN = 'TOKEN1'  # Токен GraphQL интеграции
CRM_TOKEN = 'TOKEN2'      # Токен CRM интеграции

@app.route('/webhook/enrich', methods=['POST'])
def enrich_webhook():
    try:
        payload = request.json
        logging.info(f'📥 Received webhook: {payload}')
        
        # Шаг 1: Получить данные из GraphQL
        logging.info('🔍 Fetching user data from GraphQL...')
        graphql_response = requests.post(
            f'{DMINTEGROFF_URL}/webhook/{GRAPHQL_TOKEN}',
            json=payload
        )
        graphql_response.raise_for_status()
        graphql_data = graphql_response.json()
        
        logging.info(f'✅ GraphQL response: {graphql_data}')
        
        # Шаг 2: Отправить в CRM
        logging.info('📤 Sending to CRM...')
        crm_response = requests.post(
            f'{DMINTEGROFF_URL}/webhook/{CRM_TOKEN}',
            json=graphql_data
        )
        crm_response.raise_for_status()
        crm_data = crm_response.json()
        
        logging.info(f'✅ CRM response: {crm_data}')
        
        return jsonify({
            'success': True,
            'message': 'Data enriched and sent to CRM',
            'graphql_data': graphql_data,
            'crm_response': crm_data
        })
        
    except Exception as e:
        logging.error(f'❌ Error: {str(e)}')
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500

@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'ok'})

if __name__ == '__main__':
    print('🚀 Orchestrator running on http://localhost:3000')
    print('📥 Webhook endpoint: http://localhost:3000/webhook/enrich')
    app.run(port=3000, debug=True)
```

### Установите зависимости

```bash
pip install flask requests
```

### Запустите оркестратор

```bash
python orchestrator.py
```

## 📊 Сравнение решений

| Решение | Сложность | Гибкость | Производительность |
|---------|-----------|----------|-------------------|
| Две интеграции | ⭐ Низкая | ⭐⭐ Средняя | ⭐⭐⭐ Высокая |
| Node.js оркестратор | ⭐⭐ Средняя | ⭐⭐⭐ Высокая | ⭐⭐ Средняя |
| Python оркестратор | ⭐⭐ Средняя | ⭐⭐⭐ Высокая | ⭐⭐ Средняя |

## 💡 Рекомендации

### Используйте две интеграции если:
- ✅ Простой сценарий
- ✅ Не нужна автоматизация
- ✅ Редкие запросы

### Используйте оркестратор если:
- ✅ Нужна автоматизация
- ✅ Сложная логика
- ✅ Частые запросы
- ✅ Нужна обработка ошибок
- ✅ Нужно логирование

## 🔮 Будущее: Встроенная поддержка

В будущих версиях планируется добавить встроенную поддержку цепочек:

```
┌─────────────────────────────────────────┐
│  Интеграция: User Enrichment            │
├─────────────────────────────────────────┤
│                                         │
│  Шаг 1: Получить webhook (REST)        │
│  Шаг 2: Обогатить через GraphQL         │
│  Шаг 3: Трансформировать               │
│  Шаг 4: Отправить в CRM (REST)         │
│                                         │
└─────────────────────────────────────────┘
```

Это позволит настроить всё в одной интеграции без внешних скриптов!

---

**Дата**: 2024-12-04  
**Версия**: 1.0  
**Статус**: ✅ Работает (через оркестратор)
