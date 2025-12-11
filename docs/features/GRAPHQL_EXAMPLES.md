# GraphQL Integration Examples

## Пример 1: GitHub API - Создание Issue

### Описание
Получаем webhook с данными об ошибке и создаем GitHub Issue через GraphQL API.

### Настройка интеграции

**API Type**: GraphQL 
**GraphQL Endpoint**: `https://api.github.com/graphql` 
**Auth Type**: Bearer Token 
**Bearer Token**: `ghp_your_github_token`

### GraphQL Mutation
```graphql
mutation CreateIssue($repositoryId: ID!, $title: String!, $body: String!) {
 createIssue(input: {
 repositoryId: $repositoryId
 title: $title
 body: $body
 }) {
 issue {
 id
 number
 title
 url
 }
 }
}
```

### Variable Mapping
```json
{
 "repositoryId": "repository_id",
 "title": "error.title",
 "body": "error.description"
}
```

### Webhook Payload
```json
{
 "repository_id": "MDEwOlJlcG9zaXRvcnkxMjM0NTY3ODk=",
 "error": {
 "title": "Critical Bug in Production",
 "description": "Application crashed with error: NullPointerException"
 }
}
```

### Результат
GitHub Issue будет создан с указанным заголовком и описанием.

---

## Пример 2: Shopify - Создание продукта

### Описание
Получаем webhook с данными о новом продукте и создаем его в Shopify через GraphQL.

### Настройка интеграции

**API Type**: GraphQL 
**GraphQL Endpoint**: `https://your-shop.myshopify.com/admin/api/2024-01/graphql.json` 
**Auth Type**: Bearer Token 
**Bearer Token**: `shpat_your_access_token`

### GraphQL Mutation
```graphql
mutation CreateProduct($title: String!, $description: String!, $price: String!) {
 productCreate(input: {
 title: $title
 descriptionHtml: $description
 variants: [{
 price: $price
 }]
 }) {
 product {
 id
 title
 handle
 }
 userErrors {
 field
 message
 }
 }
}
```

### Variable Mapping
```json
{
 "title": "product.name",
 "description": "product.description",
 "price": "product.price"
}
```

### Webhook Payload
```json
{
 "product": {
 "name": "Awesome T-Shirt",
 "description": "<p>High quality cotton t-shirt</p>",
 "price": "29.99"
 }
}
```

---

## Пример 3: Hasura - Вставка данных

### Описание
Получаем webhook с данными пользователя и вставляем в БД через Hasura GraphQL.

### Настройка интеграции

**API Type**: GraphQL 
**GraphQL Endpoint**: `https://your-app.hasura.app/v1/graphql` 
**Auth Type**: Bearer Token 
**Bearer Token**: `your_hasura_admin_secret`

### GraphQL Mutation
```graphql
mutation InsertUser($name: String!, $email: String!, $age: Int!) {
 insert_users_one(object: {
 name: $name
 email: $email
 age: $age
 }) {
 id
 name
 email
 created_at
 }
}
```

### Variable Mapping
```json
{
 "name": "user.name",
 "email": "user.email",
 "age": "user.age"
}
```

### Webhook Payload
```json
{
 "user": {
 "name": "John Doe",
 "email": "john@example.com",
 "age": 30
 }
}
```

---

## Пример 4: Contentful - Создание записи

### Описание
Получаем webhook с контентом и создаем запись в Contentful CMS.

### Настройка интеграции

**API Type**: GraphQL 
**GraphQL Endpoint**: `https://graphql.contentful.com/content/v1/spaces/{space_id}` 
**Auth Type**: Bearer Token 
**Bearer Token**: `your_contentful_token`

### GraphQL Mutation
```graphql
mutation CreateBlogPost($title: String!, $content: String!, $author: String!) {
 createBlogPost(data: {
 title: $title
 content: $content
 author: $author
 }) {
 id
 title
 publishedAt
 }
}
```

### Variable Mapping
```json
{
 "title": "post.title",
 "content": "post.body",
 "author": "post.author.name"
}
```

### Webhook Payload
```json
{
 "post": {
 "title": "Getting Started with GraphQL",
 "body": "GraphQL is a query language for APIs...",
 "author": {
 "name": "Jane Smith"
 }
 }
}
```

---

## Пример 5: Strapi - Обновление записи

### Описание
Получаем webhook с обновленными данными и обновляем запись в Strapi.

### Настройка интеграции

**API Type**: GraphQL 
**GraphQL Endpoint**: `https://your-strapi.com/graphql` 
**Auth Type**: Bearer Token 
**Bearer Token**: `your_strapi_jwt_token`

### GraphQL Mutation
```graphql
mutation UpdateArticle($id: ID!, $title: String!, $content: String!) {
 updateArticle(
 input: {
 where: { id: $id }
 data: {
 title: $title
 content: $content
 }
 }
 ) {
 article {
 id
 title
 updatedAt
 }
 }
}
```

### Variable Mapping
```json
{
 "id": "article.id",
 "title": "article.title",
 "content": "article.content"
}
```

### Webhook Payload
```json
{
 "article": {
 "id": "123",
 "title": "Updated Title",
 "content": "Updated content..."
 }
}
```

---

## Пример 6: Apollo Federation - Query с фрагментами

### Описание
Запрос данных из нескольких микросервисов через Apollo Federation.

### Настройка интеграции

**API Type**: GraphQL 
**GraphQL Endpoint**: `https://gateway.example.com/graphql` 
**Auth Type**: OAuth 2.0

### GraphQL Query
```graphql
query GetUserWithOrders($userId: ID!) {
 user(id: $userId) {
 id
 name
 email
 orders {
 id
 total
 items {
 productId
 quantity
 }
 }
 }
}
```

### Variable Mapping
```json
{
 "userId": "user_id"
}
```

### Webhook Payload
```json
{
 "user_id": "user_123"
}
```

---

## Пример 7: Placeholder в запросе (без переменных)

### Описание
Использование placeholders для простых случаев без переменных.

### GraphQL Query
```graphql
query {
 user(id: {{user.id}}) {
 name
 email
 posts(limit: {{limit}}) {
 title
 createdAt
 }
 }
}
```

### Webhook Payload
```json
{
 "user": {
 "id": "123"
 },
 "limit": 10
}
```

### Результат
Система заменит:
- `{{user.id}}` → `"123"`
- `{{limit}}` → `10`

---

## Пример 8: Сложный маппинг с массивами

### Описание
Передача массива объектов в GraphQL mutation.

### GraphQL Mutation
```graphql
mutation CreateOrder($userId: ID!, $items: [OrderItemInput!]!) {
 createOrder(input: {
 userId: $userId
 items: $items
 }) {
 id
 total
 }
}
```

### Variable Mapping
```json
{
 "userId": "customer.id",
 "items": "order.items"
}
```

### Webhook Payload
```json
{
 "customer": {
 "id": "user_123"
 },
 "order": {
 "items": [
 {
 "productId": "prod_1",
 "quantity": 2,
 "price": 29.99
 },
 {
 "productId": "prod_2",
 "quantity": 1,
 "price": 49.99
 }
 ]
 }
}
```

---

## Пример 9: Условная логика с фрагментами

### GraphQL Query
```graphql
query GetContent($id: ID!, $includeComments: Boolean!) {
 content(id: $id) {
 id
 title
 body
 comments @include(if: $includeComments) {
 id
 text
 author
 }
 }
}
```

### Variable Mapping
```json
{
 "id": "content_id",
 "includeComments": "include_comments"
}
```

### Webhook Payload
```json
{
 "content_id": "123",
 "include_comments": true
}
```

---

## Пример 10: Batch операции

### GraphQL Mutation
```graphql
mutation BatchCreateUsers($users: [UserInput!]!) {
 createUsers(input: $users) {
 count
 users {
 id
 name
 }
 }
}
```

### Variable Mapping
```json
{
 "users": "users"
}
```

### Webhook Payload
```json
{
 "users": [
 {
 "name": "John Doe",
 "email": "john@example.com"
 },
 {
 "name": "Jane Smith",
 "email": "jane@example.com"
 }
 ]
}
```

---

## Тестирование примеров

Для каждого примера:

1. **Создайте интеграцию** с указанными настройками
2. **Настройте аутентификацию** (получите токены от соответствующих сервисов)
3. **Используйте "Test Connection"** для проверки подключения
4. **Используйте "Test Query"** с примером webhook payload
5. **Активируйте интеграцию** после успешного теста
6. **Отправьте реальный webhook** для проверки

## Примечания

- Замените `your-shop`, `your-app`, `your_token` на реальные значения
- Для GitHub требуется Personal Access Token с правами `repo`
- Для Shopify требуется Admin API access token
- Для Hasura можно использовать admin secret или JWT token
- Все примеры используют актуальные API на момент написания (2024-12-04)

---

**Дата**: 2024-12-04 
**Версия**: 1.0 
**Статус**: Готово к использованию
