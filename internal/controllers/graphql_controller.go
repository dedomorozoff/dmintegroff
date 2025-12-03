package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/services"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// IntrospectGraphQLSchema fetches GraphQL schema using introspection
func IntrospectGraphQLSchema(c *gin.Context) {
	integrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	var integration models.Integration
	if err := database.DB.First(&integration, integrationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	endpoint := integration.GraphQLEndpoint
	if endpoint == "" {
		endpoint = integration.TargetAPI
	}

	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GraphQL endpoint not configured"})
		return
	}

	graphqlService := services.NewGraphQLService()
	schema, err := graphqlService.IntrospectSchema(endpoint, &integration)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"endpoint":       endpoint,
			"error":          err.Error(),
		}).Error("Failed to introspect GraphQL schema")
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to introspect schema: " + err.Error()})
		return
	}

	// Update schema in database
	now := time.Now().Unix()
	integration.GraphQLSchema = schema
	integration.GraphQLSchemaUpdatedAt = &now
	
	if err := database.DB.Save(&integration).Error; err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integrationID,
			"error":          err.Error(),
		}).Error("Failed to save GraphQL schema")
	}

	c.JSON(http.StatusOK, gin.H{
		"schema":     schema,
		"updated_at": now,
	})
}

// TestGraphQLConnection tests GraphQL endpoint connectivity
func TestGraphQLConnection(c *gin.Context) {
	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
		AuthType string `json:"auth_type"`
		BearerToken string `json:"bearer_token"`
		BasicAuthUser string `json:"basic_auth_user"`
		BasicAuthPass string `json:"basic_auth_pass"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create temporary integration for auth
	integration := &models.Integration{
		AuthType:      req.AuthType,
		BearerToken:   req.BearerToken,
		BasicAuthUser: req.BasicAuthUser,
		BasicAuthPass: req.BasicAuthPass,
	}

	graphqlService := services.NewGraphQLService()
	if err := graphqlService.TestGraphQLConnection(req.Endpoint, integration); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"endpoint": req.Endpoint,
			"error":    err.Error(),
		}).Warn("GraphQL connection test failed")
		
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "GraphQL connection successful",
	})
}

// TestGraphQLQuery tests a GraphQL query
func TestGraphQLQuery(c *gin.Context) {
	var req struct {
		IntegrationID uint                   `json:"integration_id" binding:"required"`
		Query         string                 `json:"query" binding:"required"`
		Variables     map[string]interface{} `json:"variables"`
		OperationName string                 `json:"operation_name"`
		TestPayload   map[string]interface{} `json:"test_payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var integration models.Integration
	if err := database.DB.First(&integration, req.IntegrationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Temporarily update query for testing
	originalQuery := integration.GraphQLQuery
	originalVars := integration.GraphQLVariables
	originalOpName := integration.GraphQLOperationName
	
	integration.GraphQLQuery = req.Query
	if req.Variables != nil {
		varsJSON, _ := json.Marshal(req.Variables)
		integration.GraphQLVariables = string(varsJSON)
	}
	integration.GraphQLOperationName = req.OperationName

	graphqlService := services.NewGraphQLService()
	
	// Use test payload or empty map
	payload := req.TestPayload
	if payload == nil {
		payload = make(map[string]interface{})
	}
	
	result, err := graphqlService.ExecuteQuery(&integration, payload)
	
	// Restore original values
	integration.GraphQLQuery = originalQuery
	integration.GraphQLVariables = originalVars
	integration.GraphQLOperationName = originalOpName
	
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": req.IntegrationID,
			"error":          err.Error(),
		}).Warn("GraphQL query test failed")
		
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetGraphQLSchema returns cached GraphQL schema
func GetGraphQLSchema(c *gin.Context) {
	integrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	var integration models.Integration
	if err := database.DB.First(&integration, integrationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	if integration.GraphQLSchema == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schema not cached. Run introspection first."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schema":     integration.GraphQLSchema,
		"updated_at": integration.GraphQLSchemaUpdatedAt,
	})
}
