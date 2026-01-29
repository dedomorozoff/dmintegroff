# Сборка с встроенными файлами (go:embed)

Проект dmIntegroff теперь поддерживает сборку с встроенными статическими файлами и шаблонами прямо в исполняемый файл.

## Преимущества

- **Единый файл**: Все статические файлы, шаблоны и integration-templates встроены в исполняемый файл
- **Простой деплой**: Нужно скопировать только один .exe файл
- **Безопасность**: Файлы не могут быть случайно изменены или удалены
- **Производительность**: Быстрая загрузка файлов из памяти

## Скрипты сборки

### build-embedded.bat
Создает production-версию с встроенными файлами:
```bash
.\build-embedded.bat
```

Результат: `dmintegroff.exe` (~48MB) - полностью автономный исполняемый файл.

### build-dev.bat  
Создает dev-версию без встроенных файлов (для разработки):
```bash
.\build-dev.bat
```

Результат: `dmintegroff-dev.exe` (~30MB) - требует наличия папок static/, templates/, integration-templates/.

## Архитектура

### Пакет internal/assets
- `embed.go` - содержит `go:embed` директивы и функции для работы с встроенными файлами
- Автоматически копирует файлы из корневых папок при сборке

### Изменения в коде
- `cmd/server/main.go` - использует `assets.GetStaticFS()`, `assets.LoadHTMLTemplates()`
- `internal/routes/routes.go` - принимает встроенные файлы как параметры

## Разработка

Во время разработки рекомендуется:

1. **Для быстрой разработки**: `go run cmd/server/main.go` (использует обычные файлы)
2. **Для тестирования production**: `.\build-embedded.bat && .\dmintegroff.exe`

## Обновление встроенных файлов

При изменении файлов в static/, templates/, integration-templates/:

1. Запустите `.\build-embedded.bat` для пересборки
2. Файлы автоматически скопируются в internal/assets/ и встроятся в бинарник

## Структура файлов

```
dmintegroff/
├── static/                          # Исходные статические файлы
├── templates/                       # Исходные HTML шаблоны  
├── integration-templates/           # Исходные шаблоны интеграций
├── internal/assets/                 # Копии для go:embed
│   ├── embed.go                     # go:embed директивы
│   ├── static/                      # Копия static/
│   ├── templates/                   # Копия templates/
│   └── integration-templates/       # Копия integration-templates/
├── build-embedded.bat               # Сборка production
├── build-dev.bat                    # Сборка dev
└── dmintegroff.exe                  # Production бинарник
```

## Примечания

- Размер исполняемого файла увеличивается на ~18MB из-за встроенных файлов
- При изменении статических файлов нужна пересборка
- В production рекомендуется использовать только embedded версию
- Dev версия полезна для быстрой разработки и отладки шаблонов