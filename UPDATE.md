# 🔄 Обновление dmIntegroff

## После `git pull` выполните:

```bash
# 1. Применить миграции
.\dmIntegroff-admin.exe --migrate

# 2. Пересобрать
go build -o dmIntegroff.exe cmd/server/main.go

# 3. Перезапустить
.\dmintegroff.exe
```

## Проверка

```powershell
# Тест логирования
.\tests\test_logs.ps1

# Открыть приложение
start http://localhost:8080
```

## Что делает миграция

GORM AutoMigrate:
- ✅ Создает новые таблицы
- ✅ Добавляет новые колонки
- ✅ Сохраняет все данные
- ❌ НЕ удаляет старые колонки

## Откат

```bash
# Восстановить БД из бэкапа
copy dmintegroff.db.backup dmintegroff.db

# Откатить код
git checkout <commit>

# Пересобрать
go build -o dmintegroff.exe cmd/server/main.go
```

## Документация

- [CHANGELOG.md](CHANGELOG.md) - История изменений
- [CLI.md](CLI.md) - CLI инструмент
- [DEVELOPMENT.md](DEVELOPMENT.md) - Для разработчиков
