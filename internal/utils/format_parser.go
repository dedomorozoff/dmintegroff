package utils

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
)

// ParsePayload парсит входящие данные в зависимости от Content-Type
func ParsePayload(contentType string, body []byte, form url.Values, multipartForm *multipart.Form) (map[string]interface{}, []byte, error) {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	
	// Убираем параметры из Content-Type (например: "application/json; charset=utf-8" -> "application/json")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	// JSON (по умолчанию)
	if contentType == "" || contentType == "application/json" || strings.Contains(contentType, "json") {
		return ParseJSON(body)
	}

	// Form-data
	if contentType == "application/x-www-form-urlencoded" {
		return ParseFormURLEncoded(form)
	}

	// Multipart form-data
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return ParseMultipartForm(multipartForm)
	}

	// XML
	if contentType == "application/xml" || contentType == "text/xml" || strings.Contains(contentType, "xml") {
		return ParseXML(body)
	}

	// Plain text
	if contentType == "text/plain" {
		return ParsePlainText(body)
	}

	// Неизвестный формат - пробуем JSON
	return ParseJSON(body)
}

// ParseJSON парсит JSON данные
func ParseJSON(body []byte) (map[string]interface{}, []byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, nil, errors.New("invalid JSON: " + err.Error())
	}
	return payload, body, nil
}

// ParseFormURLEncoded парсит application/x-www-form-urlencoded данные
func ParseFormURLEncoded(form url.Values) (map[string]interface{}, []byte, error) {
	if form == nil {
		return nil, nil, errors.New("no form data")
	}

	payload := make(map[string]interface{})
	for key, values := range form {
		if len(values) == 1 {
			payload[key] = values[0]
		} else {
			payload[key] = values
		}
	}

	// Конвертируем в JSON для единообразия
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return payload, nil, err
	}

	return payload, jsonBytes, nil
}

// ParseMultipartForm парсит multipart/form-data
func ParseMultipartForm(multipartForm *multipart.Form) (map[string]interface{}, []byte, error) {
	if multipartForm == nil {
		return nil, nil, errors.New("no multipart form data")
	}

	payload := make(map[string]interface{})

	// Обрабатываем обычные поля
	for key, values := range multipartForm.Value {
		if len(values) == 1 {
			payload[key] = values[0]
		} else {
			payload[key] = values
		}
	}

	// Обрабатываем файлы (сохраняем метаданные)
	if len(multipartForm.File) > 0 {
		files := make([]map[string]interface{}, 0)
		for fieldName, fileHeaders := range multipartForm.File {
			for _, fileHeader := range fileHeaders {
				fileInfo := map[string]interface{}{
					"field_name": fieldName,
					"filename":   fileHeader.Filename,
					"size":       fileHeader.Size,
					"content_type": fileHeader.Header.Get("Content-Type"),
				}
				files = append(files, fileInfo)
			}
		}
		payload["_files"] = files
	}

	// Конвертируем в JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return payload, nil, err
	}

	return payload, jsonBytes, nil
}

// ParseXML парсит XML данные
func ParseXML(body []byte) (map[string]interface{}, []byte, error) {
	if len(body) == 0 {
		return nil, nil, errors.New("empty XML body")
	}

	// Парсим XML в map
	decoder := xml.NewDecoder(strings.NewReader(string(body)))
	payload, err := xmlToMap(decoder)
	if err != nil {
		return nil, nil, errors.New("invalid XML: " + err.Error())
	}

	// Конвертируем в JSON для единообразия
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return payload, body, err // Возвращаем оригинальный XML как fallback
	}

	return payload, jsonBytes, nil
}

// xmlToMap конвертирует XML в map[string]interface{}
func xmlToMap(decoder *xml.Decoder) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var currentKey string
	var currentValue strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			currentKey = t.Name.Local
			// Обрабатываем атрибуты
			if len(t.Attr) > 0 {
				attrs := make(map[string]string)
				for _, attr := range t.Attr {
					attrs[attr.Name.Local] = attr.Value
				}
				result[currentKey+"_attributes"] = attrs
			}
		case xml.CharData:
			currentValue.WriteString(string(t))
		case xml.EndElement:
			if currentKey != "" && currentValue.Len() > 0 {
				value := strings.TrimSpace(currentValue.String())
				if value != "" {
					result[currentKey] = value
				}
				currentValue.Reset()
			}
		}
	}

	return result, nil
}

// ParsePlainText парсит plain text данные
func ParsePlainText(body []byte) (map[string]interface{}, []byte, error) {
	if len(body) == 0 {
		return nil, nil, errors.New("empty text body")
	}

	payload := map[string]interface{}{
		"text": string(body),
		"length": len(body),
	}

	// Пробуем разбить на строки
	lines := strings.Split(string(body), "\n")
	if len(lines) > 1 {
		payload["lines"] = lines
		payload["line_count"] = len(lines)
	}

	// Конвертируем в JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return payload, body, err
	}

	return payload, jsonBytes, nil
}

// DetectContentType пытается определить Content-Type по содержимому
func DetectContentType(body []byte) string {
	if len(body) == 0 {
		return "text/plain"
	}

	// Пробуем JSON
	var js interface{}
	if json.Unmarshal(body, &js) == nil {
		return "application/json"
	}

	// Пробуем XML
	if strings.HasPrefix(strings.TrimSpace(string(body)), "<?xml") || 
	   strings.HasPrefix(strings.TrimSpace(string(body)), "<") {
		return "application/xml"
	}

	// По умолчанию text/plain
	return "text/plain"
}
