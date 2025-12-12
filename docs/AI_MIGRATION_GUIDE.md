# 🔄 Миграция AI настроек в базу данных

Руководство по переносу AI настроек из `.env` файла в базу данных.

---

## 🎯 Зачем мигрировать?

**Преимущества хранения в БД:**
- ✅ **Без перезапуска** - изменения применяются мгновенно
- ✅ **Веб-интерфейс** - удобное управление через UI
- ✅ **Безопасность** - API ключи не в файлах конфигурации
- ✅ **Версионность** - история изменений настроек
- ✅ **Масштабируемость** - легко добавлять новые параметры

---

## 🚀 Автоматическая миграция

### 1. Запуск утилиты миграции

```bash
cd dmintegroff
go run migrate_ai_settings.go
```

**Что происходит:**
1. Читает настройки из `.env` файла
2. Создает запись в таблице `ai_settings`
3. Проверяет корректность миграции
4. Выводит отчет о результатах

### 2. Пример вывода

```
🔄 Миграция AI настроек из .env в базу данных...
📦 Найдены AI настройки в переменных окружения:
   OpenRouter API Key: sk-or-v1...8760
   OpenRouter Model: anthropic/claude-3.5-sonnet
   Max Tokens: 4000
   Temperature: 0.3
   Request Timeout: 30 сек
✅ AI настройки успешно мигрированы в базу данных!
   Теперь вы можете управлять ими через веб-интерфейс в разделе 'Настройки'
   Можно удалить AI настройки из .env файла
```

---

## 🔧 Ручная настройка

Если автоматическая миграция не подходит:

### 1. Через веб-интерфейс

1. Запустите сервер: `./dmintegroff.exe`
2. Войдите как администратор
3. Перейдите в **Настройки** → **AI Ассистент**
4. Заполните форму:
   - **OpenRouter API Key** - ваш ключ от OpenRouter
   - **Модель** - выберите из списка (рекомендуется Claude 3.5 Sonnet)
   - **Параметры** - настройте токены, температуру, таймаут
5. Нажмите **"Сохранить настройки"**
6. Протестируйте кнопкой **"Тестировать AI"**

### 2. Через API

```bash
curl -X POST http://localhost:8080/api/settings/ai \
  -H "Content-Type: application/json" \
  -H "Cookie: your-session-cookie" \
  -d '{
    "openrouter_api_key": "sk-or-v1-...",
    "openrouter_model": "anthropic/claude-3.5-sonnet",
    "max_tokens": 4000,
    "temperature": 0.3,
    "request_timeout": 30
  }'
```

---

## 📋 Проверка миграции

### 1. Через веб-интерфейс

1. Откройте **Настройки** → **AI Ассистент**
2. Проверьте что поля заполнены
3. Статус должен показывать "AI настроен и доступен"
4. Нажмите **"Тестировать AI"** для проверки

### 2. Через API

```bash
# Проверка статуса
curl http://localhost:8080/api/ai/status

# Получение настроек (только для админов)
curl http://localhost:8080/api/settings/ai
```

### 3. Через базу данных

```sql
-- SQLite
SELECT * FROM ai_settings WHERE is_active = 1;

-- MySQL
SELECT * FROM ai_settings WHERE is_active = true;
```

---

## 🧹 Очистка .env файла

После успешной миграции можно удалить AI настройки из `.env`:

```bash
# Удалите или закомментируйте эти строки:
# OPENROUTER_API_KEY=...
# OPENROUTER_MODEL=...
# OPENAI_API_KEY=...
# AI_MAX_TOKENS=...
# AI_TEMPERATURE=...
# AI_REQUEST_TIMEOUT=...
```

**Важно:** Система поддерживает fallback на `.env`, поэтому можно оставить настройки для резерва.

---

## 🐛 Устранение проблем

### Ошибка "AI не настроен"

**Причины:**
- Миграция не выполнена
- API ключи не сохранились
- База данных недоступна

**Решения:**
1. Повторите миграцию: `go run migrate_ai_settings.go`
2. Проверьте подключение к БД
3. Настройте через веб-интерфейс вручную

### Настройки не применяются

**Причины:**
- Кэширование старой конфигурации
- Ошибки в базе данных

**Решения:**
1. Перезапустите сервер
2. Проверьте логи на ошибки
3. Убедитесь что `is_active = true` в БД

### Дублирование настроек

**Причины:**
- Несколько записей в таблице `ai_settings`
- Конфликт между БД и `.env`

**Решения:**
1. Оставьте только одну активную запись в БД
2. Удалите настройки из `.env`
3. Перезапустите сервер

---

## 📊 Структура таблицы

```sql
CREATE TABLE ai_settings (
    id INTEGER PRIMARY KEY,
    created_at DATETIME,
    updated_at DATETIME,
    openrouter_api_key TEXT,
    openrouter_model VARCHAR(255) DEFAULT 'anthropic/claude-3.5-sonnet',
    openrouter_url VARCHAR(255) DEFAULT 'https://openrouter.ai/api/v1',
    openai_api_key TEXT,
    openai_model VARCHAR(255) DEFAULT 'gpt-4-1106-preview',
    max_tokens INTEGER DEFAULT 4000,
    temperature REAL DEFAULT 0.3,
    request_timeout INTEGER DEFAULT 30,
    local_llm_enabled BOOLEAN DEFAULT false,
    local_llm_url VARCHAR(255) DEFAULT 'http://localhost:11434',
    is_active BOOLEAN DEFAULT true
);
```

---

**Дата создания:** 12 декабря 2024  
**Статус:** Готово к использованию  
**Версия:** 1.0