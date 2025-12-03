# 🔐 Руководство по OAuth 2.0

## Обзор

dmIntegroff поддерживает несколько методов аутентификации для Target API:

- **None** - без аутентификации
- **OAuth 2.0** - автоматическое получение и обновление токенов
- **Bearer Token** - статический токен авторизации
- **Basic Auth** - базовая HTTP аутентификация

## OAuth 2.0

### Поддерживаемые типы

- **Client Credentials** - для server-to-server интеграций (рекомендуется)
- Другие типы можно настроить через поле Grant Type

### Настройка OAuth 2.0

1. **Перейдите в редактирование интеграции**
2. **Выберите тип аутентификации**: OAuth 2.0
3. **Заполните параметры**:
   - **Token URL** - endpoint для получения токена (например: `https://api.example.com/oauth/token`)
   - **Client ID** - идентификатор клиента
   - **Client Secret** - секретный ключ клиента
   - **Scope** - области доступа (опционально, через пробел)
   - **Grant Type** - тип авторизации (по умолчанию: `client_credentials`)

4. **Нажмите "Тест OAuth"** для проверки настроек
5. **Сохраните интеграцию**

### Как это работает

1. При первом запросе к Target API система автоматически запрашивает access token
2. Токен кэшируется и используется для всех последующих запросов
3. Перед истечением срока действия токен автоматически обновляется
4. Все токены хранятся в зашифрованном виде в базе данных

### Пример: OAuth 2.0 с Client Credentials

```bash
# Система автоматически выполняет запрос:
POST https://api.example.com/oauth/token
Authorization: Basic base64(client_id:client_secret)
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&scope=read write
```

Ответ:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

Затем система использует токен для запросов:
```bash
POST https://api.example.com/data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{"field": "value"}
```

## Bearer Token

Для API, которые используют статический токен:

1. **Выберите тип**: Bearer Token
2. **Введите токен** в поле Bearer Token
3. **Сохраните**

Токен будет добавлен в заголовок:
```
Authorization: Bearer YOUR_TOKEN
```

## Basic Auth

Для API с базовой HTTP аутентификацией:

1. **Выберите тип**: Basic Auth
2. **Введите Username и Password**
3. **Сохраните**

Учетные данные будут закодированы в Base64 и добавлены в заголовок:
```
Authorization: Basic base64(username:password)
```

## Безопасность

- Все секреты (Client Secret, Bearer Token, Password) хранятся в базе данных
- Рекомендуется использовать переменные окружения для особо чувствительных данных
- Access токены автоматически обновляются и не требуют ручного вмешательства
- В логах не отображаются токены и пароли

## Отладка

### Проверка OAuth настроек

Используйте кнопку **"Тест OAuth"** в форме редактирования интеграции для проверки:
- Корректности Token URL
- Валидности Client ID и Secret
- Доступности OAuth сервера

### Логи

Все ошибки аутентификации записываются в:
- Логи интеграции (раздел "Логи")
- Файл `dmintegroff.log`

### Типичные ошибки

**"Failed to get OAuth2 token"**
- Проверьте Token URL
- Убедитесь, что Client ID и Secret корректны
- Проверьте доступность OAuth сервера

**"Token request failed with status 401"**
- Неверные Client ID или Secret
- Проверьте учетные данные в панели управления API

**"Token request failed with status 400"**
- Неверный Grant Type
- Неверный формат Scope
- Проверьте документацию вашего API

## Примеры интеграций

### Salesforce

```
Token URL: https://login.salesforce.com/services/oauth2/token
Grant Type: client_credentials
Scope: api
```

### Microsoft Graph API

```
Token URL: https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token
Grant Type: client_credentials
Scope: https://graph.microsoft.com/.default
```

### Google APIs

```
Token URL: https://oauth2.googleapis.com/token
Grant Type: client_credentials
Scope: https://www.googleapis.com/auth/cloud-platform
```

## Дополнительная информация

- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [Client Credentials Grant](https://oauth.net/2/grant-types/client-credentials/)
