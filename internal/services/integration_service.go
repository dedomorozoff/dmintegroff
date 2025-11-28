package services

import (
	"bytes"
	"encoding/json"
	"gintegra/internal/database"
	"gintegra/internal/models"
	"net/http"
)

func ProcessWebhook(integrationID uint, payload map[string]interface{}) error {
	var integration models.Integration
	if err := database.DB.First(&integration, integrationID).Error; err != nil {
		return err
	}

	// Parse mapping config
	// Expected format: {"target_field": "source_field"}
	var mapping map[string]string
	if integration.MappingConfig != "" {
		if err := json.Unmarshal([]byte(integration.MappingConfig), &mapping); err != nil {
			// Log error but proceed? Or fail?
			// For now, let's treat invalid mapping as empty
		}
	}

	// Transform data
	transformed := make(map[string]interface{})
	if len(mapping) > 0 {
		for targetField, sourceField := range mapping {
			if val, ok := payload[sourceField]; ok {
				transformed[targetField] = val
			}
		}
	} else {
		// If no mapping, pass through all data
		transformed = payload
	}

	// Send to target
	jsonData, _ := json.Marshal(transformed)
	resp, err := http.Post(integration.TargetAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
