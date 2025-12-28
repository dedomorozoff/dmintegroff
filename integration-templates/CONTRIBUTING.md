# Добавление новых шаблонов интеграций

## Как добавить новый шаблон

### 1. Создайте JSON файл шаблона

Создайте новый файл в соответствующей категории:

```bash
# Пример для CRM системы
touch integration-templates/crm/новая-система.json
```

### 2. Заполните структуру шаблона

```json
{
  "name": "Название системы",
  "description": "Краткое описание того, что делает шаблон",
  "type": "json",
  "category": "crm",
  "template": {
    // Ваш JSON шаблон здесь
    "field1": "{{user.email}}",
    "field2": "{{user.name}}"
  }
}
```

Для текстовых шаблонов (SQL, form-data и т.д.):

```json
{
  "name": "MySQL Insert",
  "description": "SQL запрос для вставки данных",
  "type": "text",
  "category": "databases",
  "template": "INSERT INTO table (field1, field2) VALUES ('{{user.name}}', '{{user.email}}');"
}
```

### 3. Добавьте файл в список загрузки

Отредактируйте `static/js/popular-templates.js` и добавьте имя файла в соответствующий массив:

```javascript
const templateFiles = {
    crm: ['salesforce.json', 'hubspot.json', 'новая-система.json'], // <- добавьте здесь
    // ...
};
```

### 4. Доступные переменные

Используйте эти переменные в ваших шаблонах:

#### Пользователь
- `{{user.name}}` - Полное имя
- `{{user.first_name}}` - Имя
- `{{user.last_name}}` - Фамилия
- `{{user.email}}` - Email адрес
- `{{user.phone}}` - Номер телефона
- `{{user.id}}` - Уникальный ID

#### Компания
- `{{company.name}}` - Название компании
- `{{company.website}}` - Веб-сайт

#### Событие
- `{{event.type}}` - Тип события
- `{{event.data}}` - Данные события
- `{{event.id}}` - ID события

#### Сообщение
- `{{message.text}}` - Текст сообщения
- `{{message.subject}}` - Тема сообщения

#### Системные
- `{{timestamp}}` - Unix timestamp
- `{{date}}` - Дата в формате ISO
- `{{custom.*}}` - Пользовательские поля

#### Специальные для конкретных систем
- `{{lead.name}}` - Название лида (CRM)
- `{{lead.price}}` - Стоимость лида (CRM)
- `{{deal.title}}` - Название сделки (CRM)
- `{{deal.value}}` - Сумма сделки (CRM)
- `{{payment.amount}}` - Сумма платежа
- `{{payment.description}}` - Описание платежа
- `{{chat.id}}` - ID чата (мессенджеры)

### 5. Категории шаблонов

- `crm` - CRM системы (Salesforce, HubSpot, amoCRM и т.д.)
- `messengers` - Мессенджеры (Telegram, Slack, Discord и т.д.)
- `email` - Email сервисы (MailChimp, SendGrid и т.д.)
- `analytics` - Системы аналитики (Google Analytics, Mixpanel и т.д.)
- `payments` - Платежные системы (Stripe, PayPal и т.д.)
- `databases` - Базы данных (MySQL, PostgreSQL, MongoDB и т.д.)
- `other` - Другие системы (Jira, Trello, Zapier и т.д.)

### 6. Рекомендации по именованию

- Используйте kebab-case для имен файлов: `google-analytics.json`
- Имя файла должно отражать название системы
- Используйте понятные описания на русском языке

### 7. Тестирование

После добавления шаблона:

1. Перезапустите сервер
2. Откройте страницу создания интеграции
3. Проверьте, что новый шаблон появился в списке
4. Загрузите шаблон и убедитесь, что он корректно отображается

### 8. Примеры популярных систем для добавления

#### CRM
- Bitrix24
- Zoho CRM
- Microsoft Dynamics
- Freshsales

#### Мессенджеры
- Microsoft Teams
- Viber
- VK Messages

#### Email
- Constant Contact
- Campaign Monitor
- GetResponse

#### Аналитика
- Yandex.Metrica
- Adobe Analytics
- Hotjar

#### Платежи
- Robokassa
- Qiwi
- Сбербанк Эквайринг

#### Базы данных
- ClickHouse
- Redis
- Elasticsearch

#### Другие
- Asana
- Monday.com
- Airtable
- Google Sheets