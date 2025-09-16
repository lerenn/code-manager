// Package config provides configuration management functionality for the CM application.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the application configuration.
type Config struct {
	RepositoriesDir string `yaml:"repositories_dir"` // User's repositories directory (default: ~/Code/repos)
	WorkspacesDir   string `yaml:"workspaces_dir"`   // User's workspaces directory (default: ~/Code/workspaces)
	StatusFile      string `yaml:"status_file"`      // Status file path (default: ~/.cm/status.yaml)
}

// validateDirectoryAccessibility checks if a directory path is accessible and can be created.
func (c Config) validateDirectoryAccessibility(path, pathName string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Try to create the parent directory to validate permissions
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%s parent directory is not accessible: %w", pathName, err)
		}
		// Clean up the test directory
		if err := os.RemoveAll(dir); err != nil {
			// Log the error but don't fail validation for cleanup errors
			_ = err
		}
	} else if err != nil {
		return fmt.Errorf("%s parent directory is not accessible: %w", pathName, err)
	}
	return nil
}

// Validate validates the configuration values.
func (c Config) Validate() error {
	if c.RepositoriesDir == "" {
		return ErrRepositoriesDirEmpty
	}

	if c.WorkspacesDir == "" {
		return ErrWorkspacesDirEmpty
	}

	// Check if repositories directory is accessible
	if err := c.validateDirectoryAccessibility(c.RepositoriesDir, "repositories_dir"); err != nil {
		return err
	}

	// Check if workspaces directory is accessible
	if err := c.validateDirectoryAccessibility(c.WorkspacesDir, "workspaces_dir"); err != nil {
		return err
	}

	return nil
}

// expandTilde expands a single path if it starts with tilde.
func (c *Config) expandTilde(path string, homeDir string) string {
	if strings.HasPrefix(path, "~") {
		return filepath.Join(homeDir, strings.TrimPrefix(path, "~"))
	}
	return path
}

// expandTildes expands tilde (~) to the user's home directory in configuration paths.
func (c *Config) expandTildes() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	// Expand tildes in all paths
	c.RepositoriesDir = c.expandTilde(c.RepositoriesDir, homeDir)
	c.WorkspacesDir = c.expandTilde(c.WorkspacesDir, homeDir)
	c.StatusFile = c.expandTilde(c.StatusFile, homeDir)

	return nil
}
