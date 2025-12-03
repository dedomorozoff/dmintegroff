package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"

	"github.com/machinebox/graphql"
)

// GraphQLService handles GraphQL operations
type GraphQLService struct {
	client *http.Client
}

// NewGraphQLService creates a new GraphQL service
func NewGraphQLService() *GraphQLService {
	return &GraphQLService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}            `json:"data"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents error location
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ExecuteQuery executes a GraphQL query
func (s *GraphQLService) ExecuteQuery(integration *models.Integration, payload map[string]interface{}) (map[string]interface{}, error) {
	endpoint := integration.GraphQLEndpoint
	if endpoint == "" {
		endpoint = integration.TargetAPI
	}

	// Build variables from mapping
	variables, err := s.buildVariables(integration.GraphQLVariables, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to build variables: %w", err)
	}

	// Replace placeholders in query
	query := s.replacePlaceholders(integration.GraphQLQuery, payload)

	// Create GraphQL request
	gqlReq := GraphQLRequest{
		Query:         query,
		Variables:     variables,
		OperationName: integration.GraphQLOperationName,
	}

	// Marshal request
	reqBody, err := json.Marshal(gqlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Log.Debugf("GraphQL Request to %s: %s", endpoint, string(reqBody))

	// Create HTTP request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication
	if err := s.addAuthentication(req, integration); err != nil {
		return nil, fmt.Errorf("failed to add authentication: %w", err)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Log.Debugf("GraphQL Response (%d): %s", resp.StatusCode, string(respBody))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		errMsgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			errMsgs[i] = e.Message
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(errMsgs, "; "))
	}

	// Convert data to map
	result, ok := gqlResp.Data.(map[string]interface{})
	if !ok {
		// Try to convert through JSON
		dataBytes, _ := json.Marshal(gqlResp.Data)
		json.Unmarshal(dataBytes, &result)
	}

	return result, nil
}

// IntrospectSchema fetches GraphQL schema using introspection
func (s *GraphQLService) IntrospectSchema(endpoint string, integration *models.Integration) (string, error) {
	introspectionQuery := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
				directives {
					name
					description
					locations
					args {
						...InputValue
					}
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes {
				...TypeRef
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	gqlReq := GraphQLRequest{
		Query: introspectionQuery,
	}

	reqBody, err := json.Marshal(gqlReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal introspection request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication if provided
	if integration != nil {
		if err := s.addAuthentication(req, integration); err != nil {
			return "", fmt.Errorf("failed to add authentication: %w", err)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute introspection: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read introspection response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("introspection failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}

// buildVariables builds GraphQL variables from mapping and payload
func (s *GraphQLService) buildVariables(mappingJSON string, payload map[string]interface{}) (map[string]interface{}, error) {
	if mappingJSON == "" {
		return nil, nil
	}

	var mapping map[string]string
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		return nil, err
	}

	variables := make(map[string]interface{})
	for varName, fieldPath := range mapping {
		value := getNestedValue(payload, fieldPath)
		if value != nil {
			variables[varName] = value
		}
	}

	return variables, nil
}

// replacePlaceholders replaces {{field.path}} placeholders in query
func (s *GraphQLService) replacePlaceholders(query string, payload map[string]interface{}) string {
	result := query
	
	// Find all {{...}} placeholders
	start := 0
	for {
		startIdx := strings.Index(result[start:], "{{")
		if startIdx == -1 {
			break
		}
		startIdx += start
		
		endIdx := strings.Index(result[startIdx:], "}}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx
		
		placeholder := result[startIdx+2 : endIdx]
		placeholder = strings.TrimSpace(placeholder)
		
		value := getNestedValue(payload, placeholder)
		if value != nil {
			// Convert value to string
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = fmt.Sprintf(`"%s"`, v)
			case float64, int, int64, bool:
				strValue = fmt.Sprintf("%v", v)
			default:
				jsonBytes, _ := json.Marshal(v)
				strValue = string(jsonBytes)
			}
			
			result = result[:startIdx] + strValue + result[endIdx+2:]
			start = startIdx + len(strValue)
		} else {
			start = endIdx + 2
		}
	}
	
	return result
}

// addAuthentication adds authentication headers to request
func (s *GraphQLService) addAuthentication(req *http.Request, integration *models.Integration) error {
	switch integration.AuthType {
	case "oauth2":
		// Get OAuth token
		token, err := GetAccessToken(integration)
		if err != nil {
			return fmt.Errorf("failed to get OAuth token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		
	case "bearer":
		if integration.BearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+integration.BearerToken)
		}
		
	case "basic":
		if integration.BasicAuthUser != "" {
			req.SetBasicAuth(integration.BasicAuthUser, integration.BasicAuthPass)
		}
	}
	
	return nil
}

// getNestedValue retrieves nested value from map using dot notation
func getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data
	
	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	
	return current
}

// TestGraphQLConnection tests GraphQL endpoint connectivity
func (s *GraphQLService) TestGraphQLConnection(endpoint string, integration *models.Integration) error {
	ctx := context.Background()
	client := graphql.NewClient(endpoint)
	
	// Simple introspection query to test connection
	req := graphql.NewRequest(`
		query {
			__schema {
				queryType {
					name
				}
			}
		}
	`)
	
	// Add authentication
	if integration != nil {
		switch integration.AuthType {
		case "oauth2":
			token, err := GetAccessToken(integration)
			if err != nil {
				return fmt.Errorf("failed to get OAuth token: %w", err)
			}
			req.Header.Set("Authorization", "Bearer "+token)
			
		case "bearer":
			if integration.BearerToken != "" {
				req.Header.Set("Authorization", "Bearer "+integration.BearerToken)
			}
			
		case "basic":
			if integration.BasicAuthUser != "" {
				req.Header.Set("Authorization", "Basic "+basicAuth(integration.BasicAuthUser, integration.BasicAuthPass))
			}
		}
	}
	
	var response interface{}
	if err := client.Run(ctx, req, &response); err != nil {
		return fmt.Errorf("GraphQL connection test failed: %w", err)
	}
	
	return nil
}

// basicAuth creates basic auth header value
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return fmt.Sprintf("Basic %s", auth)
}
