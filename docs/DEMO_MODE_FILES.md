# Файлы демо-режима

Полный список всех файлов, связанных с демо-режимом.

## 📂 Структура файлов

```
dmintegroff/
│
├── Основная документация
│ ├── DEMO_MODE_SETUP.md # Пошаговая установка
│ ├── QUICK_START_DEMO.md # Быстрый старт (3 минуты)
│ ├── DEMO_MODE_SUMMARY.md # Сводка изменений
│ ├── CHANGELOG_DEMO_MODE.md # Детальный changelog
│ ├── DEVELOPER_NOTES_DEMO.md # Заметки для разработчиков
│ ├── DEMO_MODE_FILES.md # Этот файл
│ └── GIT_COMMIT_MESSAGE.txt # Шаблон commit message
│
├── docs/
│ ├── DEMO_MODE.md # Полная документация
│ ├── DEMO_MODE_EXAMPLE.md # Примеры использования
│ ├── DEMO_MODE_FAQ.md # Часто задаваемые вопросы
│ └── DEMO_MODE_FLOW.md # Схемы и диаграммы
│
├── 🗄️ migrations/
│ ├── 006_add_demo_users.sql # Миграция для MySQL
│ └── 006_add_demo_users_sqlite.sql # Миграция для SQLite
│
├── scripts/
│ └── apply_demo_mode_migration.go # Скрипт применения миграции
│
├── internal/
│ ├── models/
│ │ └── user.go # ✏️ Модель пользователя (изменен)
│ ├── controllers/
│ │ └── auth_controller.go # ✏️ Контроллер авторизации (изменен)
│ └── services/
│ └── integration_service.go # ✏️ Сервис интеграций (изменен)
│
├── templates/
│ └── pages/
│ └── login.html # ✏️ Страница логина (изменен)
│
└── Конфигурация
 ├── .env.example # ✏️ Пример конфигурации (изменен)
 └── README.md # ✏️ Главный README (изменен)
```

## 📖 Назначение файлов

### Документация для пользователей

| Файл | Назначение | Для кого |
|------|-----------|----------|
| `QUICK_START_DEMO.md` | Быстрый старт за 3 минуты | Новые пользователи |
| `DEMO_MODE_SETUP.md` | Пошаговая установка | Администраторы |
| `docs/DEMO_MODE.md` | Полная документация | Все пользователи |
| `docs/DEMO_MODE_EXAMPLE.md` | Примеры использования | Пользователи |
| `docs/DEMO_MODE_FAQ.md` | Часто задаваемые вопросы | Все |

### Документация для разработчиков

| Файл | Назначение | Для кого |
|------|-----------|----------|
| `DEVELOPER_NOTES_DEMO.md` | Архитектура и расширения | Разработчики |
| `docs/DEMO_MODE_FLOW.md` | Схемы и диаграммы | Разработчики |
| `CHANGELOG_DEMO_MODE.md` | Детальный changelog | Разработчики |
| `DEMO_MODE_SUMMARY.md` | Сводка изменений | Все |

### Технические файлы

| Файл | Назначение | Тип |
|------|-----------|-----|
| `migrations/006_add_demo_users.sql` | Миграция MySQL | SQL |
| `migrations/006_add_demo_users_sqlite.sql` | Миграция SQLite | SQL |
| `scripts/apply_demo_mode_migration.go` | Скрипт миграции | Go |

### Измененные файлы

| Файл | Изменения |
|------|-----------|
| `internal/models/user.go` | Добавлены поля IsDemo, ExpiresAt |
| `internal/controllers/auth_controller.go` | Добавлены функции регистрации и очистки |
| `internal/services/integration_service.go` | Добавлена валидация демо-режима |
| `templates/pages/login.html` | Автозаполнение и уведомления |
| `.env.example` | Настройки демо-режима |
| `README.md` | Информация о демо-режиме |
| `migrations/README.md` | Документация миграции 006 |

## С чего начать?

### Для новых пользователей
1. Читайте `QUICK_START_DEMO.md` - быстрый старт за 3 минуты
2. Затем `docs/DEMO_MODE_EXAMPLE.md` - примеры использования

### Для администраторов
1. Читайте `DEMO_MODE_SETUP.md` - пошаговая установка
2. Затем `docs/DEMO_MODE.md` - полная документация
3. При проблемах `docs/DEMO_MODE_FAQ.md` - FAQ

### Для разработчиков
1. Читайте `DEVELOPER_NOTES_DEMO.md` - архитектура
2. Затем `docs/DEMO_MODE_FLOW.md` - схемы работы
3. Изучите `CHANGELOG_DEMO_MODE.md` - детали изменений

## Статистика

- **Всего файлов**: 19 (12 новых + 7 измененных)
- **Строк документации**: ~3000
- **Строк кода**: ~300
- **Миграций**: 2 (MySQL + SQLite)
- **Примеров**: 10+

## Поиск информации

### Хочу узнать...

**Как установить?**
→ `DEMO_MODE_SETUP.md` или `QUICK_START_DEMO.md`

**Как использовать?**
→ `docs/DEMO_MODE_EXAMPLE.md`

**Как это работает?**
→ `docs/DEMO_MODE_FLOW.md`

**Что изменилось?**
→ `CHANGELOG_DEMO_MODE.md` или `DEMO_MODE_SUMMARY.md`

**Есть проблема**
→ `docs/DEMO_MODE_FAQ.md`

**Хочу расширить**
→ `DEVELOPER_NOTES_DEMO.md`

**Нужна полная информация**
→ `docs/DEMO_MODE.md`

## Шаблоны

### Git Commit
Используйте `GIT_COMMIT_MESSAGE.txt` как шаблон для commit message.

### Конфигурация
Скопируйте настройки из `.env.example`:
```env
DEMO_MODE=true
DEMO_SECRET=your-secret-key
DEMO_TARGET_URL=https://webhook.site/your-id
```

## Категории файлов

### По типу

**Markdown документация**: 12 файлов
- DEMO_MODE_SETUP.md
- QUICK_START_DEMO.md
- DEMO_MODE_SUMMARY.md
- CHANGELOG_DEMO_MODE.md
- DEVELOPER_NOTES_DEMO.md
- DEMO_MODE_FILES.md
- docs/DEMO_MODE.md
- docs/DEMO_MODE_EXAMPLE.md
- docs/DEMO_MODE_FAQ.md
- docs/DEMO_MODE_FLOW.md
- README.md (изменен)
- migrations/README.md (изменен)

**Go код**: 4 файла
- internal/models/user.go (изменен)
- internal/controllers/auth_controller.go (изменен)
- internal/services/integration_service.go (изменен)
- scripts/apply_demo_mode_migration.go

**SQL миграции**: 2 файла
- migrations/006_add_demo_users.sql
- migrations/006_add_demo_users_sqlite.sql

**HTML шаблоны**: 1 файл
- templates/pages/login.html (изменен)

**Конфигурация**: 1 файл
- .env.example (изменен)

**Текстовые файлы**: 1 файл
- GIT_COMMIT_MESSAGE.txt

### По назначению

**Установка и настройка**: 3 файла
- DEMO_MODE_SETUP.md
- QUICK_START_DEMO.md
- .env.example

**Использование**: 2 файла
- docs/DEMO_MODE.md
- docs/DEMO_MODE_EXAMPLE.md

**Справка**: 2 файла
- docs/DEMO_MODE_FAQ.md
- docs/DEMO_MODE_FLOW.md

**Разработка**: 2 файла
- DEVELOPER_NOTES_DEMO.md
- docs/DEMO_MODE_FLOW.md

**История изменений**: 2 файла
- CHANGELOG_DEMO_MODE.md
- DEMO_MODE_SUMMARY.md

**Миграции**: 3 файла
- migrations/006_add_demo_users.sql
- migrations/006_add_demo_users_sqlite.sql
- scripts/apply_demo_mode_migration.go

**Код**: 4 файла
- internal/models/user.go
- internal/controllers/auth_controller.go
- internal/services/integration_service.go
- templates/pages/login.html

## Приоритет чтения

### Уровень 1 (Обязательно)
1. `QUICK_START_DEMO.md` - быстрый старт
2. `docs/DEMO_MODE.md` - основная документация

### Уровень 2 (Рекомендуется)
3. `DEMO_MODE_SETUP.md` - детальная установка
4. `docs/DEMO_MODE_EXAMPLE.md` - примеры
5. `docs/DEMO_MODE_FAQ.md` - FAQ

### Уровень 3 (Опционально)
6. `docs/DEMO_MODE_FLOW.md` - схемы
7. `DEVELOPER_NOTES_DEMO.md` - для разработчиков
8. `CHANGELOG_DEMO_MODE.md` - детальный changelog

### Уровень 4 (Справочно)
9. `DEMO_MODE_SUMMARY.md` - сводка
10. `DEMO_MODE_FILES.md` - этот файл
11. `GIT_COMMIT_MESSAGE.txt` - шаблон commit

## Что включить в релиз

### Обязательные файлы
- Все файлы в `docs/DEMO_MODE*.md`
- Миграции в `migrations/006_*`
- Скрипт `scripts/apply_demo_mode_migration.go`
- Измененные файлы кода

### Рекомендуемые файлы
- `DEMO_MODE_SETUP.md`
- `QUICK_START_DEMO.md`
- `CHANGELOG_DEMO_MODE.md`

### Опциональные файлы
- `DEVELOPER_NOTES_DEMO.md`
- `DEMO_MODE_SUMMARY.md`
- `DEMO_MODE_FILES.md`
- `GIT_COMMIT_MESSAGE.txt`

## Обновление документации

При изменении функционала обновите:
1. `docs/DEMO_MODE.md` - основная документация
2. `CHANGELOG_DEMO_MODE.md` - добавьте запись
3. `docs/DEMO_MODE_FAQ.md` - если есть новые вопросы
4. `DEVELOPER_NOTES_DEMO.md` - если изменилась архитектура

## Чеклист для релиза

- [ ] Все файлы созданы
- [ ] Документация проверена
- [ ] Примеры протестированы
- [ ] Миграции работают
- [ ] Код компилируется
- [ ] README обновлен
- [ ] Changelog заполнен
- [ ] Git commit message подготовлен

---

**Версия**: 1.0 
**Дата**: 2024-12-01 
**Статус**: Готово
