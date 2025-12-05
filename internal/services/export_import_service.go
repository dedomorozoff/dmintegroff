package services

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ExportFormat represents the structure of exported data
type ExportFormat struct {
	Version      string                  `json:"version"`
	ExportedAt   time.Time               `json:"exported_at"`
	ExportedBy   string                  `json:"exported_by"`
	ProjectName  string                  `json:"project_name,omitempty"`
	Integrations []IntegrationExportData `json:"integrations"`
}

// IntegrationExportData represents integration data for export
type IntegrationExportData struct {
	Name           string `json:"name"`
	SourceAPI      string `json:"source_api,omitempty"`
	TargetAPI      string `json:"target_api"`
	HTTPMethod     string `json:"http_method"`
	Mode           string `json:"mode"`
	MappingConfig  string `json:"mapping_config,omitempty"`
	OutputTemplate string `json:"output_template,omitempty"`
	TemplateType   string `json:"template_type,omitempty"`
	
	// Authentication
	AuthType           string `json:"auth_type"`
	OAuth2TokenURL     string `json:"oauth2_token_url,omitempty"`
	OAuth2ClientID     string `json:"oauth2_client_id,omitempty"`
	OAuth2ClientSecret string `json:"oauth2_client_secret,omitempty"`
	OAuth2Scope        string `json:"oauth2_scope,omitempty"`
	OAuth2GrantType    string `json:"oauth2_grant_type,omitempty"`
	BearerToken        string `json:"bearer_token,omitempty"`
	BasicAuthUser      string `json:"basic_auth_user,omitempty"`
	BasicAuthPass      string `json:"basic_auth_pass,omitempty"`
	
	// Webhook Signatures
	WebhookSignatureEnabled   bool   `json:"webhook_signature_enabled"`
	WebhookSignatureSecret    string `json:"webhook_signature_secret,omitempty"`
	WebhookSignatureHeader    string `json:"webhook_signature_header,omitempty"`
	WebhookSignatureAlgorithm string `json:"webhook_signature_algorithm,omitempty"`
}

// ImportOptions configures import behavior
type ImportOptions struct {
	ProjectID       uint   `json:"project_id"`
	UserID          uint   `json:"user_id"`
	SkipDuplicates  bool   `json:"skip_duplicates"`
	UpdateExisting  bool   `json:"update_existing"`
	GenerateNewTokens bool `json:"generate_new_tokens"`
}

// ImportResult contains import operation results
type ImportResult struct {
	TotalCount    int      `json:"total_count"`
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	UpdatedCount  int      `json:"updated_count"`
	Errors        []string `json:"errors,omitempty"`
}

// ExportIntegrations exports integrations to JSON format
func ExportIntegrations(integrationIDs []uint, userEmail string) ([]byte, error) {
	var integrations []models.Integration
	
	query := database.DB.Where("id IN ?", integrationIDs)
	if err := query.Find(&integrations).Error; err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to fetch integrations for export")
		return nil, err
	}
	
	if len(integrations) == 0 {
		return nil, errors.New("no integrations found for export")
	}
	
	exportData := ExportFormat{
		Version:      "1.0",
		ExportedAt:   time.Now(),
		ExportedBy:   userEmail,
		Integrations: make([]IntegrationExportData, 0, len(integrations)),
	}
	
	for _, integration := range integrations {
		exportData.Integrations = append(exportData.Integrations, IntegrationExportData{
			Name:           integration.Name,
			SourceAPI:      integration.SourceAPI,
			TargetAPI:      integration.TargetAPI,
			HTTPMethod:     integration.HTTPMethod,
			Mode:           integration.Mode,
			MappingConfig:  integration.MappingConfig,
			OutputTemplate: integration.OutputTemplate,
			TemplateType:   integration.TemplateType,
			
			AuthType:           integration.AuthType,
			OAuth2TokenURL:     integration.OAuth2TokenURL,
			OAuth2ClientID:     integration.OAuth2ClientID,
			OAuth2ClientSecret: integration.OAuth2ClientSecret,
			OAuth2Scope:        integration.OAuth2Scope,
			OAuth2GrantType:    integration.OAuth2GrantType,
			BearerToken:        integration.BearerToken,
			BasicAuthUser:      integration.BasicAuthUser,
			BasicAuthPass:      integration.BasicAuthPass,
			
			WebhookSignatureEnabled:   integration.WebhookSignatureEnabled,
			WebhookSignatureSecret:    integration.WebhookSignatureSecret,
			WebhookSignatureHeader:    integration.WebhookSignatureHeader,
			WebhookSignatureAlgorithm: integration.WebhookSignatureAlgorithm,
		})
	}
	
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to marshal export data")
		return nil, err
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"count":       len(integrations),
		"exported_by": userEmail,
	}).Info("Integrations exported successfully")
	
	return jsonData, nil
}

// ExportProjectIntegrations exports all integrations from a project
func ExportProjectIntegrations(projectID uint, userEmail string) ([]byte, error) {
	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		return nil, err
	}
	
	var integrations []models.Integration
	if err := database.DB.Where("project_id = ?", projectID).Find(&integrations).Error; err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"project_id": projectID,
			"error":      err.Error(),
		}).Error("Failed to fetch project integrations for export")
		return nil, err
	}
	
	if len(integrations) == 0 {
		return nil, errors.New("no integrations found in project")
	}
	
	exportData := ExportFormat{
		Version:      "1.0",
		ExportedAt:   time.Now(),
		ExportedBy:   userEmail,
		ProjectName:  project.Name,
		Integrations: make([]IntegrationExportData, 0, len(integrations)),
	}
	
	for _, integration := range integrations {
		exportData.Integrations = append(exportData.Integrations, IntegrationExportData{
			Name:           integration.Name,
			SourceAPI:      integration.SourceAPI,
			TargetAPI:      integration.TargetAPI,
			HTTPMethod:     integration.HTTPMethod,
			Mode:           integration.Mode,
			MappingConfig:  integration.MappingConfig,
			OutputTemplate: integration.OutputTemplate,
			TemplateType:   integration.TemplateType,
			
			AuthType:           integration.AuthType,
			OAuth2TokenURL:     integration.OAuth2TokenURL,
			OAuth2ClientID:     integration.OAuth2ClientID,
			OAuth2ClientSecret: integration.OAuth2ClientSecret,
			OAuth2Scope:        integration.OAuth2Scope,
			OAuth2GrantType:    integration.OAuth2GrantType,
			BearerToken:        integration.BearerToken,
			BasicAuthUser:      integration.BasicAuthUser,
			BasicAuthPass:      integration.BasicAuthPass,
			
			WebhookSignatureEnabled:   integration.WebhookSignatureEnabled,
			WebhookSignatureSecret:    integration.WebhookSignatureSecret,
			WebhookSignatureHeader:    integration.WebhookSignatureHeader,
			WebhookSignatureAlgorithm: integration.WebhookSignatureAlgorithm,
		})
	}
	
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to marshal export data")
		return nil, err
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"project_id":  projectID,
		"count":       len(integrations),
		"exported_by": userEmail,
	}).Info("Project integrations exported successfully")
	
	return jsonData, nil
}

// ImportIntegrations imports integrations from JSON data
func ImportIntegrations(jsonData []byte, options ImportOptions) (*ImportResult, error) {
	var exportData ExportFormat
	if err := json.Unmarshal(jsonData, &exportData); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to parse import data")
		return nil, errors.New("invalid JSON format")
	}
	
	// Validate version
	if exportData.Version != "1.0" {
		return nil, fmt.Errorf("unsupported export version: %s", exportData.Version)
	}
	
	// Validate project exists
	var project models.Project
	if err := database.DB.First(&project, options.ProjectID).Error; err != nil {
		return nil, errors.New("target project not found")
	}
	
	// Validate user exists
	var user models.User
	if err := database.DB.First(&user, options.UserID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	
	result := &ImportResult{
		TotalCount: len(exportData.Integrations),
		Errors:     make([]string, 0),
	}
	
	for i, integrationData := range exportData.Integrations {
		if err := importSingleIntegration(&integrationData, options, result); err != nil {
			errMsg := fmt.Sprintf("Integration %d (%s): %s", i+1, integrationData.Name, err.Error())
			result.Errors = append(result.Errors, errMsg)
			logger.Log.WithFields(map[string]interface{}{
				"integration_name": integrationData.Name,
				"error":            err.Error(),
			}).Warn("Failed to import integration")
		}
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"total":    result.TotalCount,
		"imported": result.ImportedCount,
		"updated":  result.UpdatedCount,
		"skipped":  result.SkippedCount,
		"errors":   len(result.Errors),
	}).Info("Import completed")
	
	return result, nil
}

// importSingleIntegration imports a single integration
func importSingleIntegration(data *IntegrationExportData, options ImportOptions, result *ImportResult) error {
	// Validate required fields
	if data.Name == "" {
		return errors.New("integration name is required")
	}
	if data.TargetAPI == "" {
		return errors.New("target API is required")
	}
	
	// Check for existing integration with same name in project
	var existing models.Integration
	err := database.DB.Where("name = ? AND project_id = ?", data.Name, options.ProjectID).First(&existing).Error
	
	if err == nil {
		// Integration exists
		if options.SkipDuplicates {
			result.SkippedCount++
			return nil
		}
		
		if options.UpdateExisting {
			// Update existing integration
			return updateExistingIntegration(&existing, data, options, result)
		}
		
		return fmt.Errorf("integration with name '%s' already exists", data.Name)
	}
	
	// Create new integration
	webhookToken, err := utils.GenerateToken(16)
	if err != nil {
		return fmt.Errorf("failed to generate webhook token: %w", err)
	}
	
	if !options.GenerateNewTokens && data.WebhookSignatureSecret != "" {
		// Keep existing signature secret if not generating new tokens
	} else if data.WebhookSignatureEnabled {
		// Generate new signature secret
		secret, err := utils.GenerateToken(32)
		if err != nil {
			return fmt.Errorf("failed to generate signature secret: %w", err)
		}
		data.WebhookSignatureSecret = secret
	}
	
	integration := models.Integration{
		Name:           data.Name,
		WebhookToken:   webhookToken,
		SourceAPI:      data.SourceAPI,
		TargetAPI:      data.TargetAPI,
		HTTPMethod:     data.HTTPMethod,
		Mode:           data.Mode,
		MappingConfig:  data.MappingConfig,
		OutputTemplate: data.OutputTemplate,
		TemplateType:   data.TemplateType,
		Status:         "active",
		ProjectID:      options.ProjectID,
		CreatedByID:    options.UserID,
		
		AuthType:           data.AuthType,
		OAuth2TokenURL:     data.OAuth2TokenURL,
		OAuth2ClientID:     data.OAuth2ClientID,
		OAuth2ClientSecret: data.OAuth2ClientSecret,
		OAuth2Scope:        data.OAuth2Scope,
		OAuth2GrantType:    data.OAuth2GrantType,
		BearerToken:        data.BearerToken,
		BasicAuthUser:      data.BasicAuthUser,
		BasicAuthPass:      data.BasicAuthPass,
		
		WebhookSignatureEnabled:   data.WebhookSignatureEnabled,
		WebhookSignatureSecret:    data.WebhookSignatureSecret,
		WebhookSignatureHeader:    data.WebhookSignatureHeader,
		WebhookSignatureAlgorithm: data.WebhookSignatureAlgorithm,
	}
	
	// Set defaults
	if integration.HTTPMethod == "" {
		integration.HTTPMethod = "POST"
	}
	if integration.Mode == "" {
		integration.Mode = "listening"
	}
	if integration.AuthType == "" {
		integration.AuthType = "none"
	}
	if integration.WebhookSignatureHeader == "" && integration.WebhookSignatureEnabled {
		integration.WebhookSignatureHeader = "X-Webhook-Signature"
	}
	if integration.WebhookSignatureAlgorithm == "" && integration.WebhookSignatureEnabled {
		integration.WebhookSignatureAlgorithm = "sha256"
	}
	
	if err := database.DB.Create(&integration).Error; err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	
	result.ImportedCount++
	logger.Log.WithFields(map[string]interface{}{
		"integration_id":   integration.ID,
		"integration_name": integration.Name,
	}).Info("Integration imported successfully")
	
	return nil
}

// updateExistingIntegration updates an existing integration with imported data
func updateExistingIntegration(existing *models.Integration, data *IntegrationExportData, options ImportOptions, result *ImportResult) error {
	// Update fields
	existing.SourceAPI = data.SourceAPI
	existing.TargetAPI = data.TargetAPI
	existing.HTTPMethod = data.HTTPMethod
	existing.Mode = data.Mode
	existing.MappingConfig = data.MappingConfig
	existing.OutputTemplate = data.OutputTemplate
	existing.TemplateType = data.TemplateType
	
	existing.AuthType = data.AuthType
	existing.OAuth2TokenURL = data.OAuth2TokenURL
	existing.OAuth2ClientID = data.OAuth2ClientID
	existing.OAuth2ClientSecret = data.OAuth2ClientSecret
	existing.OAuth2Scope = data.OAuth2Scope
	existing.OAuth2GrantType = data.OAuth2GrantType
	existing.BearerToken = data.BearerToken
	existing.BasicAuthUser = data.BasicAuthUser
	existing.BasicAuthPass = data.BasicAuthPass
	
	existing.WebhookSignatureEnabled = data.WebhookSignatureEnabled
	existing.WebhookSignatureHeader = data.WebhookSignatureHeader
	existing.WebhookSignatureAlgorithm = data.WebhookSignatureAlgorithm
	
	// Handle signature secret
	if options.GenerateNewTokens && data.WebhookSignatureEnabled {
		secret, err := utils.GenerateToken(32)
		if err != nil {
			return fmt.Errorf("failed to generate signature secret: %w", err)
		}
		existing.WebhookSignatureSecret = secret
	} else if data.WebhookSignatureSecret != "" {
		existing.WebhookSignatureSecret = data.WebhookSignatureSecret
	}
	
	if err := database.DB.Save(existing).Error; err != nil {
		return fmt.Errorf("failed to update integration: %w", err)
	}
	
	result.UpdatedCount++
	logger.Log.WithFields(map[string]interface{}{
		"integration_id":   existing.ID,
		"integration_name": existing.Name,
	}).Info("Integration updated successfully")
	
	return nil
}

// ValidateImportData validates import data without importing
func ValidateImportData(jsonData []byte) (*ExportFormat, error) {
	var exportData ExportFormat
	if err := json.Unmarshal(jsonData, &exportData); err != nil {
		return nil, errors.New("invalid JSON format")
	}
	
	if exportData.Version != "1.0" {
		return nil, fmt.Errorf("unsupported export version: %s", exportData.Version)
	}
	
	if len(exportData.Integrations) == 0 {
		return nil, errors.New("no integrations found in import data")
	}
	
	// Validate each integration
	for i, integration := range exportData.Integrations {
		if integration.Name == "" {
			return nil, fmt.Errorf("integration %d: name is required", i+1)
		}
		if integration.TargetAPI == "" {
			return nil, fmt.Errorf("integration %d (%s): target API is required", i+1, integration.Name)
		}
	}
	
	return &exportData, nil
}
