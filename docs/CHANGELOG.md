# История изменений

## [2024-12-04] - Метрики и мониторинг

### Новые возможности

#### Prometheus метрики
- **12 метрик** для мониторинга всех компонентов системы
- **Webhook метрики** - total, success, failure, retries, duration
- **OAuth метрики** - token refresh, cache hit/miss, duration
- **Retry метрики** - количество попыток на запрос
- **Signature метрики** - generated, verified
- **System gauges** - active integrations, queue size
- **Labels** для детальной аналитики по интеграциям и проектам

#### Health Check endpoints
- **GET /metrics/health** - полная проверка здоровья системы
- Статус: healthy/degraded/unhealthy
- Проверка всех компонентов (БД, диск, память)
- Детальная информация и метрики
- Timestamp, uptime, версия
- **GET /metrics/health/live** - liveness probe для Kubernetes
- **GET /metrics/health/ready** - readiness probe для Kubernetes

#### Интеграции
- **Prometheus** - сбор метрик через /metrics endpoint
- **Grafana** - примеры дашбордов и панелей
- **Kubernetes** - health probes для deployment
- **Alerting** - примеры правил для критичных событий

#### Метрики по категориям

**Webhook:**
- `dmintegroff_webhook_total` - всего запросов
- `dmintegroff_webhook_success_total` - успешных
- `dmintegroff_webhook_failure_total` - неудачных (с типом ошибки)
- `dmintegroff_webhook_retries_total` - retry попыток
- `dmintegroff_webhook_duration_seconds` - время выполнения (histogram)

**OAuth:**
- `dmintegroff_oauth_token_refresh_total` - обновления токенов
- `dmintegroff_oauth_token_cache_total` - кэш hit/miss
- `dmintegroff_oauth_duration_seconds` - время получения токена

**Retry:**
- `dmintegroff_retry_attempts` - количество попыток (histogram)

**Signatures:**
- `dmintegroff_signature_generated_total` - сгенерировано
- `dmintegroff_signature_verified_total` - проверено (valid/invalid)

**System:**
- `dmintegroff_active_integrations` - активных интеграций (gauge)
- `dmintegroff_queue_size` - размер очереди (gauge)

### Технические изменения

#### Backend
- `internal/services/metrics_service.go` - сервис Prometheus метрик
- `internal/services/health_service.go` - сервис health checks
- `internal/controllers/metrics_controller.go` - HTTP контроллеры
- Структуры: `MetricsService`, `HealthService`, `HealthCheck`, `ComponentHealth`
- Методы записи метрик для всех компонентов

#### API Endpoints
- `GET /metrics` - Prometheus метрики
- `GET /metrics/health` - полная проверка здоровья
- `GET /metrics/health/live` - liveness probe
- `GET /metrics/health/ready` - readiness probe

#### Зависимости
- `github.com/prometheus/client_golang` v1.23.2 - Prometheus клиент
- `github.com/DATA-DOG/go-sqlmock` v1.5.2 - моки для тестов БД

### Документация
- `docs/METRICS_MONITORING.md` - полное руководство (1000+ строк)
- `docs/METRICS_EXAMPLES.md` - примеры использования
- `METRICS_CHECKLIST.md` - чеклист реализации
- `METRICS_SUMMARY.md` - краткая сводка

### Тесты
- `internal/services/metrics_service_test.go` - 13 тестов
- `internal/services/health_service_test.go` - 8 тестов
- Все 21 тест проходят

### Примеры использования

#### Prometheus запросы
```promql
# Успешность webhook
sum(rate(dmintegroff_webhook_success_total[5m])) /
sum(rate(dmintegroff_webhook_total[5m])) * 100

# P95 время выполнения
histogram_quantile(0.95,
 rate(dmintegroff_webhook_duration_seconds_bucket[5m])
)

# OAuth кэш hit rate
sum(rate(dmintegroff_oauth_token_cache_total{result="hit"}[5m])) /
sum(rate(dmintegroff_oauth_token_cache_total[5m])) * 100
```

#### Kubernetes deployment
```yaml
livenessProbe:
 httpGet:
 path: /metrics/health/live
 port: 8080
readinessProbe:
 httpGet:
 path: /metrics/health/ready
 port: 8080
```

---

## [2024-12-03] - Экспорт/Импорт конфигураций

### Новые возможности

#### Экспорт конфигураций
- **Экспорт выбранных интеграций** - выбор через чекбоксы на странице интеграций
- **Экспорт проекта** - экспорт всех интеграций проекта одним файлом
- **JSON формат** - удобный для чтения и редактирования формат
- **Метаданные** - версия, дата, автор, название проекта
- **Все настройки** - OAuth, подписи, маппинг, шаблоны

#### Импорт конфигураций
- **Импорт из JSON** - загрузка файла через веб-интерфейс
- **Валидация** - проверка формата и данных перед импортом
- **Предпросмотр** - информация о файле перед импортом
- **Выбор проекта** - импорт в любой доступный проект
- **Обработка дубликатов** - пропуск или обновление существующих
- **Генерация токенов** - создание новых webhook токенов и секретов
- **Детальный отчет** - статистика импорта с ошибками

#### Опции импорта
- **Пропускать дубликаты** - интеграции с существующими именами пропускаются
- **Обновлять существующие** - обновление интеграций с совпадающими именами
- **Генерировать новые токены** - создание новых токенов и секретов

#### Использование
- **Резервное копирование** - сохранение конфигураций
- **Миграция** - перенос между dev/staging/production
- **Клонирование** - создание копий интеграций
- **Версионирование** - хранение в Git
- **Обмен** - передача конфигураций между командами

### Технические изменения

#### Backend
- `internal/services/export_import_service.go` - сервис экспорта/импорта
- `internal/controllers/export_import_controller.go` - HTTP контроллеры
- Структуры: `ExportFormat`, `IntegrationExportData`, `ImportOptions`, `ImportResult`
- Функции: `ExportIntegrations()`, `ExportProjectIntegrations()`, `ImportIntegrations()`, `ValidateImportData()`

#### API Endpoints
- `GET /export/integrations?ids=1,2,3` - экспорт выбранных интеграций
- `GET /export/projects/:id` - экспорт всех интеграций проекта
- `POST /import/integrations` - импорт интеграций из файла
- `POST /import/validate` - валидация файла без импорта

#### Frontend
- Кнопки экспорта/импорта на странице интеграций
- Чекбоксы для выбора интеграций
- Модальное окно импорта с настройками
- Предпросмотр информации о файле
- Отображение результатов импорта

### Тестирование
- 11 новых unit тестов
- Тестирование валидации данных
- Тестирование формата экспорта
- Тестирование опций импорта
- Все тесты проходят (51/51)

### Документация
- `docs/EXPORT_IMPORT.md` - полное руководство
- `docs/EXPORT_IMPORT_EXAMPLE.json` - пример файла экспорта
- `EXPORT_IMPORT_SUMMARY.md` - краткая сводка
- `EXPORT_IMPORT_CHECKLIST.md` - чеклист реализации
- API Reference с примерами
- Troubleshooting guide

---

## [2024-12-03] - Webhook подписи для безопасности

### Новые возможности

#### HMAC подписи для webhook
- **Webhook подписи (HMAC)** - криптографическая защита webhook запросов
- **Поддержка алгоритмов** - SHA-256, SHA-512, SHA-1
- **Гибкая конфигурация** - настраиваемые заголовки и алгоритмы
- **Автоматическая генерация секретов** - криптографически безопасные ключи
- **Constant-time сравнение** - защита от timing атак

#### Возможности подписей
- **Аутентификация источника** - проверка, что запрос от доверенного источника
- **Целостность данных** - гарантия, что данные не изменены
- **Защита от подделки** - невозможно отправить поддельный webhook
- **Совместимость** - поддержка форматов GitHub, Stripe, Slack

#### Конфигурация
- `WebhookSignatureEnabled` - включение/выключение подписей
- `WebhookSignatureSecret` - секретный ключ для HMAC
- `WebhookSignatureHeader` - имя заголовка (по умолчанию: X-Webhook-Signature)
- `WebhookSignatureAlgorithm` - алгоритм (sha256, sha512, sha1)

### Технические изменения

#### База данных
- Миграции `008_add_webhook_signatures.sql` и `008_add_webhook_signatures_sqlite.sql`
- Добавлены поля в таблицу `integrations`:
- `webhook_signature_enabled` - флаг включения подписей
- `webhook_signature_secret` - секретный ключ
- `webhook_signature_header` - имя заголовка
- `webhook_signature_algorithm` - алгоритм хеширования

#### Новые модули
- `internal/services/webhook_signature.go` - сервис подписей
- `GenerateSignature()` - генерация HMAC подписи
- `VerifySignature()` - проверка подписи (constant-time)
- `AddSignatureToRequest()` - добавление подписи в исходящий запрос
- `VerifyIncomingSignature()` - проверка входящей подписи
- `GenerateRandomSecret()` - генерация безопасного секрета
- `ValidateSignatureConfig()` - валидация конфигурации

#### Обновленные модули
- `internal/models/integration.go` - добавлены поля webhook подписей
- `internal/services/integration_service.go` - интеграция подписей в ProcessWebhook

#### Формат подписи
```
X-Webhook-Signature: sha256=5d41402abc4b2a76b9719d911017c592
```

### Тестирование
- `internal/services/webhook_signature_test.go` - полный набор тестов
- Генерация подписи для всех алгоритмов (4 теста)
- Верификация подписи (4 теста)
- Добавление подписи в запрос (4 теста)
- Проверка входящей подписи (5 тестов)
- Генерация случайного секрета (3 теста)
- Валидация конфигурации (6 тестов)
- **Все тесты проходят успешно** 

### Документация
- `docs/WEBHOOK_SIGNATURES.md` - полное руководство
- Обзор и принцип работы
- Конфигурация и параметры
- Поддерживаемые алгоритмы
- Примеры использования
- Интеграция с популярными сервисами
- Безопасность и лучшие практики
- Troubleshooting
- API Reference

### Безопасность

#### Защита от атак
- **Timing attacks** - constant-time сравнение через `hmac.Equal()`
- **Replay attacks** - рекомендации по добавлению timestamp
- **MITM attacks** - проверка целостности данных
- **Brute force** - минимальная длина секрета 16 символов

#### Лучшие практики
- Используйте SHA-256 или SHA-512
- Генерируйте длинные секреты (32+ байта)
- Храните секреты безопасно (env, vault)
- Периодически ротируйте секреты
- Всегда проверяйте подписи на стороне получателя

### Примеры интеграций

#### GitHub Webhooks
```go
WebhookSignatureHeader: "X-Hub-Signature-256"
WebhookSignatureAlgorithm: "sha256"
```

#### Stripe Webhooks
```go
WebhookSignatureHeader: "Stripe-Signature"
WebhookSignatureAlgorithm: "sha256"
```

#### Slack Webhooks
```go
WebhookSignatureHeader: "X-Slack-Signature"
WebhookSignatureAlgorithm: "sha256"
```

### Производительность
- **Минимальные накладные расходы** - HMAC вычисляется быстро
- **Кэширование не требуется** - подпись генерируется для каждого запроса
- **Масштабируемость** - не влияет на throughput

---

## [2024-12-03] - Retry механизм с экспоненциальной задержкой

### Новые возможности

#### Автоматические повторные попытки
- **Retry с экспоненциальной задержкой** - автоматическое повторение неудачных запросов
- **Умное определение ошибок** - retry только для временных сбоев (429, 500, 502, 503, 504)
- **Настраиваемые параметры** - количество попыток, задержки, множитель
- **Защита от перегрузки** - экспоненциальное увеличение задержки предотвращает DDoS
- **Подробное логирование** - каждая попытка записывается в лог

#### Применение retry
- **OAuth2 токен запросы** - автоматический retry при получении access token
- **Webhook отправка** - retry при доставке данных в целевые API
- **Сохранение тела запроса** - корректная повторная отправка POST/PUT данных

#### Параметры по умолчанию
- Максимум **3 попытки**
- Начальная задержка **1 секунда**
- Максимальная задержка **30 секунд**
- Множитель **2.0** (экспоненциальный рост)
- Retry для статусов: 429, 500, 502, 503, 504

### Технические изменения

#### Новые структуры и функции
- `RetryConfig` - конфигурация retry механизма
- `DefaultRetryConfig()` - параметры по умолчанию
- `retryWithBackoff()` - основная функция retry с экспоненциальной задержкой
- `isRetryableStatus()` - проверка, нужен ли retry для статуса
- `calculateDelay()` - расчет задержки для попытки

#### Обновленные модули
- `internal/services/oauth_service.go` - добавлен retry для OAuth запросов
- `internal/services/integration_service.go` - добавлен retry для webhook
- Добавлен импорт `math` для экспоненциальных вычислений

#### Логирование
- **Warning** - при каждой retry попытке с деталями (attempt, max, status_code, url)
- **Debug** - информация о задержке перед retry (delay_seconds)
- **Error** - финальная ошибка после исчерпания всех попыток

### Тестирование
- `internal/services/retry_test.go` - полный набор тестов
- Проверка retryable статусов (9 тестов)
- Расчет экспоненциальной задержки (6 тестов)
- Успешный запрос без retry
- Успех после нескольких retry
- Исчерпание всех попыток
- Не-повторяемые статусы (400, 401, 404)
- Корректная обработка тела запроса при retry
- **Все тесты проходят успешно** 

### Документация
- `docs/RETRY_MECHANISM.md` - полное руководство по retry механизму
- Обзор и основные возможности
- Конфигурация и параметры
- Примеры использования
- Логирование и мониторинг
- Лучшие практики
- Troubleshooting
- Roadmap

### Производительность
- **Минимизация нагрузки** - retry только для временных ошибок
- **Экспоненциальная задержка** - защита от перегрузки API
- **Эффективное использование ресурсов** - нет retry для постоянных ошибок (4xx)

### Надежность
- **Автоматическое восстановление** - система справляется с временными сбоями
- **Сохранение данных** - тело запроса корректно повторяется
- **Прозрачность** - все попытки логируются для анализа

### Примеры логов

```
level=warning msg="Request returned retryable status, will retry" 
 attempt=1 max=3 status_code=503 url="https://api.example.com/webhook"

level=debug msg="Waiting before retry" delay_seconds=1.0

level=warning msg="Request returned retryable status, will retry" 
 attempt=2 max=3 status_code=503 url="https://api.example.com/webhook"

level=debug msg="Waiting before retry" delay_seconds=2.0

level=info msg="Webhook processed successfully" 
 integration_id=123 target_api="https://api.example.com/webhook" status_code=200
```

### Roadmap для retry
- [ ] Jitter (случайная вариация задержки)
- [ ] Circuit breaker pattern
- [ ] Метрики retry в Prometheus
- [ ] Настройка retry через UI
- [ ] Retry для конкретных интеграций

---

## [2024-12-03] - OAuth 2.0 и аутентификация

### Новые возможности

#### Аутентификация
- **OAuth 2.0** - автоматическое получение и обновление access токенов
- **Bearer Token** - поддержка статических токенов авторизации
- **Basic Auth** - HTTP базовая аутентификация
- **Тестирование OAuth** - проверка настроек перед активацией
- **Автоматическое обновление токенов** - система следит за сроком действия
- **Безопасное хранение** - все секреты хранятся в БД

#### OAuth 2.0 функции
- **Client Credentials Grant** - для server-to-server интеграций
- **Кэширование токенов** - минимизация запросов к OAuth серверу
- **Автоматический refresh** - токены обновляются за 60 секунд до истечения
- **Поддержка Scope** - настройка областей доступа
- **Гибкие Grant Types** - client_credentials, password и другие

#### UI улучшения
- **Секция аутентификации** в формах создания/редактирования
- **Динамические поля** - показываются только для выбранного типа
- **Кнопка "Тест OAuth"** - проверка настроек в один клик
- **Визуальная обратная связь** - success/error сообщения
- **Защита паролей** - поля типа password для секретов

### Технические изменения

#### База данных
- Добавлены поля аутентификации в таблицу `integrations`:
- `auth_type` - тип аутентификации (none, oauth2, bearer, basic)
- `oauth2_token_url` - endpoint для получения токена
- `oauth2_client_id` - идентификатор клиента
- `oauth2_client_secret` - секретный ключ
- `oauth2_scope` - области доступа
- `oauth2_grant_type` - тип авторизации
- `bearer_token` - статический токен
- `basic_auth_user` - имя пользователя для Basic Auth
- `basic_auth_pass` - пароль для Basic Auth
- `oauth2_access_token` - текущий access token (runtime)
- `oauth2_refresh_token` - refresh token (runtime)
- `oauth2_expires_at` - время истечения токена (runtime)
- Миграции: `007_add_oauth_support_sqlite.sql`, `007_add_oauth_support.sql`

#### Новые модули
- `internal/services/oauth_service.go` - OAuth 2.0 логика
- `GetAccessToken()` - получение валидного токена
- `fetchNewAccessToken()` - запрос нового токена
- `AddAuthHeaders()` - добавление заголовков аутентификации
- `TestOAuth2Connection()` - тестирование настроек

#### Обновленные модули
- `internal/models/integration.go` - добавлены поля OAuth
- `internal/services/integration_service.go` - интеграция с OAuth
- `internal/controllers/integration_controller.go` - обработка OAuth полей
- `internal/routes/routes.go` - новый endpoint `/api/integrations/:id/test-oauth`
- `templates/pages/integration_edit.html` - UI для OAuth
- `templates/pages/integration_create.html` - UI для OAuth
- `static/css/modern.css` - стили для success-box

### Документация
- `docs/OAUTH_GUIDE.md` - полное руководство по OAuth
- `docs/OAUTH_EXAMPLES.md` - примеры настройки популярных API
- `docs/OAUTH_MIGRATION.md` - инструкции по обновлению
- `README.md` - обновлен roadmap и список возможностей

### Безопасность
- Секреты не отображаются в JSON API
- Токены автоматически обновляются
- Поддержка HTTPS для production
- Логи не содержат токены и пароли

### Примеры интеграций
Добавлены примеры для:
- Salesforce
- Microsoft Dynamics 365
- HubSpot
- Zoho CRM
- Pipedrive
- Slack
- Google Sheets API
- Airtable
- Notion
- Mailchimp

### Производительность
- Кэширование токенов снижает нагрузку на OAuth серверы
- Токены обновляются только при необходимости
- Минимальное количество запросов к БД

---

## [2024-11-30] - Кастомные JSON шаблоны

### Новые возможности

#### JSON Шаблоны
- **Кастомные JSON шаблоны** с подстановкой значений через `{{field.path}}`
- **Wildcard для массивов** `{{array.*}}` для копирования массивов неизвестной длины
- **Два режима трансформации**: простой маппинг или кастомный шаблон
- **Валидация шаблонов** с кнопкой "Проверить шаблон"
- **Обратная совместимость** со старым маппингом

#### Подсветка синтаксиса
- **Цветной JSON** в редакторах (One Dark theme)
- **Подсветка плейсхолдеров** `{{field.path}}` синим цветом с фоном
- **Подсветка в реальном времени** при вводе текста
- **Работает в двух редакторах**: шаблоны и Sample Payload

#### UI улучшения
- **Автоформатирование JSON** кнопкой "Автоотступы"
- **Копирование полей** иконками в списке доступных полей
- **Hover эффекты** на строках с полями
- **Обновленные подсказки** со всеми новыми функциями

### Технические изменения

#### База данных
- Добавлено поле `output_template` в таблицу `integrations`
- Миграция `005_add_output_template_sqlite.sql`

#### Новые модули
- `internal/utils/template_processor.go` - обработка шаблонов
- `internal/utils/json_parser.go` - парсинг JSON (обновлен)
- `static/js/json-highlight.js` - подсветка синтаксиса
- `static/css/json-highlight.css` - стили подсветки

#### Обновленные модули
- `internal/models/integration.go` - добавлено поле OutputTemplate
- `internal/services/integration_service.go` - поддержка шаблонов
- `internal/controllers/integration_controller.go` - обработка шаблонов
- `templates/pages/integration_configure.html` - UI для шаблонов

### Документация
- `docs/TEMPLATE_GUIDE.md` - руководство по шаблонам
- `docs/ARRAY_WILDCARD.md` - работа с массивами
- `docs/JSON_HIGHLIGHTING.md` - подсветка синтаксиса
- `docs/EXAMPLES.md` - обновлены примеры

### Тестирование
- 24 теста для template_processor
- 3 теста для array wildcard
- Все тесты проходят успешно

---

## [Предыдущие версии] - 2024

### Новые возможности

#### Управление проектами
- **Каскадное удаление интеграций**
- При удалении проекта автоматически удаляются все его интеграции
- Реализовано через `ON DELETE CASCADE` на уровне БД
- Упрощена логика контроллера проектов

#### Интерфейс и UX
- **Страница настроек пользователя**
- Смена пароля через веб-интерфейс
- Отображение информации о текущем пользователе
- Валидация пароля (минимум 6 символов)
- Проверка совпадения паролей на клиенте и сервере

- **Интерактивная статистика**
- График запросов за последние 30 дней
- Визуализация с помощью Chart.js
- Открывается при клике на блок "Обработано"
- Адаптирован для темной темы интерфейса

- **Улучшенный дашборд**
- Карточки со статистикой интеграций
- Блок последней активности с автообновлением
- Прогресс-бар для новых пользователей
- Быстрые действия для основных операций

#### Логирование и мониторинг
- **Оптимизированное логирование**
- Автоматическое ограничение до 50 записей в БД
- Унифицированные типы логов (webhook, test, error)
- Раздельный учет webhook и тестовых запросов
- Улучшенная производительность за счет автоочистки

- **Статистика в реальном времени**
- API endpoint для получения статистики по дням
- Группировка запросов по датам
- Подсчет только webhook запросов (без тестовых)

#### Безопасность
- **Улучшенная аутентификация**
- Хеширование паролей через bcrypt
- Проверка текущего пароля при смене
- Валидация минимальной длины пароля
- Защита от несанкционированного доступа

### Улучшения

#### Производительность
- Оптимизация SQL запросов для статистики
- Использование `CreateLogWithLimit` для всех типов логов
- Автоматическая очистка старых записей

#### Код и архитектура
- Рефакторинг контроллера настроек
- Унификация типов логов по всей системе
- Улучшенная обработка ошибок
- Добавлены комментарии и документация

#### UI/UX
- Улучшенная видимость графика статистики
- Адаптивный дизайн модальных окон
- Плавные анимации и переходы
- Консистентная цветовая схема

### Исправления

- **Дублирование запросов в логах**
- Исправлена проблема с разными типами логов (incoming/outgoing → webhook)
- Унифицирован подсчет запросов на дашборде
- Корректный учет статистики

- **Отображение графика**
- Исправлена прозрачность элементов графика
- Добавлен контрастный фон для лучшей видимости
- Улучшены цвета для темной темы

- **Счетчик "Обработано"**
- Теперь корректно показывает количество webhook запросов
- Исключены тестовые запросы из общей статистики

### 🗑️ Удалено

- Убрана ссылка "Пользователи" из главного меню (функционал не реализован)
- Удалено управление .env через веб-интерфейс (настройки редактируются вручную)

### Документация

- Обновлен README.md с новыми возможностями
- Добавлена документация по настройкам
- Улучшены примеры конфигурации
- Добавлен CHANGELOG.md

---

## Предыдущие версии

### Начальная версия

#### Основной функционал
- Создание и управление интеграциями
- Система проектов для группировки
- Webhook обработка с маппингом полей
- Режим прослушивания для захвата структуры данных
- Логирование запросов и ошибок
- CLI инструмент для администрирования
- Ролевая модель (admin/specialist)
- Темная тема интерфейса
- Поддержка SQLite и MySQL

