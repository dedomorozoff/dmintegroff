# 🤝 Участие в развитии dmIntegroff

Мы приветствуем вклад в развитие проекта! Этот документ поможет вам начать.

## 🎯 Как помочь проекту

### 💡 Предложить идею
- Создайте [Issue](https://github.com/dedomorozoff/dmintegroff/issues) с тегом `enhancement`
- Опишите проблему и предложите решение
- Обсудите с сообществом

### 🐛 Сообщить о баге
- Создайте [Issue](https://github.com/dedomorozoff/dmintegroff/issues) с тегом `bug`
- Приложите скриншоты и логи
- Укажите шаги для воспроизведения

### 💻 Внести код
- Fork репозитория
- Создайте feature branch
- Внесите изменения
- Создайте Pull Request

### 📖 Улучшить документацию
- Исправьте опечатки
- Добавьте примеры
- Переведите на другие языки

## 🛠 Настройка окружения разработки

### Требования

- **Go 1.25+**
- **Git**
- **Редактор кода** (VS Code, GoLand, Vim)

### Клонирование и настройка

```bash
# Fork репозитория на GitHub, затем клонируйте свой fork
git clone https://github.com/YOUR_USERNAME/dmintegroff.git
cd dmIntegroff

# Добавьте upstream remote
git remote add upstream https://github.com/dedomorozoff/dmintegroff.git

# Установите зависимости
go mod tidy

# Скопируйте конфигурацию
cp .env.example .env
```

### Запуск в режиме разработки

```bash
# Запуск сервера с автоперезагрузкой (если установлен air)
air

# Или обычный запуск
go run cmd/server/main.go
```

### Запуск тестов

```bash
# Все тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Тесты конкретного пакета
go test ./internal/controllers/
```

## 📝 Стандарты кода

### Go код

```go
// Хорошо
func CreateIntegration(name string, targetURL string) (*Integration, error) {
    if name == "" {
        return nil, errors.New("name is required")
    }
    
    integration := &Integration{
        Name:      name,
        TargetURL: targetURL,
        CreatedAt: time.Now(),
    }
    
    return integration, nil
}

// Плохо
func create(n string, u string) *Integration {
    return &Integration{n, u, time.Now()}
}
```

### Комментарии

```go
// CreateIntegration создает новую интеграцию с валидацией параметров
func CreateIntegration(name string, targetURL string) (*Integration, error) {
    // Валидация обязательных полей
    if name == "" {
        return nil, errors.New("name is required")
    }
    
    // Создание объекта интеграции
    integration := &Integration{
        Name:      name,
        TargetURL: targetURL,
        CreatedAt: time.Now(),
    }
    
    return integration, nil
}
```

### HTML/CSS

```html
<!-- Хорошо: семантичная разметка -->
<article class="integration-card">
    <header class="integration-header">
        <h2 class="integration-title">{{.Name}}</h2>
        <span class="integration-status {{.Status}}">{{.Status}}</span>
    </header>
    <div class="integration-content">
        <p class="integration-description">{{.Description}}</p>
    </div>
</article>

<!-- Плохо: неясная структура -->
<div class="card">
    <div>{{.Name}} - {{.Status}}</div>
    <div>{{.Description}}</div>
</div>
```

### JavaScript

```javascript
// Хорошо: современный ES6+
class IntegrationManager {
    constructor(apiUrl) {
        this.apiUrl = apiUrl;
    }
    
    async createIntegration(data) {
        try {
            const response = await fetch(`${this.apiUrl}/integrations`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Failed to create integration:', error);
            throw error;
        }
    }
}

// Плохо: устаревший подход
function createIntegration(data) {
    $.post('/integrations', data, function(result) {
        console.log(result);
    });
}
```

## 🔄 Процесс разработки

### 1. Планирование

1. Выберите задачу из [Issues](https://github.com/dedomorozoff/dmintegroff/issues)
2. Обсудите подход в комментариях
3. Получите одобрение от мейнтейнеров

### 2. Разработка

```bash
# Создайте feature branch
git checkout -b feature/amazing-feature

# Внесите изменения
# ... код ...

# Коммитьте изменения
git add .
git commit -m "feat: add amazing feature"

# Отправьте в свой fork
git push origin feature/amazing-feature
```

### 3. Тестирование

```bash
# Запустите тесты
go test ./...

# Проверьте форматирование
go fmt ./...

# Проверьте линтером (если установлен)
golangci-lint run
```

### 4. Pull Request

1. Создайте PR из вашего fork в основной репозиторий
2. Заполните шаблон PR
3. Дождитесь ревью
4. Внесите правки при необходимости

## 📋 Шаблон Pull Request

```markdown
## Описание
Краткое описание изменений.

## Тип изменений
- [ ] Исправление бага
- [ ] Новая функция
- [ ] Улучшение производительности
- [ ] Рефакторинг
- [ ] Документация

## Тестирование
- [ ] Добавлены новые тесты
- [ ] Все тесты проходят
- [ ] Проверено вручную

## Скриншоты (если применимо)
Добавьте скриншоты изменений в UI.

## Чеклист
- [ ] Код соответствует стандартам проекта
- [ ] Добавлены комментарии к сложным участкам
- [ ] Обновлена документация
- [ ] Изменения протестированы
```

## 🏗 Архитектура проекта

### Структура кода

```
dmIntegroff/
├── cmd/                    # Точки входа
│   ├── server/            # Основной сервер
│   └── admin/             # CLI инструменты
├── internal/              # Внутренняя логика
│   ├── controllers/       # HTTP обработчики
│   ├── models/           # Модели данных
│   ├── services/         # Бизнес-логика
│   ├── middleware/       # Middleware
│   └── utils/            # Утилиты
├── templates/            # HTML шаблоны
├── static/              # Статические файлы
└── docs/               # Документация
```

### Принципы

1. **Разделение ответственности** - каждый пакет имеет четкую роль
2. **Dependency Injection** - зависимости передаются явно
3. **Тестируемость** - код легко покрывается тестами
4. **Читаемость** - код должен быть понятен без комментариев

## 🎨 UI/UX Guidelines

### Дизайн система

- **Цветовая схема:** Темная тема с акцентами
- **Типографика:** Системные шрифты
- **Компоненты:** Карточки, кнопки, формы
- **Адаптивность:** Mobile-first подход

### Принципы UX

1. **Простота** - минимум кликов для достижения цели
2. **Понятность** - интуитивная навигация
3. **Обратная связь** - пользователь всегда знает что происходит
4. **Доступность** - поддержка клавиатуры и скринридеров

## 🌍 Локализация

### Языки интерфейса

- **Основной:** Русский (обязательно)
- **Дополнительный:** Английский (желательно)

### Правила перевода

```go
// Хорошо: используйте константы для текста
const (
    MsgIntegrationCreated = "Интеграция успешно создана"
    MsgIntegrationFailed  = "Ошибка создания интеграции"
)

// Плохо: хардкод текста в коде
fmt.Println("Интеграция создана")
```

## 🚀 Приоритетные направления

### Высокий приоритет
- 🤖 **AI-ассистент** - революционная функция
- 🔍 **История запросов** - критично для продакшена
- 🔧 **Валидация конфигурации** - UX улучшение

### Средний приоритет
- 🎨 **Готовые шаблоны** - ускорение внедрения
- 🔄 **Retry механизм** - надежность
- 📊 **Расширенная аналитика** - инсайты

### Низкий приоритет
- 🐳 **Docker** - развертывание
- 🎨 **Визуальный редактор** - продвинутый UX
- 🔗 **Webhook цепочки** - сложные сценарии

## 📞 Связь с командой

- **GitHub Issues** - основной канал для обсуждений
- **Pull Requests** - для ревью кода
- **Email** - для приватных вопросов

## 📄 Лицензия

Внося вклад в проект, вы соглашаетесь с тем, что ваш код будет лицензирован под [MIT License](LICENSE).

---

**Спасибо за участие в развитии dmIntegroff! 🚀**