# 🤝 Участие в разработке Roadmap функций

Руководство для разработчиков, желающих внести вклад в реализацию функций из roadmap.

---

## 📋 Как выбрать задачу

1. Изучите [Roadmap](ROADMAP.md) и выберите интересующую функцию
2. Проверьте, что функция еще не в разработке (статус 📋 Запланировано)
3. Создайте issue с названием: `[Roadmap] Название функции`
4. Опишите ваш план реализации
5. Дождитесь одобрения от мейнтейнеров

---

## 🎯 Приоритетные задачи для новых контрибьюторов

### Легкие задачи (Good First Issue)

#### 1. Исправление бага в редакторе образца данных
**Сложность:** 🟢 Легкая  
**Время:** 2-4 часа  
**Файлы:**
- `templates/pages/integration_configure.html`
- `static/js/json-highlight.js`

**Что нужно сделать:**
- Исправить инициализацию редактора для пустого payload
- Улучшить обработку paste событий
- Исправить баг с вставкой HTML-мусора

**Тесты:**
1. Создать интеграцию без webhook
2. Открыть настройку маппинга
3. Вставить JSON из буфера обмена
4. Проверить корректность отображения

---

### Средние задачи

#### 2. Использование запроса в качестве образца данных
**Сложность:** 🟡 Средняя  
**Время:** 4-8 часов  
**Файлы:**
- `internal/controllers/webhook_test_controller.go`
- `templates/pages/webhook_test.html`
- Новый endpoint: `POST /api/webhook-test/:token/use-as-sample`

**Что нужно сделать:**
1. Добавить кнопку "Использовать как образец" в UI тестового хука
2. Создать endpoint для сохранения запроса как образца
3. Связать тестовый хук с интеграцией
4. Перенаправить на страницу настройки маппинга

**Дизайн-документ:** [ROADMAP.md - Пункт 1](ROADMAP.md#1-использование-запроса-в-качестве-образца-данных-в-тестовом-хуке)

---

#### 3. Тестирование запроса после настройки маппинга
**Сложность:** 🟡 Средняя  
**Время:** 6-10 часов  
**Файлы:**
- `templates/pages/integration_configure.html`
- `internal/controllers/integration_controller.go`
- Новый endpoint: `POST /api/integrations/:id/test-mapping`

**Что нужно сделать:**
1. Добавить кнопку "Протестировать запрос"
2. Создать модальное окно с результатами
3. Реализовать endpoint для тестирования
4. Добавить кнопку "Посмотреть в тесте интеграций"

**Дизайн-документ:** [ROADMAP.md - Пункт 3](ROADMAP.md#3-тестирование-запроса-после-настройки-маппинга)

---

### Сложные задачи

#### 4. Поддержка form-data и XML
**Сложность:** 🔴 Сложная  
**Время:** 16-24 часа  
**Файлы:**
- `internal/controllers/integration_controller.go`
- `internal/utils/parser.go` (новый)
- `internal/models/integration.go`
- `templates/pages/integration_configure.html`

**Что нужно сделать:**
1. Создать парсеры для form-data и XML
2. Добавить автоопределение Content-Type
3. Преобразование в JSON для маппинга
4. UI для выбора формата входных данных
5. Тесты для всех форматов

**Дизайн-документ:** [ROADMAP.md - Пункт 2](ROADMAP.md#2-поддержка-различных-форматов-входящих-данных)

---

#### 5. Множественные маппинги на один webhook
**Сложность:** 🔴 Очень сложная  
**Время:** 40-60 часов  
**Релиз:** v2.0.0  
**Файлы:**
- Множество файлов (см. дизайн-документ)
- Новая модель `IntegrationOutput`
- Миграция БД
- Новые страницы UI

**Что нужно сделать:**
1. Создать модель данных
2. Миграция БД
3. CRUD операции для выходов
4. Логика обработки множественных выходов
5. UI для управления выходами
6. Условное выполнение
7. Параллельная обработка
8. Логирование и статистика

**Дизайн-документ:** [MULTIPLE_OUTPUTS_DESIGN.md](MULTIPLE_OUTPUTS_DESIGN.md)

---

## 🛠️ Процесс разработки

### 1. Подготовка

```bash
# Форкнуть репозиторий
git clone https://github.com/your-username/dmintegroff.git
cd dmintegroff

# Создать ветку для функции
git checkout -b feature/roadmap-function-name

# Установить зависимости
go mod download
```

### 2. Разработка

1. Изучите существующий код
2. Следуйте стилю кодирования проекта
3. Пишите тесты для новой функциональности
4. Обновите документацию

### 3. Тестирование

```bash
# Запустить тесты
go test ./...

# Запустить приложение
go run cmd/server/main.go

# Проверить функциональность вручную
```

### 4. Коммит и PR

```bash
# Коммит изменений
git add .
git commit -m "[Roadmap] Название функции: краткое описание"

# Пуш в форк
git push origin feature/roadmap-function-name

# Создать Pull Request на GitHub
```

### 5. Code Review

- Дождитесь review от мейнтейнеров
- Внесите необходимые изменения
- После одобрения - merge в main

---

## 📝 Стандарты кодирования

### Go код

```go
// Комментарии на русском языке
// Функции с понятными названиями
func ProcessWebhookWithOutputs(integrationID uint, payload map[string]interface{}) error {
    // Логирование важных событий
    logger.Log.WithFields(map[string]interface{}{
        "integration_id": integrationID,
        "payload_size":   len(payload),
    }).Info("Processing webhook with multiple outputs")
    
    // Обработка ошибок
    if err := validatePayload(payload); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }
    
    // ...
}
```

### HTML/JavaScript

```html
<!-- Комментарии на русском -->
<!-- Использование Lucide иконок -->
<button class="btn btn-primary" onclick="testMapping()">
    <i data-lucide="play-circle"></i> Протестировать запрос
</button>

<script>
// Функции с понятными названиями
function testMapping() {
    // Валидация перед отправкой
    if (!validateForm()) {
        showNotification('Исправьте ошибки в форме', 'error');
        return;
    }
    
    // Отправка запроса
    fetch(`/api/integrations/${integrationId}/test-mapping`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({...})
    })
    .then(response => response.json())
    .then(data => showTestResults(data))
    .catch(error => showNotification('Ошибка тестирования', 'error'));
}
</script>
```

---

## 🧪 Тестирование

### Unit тесты

```go
func TestProcessWebhookWithOutputs(t *testing.T) {
    // Arrange
    integration := createTestIntegration()
    payload := map[string]interface{}{
        "user": map[string]interface{}{
            "name": "John Doe",
            "email": "john@example.com",
        },
    }
    
    // Act
    err := ProcessWebhookWithOutputs(integration.ID, payload)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 3, len(getProcessedOutputs(integration.ID)))
}
```

### Integration тесты

```go
func TestWebhookEndToEnd(t *testing.T) {
    // Создать тестовую интеграцию
    integration := createTestIntegration()
    
    // Отправить webhook
    response := sendTestWebhook(integration.WebhookToken, testPayload)
    
    // Проверить результат
    assert.Equal(t, 200, response.StatusCode)
    assert.True(t, allOutputsProcessed(integration.ID))
}
```

---

## 📚 Полезные ресурсы

### Документация проекта
- [README](../README.md) - основная документация
- [Roadmap](ROADMAP.md) - детальный roadmap
- [Changelog](CHANGELOG.md) - история изменений
- [Technical Docs](TECHNICAL_DOCS.md) - техническая документация
- [Programmer Guide](PROGRAMMER_GUIDE.md) - руководство программиста

### Внешние ресурсы
- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM](https://gorm.io/docs/)
- [Lucide Icons](https://lucide.dev/)

---

## 💬 Связь с командой

- **GitHub Issues** - для обсуждения функций и багов
- **Pull Requests** - для code review
- **Discussions** - для общих вопросов

---

## 🎉 Благодарности

Спасибо всем контрибьюторам, которые помогают развивать dmIntegroff!

Ваш вклад делает проект лучше для всех пользователей.

---

**Последнее обновление:** 8 декабря 2024
