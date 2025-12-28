# 📚 Документация dmIntegroff

Добро пожаловать в документацию dmIntegroff - современной системы управления интеграциями REST API.

## 🚀 Быстрый старт

- **[Быстрый старт](../QUICKSTART.md)** - запуск за 5 минут
- **[Руководство пользователя](guides/USER_GUIDE.md)** - полное руководство по использованию
- **[Примеры интеграций](guides/EXAMPLES.md)** - реальные сценарии использования

## 📖 Руководства пользователя

### Основы
- **[Руководство пользователя](guides/USER_GUIDE.md)** - полное руководство по использованию системы
- **[Работа с проектами](guides/PROJECTS.md)** - организация интеграций в проекты
- **[Настройки системы](guides/SETTINGS_GUIDE.md)** - конфигурация и персонализация

### Примеры и сценарии
- **[Примеры интеграций](guides/EXAMPLES.md)** - реальные сценарии использования
- **[OAuth примеры](features/OAUTH_EXAMPLES.md)** - примеры настройки OAuth
- **[GraphQL примеры](features/GRAPHQL_EXAMPLES.md)** - работа с GraphQL API

## 🔧 Функции и возможности

### Интеграции
- **[Множественные выходы](features/MULTIPLE_OUTPUTS_GUIDE.md)** - один webhook → несколько API
- **[Тестирование webhook](features/WEBHOOK_TEST.md)** - встроенные инструменты отладки
- **[Экспорт/импорт](features/EXPORT_IMPORT.md)** - резервное копирование конфигураций

### Аутентификация
- **[OAuth 2.0](features/OAUTH_GUIDE.md)** - настройка OAuth аутентификации
- **[GraphQL интеграции](features/GRAPHQL_GUIDE.md)** - работа с GraphQL API
- **[GraphQL + REST](features/GRAPHQL_REST_ENRICHMENT.md)** - обогащение данных

### Мониторинг
- **[Метрики и мониторинг](monitoring/METRICS_MONITORING.md)** - Prometheus интеграция
- **[Примеры метрик](monitoring/METRICS_EXAMPLES.md)** - практические примеры

## 🛠 Для разработчиков

### Техническая документация
- **[Техническая документация](development/TECHNICAL_DOCS.md)** - архитектура и API
- **[Руководство программиста](development/PROGRAMMER_GUIDE.md)** - разработка и расширение
- **[План разработки](development/DEVELOPMENT_PLAN.md)** - техническое планирование

### Настройка и развертывание
- **[Настройка CORS](setup/CORS_GUIDE.md)** - конфигурация CORS
- **[Настройка прокси](setup/PROXY_SETUP.md)** - развертывание за прокси
- **[Безопасность прокси](setup/PROXY_SECURITY.md)** - безопасная конфигурация
- **[Настройка Redis](setup/REDIS_SETUP.md)** - кеширование и производительность
- **[Демо режим](setup/DEMO_MODE.md)** - настройка демонстрационного режима

## 🤖 AI и планы развития

### AI интеграция
- **[AI резюме](AI_SUMMARY.md)** - краткое описание AI возможностей
- **[Roadmap](ROADMAP.md)** - планируемые функции и улучшения
- **[TODO список](TODO_NEXT.md)** - ближайшие задачи

### История изменений
- **[Changelog](CHANGELOG.md)** - история всех изменений в проекте

## 🧪 Тестирование

- **[Тестирование кастомных заголовков](testing/CUSTOM_HEADERS_TESTING.md)**
- **[Тестирование множественных выходов](testing/test_multiple_outputs.md)**
- **[Использование запроса как образца](testing/test_use_as_sample.md)**

## 📁 Структура документации

```
docs/
├── guides/              # Руководства пользователя
├── features/           # Описание функций
├── development/        # Техническая документация
├── setup/             # Настройка и развертывание
├── monitoring/        # Мониторинг и метрики
├── testing/           # Тестирование
├── AI_SUMMARY.md      # AI возможности
├── ROADMAP.md         # План развития
├── TODO_NEXT.md       # Ближайшие задачи
└── CHANGELOG.md       # История изменений
```

## 🎯 Рекомендуемый порядок изучения

### Для новых пользователей
1. **[Быстрый старт](../QUICKSTART.md)** - запуск системы
2. **[Руководство пользователя](guides/USER_GUIDE.md)** - основы работы
3. **[Примеры интеграций](guides/EXAMPLES.md)** - практические сценарии

### Для продвинутых пользователей
1. **[OAuth настройка](features/OAUTH_GUIDE.md)** - аутентификация
2. **[GraphQL интеграции](features/GRAPHQL_GUIDE.md)** - работа с GraphQL
3. **[Множественные выходы](features/MULTIPLE_OUTPUTS_GUIDE.md)** - сложные сценарии

### Для разработчиков
1. **[Техническая документация](development/TECHNICAL_DOCS.md)** - архитектура
2. **[Руководство программиста](development/PROGRAMMER_GUIDE.md)** - разработка
3. **[Настройка окружения](setup/)** - развертывание

## 🤝 Участие в развитии

- **[Contributing](../CONTRIBUTING.md)** - как участвовать в развитии проекта
- **[GitHub Issues](https://github.com/dedomorozoff/dmintegroff/issues)** - сообщить о проблеме или предложить улучшение

## 📞 Поддержка

Если вы не нашли ответ в документации:

1. Проверьте [FAQ в руководстве пользователя](guides/USER_GUIDE.md#faq)
2. Создайте [Issue на GitHub](https://github.com/dedomorozoff/dmintegroff/issues)
3. Изучите [примеры интеграций](guides/EXAMPLES.md)

---

**Последнее обновление:** 28 декабря 2024