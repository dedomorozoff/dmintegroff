package services

import (
	"encoding/json"
	"testing"
	"time"
)

func TestExportFormat(t *testing.T) {
	exportData := ExportFormat{
		Version:    "1.0",
		ExportedAt: time.Now(),
		ExportedBy: "test@example.com",
		Integrations: []IntegrationExportData{
			{
				Name:       "Test Integration",
				TargetAPI:  "https://api.example.com/webhook",
				HTTPMethod: "POST",
				Mode:       "listening",
				AuthType:   "oauth2",
			},
		},
	}
	
	// Test JSON marshaling
	jsonData, err := json.Marshal(exportData)
	if err != nil {
		t.Fatalf("Failed to marshal export data: %v", err)
	}
	
	// Test JSON unmarshaling
	var decoded ExportFormat
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal export data: %v", err)
	}
	
	if decoded.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", decoded.Version)
	}
	
	if len(decoded.Integrations) != 1 {
		t.Errorf("Expected 1 integration, got %d", len(decoded.Integrations))
	}
	
	if decoded.Integrations[0].Name != "Test Integration" {
		t.Errorf("Expected name 'Test Integration', got %s", decoded.Integrations[0].Name)
	}
}

func TestValidateImportData_ValidData(t *testing.T) {
	exportData := ExportFormat{
		Version:    "1.0",
		ExportedAt: time.Now(),
		ExportedBy: "test@example.com",
		Integrations: []IntegrationExportData{
			{
				Name:       "Test Integration",
				TargetAPI:  "https://api.example.com/webhook",
				HTTPMethod: "POST",
			},
		},
	}
	
	jsonData, _ := json.Marshal(exportData)
	
	result, err := ValidateImportData(jsonData)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	
	if result.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", result.Version)
	}
	
	if len(result.Integrations) != 1 {
		t.Errorf("Expected 1 integration, got %d", len(result.Integrations))
	}
}

func TestValidateImportData_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{"invalid": json}`)
	
	_, err := ValidateImportData(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestValidateImportData_UnsupportedVersion(t *testing.T) {
	exportData := ExportFormat{
		Version:      "2.0",
		Integrations: []IntegrationExportData{},
	}
	
	jsonData, _ := json.Marshal(exportData)
	
	_, err := ValidateImportData(jsonData)
	if err == nil {
		t.Error("Expected error for unsupported version, got nil")
	}
}

func TestValidateImportData_NoIntegrations(t *testing.T) {
	exportData := ExportFormat{
		Version:      "1.0",
		Integrations: []IntegrationExportData{},
	}
	
	jsonData, _ := json.Marshal(exportData)
	
	_, err := ValidateImportData(jsonData)
	if err == nil {
		t.Error("Expected error for no integrations, got nil")
	}
}

func TestValidateImportData_MissingName(t *testing.T) {
	exportData := ExportFormat{
		Version: "1.0",
		Integrations: []IntegrationExportData{
			{
				Name:      "", // Missing name
				TargetAPI: "https://api.example.com",
			},
		},
	}
	
	jsonData, _ := json.Marshal(exportData)
	
	_, err := ValidateImportData(jsonData)
	if err == nil {
		t.Error("Expected error for missing name, got nil")
	}
}

func TestValidateImportData_MissingTargetAPI(t *testing.T) {
	exportData := ExportFormat{
		Version: "1.0",
		Integrations: []IntegrationExportData{
			{
				Name:      "Test",
				TargetAPI: "", // Missing target API
			},
		},
	}
	
	jsonData, _ := json.Marshal(exportData)
	
	_, err := ValidateImportData(jsonData)
	if err == nil {
		t.Error("Expected error for missing target API, got nil")
	}
}

func TestIntegrationExportData_AllFields(t *testing.T) {
	data := IntegrationExportData{
		Name:           "Full Integration",
		SourceAPI:      "https://source.example.com",
		TargetAPI:      "https://target.example.com",
		HTTPMethod:     "POST",
		Mode:           "listening",
		MappingConfig:  `{"field1": "value1"}`,
		OutputTemplate: `{"output": "{{field1}}"}`,
		
		AuthType:           "oauth2",
		OAuth2TokenURL:     "https://oauth.example.com/token",
		OAuth2ClientID:     "client-id",
		OAuth2ClientSecret: "client-secret",
		OAuth2Scope:        "read write",
		OAuth2GrantType:    "client_credentials",
		BearerToken:        "bearer-token",
		BasicAuthUser:      "user",
		BasicAuthPass:      "pass",
		
		WebhookSignatureEnabled:   true,
		WebhookSignatureSecret:    "secret-key",
		WebhookSignatureHeader:    "X-Signature",
		WebhookSignatureAlgorithm: "sha256",
	}
	
	// Test JSON marshaling
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal integration data: %v", err)
	}
	
	// Test JSON unmarshaling
	var decoded IntegrationExportData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal integration data: %v", err)
	}
	
	// Verify all fields
	if decoded.Name != data.Name {
		t.Errorf("Name mismatch: expected %s, got %s", data.Name, decoded.Name)
	}
	if decoded.AuthType != data.AuthType {
		t.Errorf("AuthType mismatch: expected %s, got %s", data.AuthType, decoded.AuthType)
	}
	if decoded.WebhookSignatureEnabled != data.WebhookSignatureEnabled {
		t.Errorf("WebhookSignatureEnabled mismatch")
	}
}

func TestImportOptions_Defaults(t *testing.T) {
	options := ImportOptions{
		ProjectID:         1,
		UserID:            1,
		SkipDuplicates:    false,
		UpdateExisting:    false,
		GenerateNewTokens: false,
	}
	
	if options.ProjectID != 1 {
		t.Errorf("Expected ProjectID 1, got %d", options.ProjectID)
	}
	
	if options.SkipDuplicates {
		t.Error("Expected SkipDuplicates to be false")
	}
}

func TestImportResult_Structure(t *testing.T) {
	result := ImportResult{
		TotalCount:    10,
		ImportedCount: 7,
		SkippedCount:  2,
		UpdatedCount:  1,
		Errors:        []string{"error1", "error2"},
	}
	
	if result.TotalCount != 10 {
		t.Errorf("Expected TotalCount 10, got %d", result.TotalCount)
	}
	
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
	
	// Test JSON marshaling
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}
	
	var decoded ImportResult
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	
	if decoded.ImportedCount != result.ImportedCount {
		t.Errorf("ImportedCount mismatch")
	}
}

func TestExportFormat_WithProjectName(t *testing.T) {
	exportData := ExportFormat{
		Version:     "1.0",
		ExportedAt:  time.Now(),
		ExportedBy:  "test@example.com",
		ProjectName: "My Project",
		Integrations: []IntegrationExportData{
			{
				Name:      "Integration 1",
				TargetAPI: "https://api1.example.com",
			},
			{
				Name:      "Integration 2",
				TargetAPI: "https://api2.example.com",
			},
		},
	}
	
	jsonData, _ := json.Marshal(exportData)
	
	var decoded ExportFormat
	json.Unmarshal(jsonData, &decoded)
	
	if decoded.ProjectName != "My Project" {
		t.Errorf("Expected ProjectName 'My Project', got %s", decoded.ProjectName)
	}
	
	if len(decoded.Integrations) != 2 {
		t.Errorf("Expected 2 integrations, got %d", len(decoded.Integrations))
	}
}
