# Экспорт в Git репозиторий

## Обзор

Функция экспорта в Git позволяет автоматически сохранять конфигурации интеграций в Git репозиторий с автоматическим коммитом и опциональным push.

## Возможности

- ✅ Экспорт напрямую в локальный Git репозиторий
- ✅ Автоматический коммит с настраиваемым сообщением
- ✅ Опциональный автоматический push
- ✅ Выбор ветки для коммита
- ✅ Настраиваемый путь к файлу в репозитории
- ✅ Проверка наличия изменений (не создает пустые коммиты)

## Использование

### Через веб-интерфейс

1. Выберите интеграции с помощью чекбоксов
2. Нажмите кнопку **Экспорт в Git**
3. Заполните форму:
   - **Путь к Git репозиторию** - абсолютный путь (например: `/home/user/my-repo`)
   - **Путь к файлу** - относительный путь в репозитории (например: `configs/integrations.json`)
   - **Ветка** - ветка для коммита (по умолчанию: `main`)
   - **Сообщение коммита** - опционально, будет сгенерировано автоматически
   - **Автоматический push** - отправить изменения на remote
4. Нажмите **Экспортировать в Git**

### Через API

```bash
curl -X POST "http://localhost:8080/export/git/integrations?ids=1,2,3" \
  -H "Cookie: mysession=..." \
  -F "repo_path=/path/to/repo" \
  -F "file_path=configs/integrations.json" \
  -F "branch=main" \
  -F "auto_push=true"
```


## Примеры

### Экспорт с автоматическим push

```bash
curl -X POST "http://localhost:8080/export/git/integrations?ids=1,2,3" \
  -F "repo_path=/home/user/config-repo" \
  -F "file_path=production/integrations.json" \
  -F "branch=main" \
  -F "commit_msg=Update production integrations" \
  -F "auto_push=true"
```

### Экспорт проекта

```bash
curl -X POST "http://localhost:8080/export/git/projects/1" \
  -F "repo_path=/home/user/config-repo" \
  -F "file_path=projects/project-1.json" \
  -F "branch=develop"
```

## Workflow примеры

### CI/CD интеграция

```bash
#!/bin/bash
# Автоматический экспорт конфигураций в Git при изменениях

REPO_PATH="/var/config-repo"
FILE_PATH="configs/integrations.json"

# Экспорт в Git
curl -X POST "http://localhost:8080/export/git/integrations?ids=1,2,3" \
  -F "repo_path=$REPO_PATH" \
  -F "file_path=$FILE_PATH" \
  -F "branch=main" \
  -F "auto_push=true"
```

### Версионирование по окружениям

```bash
# Development
curl -X POST "http://localhost:8080/export/git/projects/1" \
  -F "repo_path=/repo" \
  -F "file_path=dev/integrations.json" \
  -F "branch=develop"

# Staging
curl -X POST "http://localhost:8080/export/git/projects/2" \
  -F "repo_path=/repo" \
  -F "file_path=staging/integrations.json" \
  -F "branch=staging"

# Production
curl -X POST "http://localhost:8080/export/git/projects/3" \
  -F "repo_path=/repo" \
  -F "file_path=production/integrations.json" \
  -F "branch=main" \
  -F "auto_push=true"
```

## Требования

- Git должен быть установлен на сервере
- Репозиторий должен быть инициализирован (`git init`)
- Для auto_push: настроен remote и есть права на push

## Безопасность

⚠️ **Важно:**
- Файлы содержат чувствительные данные (OAuth секреты, токены)
- Используйте приватные репозитории
- Настройте `.gitignore` если нужно исключить секреты
- Рассмотрите использование Git-crypt или git-secret для шифрования

## Troubleshooting

### Ошибка: "not a git repository"
Убедитесь, что путь указывает на корень Git репозитория (где находится `.git`)

### Ошибка: "git push failed"
Проверьте:
- Настроен ли remote: `git remote -v`
- Есть ли права на push
- Правильно ли настроена аутентификация (SSH keys или credentials)

### Ошибка: "repository path does not exist"
Проверьте абсолютный путь к репозиторию

## См. также

- [Экспорт/Импорт конфигураций](EXPORT_IMPORT.md)
- [OAuth 2.0 Guide](OAUTH_GUIDE.md)
- [Webhook Signatures](WEBHOOK_SIGNATURES.md)
