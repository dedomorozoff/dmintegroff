# GraphQL Support Guide

## Обзор

dmIntegroff теперь поддерживает GraphQL API в дополнение к REST. Вы можете:
- Отправлять GraphQL запросы к внешним API
- Трансформировать REST webhook → GraphQL mutation
- Использовать интроспекцию схемы
- Тестировать GraphQL запросы перед активацией

## Возможности

### 1. GraphQL Client
- Отправка queries и mutations
- Поддержка переменных
- Поддержка operation names
- Автоматическая аутентификация (OAuth 2.0, Bearer, Basic Auth)

### 2. Schema Introspection
- Автоматическое получение схемы GraphQL API
- Кэширование схемы в БД
- Просмотр доступных типов и полей

### 3. Variable Mapping
- Маппинг полей из webhook payload в GraphQL переменные
- Поддержка вложенных полей через dot notation
- Динамическая подстановка значений

### 4. Testing
- Тест подключения к GraphQL endpoint
- Тест выполнения запросов с тестовыми данными
- Просмотр результатов в реальном времени

## Создание GraphQL интеграции

### Шаг 1: Создание интеграции

1. Перейдите в **"Интеграции"** → **"Создать"**
2. Выберите **API Type: GraphQL**
3. Укажите:
- **Название**: Описательное имя интеграции
- **GraphQL Endpoint**: URL GraphQL API (например, `https://api.example.com/graphql`)
- **Проект**: Выберите проект для организации

### Шаг 2: Настройка аутентификации (опционально)

Если GraphQL API требует аутентификацию:

#### OAuth 2.0
```
Auth Type: OAuth 2.0
Token URL: https://oauth.example.com/token
Client ID: your-client-id
Client Secret: your-client-secret
Scope: graphql:read graphql:write
```

#### Bearer Token
```
Auth Type: Bearer Token
Token: your-static-token
```

#### Basic Auth
```
Auth Type: Basic Auth
Username: your-username
Password: your-password
```

### Шаг 3: Интроспекция схемы (опционально)

Нажмите **"Introspect Schema"** для автоматического получения схемы API:
- Система выполнит introspection query
- Схема будет сохранена в БД
- Вы сможете просмотреть доступные типы и поля

### Шаг 4: Настройка GraphQL запроса

#### Пример Query
```graphql
query GetUser($userId: ID!) {
 user(id: $userId) {
 id
 name
 email
 profile {
 age
 city
 }
 }
}
```

#### Пример Mutation
```graphql
mutation CreateUser($input: CreateUserInput!) {
 createUser(input: $input) {
 id
 name
 email
 }
}
```

### Шаг 5: Настройка переменных

Укажите маппинг полей из webhook payload в GraphQL переменные:

```json
{
 "userId": "user.id",
 "input": "user"
}
```

Это означает:
- `userId` будет взят из `payload.user.id`
- `input` будет взят из `payload.user`

### Шаг 6: Тестирование

1. Нажмите **"Test Query"**
2. Укажите тестовый payload:
```json
{
 "user": {
 "id": "123",
 "name": "John Doe",
 "email": "john@example.com"
 }
}
```
3. Просмотрите результат выполнения

### Шаг 7: Активация

Нажмите **"Activate"** для активации интеграции.

## Примеры использования

### Пример 1: REST → GraphQL Query

**Сценарий**: Получаем webhook с ID пользователя, запрашиваем данные из GraphQL API

**Webhook payload**:
```json
{
 "user_id": "123"
}
```

**GraphQL Query**:
```graphql
query GetUser($id: ID!) {
 user(id: $id) {
 id
 name
 email
 createdAt
 }
}
```

**Variable Mapping**:
```json
{
 "id": "user_id"
}
```

**Результат**: Система выполнит GraphQL запрос с переменной `id = "123"`

### Пример 2: REST → GraphQL Mutation

**Сценарий**: Получаем webhook с данными пользователя, создаем пользователя через GraphQL

**Webhook payload**:
```json
{
 "name": "John Doe",
 "email": "john@example.com",
 "age": 30
}
```

**GraphQL Mutation**:
```graphql
mutation CreateUser($name: String!, $email: String!, $age: Int!) {
 createUser(input: {
 name: $name
 email: $email
 age: $age
 }) {
 id
 name
 email
 }
}
```

**Variable Mapping**:
```json
{
 "name": "name",
 "email": "email",
 "age": "age"
}
```

### Пример 3: Вложенные данные

**Webhook payload**:
```json
{
 "user": {
 "profile": {
 "firstName": "John",
 "lastName": "Doe"
 },
 "contact": {
 "email": "john@example.com"
 }
 }
}
```

**GraphQL Mutation**:
```graphql
mutation UpdateUser($firstName: String!, $lastName: String!, $email: String!) {
 updateUser(input: {
 firstName: $firstName
 lastName: $lastName
 email: $email
 }) {
 id
 }
}
```

**Variable Mapping**:
```json
{
 "firstName": "user.profile.firstName",
 "lastName": "user.profile.lastName",
 "email": "user.contact.email"
}
```

### Пример 4: Placeholder в запросе

Вместо переменных можно использовать placeholders прямо в запросе:

**GraphQL Query**:
```graphql
query {
 user(id: {{user.id}}) {
 name
 email
 }
}
```

**Webhook payload**:
```json
{
 "user": {
 "id": "123"
 }
}
```

Система автоматически заменит `{{user.id}}` на `"123"`.

## API Endpoints

### Introspect Schema
```http
POST /api/graphql/introspect/:id
```

Выполняет introspection GraphQL API и сохраняет схему.

**Response**:
```json
{
 "schema": "...",
 "updated_at": 1234567890
}
```

### Get Cached Schema
```http
GET /api/graphql/schema/:id
```

Возвращает кэшированную схему.

**Response**:
```json
{
 "schema": "...",
 "updated_at": 1234567890
}
```

### Test Connection
```http
POST /api/graphql/test-connection
Content-Type: application/json

{
 "endpoint": "https://api.example.com/graphql",
 "auth_type": "bearer",
 "bearer_token": "your-token"
}
```

**Response**:
```json
{
 "success": true,
 "message": "GraphQL connection successful"
}
```

### Test Query
```http
POST /api/graphql/test-query
Content-Type: application/json

{
 "integration_id": 1,
 "query": "query { user(id: \"123\") { name } }",
 "variables": {},
 "test_payload": {
 "user_id": "123"
 }
}
```

**Response**:
```json
{
 "success": true,
 "data": {
 "user": {
 "name": "John Doe"
 }
 }
}
```

## UI Components

### GraphQL Query Editor
- Syntax highlighting для GraphQL
- Автодополнение (если схема загружена)
- Валидация синтаксиса
- Форматирование кода

### Variable Mapping Editor
- JSON редактор для маппинга
- Подсказки по доступным полям
- Валидация JSON

### Schema Viewer
- Просмотр типов и полей
- Поиск по схеме
- Документация полей

## Troubleshooting

### Ошибка: "GraphQL connection test failed"

**Причины**:
- Неверный endpoint URL
- Проблемы с аутентификацией
- API недоступен

**Решение**:
1. Проверьте URL endpoint
2. Проверьте настройки аутентификации
3. Используйте "Test Connection" для диагностики

### Ошибка: "GraphQL errors: Field not found"

**Причины**:
- Неверное имя поля в запросе
- Поле не существует в схеме

**Решение**:
1. Выполните introspection для получения актуальной схемы
2. Проверьте имена полей в запросе
3. Используйте "Test Query" для проверки

### Ошибка: "Failed to build variables"

**Причины**:
- Неверный JSON в variable mapping
- Поле не найдено в payload

**Решение**:
1. Проверьте JSON синтаксис в variable mapping
2. Убедитесь, что пути к полям корректны
3. Используйте "Test Query" с тестовым payload

## Логирование

Все GraphQL запросы логируются в разделе **"Логи"**:
- Тип: `graphql`
- URL: GraphQL endpoint
- Request Body: GraphQL query
- Response Body: Результат выполнения
- Status Code: HTTP статус

## Безопасность

### Аутентификация
- Поддержка OAuth 2.0 с автоматическим обновлением токенов
- Bearer tokens для статической аутентификации
- Basic Auth для простых случаев

### Защита данных
- Секреты хранятся в БД (рекомендуется шифрование)
- HTTPS для production
- Валидация всех входных данных

## Best Practices

### 1. Используйте переменные
Вместо:
```graphql
query {
 user(id: "123") { name }
}
```

Используйте:
```graphql
query GetUser($id: ID!) {
 user(id: $id) { name }
}
```

### 2. Указывайте operation names
```graphql
query GetUser($id: ID!) {
 user(id: $id) { name }
}
```

### 3. Кэшируйте схему
- Выполняйте introspection периодически
- Используйте кэшированную схему для валидации

### 4. Тестируйте перед активацией
- Используйте "Test Connection"
- Используйте "Test Query" с реальными данными
- Проверяйте логи после активации

### 5. Обрабатывайте ошибки
- GraphQL может вернуть частичные данные с ошибками
- Проверяйте поле `errors` в ответе
- Логируйте все ошибки для отладки

## Дополнительные ресурсы

- [GraphQL Official Documentation](https://graphql.org/)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [GraphQL Schema Design](https://graphql.org/learn/schema/)

## Примеры популярных GraphQL API

### GitHub GraphQL API
```
Endpoint: https://api.github.com/graphql
Auth: Bearer Token (Personal Access Token)
```

### Shopify GraphQL API
```
Endpoint: https://{shop}.myshopify.com/admin/api/2024-01/graphql.json
Auth: Bearer Token (Access Token)
```

### Hasura GraphQL API
```
Endpoint: https://your-app.hasura.app/v1/graphql
Auth: Bearer Token or Admin Secret
```

---

**Дата**: 2024-12-04 
**Версия**: 1.0 
**Статус**: Готово к использованию
