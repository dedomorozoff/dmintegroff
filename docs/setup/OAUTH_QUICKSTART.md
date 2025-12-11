# Быстрый старт с OAuth 2.0

## За 5 минут

### Шаг 1: Создайте интеграцию

1. Перейдите в **Интеграции** → **Создать**
2. Заполните основные поля:
- Название: "Моя OAuth интеграция"
- Target API URL: `https://api.example.com/data`
- HTTP метод: POST
- Проект: выберите существующий

### Шаг 2: Настройте OAuth

3. Прокрутите до **"Аутентификация Target API"**
4. Выберите тип: **OAuth 2.0**
5. Заполните поля:
 ```
 Token URL: https://api.example.com/oauth/token
 Client ID: your_client_id
 Client Secret: your_client_secret
 Scope: read write (опционально)
 Grant Type: client_credentials
 ```

### Шаг 3: Проверьте настройки

6. Нажмите **"Тест OAuth"**
7. Дождитесь результата:
- Успех → переходите к шагу 4
- Ошибка → проверьте учетные данные

### Шаг 4: Сохраните и активируйте

8. Нажмите **"Создать интеграцию"**
9. Скопируйте Webhook URL
10. Отправьте тестовый запрос:

```bash
curl -X POST http://localhost:8080/webhook/YOUR_TOKEN \
-H "Content-Type: application/json" \
-d '{"name": "John", "email": "john@example.com"}'
```

### Шаг 5: Настройте маппинг

11. Перейдите в **"Настроить маппинг"**
12. Выберите поля для трансформации
13. Сохраните и активируйте

## Готово! 

Теперь все запросы на ваш webhook будут:
1. Трансформироваться согласно маппингу
2. Автоматически получать OAuth токен
3. Отправляться на Target API с Authorization заголовком

## Альтернативные методы

### Bearer Token (проще)

Если у вас уже есть статический токен:

```
Тип аутентификации: Bearer Token
Bearer Token: your_static_token
```

### Basic Auth (самый простой)

Для API с базовой аутентификацией:

```
Тип аутентификации: Basic Auth
Username: your_username
Password: your_password
```

## Что дальше?

- [📖 Полное руководство по OAuth](OAUTH_GUIDE.md)
- [ Примеры популярных API](OAUTH_EXAMPLES.md)
- [ Миграция существующих интеграций](OAUTH_MIGRATION.md)

## Помощь

**Не работает?**
1. Проверьте логи в разделе "Логи"
2. Проверьте файл `dmintegroff.log`
3. Используйте кнопку "Тест OAuth"

**Типичные ошибки:**
- 401 → Неверные Client ID/Secret
- 403 → Недостаточно прав (scope)
- 400 → Неверный Grant Type или Scope
