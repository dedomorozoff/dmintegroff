# Примеры настройки OAuth 2.0

## Популярные API

### 1. Salesforce

**Получение учетных данных:**
1. Войдите в Salesforce Setup
2. Apps → App Manager → New Connected App
3. Enable OAuth Settings
4. Скопируйте Consumer Key и Consumer Secret

**Настройки в dmIntegroff:**
```
Тип аутентификации: OAuth 2.0
Token URL: https://login.salesforce.com/services/oauth2/token
Client ID: [Consumer Key]
Client Secret: [Consumer Secret]
Scope: api refresh_token
Grant Type: client_credentials
```

**Target API URL:**
```
https://[instance].salesforce.com/services/data/v58.0/sobjects/Lead
```

---

### 2. Microsoft Dynamics 365

**Получение учетных данных:**
1. Azure Portal → App registrations → New registration
2. Certificates & secrets → New client secret
3. API permissions → Add Dynamics CRM permissions

**Настройки в dmIntegroff:**
```
Тип аутентификации: OAuth 2.0
Token URL: https://login.microsoftonline.com/[tenant-id]/oauth2/v2.0/token
Client ID: [Application ID]
Client Secret: [Client Secret Value]
Scope: https://[org].crm.dynamics.com/.default
Grant Type: client_credentials
```

**Target API URL:**
```
https://[org].api.crm.dynamics.com/api/data/v9.2/leads
```

---

### 3. HubSpot

**Получение учетных данных:**
1. HubSpot → Settings → Integrations → Private Apps
2. Create private app
3. Скопируйте Access Token

**Настройки в dmIntegroff:**
```
Тип аутентификации: Bearer Token
Bearer Token: [Access Token]
```

**Target API URL:**
```
https://api.hubapi.com/crm/v3/objects/contacts
```

---

### 4. Zoho CRM

**Получение учетных данных:**
1. Zoho API Console → Add Client
2. Client Type: Server-based Applications
3. Скопируйте Client ID и Client Secret

**Настройки в dmIntegroff:**
```
Тип аутентификации: OAuth 2.0
Token URL: https://accounts.zoho.com/oauth/v2/token
Client ID: [Client ID]
Client Secret: [Client Secret]
Scope: ZohoCRM.modules.ALL
Grant Type: client_credentials
```

**Target API URL:**
```
https://www.zohoapis.com/crm/v3/Leads
```

---

### 5. Pipedrive

**Получение учетных данных:**
1. Pipedrive → Settings → Personal preferences → API
2. Generate new token

**Настройки в dmIntegroff:**
```
Тип аутентификации: Bearer Token
Bearer Token: [API Token]
```

**Target API URL:**
```
https://api.pipedrive.com/v1/persons
```

---

### 6. Slack

**Получение учетных данных:**
1. api.slack.com → Your Apps → Create New App
2. OAuth & Permissions → Bot Token Scopes
3. Install App to Workspace

**Настройки в dmIntegroff:**
```
Тип аутентификации: Bearer Token
Bearer Token: [Bot User OAuth Token]
```

**Target API URL:**
```
https://slack.com/api/chat.postMessage
```

---

### 7. Google Sheets API

**Получение учетных данных:**
1. Google Cloud Console → APIs & Services → Credentials
2. Create Credentials → Service Account
3. Create Key (JSON) → Скопируйте client_email и private_key

**Настройки в dmIntegroff:**
```
Тип аутентификации: OAuth 2.0
Token URL: https://oauth2.googleapis.com/token
Client ID: [client_email]
Client Secret: [private_key]
Scope: https://www.googleapis.com/auth/spreadsheets
Grant Type: client_credentials
```

**Target API URL:**
```
https://sheets.googleapis.com/v4/spreadsheets/[spreadsheet-id]/values/Sheet1!A1:append
```

---

### 8. Airtable

**Получение учетных данных:**
1. Airtable → Account → Generate API key

**Настройки в dmIntegroff:**
```
Тип аутентификации: Bearer Token
Bearer Token: [API Key]
```

**Target API URL:**
```
https://api.airtable.com/v0/[base-id]/[table-name]
```

---

### 9. Notion

**Получение учетных данных:**
1. notion.so/my-integrations → New integration
2. Скопируйте Internal Integration Token

**Настройки в dmIntegroff:**
```
Тип аутентификации: Bearer Token
Bearer Token: [Integration Token]
```

**Target API URL:**
```
https://api.notion.com/v1/pages
```

---

### 10. Mailchimp

**Получение учетных данных:**
1. Mailchimp → Account → Extras → API keys
2. Create A Key

**Настройки в dmIntegroff:**
```
Тип аутентификации: Basic Auth
Username: anystring
Password: [API Key]
```

**Target API URL:**
```
https://[dc].api.mailchimp.com/3.0/lists/[list-id]/members
```

---

## Тестирование настроек

После настройки OAuth:

1. **Нажмите "Тест OAuth"** в форме редактирования
2. Проверьте результат:
- Успех - настройки корректны
- Ошибка - проверьте учетные данные

3. **Отправьте тестовый запрос:**
```bash
curl -X POST http://localhost:8080/webhook/YOUR_TOKEN \
-H "Content-Type: application/json" \
-d '{"name": "Test", "email": "test@example.com"}'
```

4. **Проверьте логи** в разделе "Логи"

---

## Отладка проблем

### Ошибка 401 Unauthorized
- Проверьте Client ID и Secret
- Убедитесь, что токен не истек
- Проверьте права доступа (scopes)

### Ошибка 403 Forbidden
- Недостаточно прав (scopes)
- Проверьте настройки приложения в API провайдере

### Ошибка 400 Bad Request
- Неверный Grant Type
- Неверный формат Scope
- Проверьте документацию API

### Токен не обновляется
- Проверьте Token URL
- Убедитесь, что expires_in возвращается в ответе
- Проверьте логи dmintegroff.log

---

## Безопасность

 **Важно:**
- Никогда не публикуйте Client Secret
- Используйте HTTPS в production
- Регулярно обновляйте токены
- Ограничивайте scopes минимально необходимыми
- Храните резервные копии учетных данных в безопасном месте
