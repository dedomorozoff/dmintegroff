# Быстрый старт демо-режима

## За 3 минуты

### 1️⃣ Обновите базу данных (30 сек)

```bash
go run cmd/admin/main.go
```
Выберите: `1. Выполнить миграцию базы данных`

### 2️⃣ Настройте .env (30 сек)

Добавьте в конец файла `.env`:

```env
DEMO_MODE=true
DEMO_SECRET=demo123
DEMO_TARGET_URL=https://webhook.site/unique-id
```

 Получите свой webhook.site URL: https://webhook.site

### 3️⃣ Запустите сервер (10 сек)

```bash
go run cmd/server/main.go
```

### 4️⃣ Откройте демо-ссылку (10 сек)

```
http://localhost:8080/login?demo_key=demo123&username=demo&password=demo123
```

### 5️⃣ Войдите в систему (10 сек)

Поля уже заполнены → нажмите "Войти в систему"

### 6️⃣ Создайте интеграцию (60 сек)

1. Интеграции → Создать интеграцию
2. Название: `Тест`
3. Target API: `https://webhook.site/unique-id` (ваш URL)
4. Сохранить

### 7️⃣ Отправьте тестовый вебхук (30 сек)

Скопируйте Webhook URL из интеграции и выполните:

```bash
curl -X POST http://localhost:8080/webhook/YOUR_TOKEN ^
-H "Content-Type: application/json" ^
-d "{\"test\": \"hello from demo\"}"
```

### Готово!

Проверьте webhook.site - вы должны увидеть ваш запрос!

---

## Что дальше?

- 📖 Полная документация: `docs/DEMO_MODE.md`
- Примеры: `docs/DEMO_MODE_EXAMPLE.md`
- Настройка: `DEMO_MODE_SETUP.md`

## Важно

- Демо-пользователи удаляются через 24 часа
- Вебхуки работают только с URL из `DEMO_TARGET_URL`
- Для продакшена установите `DEMO_MODE=false`
