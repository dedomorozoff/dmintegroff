package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"dmintegroff/internal/services"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ExportIntegrationsHandler exports selected integrations
func ExportIntegrationsHandler(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	if userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user session"})
		return
	}
	
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	
	// Get integration IDs from query parameter
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Integration IDs are required"})
		return
	}
	
	// Parse IDs
	idStrings := strings.Split(idsParam, ",")
	integrationIDs := make([]uint, 0, len(idStrings))
	for _, idStr := range idStrings {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid integration ID: %s", idStr)})
			return
		}
		integrationIDs = append(integrationIDs, uint(id))
	}
	
	// Export integrations
	jsonData, err := services.ExportIntegrations(integrationIDs, user.Username)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to export integrations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Set headers for file download
	filename := fmt.Sprintf("integrations_export_%s.json", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jsonData)
}

// ExportProjectIntegrationsHandler exports all integrations from a project
func ExportProjectIntegrationsHandler(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	if userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user session"})
		return
	}
	
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	// Export project integrations
	jsonData, err := services.ExportProjectIntegrations(uint(projectID), user.Username)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":    userID,
			"project_id": projectID,
			"error":      err.Error(),
		}).Error("Failed to export project integrations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Set headers for file download
	filename := fmt.Sprintf("project_%d_export_%s.json", projectID, time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jsonData)
}

// ImportIntegrationsHandler imports integrations from JSON file
func ImportIntegrationsHandler(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	if userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user session"})
		return
	}
	
	// Get project ID
	projectID, err := strconv.ParseUint(c.PostForm("project_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	// Get import options
	skipDuplicates := c.PostForm("skip_duplicates") == "true"
	updateExisting := c.PostForm("update_existing") == "true"
	generateNewTokens := c.PostForm("generate_new_tokens") == "true"
	
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	
	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".json") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JSON files are allowed"})
		return
	}
	
	// Read file content
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer fileContent.Close()
	
	jsonData, err := io.ReadAll(fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
		return
	}
	
	// Import integrations
	options := services.ImportOptions{
		ProjectID:         uint(projectID),
		UserID:            userID,
		SkipDuplicates:    skipDuplicates,
		UpdateExisting:    updateExisting,
		GenerateNewTokens: generateNewTokens,
	}
	
	result, err := services.ImportIntegrations(jsonData, options)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":    userID,
			"project_id": projectID,
			"error":      err.Error(),
		}).Error("Failed to import integrations")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"user_id":    userID,
		"project_id": projectID,
		"imported":   result.ImportedCount,
		"updated":    result.UpdatedCount,
		"skipped":    result.SkippedCount,
	}).Info("Integrations imported successfully")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Import completed",
		"result":  result,
	})
}

// ValidateImportHandler validates import file without importing
func ValidateImportHandler(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	if userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	
	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".json") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JSON files are allowed"})
		return
	}
	
	// Read file content
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer fileContent.Close()
	
	jsonData, err := io.ReadAll(fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
		return
	}
	
	// Validate import data
	exportData, err := services.ValidateImportData(jsonData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"valid":             true,
		"version":           exportData.Version,
		"exported_at":       exportData.ExportedAt,
		"exported_by":       exportData.ExportedBy,
		"project_name":      exportData.ProjectName,
		"integrations_count": len(exportData.Integrations),
	})
}


// ExportToGitHandler exports integrations to Git repository
func ExportToGitHandler(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	if userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user session"})
		return
	}
	
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	
	// Get integration IDs from query parameter
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Integration IDs are required"})
		return
	}
	
	// Parse IDs
	idStrings := strings.Split(idsParam, ",")
	integrationIDs := make([]uint, 0, len(idStrings))
	for _, idStr := range idStrings {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid integration ID: %s", idStr)})
			return
		}
		integrationIDs = append(integrationIDs, uint(id))
	}
	
	// Get Git config from form
	repoPath := c.PostForm("repo_path")
	filePath := c.PostForm("file_path")
	branch := c.PostForm("branch")
	commitMsg := c.PostForm("commit_msg")
	autoPush := c.PostForm("auto_push") == "true"
	
	if repoPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository path is required"})
		return
	}
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}
	
	config := services.GitExportConfig{
		RepoPath:    repoPath,
		FilePath:    filePath,
		Branch:      branch,
		CommitMsg:   commitMsg,
		AutoPush:    autoPush,
		AuthorEmail: user.Username,
	}
	
	// Export to Git
	result, err := services.ExportToGit(integrationIDs, user.Username, config)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":   userID,
			"repo_path": repoPath,
			"error":     err.Error(),
		}).Error("Failed to export to Git")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Export to Git completed",
		"result":  result,
	})
}

// ExportProjectToGitHandler exports project integrations to Git repository
func ExportProjectToGitHandler(c *gin.Context) {
	session := sessions.Default(c)
	userIDInterface := session.Get("user_id")
	if userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user session"})
		return
	}
	
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	// Get Git config from form
	repoPath := c.PostForm("repo_path")
	filePath := c.PostForm("file_path")
	branch := c.PostForm("branch")
	commitMsg := c.PostForm("commit_msg")
	autoPush := c.PostForm("auto_push") == "true"
	
	if repoPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository path is required"})
		return
	}
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}
	
	config := services.GitExportConfig{
		RepoPath:    repoPath,
		FilePath:    filePath,
		Branch:      branch,
		CommitMsg:   commitMsg,
		AutoPush:    autoPush,
		AuthorEmail: user.Username,
	}
	
	// Export to Git
	result, err := services.ExportProjectToGit(uint(projectID), user.Username, config)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":    userID,
			"project_id": projectID,
			"repo_path":  repoPath,
			"error":      err.Error(),
		}).Error("Failed to export project to Git")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Export to Git completed",
		"result":  result,
	})
}
