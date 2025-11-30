package utils

import (
	"encoding/json"
	"testing"
)

func TestTemplateProcessor_ProcessTemplate(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name       string
		template   string
		sourceData map[string]interface{}
		want       string
		wantErr    bool
	}{
		{
			name:     "Simple string substitution",
			template: `{"name": "{{user.name}}", "email": "{{user.email}}"}`,
			sourceData: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "John",
					"email": "john@example.com",
				},
			},
			want:    `{"name":"John","email":"john@example.com"}`,
			wantErr: false,
		},
		{
			name:     "Number substitution",
			template: `{"age": {{user.age}}, "count": {{count}}}`,
			sourceData: map[string]interface{}{
				"user": map[string]interface{}{
					"age": float64(25),
				},
				"count": float64(10),
			},
			want:    `{"age":25,"count":10}`,
			wantErr: false,
		},
		{
			name:     "Boolean substitution",
			template: `{"active": {{is_active}}, "verified": {{is_verified}}}`,
			sourceData: map[string]interface{}{
				"is_active":   true,
				"is_verified": false,
			},
			want:    `{"active":true,"verified":false}`,
			wantErr: false,
		},
		{
			name:     "Nested object",
			template: `{"user": {"name": "{{user.profile.name}}", "age": {{user.profile.age}}}}`,
			sourceData: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"name": "Alice",
						"age":  float64(30),
					},
				},
			},
			want:    `{"user":{"name":"Alice","age":30}}`,
			wantErr: false,
		},
		{
			name:     "Array access",
			template: `{"first": "{{items[0].name}}", "second": "{{items[1].name}}"}`,
			sourceData: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "Item1"},
					map[string]interface{}{"name": "Item2"},
				},
			},
			want:    `{"first":"Item1","second":"Item2"}`,
			wantErr: false,
		},
		{
			name:     "Mixed static and dynamic",
			template: `{"name": "{{user.name}}", "source": "webhook", "version": "1.0"}`,
			sourceData: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Bob",
				},
			},
			want:    `{"name":"Bob","source":"webhook","version":"1.0"}`,
			wantErr: false,
		},
		{
			name:     "Missing field returns null",
			template: `{"name": "{{user.name}}", "missing": {{missing.field}}}`,
			sourceData: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Charlie",
				},
			},
			want:    `{"name":"Charlie","missing":null}`,
			wantErr: false,
		},
		{
			name:       "Empty template",
			template:   "",
			sourceData: map[string]interface{}{},
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processor.ProcessTemplate(tt.template, tt.sourceData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Сравниваем JSON (игнорируя форматирование)
				var gotJSON, wantJSON interface{}
				json.Unmarshal([]byte(tt.want), &wantJSON)
				gotBytes, _ := json.Marshal(got)
				json.Unmarshal(gotBytes, &gotJSON)

				gotStr, _ := json.Marshal(gotJSON)
				wantStr, _ := json.Marshal(wantJSON)

				if string(gotStr) != string(wantStr) {
					t.Errorf("ProcessTemplate() = %v, want %v", string(gotStr), string(wantStr))
				}
			}
		})
	}
}

func TestTemplateProcessor_ValidateTemplate(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "Valid template",
			template: `{"name": "{{user.name}}", "age": {{user.age}}}`,
			wantErr:  false,
		},
		{
			name:     "Valid nested template",
			template: `{"user": {"profile": {"name": "{{name}}"}}}`,
			wantErr:  false,
		},
		{
			name:     "Invalid JSON - missing comma",
			template: `{"name": "{{user.name}}" "age": {{user.age}}}`,
			wantErr:  true,
		},
		{
			name:     "Invalid JSON - trailing comma",
			template: `{"name": "{{user.name}}",}`,
			wantErr:  true,
		},
		{
			name:     "Empty template",
			template: "",
			wantErr:  true,
		},
		{
			name:     "Invalid JSON - missing bracket",
			template: `{"name": "{{user.name}}"`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateProcessor_ExtractPlaceholders(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "Single placeholder",
			template: `{"name": "{{user.name}}"}`,
			want:     []string{"user.name"},
		},
		{
			name:     "Multiple placeholders",
			template: `{"name": "{{user.name}}", "email": "{{user.email}}", "age": {{user.age}}}`,
			want:     []string{"user.name", "user.email", "user.age"},
		},
		{
			name:     "Nested placeholders",
			template: `{"user": {"name": "{{user.profile.name}}", "city": "{{user.address.city}}"}}`,
			want:     []string{"user.profile.name", "user.address.city"},
		},
		{
			name:     "Array placeholders",
			template: `{"first": "{{items[0].name}}", "second": "{{items[1].name}}"}`,
			want:     []string{"items[0].name", "items[1].name"},
		},
		{
			name:     "No placeholders",
			template: `{"name": "static", "value": 123}`,
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processor.ExtractPlaceholders(tt.template)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractPlaceholders() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractPlaceholders()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}


func TestTemplateProcessor_ArrayWildcard(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name       string
		template   string
		sourceData map[string]interface{}
		want       string
		wantErr    bool
	}{
		{
			name:     "Copy entire array",
			template: `{"products": {{items.*}}}`,
			sourceData: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "Item1", "price": float64(100)},
					map[string]interface{}{"name": "Item2", "price": float64(200)},
				},
			},
			want:    `{"products":[{"name":"Item1","price":100},{"name":"Item2","price":200}]}`,
			wantErr: false,
		},
		{
			name:     "Copy nested array",
			template: `{"data": {{user.orders.*}}}`,
			sourceData: map[string]interface{}{
				"user": map[string]interface{}{
					"orders": []interface{}{
						map[string]interface{}{"id": "1", "total": float64(500)},
						map[string]interface{}{"id": "2", "total": float64(300)},
					},
				},
			},
			want:    `{"data":[{"id":"1","total":500},{"id":"2","total":300}]}`,
			wantErr: false,
		},
		{
			name:     "Copy simple array",
			template: `{"tags": {{tags.*}}}`,
			sourceData: map[string]interface{}{
				"tags": []interface{}{"vip", "active", "premium"},
			},
			want:    `{"tags":["vip","active","premium"]}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processor.ProcessTemplate(tt.template, tt.sourceData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotBytes, _ := json.Marshal(got)
				var wantJSON interface{}
				json.Unmarshal([]byte(tt.want), &wantJSON)
				wantBytes, _ := json.Marshal(wantJSON)

				if string(gotBytes) != string(wantBytes) {
					t.Errorf("ProcessTemplate() = %v, want %v", string(gotBytes), string(wantBytes))
				}
			}
		})
	}
}
