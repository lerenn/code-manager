// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/status"
)

// DetectWorkspaceFiles checks if the current directory contains workspace files.
func (w *realWorkspace) DetectWorkspaceFiles() ([]string, error) {
	w.verboseLogf("Checking for .code-workspace files...")

	// Check for workspace files
	workspaceFiles, err := w.fs.Glob("*.code-workspace")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToCheckWorkspaceFiles, err)
	}

	if len(workspaceFiles) == 0 {
		w.verboseLogf("No .code-workspace files found")
		return nil, nil
	}

	w.verboseLogf("Found %d workspace file(s)", len(workspaceFiles))
	return workspaceFiles, nil
}

// ParseFile parses a workspace configuration file.
func (w *realWorkspace) ParseFile(filename string) (*Config, error) {
	w.verboseLogf("Parsing workspace configuration...")

	// Read workspace file
	content, err := w.fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWorkspaceFileNotFound, err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, ErrInvalidWorkspaceFile
	}

	// Validate folders array
	if config.Folders == nil {
		return nil, ErrNoRepositoriesFound
	}

	// Filter out null values and validate structure
	var validFolders []Folder
	for _, folder := range config.Folders {
		// Skip null values
		if folder.Path == "" {
			continue
		}

		// Validate path field
		if folder.Path == "" {
			return nil, fmt.Errorf("workspace folder must contain path field")
		}

		// Validate name field if present
		// Name field is optional, but if present it should be a string
		// This is already handled by JSON unmarshaling

		validFolders = append(validFolders, folder)
	}

	// Check if we have any valid folders after filtering
	if len(validFolders) == 0 {
		return nil, ErrNoRepositoriesFound
	}

	config.Folders = validFolders
	return &config, nil
}

// GetName extracts the workspace name from configuration or filename.
func (w *realWorkspace) GetName(config *Config, filename string) string {
	// First try to get name from workspace configuration
	if config.Name != "" {
		return config.Name
	}

	// Fallback to filename without extension
	return strings.TrimSuffix(filepath.Base(filename), ".code-workspace")
}

// getWorkspacePath gets the absolute path for the workspace file.
func (w *realWorkspace) getWorkspacePath() (string, error) {
	workspacePath, err := filepath.Abs(w.OriginalFile)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}
	return workspacePath, nil
}

// createWorktreeWorkspaceFileParams contains parameters for creating a worktree workspace file.
type createWorktreeWorkspaceFileParams struct {
	WorkspaceConfig       *Config
	WorkspaceName         string
	Branch                string
	WorktreeWorkspacePath string
}

// createWorktreeWorkspaceFile creates the worktree-specific workspace file.
func (w *realWorkspace) createWorktreeWorkspaceFile(params createWorktreeWorkspaceFileParams) error {
	w.verboseLogf("Creating worktree-specific workspace file")

	// Ensure workspaces directory exists
	workspacesDir := filepath.Dir(params.WorktreeWorkspacePath)
	if err := w.fs.MkdirAll(workspacesDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspaces directory: %w", err)
	}

	// Sanitize branch name for workspace name (replace slashes with hyphens)
	sanitizedBranchForName := strings.ReplaceAll(params.Branch, "/", "-")

	// Create worktree workspace configuration
	worktreeConfig := struct {
		Name    string   `json:"name,omitempty"`
		Folders []Folder `json:"folders"`
	}{
		Name:    fmt.Sprintf("%s-%s", params.WorkspaceName, sanitizedBranchForName),
		Folders: make([]Folder, len(params.WorkspaceConfig.Folders)),
	}

	// Update folder paths to point to worktree directories
	for i, folder := range params.WorkspaceConfig.Folders {
		// Get repository URL for this folder
		resolvedPath := filepath.Join(filepath.Dir(w.OriginalFile), folder.Path)
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}

		worktreeConfig.Folders[i] = Folder{
			Name: folder.Name,
			Path: w.buildWorktreePath(repoURL, "origin", params.Branch),
		}
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(worktreeConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal worktree workspace config: %w", err)
	}

	// Write worktree workspace file
	if err := w.fs.WriteFileAtomic(params.WorktreeWorkspacePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write worktree workspace file: %w", err)
	}

	w.verboseLogf("Worktree workspace file created: %s", params.WorktreeWorkspacePath)
	return nil
}

// ensureWorkspaceInStatus ensures the workspace is added to the status file.
func (w *realWorkspace) ensureWorkspaceInStatus(workspaceConfig *Config, workspaceDir string) error {
	// Get workspace path
	workspacePath, err := filepath.Abs(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	// Check if workspace already exists in status
	_, err = w.statusManager.GetWorkspace(workspacePath)
	if err == nil {
		// Workspace already exists, no need to add it
		return nil
	}

	// Collect repository URLs for the workspace
	var repoURLs []string
	for _, folder := range workspaceConfig.Folders {
		resolvedPath := filepath.Join(workspaceDir, folder.Path)
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}
		repoURLs = append(repoURLs, repoURL)
	}

	// Add workspace to status file with worktree reference
	// The worktree reference will be updated when worktrees are created
	if err := w.statusManager.AddWorkspace(workspacePath, status.AddWorkspaceParams{
		Worktree:     "", // Will be set when worktrees are created
		Repositories: repoURLs,
	}); err != nil {
		return fmt.Errorf("failed to add workspace to status file: %w", err)
	}

	return nil
}
