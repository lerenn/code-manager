// Package config provides configuration management functionality for the WTM application.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:generate mockgen -source=config.go -destination=mockconfig.gen.go -package=config

// Config represents the application configuration.
type Config struct {
	BasePath     string `yaml:"base_path"`
	StatusFile   string `yaml:"status_file"`
	WorktreesDir string `yaml:"worktrees_dir"`
}

// Manager interface provides configuration management functionality.
type Manager interface {
	LoadConfig(configPath string) (*Config, error)
	DefaultConfig() *Config
}

type realManager struct {
	// No fields needed for basic configuration operations
}

// NewManager creates a new Manager instance.
func NewManager() Manager {
	return &realManager{}
}

// LoadConfig loads configuration from the specified file path.
func (c *realManager) LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrConfigFileNotFound, configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigFileParse, err)
	}

	// Expand tildes in configuration paths
	if err := config.expandTildes(); err != nil {
		return nil, fmt.Errorf("failed to expand tildes in configuration: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// DefaultConfig returns the default configuration.
func (c *realManager) DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory cannot be determined
		homeDir = "."
	}

	basePath := filepath.Join(homeDir, ".wtm")
	statusFile := filepath.Join(basePath, "status.yaml")
	worktreesDir := filepath.Join(basePath, "worktrees")

	return &Config{
		BasePath:     basePath,
		StatusFile:   statusFile,
		WorktreesDir: worktreesDir,
	}
}

// Validate validates the configuration values.
func (c *Config) Validate() error {
	if c.BasePath == "" {
		return ErrBasePathEmpty
	}

	// Check if base path is accessible (can be created if it doesn't exist)
	dir := filepath.Dir(c.BasePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Try to create the parent directory to validate permissions
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("base_path parent directory is not accessible: %w", err)
		}
		// Clean up the test directory
		if err := os.RemoveAll(dir); err != nil {
			// Log the error but don't fail validation for cleanup errors
			_ = err
		}
	} else if err != nil {
		return fmt.Errorf("base_path parent directory is not accessible: %w", err)
	}

	// Validate worktrees directory if specified
	if c.WorktreesDir != "" {
		worktreesDir := filepath.Dir(c.WorktreesDir)
		if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
			// Try to create the parent directory to validate permissions
			if err := os.MkdirAll(worktreesDir, 0755); err != nil {
				return fmt.Errorf("worktrees_dir parent directory is not accessible: %w", err)
			}
			// Clean up the test directory
			if err := os.RemoveAll(worktreesDir); err != nil {
				// Log the error but don't fail validation for cleanup errors
				_ = err
			}
		} else if err != nil {
			return fmt.Errorf("worktrees_dir parent directory is not accessible: %w", err)
		}
	}

	return nil
}

// expandTildes expands tilde (~) to the user's home directory in configuration paths.
func (c *Config) expandTildes() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	// Expand tilde in BasePath
	if strings.HasPrefix(c.BasePath, "~") {
		c.BasePath = filepath.Join(homeDir, strings.TrimPrefix(c.BasePath, "~"))
	}

	// Expand tilde in StatusFile
	if strings.HasPrefix(c.StatusFile, "~") {
		c.StatusFile = filepath.Join(homeDir, strings.TrimPrefix(c.StatusFile, "~"))
	}

	// Expand tilde in WorktreesDir
	if strings.HasPrefix(c.WorktreesDir, "~") {
		c.WorktreesDir = filepath.Join(homeDir, strings.TrimPrefix(c.WorktreesDir, "~"))
	}

	return nil
}

// LoadConfigWithFallback loads configuration from file with fallback to default.
func LoadConfigWithFallback(configPath string) (*Config, error) {
	manager := NewManager()

	// Try to load from file first
	if config, err := manager.LoadConfig(configPath); err == nil {
		return config, nil
	}

	// Fallback to default configuration
	return manager.DefaultConfig(), nil
}
