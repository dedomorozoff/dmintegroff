# Руководство по go:embed в dmIntegroff

## Обзор

Проект dmIntegroff теперь поддерживает встраивание статических файлов прямо в исполняемый файл с помощью `go:embed`. Это значительно упрощает деплой и делает приложение полностью автономным.

## Преимущества go:embed

✅ **Единый исполняемый файл** - все ресурсы встроены в бинарник  
✅ **Простой деплой** - нужно скопировать только один .exe файл  
✅ **Безопасность** - файлы не могут быть изменены извне  
✅ **Производительность** - быстрая загрузка из памяти  
✅ **Надежность** - нет проблем с отсутствующими файлами  

## Структура проекта

```
dmintegroff/
├── static/                          # Исходные статические файлы
├── templates/                       # Исходные HTML шаблоны  
├── integration-templates/           # Исходные шаблоны интеграций
├── internal/assets/                 # Пакет для go:embed
│   ├── embed.go                     # go:embed директивы
│   ├── static/                      # Копия для встраивания
│   ├── templates/                   # Копия для встраивания
│   └── integration-templates/       # Копия для встраивания
├── build-embedded.bat               # Сборка production
├── run-dev.bat                      # Запуск dev режима
└── dmintegroff.exe                  # Production бинарник
```

## Режимы работы

### 1. Production режим (embedded)
```bash
# Сборка с встроенными файлами
.\build-embedded.bat

# Запуск
.\dmintegroff.exe
```

**Характеристики:**
- Размер: ~48MB (включает все ресурсы)
- Автономность: 100%
- Скорость: Максимальная
- Использование: Production, деплой

### 2. Development режим (filesystem)
```bash
# Запуск с файловой системой
set USE_EMBEDDED_FILES=false
go run cmd/server/main.go

# Или через скрипт
.\run-dev.bat
```

**Характеристики:**
- Размер: ~30MB (без ресурсов)
- Автономность: Требует папки static/, templates/, integration-templates/
- Скорость: Быстрая разработка (нет пересборки при изменении шаблонов)
- Использование: Разработка, отладка

## Переменные окружения

```bash
# Использовать встроенные файлы (по умолчанию true)
USE_EMBEDDED_FILES=true   # Production режим
USE_EMBEDDED_FILES=false  # Development режим
```

## Архитектура кода

### internal/assets/embed.go
```go
//go:embed static
var StaticFiles embed.FS

//go:embed templates  
var TemplateFiles embed.FS

//go:embed integration-templates
var IntegrationTemplates embed.FS
```

### cmd/server/main.go
```go
// Автоматическое переключение режимов
useEmbedded := os.Getenv("USE_EMBEDDED_FILES")
if useEmbedded == "" {
    useEmbedded = "true" // По умолчанию embedded
}

if useEmbedded == "true" {
    // Используем встроенные файлы
    staticFS = assets.GetStaticFS()
    htmlTemplates = assets.LoadHTMLTemplates()
} else {
    // Используем файловую систему
    staticFS = os.DirFS("static")
    htmlTemplates = nil // LoadHTMLGlob в routes
}
```

### internal/routes/routes.go
```go
// Поддержка обоих режимов
if staticFS != nil {
    r.StaticFS("/static", http.FS(staticFS))  // Embedded
} else {
    r.Static("/static", "./static")           // Filesystem
}

if htmlTemplates != nil {
    r.SetHTMLTemplate(htmlTemplates)          // Embedded
} else {
    r.LoadHTMLGlob("templates/*/*")          // Filesystem
}
```

## Workflow разработки

### Быстрая разработка
```bash
# 1. Запуск в dev режиме
.\run-dev.bat

# 2. Редактирование шаблонов/CSS
# Изменения видны сразу после перезагрузки страницы

# 3. Тестирование изменений
# Перезапуск сервера: Ctrl+C, .\run-dev.bat
```

### Подготовка к деплою
```bash
# 1. Финальные изменения в static/, templates/
# 2. Сборка production версии
.\build-embedded.bat

# 3. Тестирование embedded версии
.\dmintegroff.exe

# 4. Деплой одного файла
# Копируем dmintegroff.exe на сервер
```

## Сравнение размеров

| Режим | Размер | Зависимости | Скорость загрузки |
|-------|--------|-------------|-------------------|
| Embedded | ~48MB | Нет | Максимальная |
| Filesystem | ~30MB | static/, templates/ | Быстрая |

## Лучшие практики

### Для разработки
- Используйте `USE_EMBEDDED_FILES=false`
- Редактируйте файлы в исходных папках (static/, templates/)
- Перезапускайте сервер только при изменении Go кода

### Для production
- Всегда используйте `.\build-embedded.bat`
- Тестируйте embedded версию перед деплоем
- Деплойте только dmintegroff.exe

### Для CI/CD
```bash
# В pipeline
.\build-embedded.bat
# Тестирование
.\dmintegroff.exe &
# Деплой dmintegroff.exe
```

## Устранение проблем

### Проблема: Изменения в шаблонах не видны
**Решение:** Убедитесь что `USE_EMBEDDED_FILES=false` или пересоберите с `.\build-embedded.bat`

### Проблема: Файлы не найдены в embedded режиме
**Решение:** Проверьте что файлы скопированы в `internal/assets/` при сборке

### Проблема: Большой размер исполняемого файла
**Решение:** Это нормально для embedded режима. Для dev используйте filesystem режим

## Заключение

Система `go:embed` в dmIntegroff обеспечивает:
- Гибкость разработки (filesystem режим)
- Простоту деплоя (embedded режим)  
- Автоматическое переключение режимов
- Обратную совместимость

Используйте filesystem режим для разработки и embedded для production!