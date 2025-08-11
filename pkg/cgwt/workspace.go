package cgwt

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// WorkspaceConfig represents a VS Code/Cursor workspace configuration.
type WorkspaceConfig struct {
	Name       string                 `json:"name,omitempty"`
	Folders    []WorkspaceFolder      `json:"folders"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// WorkspaceFolder represents a folder in a workspace configuration.
type WorkspaceFolder struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

// detectWorkspaceMode checks if the current directory contains workspace files.
func (c *cgwt) detectWorkspaceMode() ([]string, error) {
	c.verbosePrint("Checking for .code-workspace files...")

	// Check for workspace files
	workspaceFiles, err := c.fs.Glob("*.code-workspace")
	if err != nil {
		return nil, fmt.Errorf("failed to check for workspace files: %w", err)
	}

	if len(workspaceFiles) == 0 {
		c.verbosePrint("No .code-workspace files found")
		return nil, nil
	}

	c.verbosePrint(fmt.Sprintf("Found %d workspace file(s)", len(workspaceFiles)))
	return workspaceFiles, nil
}

// getWorkspaceInfo parses and validates a workspace configuration file.
func (c *cgwt) getWorkspaceInfo(workspaceFile string) (*WorkspaceConfig, error) {
	// Parse workspace configuration
	workspaceConfig, err := c.parseWorkspaceFile(workspaceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Validate workspace repositories
	err = c.validateWorkspaceRepositories(workspaceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to validate workspace repositories: %w", err)
	}

	c.verbosePrint("Workspace mode detected")
	return workspaceConfig, nil
}

// parseWorkspaceFile parses a workspace configuration file.
func (c *cgwt) parseWorkspaceFile(filename string) (*WorkspaceConfig, error) {
	c.verbosePrint("Parsing workspace configuration...")

	// Read workspace file
	content, err := c.fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace file: %w", err)
	}

	// Parse JSON
	var config WorkspaceConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("invalid .code-workspace file: malformed JSON")
	}

	// Validate folders array
	if config.Folders == nil {
		return nil, fmt.Errorf("workspace file must contain non-empty folders array")
	}

	// Filter out null values and validate structure
	var validFolders []WorkspaceFolder
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
		return nil, fmt.Errorf("workspace file must contain non-empty folders array")
	}

	config.Folders = validFolders
	return &config, nil
}

// validateWorkspaceRepositories validates all repository paths in the workspace.
func (c *cgwt) validateWorkspaceRepositories(config *WorkspaceConfig) error {
	c.verbosePrint("Validating repository paths...")

	// Track resolved paths to check for duplicates
	resolvedPaths := make(map[string]bool)

	for _, folder := range config.Folders {
		// Resolve relative paths relative to current working directory
		resolvedPath := folder.Path
		if !filepath.IsAbs(folder.Path) {
			resolvedPath = filepath.Join(".", folder.Path)
		}

		// Normalize path separators
		cleanPath := filepath.Clean(resolvedPath)

		// Check for duplicates
		if resolvedPaths[cleanPath] {
			return fmt.Errorf("duplicate repository paths found after resolution: %s", folder.Path)
		}
		resolvedPaths[cleanPath] = true

		// Validate that the path exists
		exists, err := c.fs.Exists(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to check repository path: %w", err)
		}
		if !exists {
			return fmt.Errorf("workspace repository not found: %s", folder.Path)
		}

		// Validate that it's a directory
		isDir, err := c.fs.IsDir(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to check repository directory: %w", err)
		}
		if !isDir {
			return fmt.Errorf("workspace repository is not a git repository: %s", folder.Path)
		}

		// Check for .git directory
		gitPath := filepath.Join(cleanPath, ".git")
		gitExists, err := c.fs.Exists(gitPath)
		if err != nil {
			return fmt.Errorf("failed to check .git directory: %w", err)
		}
		if !gitExists {
			return fmt.Errorf("workspace repository is not a git repository: %s", folder.Path)
		}

		c.verbosePrint(fmt.Sprintf("Validated repository: %s", folder.Path))
	}

	return nil
}

// getWorkspaceName extracts the workspace name from configuration or filename.
func (c *cgwt) getWorkspaceName(config *WorkspaceConfig, filename string) string {
	// First try to get name from workspace configuration
	if config.Name != "" {
		return config.Name
	}

	// Fallback to filename without extension
	return strings.TrimSuffix(filepath.Base(filename), ".code-workspace")
}
