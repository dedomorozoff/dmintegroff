package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dmintegroff/internal/models"
)

func TestGraphQLService_ExecuteQuery(t *testing.T) {
	// Create mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		
		// Parse request
		var req GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}
		
		// Send response
		response := GraphQLResponse{
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "John Doe",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	service := NewGraphQLService()
	integration := &models.Integration{
		GraphQLEndpoint: server.URL,
		GraphQLQuery: `
			query GetUser($id: ID!) {
				user(id: $id) {
					id
					name
				}
			}
		`,
		GraphQLVariables: `{"id": "user_id"}`,
		AuthType:        "none",
	}
	
	payload := map[string]interface{}{
		"user_id": "123",
	}
	
	result, err := service.ExecuteQuery(integration, payload)
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}
	
	// Verify result
	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user object in result")
	}
	
	if user["id"] != "123" {
		t.Errorf("Expected user id 123, got %v", user["id"])
	}
	
	if user["name"] != "John Doe" {
		t.Errorf("Expected user name John Doe, got %v", user["name"])
	}
}

func TestGraphQLService_ExecuteQueryWithErrors(t *testing.T) {
	// Create mock GraphQL server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Errors: []GraphQLError{
				{
					Message: "Field 'user' not found",
					Locations: []GraphQLErrorLocation{
						{Line: 2, Column: 3},
					},
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	service := NewGraphQLService()
	integration := &models.Integration{
		GraphQLEndpoint: server.URL,
		GraphQLQuery:    `query { user { id } }`,
		AuthType:        "none",
	}
	
	payload := map[string]interface{}{}
	
	_, err := service.ExecuteQuery(integration, payload)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if err.Error() != "GraphQL errors: Field 'user' not found" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestGraphQLService_BuildVariables(t *testing.T) {
	service := NewGraphQLService()
	
	mappingJSON := `{
		"userId": "user.id",
		"email": "user.email",
		"age": "user.profile.age"
	}`
	
	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    "123",
			"email": "john@example.com",
			"profile": map[string]interface{}{
				"age": 30,
			},
		},
	}
	
	variables, err := service.buildVariables(mappingJSON, payload)
	if err != nil {
		t.Fatalf("buildVariables failed: %v", err)
	}
	
	if variables["userId"] != "123" {
		t.Errorf("Expected userId 123, got %v", variables["userId"])
	}
	
	if variables["email"] != "john@example.com" {
		t.Errorf("Expected email john@example.com, got %v", variables["email"])
	}
	
	// Age can be int or float64 depending on JSON unmarshaling
	age, ok := variables["age"]
	if !ok {
		t.Error("Expected age variable")
	}
	switch v := age.(type) {
	case int:
		if v != 30 {
			t.Errorf("Expected age 30, got %v", v)
		}
	case float64:
		if v != 30.0 {
			t.Errorf("Expected age 30, got %v", v)
		}
	default:
		t.Errorf("Expected age to be int or float64, got %T", v)
	}
}

func TestGraphQLService_ReplacePlaceholders(t *testing.T) {
	service := NewGraphQLService()
	
	query := `
		query {
			user(id: {{user.id}}, name: {{user.name}}) {
				id
				name
			}
		}
	`
	
	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   123,
			"name": "John Doe",
		},
	}
	
	result := service.replacePlaceholders(query, payload)
	
	if !contains(result, `id: 123`) {
		t.Errorf("Expected id: 123 in result, got: %s", result)
	}
	
	if !contains(result, `name: "John Doe"`) {
		t.Errorf("Expected name: \"John Doe\" in result, got: %s", result)
	}
}

func TestGraphQLService_IntrospectSchema(t *testing.T) {
	// Create mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType": map[string]interface{}{
						"name": "Query",
					},
					"types": []interface{}{
						map[string]interface{}{
							"kind": "OBJECT",
							"name": "User",
							"fields": []interface{}{
								map[string]interface{}{
									"name": "id",
									"type": map[string]interface{}{
										"kind": "SCALAR",
										"name": "ID",
									},
								},
							},
						},
					},
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	service := NewGraphQLService()
	
	schema, err := service.IntrospectSchema(server.URL, nil)
	if err != nil {
		t.Fatalf("IntrospectSchema failed: %v", err)
	}
	
	if schema == "" {
		t.Error("Expected non-empty schema")
	}
	
	if !contains(schema, "Query") {
		t.Error("Expected schema to contain Query type")
	}
}

func TestGraphQLService_TestConnection(t *testing.T) {
	// Create mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType": map[string]interface{}{
						"name": "Query",
					},
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	service := NewGraphQLService()
	
	err := service.TestGraphQLConnection(server.URL, nil)
	if err != nil {
		t.Fatalf("TestGraphQLConnection failed: %v", err)
	}
}

func TestGraphQLService_WithBearerAuth(t *testing.T) {
	// Create mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization: Bearer test-token, got %s", auth)
		}
		
		response := GraphQLResponse{
			Data: map[string]interface{}{
				"result": "success",
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	service := NewGraphQLService()
	integration := &models.Integration{
		GraphQLEndpoint: server.URL,
		GraphQLQuery:    `query { result }`,
		AuthType:        "bearer",
		BearerToken:     "test-token",
	}
	
	payload := map[string]interface{}{}
	
	_, err := service.ExecuteQuery(integration, payload)
	if err != nil {
		t.Fatalf("ExecuteQuery with bearer auth failed: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
