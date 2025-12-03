# Экспорт/Импорт конфигураций

## Обзор

Функция экспорта/импорта позволяет сохранять и восстанавливать конфигурации интеграций в формате JSON. Это полезно для:

- **Резервного копирования** конфигураций
- **Миграции** между окружениями (dev → staging → production)
- **Клонирования** настроек интеграций
- **Обмена** конфигурациями между проектами
- **Версионирования** настроек в системе контроля версий

## Возможности

### Экспорт

- ✅ Экспорт выбранных интеграций
- ✅ Экспорт всех интеграций проекта
- ✅ Сохранение всех настроек (OAuth, подписи, маппинг)
- ✅ Метаданные экспорта (дата, автор, версия)
- ✅ Формат JSON для удобного чтения и редактирования

### Импорт

- ✅ Импорт из JSON файла
- ✅ Валидация перед импортом
- ✅ Выбор целевого проекта
- ✅ Обработка дубликатов (пропуск или обновление)
- ✅ Генерация новых токенов и секретов
- ✅ Детальный отчет о результатах импорта

## Использование

### Экспорт интеграций

#### Через веб-интерфейс

1. Перейдите на страницу **Интеграции**
2. Выберите интеграции с помощью чекбоксов
3. Нажмите кнопку **Экспорт**
4. Файл будет автоматически загружен

#### Экспорт всего проекта

1. Откройте страницу проекта
2. Нажмите кнопку **Экспорт**
3. Все интеграции проекта будут экспортированы

#### Через API

```bash
# Экспорт выбранных интеграций
curl -X GET "http://localhost:8080/export/integrations?ids=1,2,3" \
  -H "Cookie: mysession=..." \
  -o integrations_export.json

# Экспорт проекта
curl -X GET "http://localhost:8080/export/projects/1" \
  -H "Cookie: mysession=..." \
  -o project_export.json
```

### Импорт интеграций

#### Через веб-интерфейс

1. Перейдите на страницу **Интеграции**
2. Нажмите кнопку **Импорт**
3. Выберите JSON файл
4. Система покажет информацию о файле
5. Выберите целевой проект
6. Настройте опции импорта:
   - **Пропускать дубликаты** - интеграции с существующими именами будут пропущены
   - **Обновлять существующие** - обновить интеграции с совпадающими именами
   - **Генерировать новые токены** - создать новые webhook токены и секреты
7. Нажмите **Импортировать**

#### Через API

```bash
curl -X POST "http://localhost:8080/import/integrations" \
  -H "Cookie: mysession=..." \
  -F "file=@integrations_export.json" \
  -F "project_id=1" \
  -F "skip_duplicates=true" \
  -F "generate_new_tokens=true"
```

## Формат файла экспорта

### Структура JSON

```json
{
  "version": "1.0",
  "exported_at": "2024-12-03T10:30:00Z",
  "exported_by": "admin",
  "project_name": "My Project",
  "integrations": [
    {
      "name": "GitHub to Slack",
      "source_api": "https://github.com",
      "target_api": "https://hooks.slack.com/services/...",
      "http_method": "POST",
      "mode": "listening",
      "mapping_config": "{\"text\": \"message\"}",
      "output_template": "{\"text\": \"{{message}}\"}",
      
      "auth_type": "oauth2",
      "oauth2_token_url": "https://oauth.example.com/token",
      "oauth2_client_id": "client-id",
      "oauth2_client_secret": "client-secret",
      "oauth2_scope": "webhooks:write",
      "oauth2_grant_type": "client_credentials",
      
      "webhook_signature_enabled": true,
      "webhook_signature_secret": "secret-key",
      "webhook_signature_header": "X-Webhook-Signature",
      "webhook_signature_algorithm": "sha256"
    }
  ]
}
```

### Поля

#### Метаданные

- `version` - версия формата экспорта (текущая: "1.0")
- `exported_at` - дата и время экспорта
- `exported_by` - пользователь, выполнивший экспорт
- `project_name` - название проекта (опционально)

#### Интеграция

**Основные настройки:**
- `name` - название интеграции (обязательно)
- `source_api` - URL источника (опционально)
- `target_api` - URL целевого API (обязательно)
- `http_method` - HTTP метод (GET, POST, PUT, PATCH, DELETE)
- `mode` - режим работы (listening, active, inactive)
- `mapping_config` - конфигурация маппинга полей (JSON)
- `output_template` - шаблон вывода (JSON с плейсхолдерами)

**Аутентификация:**
- `auth_type` - тип аутентификации (none, oauth2, bearer, basic)
- `oauth2_token_url` - URL для получения OAuth токена
- `oauth2_client_id` - OAuth Client ID
- `oauth2_client_secret` - OAuth Client Secret
- `oauth2_scope` - OAuth scopes
- `oauth2_grant_type` - OAuth grant type
- `bearer_token` - статический Bearer токен
- `basic_auth_user` - Basic Auth username
- `basic_auth_pass` - Basic Auth password

**Webhook подписи:**
- `webhook_signature_enabled` - включены ли подписи
- `webhook_signature_secret` - секретный ключ для HMAC
- `webhook_signature_header` - название заголовка с подписью
- `webhook_signature_algorithm` - алгоритм (sha256, sha512, sha1)

## Опции импорта

### Пропускать дубликаты (Skip Duplicates)

Если включено, интеграции с именами, которые уже существуют в целевом проекте, будут пропущены.

```json
{
  "skip_duplicates": true
}
```

**Результат:**
- Существующие интеграции не изменяются
- Новые интеграции создаются
- В отчете указывается количество пропущенных

### Обновлять существующие (Update Existing)

Если включено, интеграции с совпадающими именами будут обновлены.

```json
{
  "update_existing": true
}
```

**Результат:**
- Существующие интеграции обновляются
- Новые интеграции создаются
- Webhook токены сохраняются

⚠️ **Внимание:** Нельзя одновременно включить `skip_duplicates` и `update_existing`.

### Генерировать новые токены (Generate New Tokens)

Если включено, для всех интеграций будут созданы новые:
- Webhook токены
- Секреты для подписей

```json
{
  "generate_new_tokens": true
}
```

**Рекомендуется включать при:**
- Импорте в другое окружение
- Клонировании интеграций
- Создании копий для тестирования

**Рекомендуется выключать при:**
- Восстановлении из резервной копии
- Миграции с сохранением URL

## Результаты импорта

После импорта система предоставляет детальный отчет:

```json
{
  "total_count": 10,
  "imported_count": 7,
  "updated_count": 2,
  "skipped_count": 1,
  "errors": [
    "Integration 5 (Invalid API): target API is required"
  ]
}
```

### Поля отчета

- `total_count` - всего интеграций в файле
- `imported_count` - успешно импортировано новых
- `updated_count` - успешно обновлено существующих
- `skipped_count` - пропущено (дубликаты)
- `errors` - список ошибок с описанием

## Примеры использования

### Резервное копирование

```bash
# Экспорт всех интеграций проекта
curl -X GET "http://localhost:8080/export/projects/1" \
  -H "Cookie: mysession=..." \
  -o backup_$(date +%Y%m%d).json
```

### Миграция между окружениями

```bash
# 1. Экспорт из dev
curl -X GET "http://localhost:8080/export/projects/1" \
  -H "Cookie: mysession=..." \
  -o dev_export.json

# 2. Редактирование URL в файле (опционально)
sed -i 's/dev.example.com/prod.example.com/g' dev_export.json

# 3. Импорт в production
curl -X POST "https://prod.example.com/import/integrations" \
  -H "Cookie: mysession=..." \
  -F "file=@dev_export.json" \
  -F "project_id=1" \
  -F "generate_new_tokens=true"
```

### Клонирование интеграции

```bash
# 1. Экспорт одной интеграции
curl -X GET "http://localhost:8080/export/integrations?ids=5" \
  -o integration.json

# 2. Редактирование названия в файле
jq '.integrations[0].name = "Copy of " + .integrations[0].name' \
  integration.json > integration_copy.json

# 3. Импорт в тот же проект
curl -X POST "http://localhost:8080/import/integrations" \
  -F "file=@integration_copy.json" \
  -F "project_id=1" \
  -F "generate_new_tokens=true"
```

### Версионирование в Git

```bash
# Экспорт конфигураций
curl -X GET "http://localhost:8080/export/projects/1" \
  -o config/integrations.json

# Коммит в Git
git add config/integrations.json
git commit -m "Update integration configs"
git push
```

## Безопасность

### Чувствительные данные

Файлы экспорта содержат чувствительные данные:
- OAuth Client Secrets
- Bearer токены
- Basic Auth пароли
- Webhook секреты

⚠️ **Рекомендации:**
- Не храните файлы экспорта в публичных репозиториях
- Используйте шифрование для хранения
- Ограничьте доступ к файлам экспорта
- Регулярно ротируйте секреты

### Валидация

Система выполняет валидацию при импорте:
- Проверка формата JSON
- Проверка версии формата
- Проверка обязательных полей
- Проверка существования проекта
- Проверка прав доступа пользователя

## Ограничения

- Максимальный размер файла импорта: зависит от настроек сервера
- Формат версии: только "1.0" поддерживается
- Webhook токены: генерируются новые при импорте (если включена опция)
- Sample payload: не экспортируется (создается при первом запросе)
- Request logs: не экспортируются

## Troubleshooting

### Ошибка: "Invalid JSON format"

**Причина:** Файл поврежден или имеет неверный формат

**Решение:**
```bash
# Проверка валидности JSON
jq . export.json

# Форматирование JSON
jq . export.json > export_formatted.json
```

### Ошибка: "Unsupported export version"

**Причина:** Файл создан в другой версии системы

**Решение:** Обновите систему или используйте совместимую версию

### Ошибка: "Integration with name 'X' already exists"

**Причина:** Интеграция с таким именем уже существует

**Решение:**
- Включите опцию "Пропускать дубликаты"
- Или включите "Обновлять существующие"
- Или измените имя в файле экспорта

### Ошибка: "Target project not found"

**Причина:** Указан несуществующий проект

**Решение:** Проверьте ID проекта или создайте новый проект

## API Reference

### GET /export/integrations

Экспорт выбранных интеграций

**Query Parameters:**
- `ids` (required) - список ID интеграций через запятую

**Response:**
- Content-Type: `application/json`
- Content-Disposition: `attachment; filename=integrations_export_YYYYMMDD_HHMMSS.json`

### GET /export/projects/:id

Экспорт всех интеграций проекта

**Path Parameters:**
- `id` (required) - ID проекта

**Response:**
- Content-Type: `application/json`
- Content-Disposition: `attachment; filename=project_ID_export_YYYYMMDD_HHMMSS.json`

### POST /import/integrations

Импорт интеграций из файла

**Form Data:**
- `file` (required) - JSON файл экспорта
- `project_id` (required) - ID целевого проекта
- `skip_duplicates` (optional) - пропускать дубликаты (true/false)
- `update_existing` (optional) - обновлять существующие (true/false)
- `generate_new_tokens` (optional) - генерировать новые токены (true/false)

**Response:**
```json
{
  "message": "Import completed",
  "result": {
    "total_count": 10,
    "imported_count": 7,
    "updated_count": 2,
    "skipped_count": 1,
    "errors": []
  }
}
```

### POST /import/validate

Валидация файла импорта без импорта

**Form Data:**
- `file` (required) - JSON файл экспорта

**Response:**
```json
{
  "valid": true,
  "version": "1.0",
  "exported_at": "2024-12-03T10:30:00Z",
  "exported_by": "admin",
  "project_name": "My Project",
  "integrations_count": 10
}
```

## Changelog

### Version 1.0 (2024-12-03)

- ✅ Экспорт выбранных интеграций
- ✅ Экспорт проекта
- ✅ Импорт с валидацией
- ✅ Обработка дубликатов
- ✅ Генерация новых токенов
- ✅ Веб-интерфейс
- ✅ API endpoints
- ✅ Детальные отчеты
- ✅ 11 unit тестов

## См. также

- [OAuth 2.0 Guide](OAUTH_GUIDE.md)
- [Webhook Signatures](WEBHOOK_SIGNATURES.md)
- [Retry Mechanism](RETRY_MECHANISM.md)
