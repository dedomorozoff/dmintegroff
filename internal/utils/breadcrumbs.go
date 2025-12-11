package utils

import (
	"fmt"
	"strings"
)

// Breadcrumb представляет элемент навигационной цепочки
type Breadcrumb struct {
	Title  string `json:"title"`
	URL    string `json:"url,omitempty"`
	Icon   string `json:"icon,omitempty"`
	Active bool   `json:"active"`
}

// BreadcrumbBuilder помогает строить навигационные цепочки
type BreadcrumbBuilder struct {
	crumbs []Breadcrumb
}

// NewBreadcrumbBuilder создает новый builder
func NewBreadcrumbBuilder() *BreadcrumbBuilder {
	return &BreadcrumbBuilder{
		crumbs: make([]Breadcrumb, 0),
	}
}

// Add добавляет элемент в цепочку
func (b *BreadcrumbBuilder) Add(title, url string) *BreadcrumbBuilder {
	b.crumbs = append(b.crumbs, Breadcrumb{
		Title: title,
		URL:   url,
	})
	return b
}

// AddWithIcon добавляет элемент с иконкой
func (b *BreadcrumbBuilder) AddWithIcon(title, url, icon string) *BreadcrumbBuilder {
	b.crumbs = append(b.crumbs, Breadcrumb{
		Title: title,
		URL:   url,
		Icon:  icon,
	})
	return b
}

// AddActive добавляет активный (текущий) элемент
func (b *BreadcrumbBuilder) AddActive(title string) *BreadcrumbBuilder {
	b.crumbs = append(b.crumbs, Breadcrumb{
		Title:  title,
		Active: true,
	})
	return b
}

// AddActiveWithIcon добавляет активный элемент с иконкой
func (b *BreadcrumbBuilder) AddActiveWithIcon(title, icon string) *BreadcrumbBuilder {
	b.crumbs = append(b.crumbs, Breadcrumb{
		Title:  title,
		Icon:   icon,
		Active: true,
	})
	return b
}

// Build возвращает готовую цепочку
func (b *BreadcrumbBuilder) Build() []Breadcrumb {
	return b.crumbs
}

// Предустановленные цепочки для основных страниц

// DashboardBreadcrumbs возвращает хлебные крошки для главной страницы
func DashboardBreadcrumbs() []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddActiveWithIcon("Главная", "home").
		Build()
}

// ProjectsBreadcrumbs возвращает хлебные крошки для страницы проектов
func ProjectsBreadcrumbs() []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddActiveWithIcon("Проекты", "folder-kanban").
		Build()
}

// ProjectBreadcrumbs возвращает хлебные крошки для конкретного проекта
func ProjectBreadcrumbs(projectName string, projectID uint) []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddWithIcon("Проекты", "/projects", "folder-kanban").
		AddActive(projectName).
		Build()
}

// ProjectCreateBreadcrumbs возвращает хлебные крошки для создания проекта
func ProjectCreateBreadcrumbs() []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddWithIcon("Проекты", "/projects", "folder-kanban").
		AddActiveWithIcon("Создание проекта", "plus").
		Build()
}

// IntegrationsBreadcrumbs возвращает хлебные крошки для страницы интеграций
func IntegrationsBreadcrumbs() []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddActiveWithIcon("Интеграции", "webhook").
		Build()
}

// IntegrationBreadcrumbs возвращает хлебные крошки для конкретной интеграции
func IntegrationBreadcrumbs(integrationName string, integrationID uint, projectName string, projectID uint) []Breadcrumb {
	builder := NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home")

	if projectName != "" {
		builder.AddWithIcon("Проекты", "/projects", "folder-kanban").
			AddWithIcon(projectName, fmt.Sprintf("/projects/%d", projectID), "folder")
	} else {
		builder.AddWithIcon("Интеграции", "/integrations", "webhook")
	}

	return builder.AddActive(integrationName).Build()
}

// IntegrationCreateBreadcrumbs возвращает хлебные крошки для создания интеграции
func IntegrationCreateBreadcrumbs(projectName string, projectID uint) []Breadcrumb {
	builder := NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home")

	if projectName != "" {
		builder.AddWithIcon("Проекты", "/projects", "folder-kanban").
			AddWithIcon(projectName, fmt.Sprintf("/projects/%d", projectID), "folder").
			AddActiveWithIcon("Создание интеграции", "plus")
	} else {
		builder.AddWithIcon("Интеграции", "/integrations", "webhook").
			AddActiveWithIcon("Создание интеграции", "plus")
	}

	return builder.Build()
}

// IntegrationConfigureBreadcrumbs возвращает хлебные крошки для настройки интеграции
func IntegrationConfigureBreadcrumbs(integrationName string, integrationID uint, projectName string, projectID uint) []Breadcrumb {
	builder := NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home")

	if projectName != "" {
		builder.AddWithIcon("Проекты", "/projects", "folder-kanban").
			AddWithIcon(projectName, fmt.Sprintf("/projects/%d", projectID), "folder").
			AddWithIcon(integrationName, fmt.Sprintf("/integrations/%d/edit", integrationID), "webhook")
	} else {
		builder.AddWithIcon("Интеграции", "/integrations", "webhook").
			AddWithIcon(integrationName, fmt.Sprintf("/integrations/%d/edit", integrationID), "webhook")
	}

	return builder.AddActiveWithIcon("Настройка маппинга", "settings").Build()
}

// IntegrationOutputsBreadcrumbs возвращает хлебные крошки для управления выходами
func IntegrationOutputsBreadcrumbs(integrationName string, integrationID uint, projectName string, projectID uint) []Breadcrumb {
	builder := NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home")

	if projectName != "" {
		builder.AddWithIcon("Проекты", "/projects", "folder-kanban").
			AddWithIcon(projectName, fmt.Sprintf("/projects/%d", projectID), "folder").
			AddWithIcon(integrationName, fmt.Sprintf("/integrations/%d/edit", integrationID), "webhook")
	} else {
		builder.AddWithIcon("Интеграции", "/integrations", "webhook").
			AddWithIcon(integrationName, fmt.Sprintf("/integrations/%d/edit", integrationID), "webhook")
	}

	return builder.AddActiveWithIcon("Управление выходами", "git-branch").Build()
}

// LogsBreadcrumbs возвращает хлебные крошки для страницы логов
func LogsBreadcrumbs() []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddActiveWithIcon("Логи", "file-text").
		Build()
}

// SettingsBreadcrumbs возвращает хлебные крошки для настроек
func SettingsBreadcrumbs() []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddActiveWithIcon("Настройки", "settings").
		Build()
}

// HelpBreadcrumbs возвращает хлебные крошки для справки
func HelpBreadcrumbs() []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddActiveWithIcon("Справка", "help-circle").
		Build()
}

// WebhookTestBreadcrumbs возвращает хлебные крошки для тестового webhook
func WebhookTestBreadcrumbs(token string) []Breadcrumb {
	return NewBreadcrumbBuilder().
		AddWithIcon("Главная", "/dashboard", "home").
		AddActiveWithIcon("Тестовый Webhook", "flask-conical").
		Build()
}

// GetBreadcrumbsForPath автоматически определяет хлебные крошки по пути
func GetBreadcrumbsForPath(path string) []Breadcrumb {
	// Убираем ведущий слэш и разбиваем путь
	path = strings.TrimPrefix(path, "/")
	segments := strings.Split(path, "/")

	if len(segments) == 0 || segments[0] == "" {
		return DashboardBreadcrumbs()
	}

	switch segments[0] {
	case "dashboard":
		return DashboardBreadcrumbs()
	case "projects":
		if len(segments) == 1 {
			return ProjectsBreadcrumbs()
		}
		// Для конкретных проектов нужна дополнительная информация из БД
		return ProjectsBreadcrumbs()
	case "integrations":
		if len(segments) == 1 {
			return IntegrationsBreadcrumbs()
		}
		// Для конкретных интеграций нужна дополнительная информация из БД
		return IntegrationsBreadcrumbs()
	case "logs":
		return LogsBreadcrumbs()
	case "settings":
		return SettingsBreadcrumbs()
	case "help":
		return HelpBreadcrumbs()
	case "webhook-test":
		if len(segments) > 1 {
			return WebhookTestBreadcrumbs(segments[1])
		}
		return DashboardBreadcrumbs()
	default:
		return DashboardBreadcrumbs()
	}
}