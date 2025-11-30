package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// TemplateProcessor обрабатывает JSON шаблоны с подстановкой значений
type TemplateProcessor struct {
	placeholderRegex *regexp.Regexp
}

// NewTemplateProcessor создает новый процессор шаблонов
func NewTemplateProcessor() *TemplateProcessor {
	return &TemplateProcessor{
		placeholderRegex: regexp.MustCompile(`\{\{([^}]+)\}\}`),
	}
}

// ProcessTemplate обрабатывает шаблон, подставляя значения из sourceData
// template - JSON строка с плейсхолдерами вида {{field.path}}
// sourceData - исходные данные для подстановки
func (tp *TemplateProcessor) ProcessTemplate(template string, sourceData map[string]interface{}) (map[string]interface{}, error) {
	if template == "" {
		return nil, fmt.Errorf("template is empty")
	}

	// Заменяем все плейсхолдеры на значения
	processedTemplate := tp.replacePlaceholders(template, sourceData)

	// Парсим результат в JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(processedTemplate), &result); err != nil {
		return nil, fmt.Errorf("failed to parse processed template: %w", err)
	}

	return result, nil
}

// replacePlaceholders заменяет все плейсхолдеры {{field.path}} на значения из sourceData
func (tp *TemplateProcessor) replacePlaceholders(template string, sourceData map[string]interface{}) string {
	// Используем расширенный regex для захвата кавычек вокруг плейсхолдера
	extendedRegex := regexp.MustCompile(`"?\{\{([^}]+)\}\}"?`)
	
	type replacement struct {
		start int
		end   int
		value string
	}
	
	var replacements []replacement
	matches := extendedRegex.FindAllStringSubmatchIndex(template, -1)
	
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		
		// Извлекаем путь из группы захвата
		path := strings.TrimSpace(template[match[2]:match[3]])
		
		// Проверяем, есть ли кавычки вокруг плейсхолдера
		fullMatch := template[match[0]:match[1]]
		hasQuotes := strings.HasPrefix(fullMatch, `"`) && strings.HasSuffix(fullMatch, `"`)
		
		// Проверяем, это запрос на весь массив (path.*)
		isArrayWildcard := strings.HasSuffix(path, ".*")
		if isArrayWildcard {
			// Убираем .* из пути
			path = strings.TrimSuffix(path, ".*")
		}
		
		// Получаем значение по пути
		value, err := GetValueByPath(sourceData, path)
		if err != nil {
			replacements = append(replacements, replacement{
				start: match[0],
				end:   match[1],
				value: "null",
			})
			continue
		}
		
		// Конвертируем значение в JSON
		jsonValue, err := json.Marshal(value)
		if err != nil {
			replacements = append(replacements, replacement{
				start: match[0],
				end:   match[1],
				value: "null",
			})
			continue
		}
		
		jsonStr := string(jsonValue)
		
		// Если это массив с wildcard, вставляем весь массив без кавычек
		if isArrayWildcard {
			replacements = append(replacements, replacement{
				start: match[0],
				end:   match[1],
				value: jsonStr, // Вставляем массив как есть
			})
			continue
		}
		
		// Если в шаблоне уже есть кавычки и значение - строка,
		// убираем кавычки из JSON значения
		if hasQuotes && strings.HasPrefix(jsonStr, `"`) && strings.HasSuffix(jsonStr, `"`) {
			// Оставляем кавычки из шаблона, убираем из значения
			jsonStr = jsonStr[1 : len(jsonStr)-1]
			// Но нужно вернуть кавычки обратно, так как они уже в шаблоне
			replacements = append(replacements, replacement{
				start: match[0] + 1, // Пропускаем открывающую кавычку
				end:   match[1] - 1, // Пропускаем закрывающую кавычку
				value: jsonStr,
			})
		} else {
			// Для чисел, булевых и т.д. просто заменяем
			replacements = append(replacements, replacement{
				start: match[0],
				end:   match[1],
				value: jsonStr,
			})
		}
	}
	
	// Применяем замены в обратном порядке
	result := template
	for i := len(replacements) - 1; i >= 0; i-- {
		r := replacements[i]
		result = result[:r.start] + r.value + result[r.end:]
	}
	
	return result
}

// ValidateTemplate проверяет корректность шаблона
func (tp *TemplateProcessor) ValidateTemplate(template string) error {
	if template == "" {
		return fmt.Errorf("template is empty")
	}

	// Проверяем, что это валидный JSON
	// Заменяем плейсхолдеры на тестовые значения
	extendedRegex := regexp.MustCompile(`"?\{\{([^}]+)\}\}"?`)
	
	testTemplate := extendedRegex.ReplaceAllStringFunc(template, func(match string) string {
		// Проверяем, есть ли кавычки вокруг плейсхолдера
		hasQuotes := strings.HasPrefix(match, `"`) && strings.HasSuffix(match, `"`)
		
		if hasQuotes {
			// Для строк возвращаем значение в кавычках
			return `"test_value"`
		}
		// Для чисел/булевых возвращаем число
		return `123`
	})

	var testData interface{}
	if err := json.Unmarshal([]byte(testTemplate), &testData); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	return nil
}

// ExtractPlaceholders извлекает все плейсхолдеры из шаблона
func (tp *TemplateProcessor) ExtractPlaceholders(template string) []string {
	matches := tp.placeholderRegex.FindAllStringSubmatch(template, -1)
	placeholders := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) > 1 {
			placeholders = append(placeholders, strings.TrimSpace(match[1]))
		}
	}

	return placeholders
}
