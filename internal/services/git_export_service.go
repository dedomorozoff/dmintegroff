package services

import (
	"dmintegroff/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// GitExportConfig contains Git repository configuration
type GitExportConfig struct {
	RepoPath   string // Path to Git repository
	Branch     string // Branch to commit to (default: main)
	FilePath   string // Path within repo (e.g., "configs/integrations.json")
	CommitMsg  string // Commit message template
	AutoPush   bool   // Automatically push after commit
	AuthorName string // Git author name
	AuthorEmail string // Git author email
}

// GitExportResult contains the result of Git export operation
type GitExportResult struct {
	Success    bool   `json:"success"`
	CommitHash string `json:"commit_hash,omitempty"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
}

// ExportToGit exports integrations to a Git repository
func ExportToGit(integrationIDs []uint, userEmail string, config GitExportConfig) (*GitExportResult, error) {
	// Validate config
	if config.RepoPath == "" {
		return nil, fmt.Errorf("repository path is required")
	}
	if config.FilePath == "" {
		return nil, fmt.Errorf("file path is required")
	}
	
	// Set defaults
	if config.Branch == "" {
		config.Branch = "main"
	}
	if config.CommitMsg == "" {
		config.CommitMsg = fmt.Sprintf("Update integrations config - %s", time.Now().Format("2006-01-02 15:04:05"))
	}
	if config.AuthorName == "" {
		config.AuthorName = "dmIntegroff"
	}
	if config.AuthorEmail == "" {
		config.AuthorEmail = userEmail
	}
	
	// Check if repo exists
	if _, err := os.Stat(config.RepoPath); os.IsNotExist(err) {
		return &GitExportResult{
			Success: false,
			Error:   "repository path does not exist",
		}, fmt.Errorf("repository path does not exist: %s", config.RepoPath)
	}
	
	// Check if it's a git repo
	gitDir := filepath.Join(config.RepoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return &GitExportResult{
			Success: false,
			Error:   "not a git repository",
		}, fmt.Errorf("not a git repository: %s", config.RepoPath)
	}
	
	// Export integrations to JSON
	jsonData, err := ExportIntegrations(integrationIDs, userEmail)
	if err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("export failed: %s", err.Error()),
		}, err
	}
	
	// Write to file in repo
	fullPath := filepath.Join(config.RepoPath, config.FilePath)
	
	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create directory: %s", err.Error()),
		}, err
	}
	
	// Write file
	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %s", err.Error()),
		}, err
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"repo_path": config.RepoPath,
		"file_path": config.FilePath,
		"count":     len(integrationIDs),
	}).Info("File written to Git repository")
	
	// Git add
	if err := gitAdd(config.RepoPath, config.FilePath); err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("git add failed: %s", err.Error()),
		}, err
	}
	
	// Check if there are changes to commit
	hasChanges, err := gitHasChanges(config.RepoPath)
	if err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("failed to check git status: %s", err.Error()),
		}, err
	}
	
	if !hasChanges {
		return &GitExportResult{
			Success: true,
			Message: "No changes to commit (file unchanged)",
		}, nil
	}
	
	// Git commit
	commitHash, err := gitCommit(config.RepoPath, config.CommitMsg, config.AuthorName, config.AuthorEmail)
	if err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("git commit failed: %s", err.Error()),
		}, err
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"repo_path":   config.RepoPath,
		"commit_hash": commitHash,
		"author":      config.AuthorEmail,
	}).Info("Changes committed to Git")
	
	result := &GitExportResult{
		Success:    true,
		CommitHash: commitHash,
		Message:    fmt.Sprintf("Successfully committed to %s", config.Branch),
	}
	
	// Git push if enabled
	if config.AutoPush {
		if err := gitPush(config.RepoPath, config.Branch); err != nil {
			result.Message += fmt.Sprintf(" (push failed: %s)", err.Error())
			logger.Log.WithFields(map[string]interface{}{
				"repo_path": config.RepoPath,
				"branch":    config.Branch,
				"error":     err.Error(),
			}).Warn("Git push failed")
		} else {
			result.Message += " and pushed"
			logger.Log.WithFields(map[string]interface{}{
				"repo_path": config.RepoPath,
				"branch":    config.Branch,
			}).Info("Changes pushed to remote")
		}
	}
	
	return result, nil
}

// ExportProjectToGit exports all project integrations to Git
func ExportProjectToGit(projectID uint, userEmail string, config GitExportConfig) (*GitExportResult, error) {
	// Validate config
	if config.RepoPath == "" {
		return nil, fmt.Errorf("repository path is required")
	}
	if config.FilePath == "" {
		return nil, fmt.Errorf("file path is required")
	}
	
	// Set defaults
	if config.Branch == "" {
		config.Branch = "main"
	}
	if config.CommitMsg == "" {
		config.CommitMsg = fmt.Sprintf("Update project integrations - %s", time.Now().Format("2006-01-02 15:04:05"))
	}
	if config.AuthorName == "" {
		config.AuthorName = "dmIntegroff"
	}
	if config.AuthorEmail == "" {
		config.AuthorEmail = userEmail
	}
	
	// Check if repo exists
	if _, err := os.Stat(config.RepoPath); os.IsNotExist(err) {
		return &GitExportResult{
			Success: false,
			Error:   "repository path does not exist",
		}, fmt.Errorf("repository path does not exist: %s", config.RepoPath)
	}
	
	// Export project integrations to JSON
	jsonData, err := ExportProjectIntegrations(projectID, userEmail)
	if err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("export failed: %s", err.Error()),
		}, err
	}
	
	// Write to file in repo
	fullPath := filepath.Join(config.RepoPath, config.FilePath)
	
	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create directory: %s", err.Error()),
		}, err
	}
	
	// Write file
	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %s", err.Error()),
		}, err
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"repo_path":  config.RepoPath,
		"file_path":  config.FilePath,
		"project_id": projectID,
	}).Info("Project file written to Git repository")
	
	// Git add
	if err := gitAdd(config.RepoPath, config.FilePath); err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("git add failed: %s", err.Error()),
		}, err
	}
	
	// Check if there are changes to commit
	hasChanges, err := gitHasChanges(config.RepoPath)
	if err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("failed to check git status: %s", err.Error()),
		}, err
	}
	
	if !hasChanges {
		return &GitExportResult{
			Success: true,
			Message: "No changes to commit (file unchanged)",
		}, nil
	}
	
	// Git commit
	commitHash, err := gitCommit(config.RepoPath, config.CommitMsg, config.AuthorName, config.AuthorEmail)
	if err != nil {
		return &GitExportResult{
			Success: false,
			Error:   fmt.Sprintf("git commit failed: %s", err.Error()),
		}, err
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"repo_path":   config.RepoPath,
		"commit_hash": commitHash,
		"project_id":  projectID,
	}).Info("Project changes committed to Git")
	
	result := &GitExportResult{
		Success:    true,
		CommitHash: commitHash,
		Message:    fmt.Sprintf("Successfully committed to %s", config.Branch),
	}
	
	// Git push if enabled
	if config.AutoPush {
		if err := gitPush(config.RepoPath, config.Branch); err != nil {
			result.Message += fmt.Sprintf(" (push failed: %s)", err.Error())
			logger.Log.WithFields(map[string]interface{}{
				"repo_path": config.RepoPath,
				"branch":    config.Branch,
				"error":     err.Error(),
			}).Warn("Git push failed")
		} else {
			result.Message += " and pushed"
			logger.Log.WithFields(map[string]interface{}{
				"repo_path": config.RepoPath,
				"branch":    config.Branch,
			}).Info("Project changes pushed to remote")
		}
	}
	
	return result, nil
}

// Git helper functions

func gitAdd(repoPath, filePath string) error {
	cmd := exec.Command("git", "add", filePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %s - %s", err.Error(), string(output))
	}
	return nil
}

func gitHasChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %s", err.Error())
	}
	return len(output) > 0, nil
}

func gitCommit(repoPath, message, authorName, authorEmail string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", message, "--author", fmt.Sprintf("%s <%s>", authorName, authorEmail))
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git commit failed: %s - %s", err.Error(), string(output))
	}
	
	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	hashOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %s", err.Error())
	}
	
	return string(hashOutput[:7]), nil // Return short hash
}

func gitPush(repoPath, branch string) error {
	cmd := exec.Command("git", "push", "origin", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %s - %s", err.Error(), string(output))
	}
	return nil
}
