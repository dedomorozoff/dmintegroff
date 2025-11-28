# Организация статических файлов

## Структура

```
static/
├── css/
│   └── common.css    # Общие стили для всех страниц
└── js/
    └── common.js     # Общие JavaScript функции
```

## CSS (common.css)

Содержит:
- Общие стили (reset, body, container)
- Стили кнопок (.btn, .btn-primary, .btn-secondary, .btn-danger)
- Стили форм (input, textarea, label)
- Стили таблиц
- Стили карточек (.card)
- Информационные блоки (.info-box, .warning-box, .error-box)
- Бейджи (.badge, .badge-listening, .badge-active, .badge-inactive)
- Утилитарные классы (.mt-1, .mb-1, .text-center)

## JavaScript (common.js)

Содержит функции:
- `copyToClipboard(text)` - копирование в буфер обмена
- `confirmDelete(message)` - подтверждение удаления
- `generateMapping(event)` - генерация JSON маппинга из формы

## Использование в шаблонах

### Подключение CSS
```html
<link rel="stylesheet" href="/static/css/common.css">
```

### Подключение JS
```html
<script src="/static/js/common.js"></script>
```

### Дополнительные стили
Если странице нужны специфичные стили, добавьте `<style>` блок после подключения common.css:
```html
<link rel="stylesheet" href="/static/css/common.css">
<style>
    /* Специфичные стили для этой страницы */
</style>
```

## Добавление новых файлов

1. Создайте файл в соответствующей папке (`css/` или `js/`)
2. Подключите в нужных шаблонах
3. Убедитесь, что сервер перезапущен (при разработке)

## Production

Для production рекомендуется:
- Минифицировать CSS и JS
- Использовать CDN для статических файлов
- Включить кеширование статических ресурсов
