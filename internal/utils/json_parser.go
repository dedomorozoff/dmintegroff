package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FieldInfo содержит информацию о поле JSON
type FieldInfo struct {
	Path     string      // Путь к полю (например, "user.name", "items[0].id")
	Value    interface{} // Значение поля
	Type     string      // Тип: "string", "number", "boolean", "object", "array", "null"
	FullPath string      // Полный путь для отображения
}

// FlattenJSON рекурсивно разбирает JSON и возвращает все поля с их путями
func FlattenJSON(data interface{}, prefix string) []FieldInfo {
	var fields []FieldInfo

	switch v := data.(type) {
	case map[string]interface{}:
		// Обрабатываем объект
		for key, value := range v {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}

			// Если значение - объект или массив, рекурсивно обрабатываем его поля
			if isComplexType(value) {
				fields = append(fields, FlattenJSON(value, path)...)
			} else {
				// Для простых типов создаем поле
				fields = append(fields, FieldInfo{
					Path:     path,
					Value:    value,
					Type:     getType(value),
					FullPath: path,
				})
			}
		}

	case []interface{}:
		// Обрабатываем массив - показываем первый элемент как пример для структуры
		if len(v) > 0 {
			// Определяем путь для первого элемента массива
			var path string
			if prefix == "" {
				path = "[0]" // Для корневого массива
			} else {
				// Если массив внутри объекта, используем путь с индексом [0]
				path = prefix + "[0]"
			}

			item := v[0]

			// Если элемент - объект или массив, рекурсивно обрабатываем его поля
			if isComplexType(item) {
				// Для сложных типов обрабатываем вложенные поля
				fields = append(fields, FlattenJSON(item, path)...)
			} else {
				// Для простых типов создаем поле с путем массива с индексом
				fields = append(fields, FieldInfo{
					Path:     path,
					Value:    item,
					Type:     getType(item),
					FullPath: path,
				})
			}
		}

	default:
		// Простое значение
		fields = append(fields, FieldInfo{
			Path:     prefix,
			Value:    v,
			Type:     getType(v),
			FullPath: prefix,
		})
	}

	return fields
}

// ParseJSONString парсит JSON строку и возвращает плоский список полей
func ParseJSONString(jsonStr string) ([]FieldInfo, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	fields := FlattenJSON(data, "")
	return deduplicateFields(fields), nil
}

// getType определяет тип значения
func getType(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

// isComplexType проверяет, является ли значение сложным типом (объект или массив)
func isComplexType(v interface{}) bool {
	_, isMap := v.(map[string]interface{})
	_, isArray := v.([]interface{})
	return isMap || isArray
}

// deduplicateFields удаляет дубликаты полей
func deduplicateFields(fields []FieldInfo) []FieldInfo {
	seen := make(map[string]bool)
	var result []FieldInfo

	for _, field := range fields {
		// Пропускаем только объекты (их поля уже обработаны рекурсивно)
		if field.Type == "object" {
			continue
		}
		// Пропускаем массивы, которые содержат сложные типы (их элементы уже обработаны рекурсивно)
		// Но оставляем массивы с простыми значениями
		if field.Type == "array" {
			// Проверяем, содержит ли массив сложные типы
			if arr, ok := field.Value.([]interface{}); ok && len(arr) > 0 {
				if isComplexType(arr[0]) {
					// Массив содержит объекты/массивы - пропускаем, т.к. их поля уже обработаны
					continue
				}
				// Массив содержит простые значения - оставляем
			} else {
				// Пустой массив или не массив - пропускаем
				continue
			}
		}
		// Добавляем поле, если его еще не было
		if !seen[field.Path] {
			seen[field.Path] = true
			result = append(result, field)
		}
	}

	return result
}

// GetValueByPath извлекает значение из JSON по пути
func GetValueByPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := parsePath(path)
	current := data

	for _, part := range parts {
		switch p := part.(type) {
		case string:
			// Доступ к полю объекта
			if obj, ok := current.(map[string]interface{}); ok {
				if val, exists := obj[p]; exists {
					current = val
				} else {
					return nil, fmt.Errorf("field %s not found", p)
				}
			} else {
				return nil, fmt.Errorf("cannot access field %s on non-object", p)
			}
		case int:
			// Доступ к элементу массива
			if arr, ok := current.([]interface{}); ok {
				if p >= 0 && p < len(arr) {
					current = arr[p]
				} else {
					return nil, fmt.Errorf("array index %d out of bounds", p)
				}
			} else {
				return nil, fmt.Errorf("cannot access index %d on non-array", p)
			}
		}
	}

	return current, nil
}

// parsePath разбирает путь на части (например, "user.items[0].name" -> ["user", "items", 0, "name"])
func parsePath(path string) []interface{} {
	var parts []interface{}
	var current string

	for i := 0; i < len(path); i++ {
		char := path[i]

		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else if char == '[' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			// Ищем закрывающую скобку
			i++
			indexStr := ""
			for i < len(path) && path[i] != ']' {
				indexStr += string(path[i])
				i++
			}
			if index, err := strconv.Atoi(indexStr); err == nil {
				parts = append(parts, index)
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
