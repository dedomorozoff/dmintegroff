# 🧪 Тестовые скрипты

Эта папка содержит скрипты для тестирования функциональности dmIntegroff.

## test_logs.ps1

Тестирует логирование HTTP запросов с заголовками.

### Использование

```powershell
.\tests\test_logs.ps1
```

### Что делает

1. Отправляет POST запрос на `/test` endpoint
2. Включает кастомные HTTP заголовки:
- `Content-Type: application/json`
- `X-Custom-Header: test-value`
- `User-Agent: TestScript/1.0`
3. Отправляет JSON данные с timestamp

### Проверка результата

После выполнения скрипта откройте:
```
http://localhost:8080/logs
```

Вы должны увидеть:
- 📦 Тело запроса (JSON данные)
- 📋 Заголовки запроса (все HTTP заголовки)

### Требования

- Приложение должно быть запущено (`.\dmintegroff.exe`)
- PowerShell 5.0 или выше

### Пример вывода

```
Отправка тестового запроса на /test endpoint...

Ответ сервера:
{
 "status": "success",
 "message": "Test request received",
 "data": {...}
}

Тестовый запрос отправлен успешно!
Откройте http://localhost:8080/logs для просмотра результата
```

## Добавление новых тестов

Создайте новый скрипт в этой папке:

```powershell
# tests/test_new_feature.ps1
Write-Host "Тестирование новой функции..." -ForegroundColor Green

# Ваш код тестирования
```

Обновите этот README с описанием нового теста.
