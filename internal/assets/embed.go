package assets

import (
	"embed"
	"html/template"
	"io/fs"
	"strings"
)

//go:embed static
var StaticFiles embed.FS

//go:embed templates
var TemplateFiles embed.FS

//go:embed integration-templates
var IntegrationTemplates embed.FS

// GetStaticFS возвращает файловую систему для статических файлов
func GetStaticFS() fs.FS {
	staticFS, err := fs.Sub(StaticFiles, "static")
	if err != nil {
		panic("failed to create static sub filesystem: " + err.Error())
	}
	return staticFS
}

// GetIntegrationTemplatesFS возвращает файловую систему для шаблонов интеграций
func GetIntegrationTemplatesFS() fs.FS {
	templatesFS, err := fs.Sub(IntegrationTemplates, "integration-templates")
	if err != nil {
		panic("failed to create integration-templates sub filesystem: " + err.Error())
	}
	return templatesFS
}

// LoadHTMLTemplates загружает HTML шаблоны из встроенной файловой системы
func LoadHTMLTemplates() *template.Template {
	tmpl := template.New("")
	
	// Проходим по всем файлам в templates/
	err := fs.WalkDir(TemplateFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Обрабатываем только .html файлы
		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			// Читаем содержимое файла
			content, err := fs.ReadFile(TemplateFiles, path)
			if err != nil {
				return err
			}
			
			// Создаем имя шаблона (убираем префикс "templates/")
			templateName := strings.TrimPrefix(path, "templates/")
			
			// Парсим шаблон
			_, err = tmpl.New(templateName).Parse(string(content))
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		panic("failed to load HTML templates: " + err.Error())
	}
	
	return tmpl
}

// GetTemplateNames возвращает список всех доступных шаблонов для отладки
func GetTemplateNames() []string {
	var names []string
	
	err := fs.WalkDir(TemplateFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			templateName := strings.TrimPrefix(path, "templates/")
			names = append(names, templateName)
		}
		
		return nil
	})
	
	if err != nil {
		panic("failed to get template names: " + err.Error())
	}
	
	return names
}