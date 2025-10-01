// Package config provides configuration management functionality for the CM application.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -source=manager.go -destination=mocks/manager.gen.go -package=mocks

// Manager interface provides configuration management functionality with an embedded config path.
type Manager interface {
	GetConfig() (Config, error)
	GetConfigStrict() (Config, error)
	GetConfigWithFallback() (Config, error)
	SaveConfig(config Config) error
	CreateConfigDirectory() error
	GetConfigPath() string
	SetConfigPath(configPath string)
	DefaultConfig() Config
	ValidateRepositoriesDir(repositoriesDir string) error
	ValidateWorkspacesDir(workspacesDir string) error
	ValidateStatusFile(statusFile string) error
}

// realManager manages configuration with an embedded config path.
type realManager struct {
	configPath string
}

// NewManager creates a new Manager instance with the specified config path.
func NewManager(configPath string) Manager {
	return &realManager{
		configPath: configPath,
	}
}

// NewConfigManager creates a new Manager instance with the specified config path.
func NewConfigManager(configPath string) Manager {
	return &realManager{
		configPath: configPath,
	}
}

// GetConfig loads configuration from the embedded config path.
func (c *realManager) GetConfig() (Config, error) {
	// Check if config file exists
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("%w: %s", ErrConfigNotInitialized, c.configPath)
	}

	// Read config file
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrConfigFileParse, err)
	}

	// Expand tildes in configuration paths
	if err := config.expandTildes(); err != nil {
		return Config{}, fmt.Errorf("failed to expand tildes in configuration: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// GetConfigStrict loads configuration and returns an error if the file is missing.
func (c *realManager) GetConfigStrict() (Config, error) {
	return c.GetConfig()
}

// GetConfigWithFallback loads the configuration from the embedded config path, falling back to default if not found.
func (c *realManager) GetConfigWithFallback() (Config, error) {
	// Try to load from file first
	if config, err := c.GetConfig(); err == nil {
		return config, nil
	}

	// Fallback to default configuration
	return c.DefaultConfig(), nil
}

// SaveConfig saves configuration to the embedded config path.
func (c *realManager) SaveConfig(config Config) error {
	// Create config directory if it doesn't exist
	if err := c.CreateConfigDirectory(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write configuration file atomically
	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// CreateConfigDirectory creates the configuration directory structure.
func (c *realManager) CreateConfigDirectory() error {
	configDir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return nil
}

// GetConfigPath returns the embedded config path.
func (c *realManager) GetConfigPath() string {
	return c.configPath
}

// SetConfigPath updates the embedded config path.
func (c *realManager) SetConfigPath(configPath string) {
	c.configPath = configPath
}

// DefaultConfig returns the default configuration.
func (c *realManager) DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory cannot be determined
		homeDir = "."
	}

	repositoriesDir := filepath.Join(homeDir, "Code", "repos")
	workspacesDir := filepath.Join(homeDir, "Code", "workspaces")
	statusFile := filepath.Join(homeDir, ".cm", "status.yaml")

	return Config{
		RepositoriesDir: repositoriesDir,
		WorkspacesDir:   workspacesDir,
		StatusFile:      statusFile,
	}
}

// ValidateRepositoriesDir validates the repositories directory for accessibility and permissions.
func (c *realManager) ValidateRepositoriesDir(repositoriesDir string) error {
	if repositoriesDir == "" {
		return ErrRepositoriesDirEmpty
	}

	// Check if directory exists
	if _, err := os.Stat(repositoriesDir); os.IsNotExist(err) {
		// Try to create the directory to validate permissions
		if err := os.MkdirAll(repositoriesDir, 0755); err != nil {
			return fmt.Errorf("repositories directory is not accessible: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("repositories directory is not accessible: %w", err)
	}

	// Check if directory is writable
	if err := c.validateDirectoryWritable(repositoriesDir); err != nil {
		return fmt.Errorf("repositories directory is not writable: %w", err)
	}

	return nil
}

// ValidateWorkspacesDir validates the workspaces directory for accessibility and permissions.
func (c *realManager) ValidateWorkspacesDir(workspacesDir string) error {
	if workspacesDir == "" {
		return ErrWorkspacesDirEmpty
	}

	// Check if directory exists
	if _, err := os.Stat(workspacesDir); os.IsNotExist(err) {
		// Try to create the directory to validate permissions
		if err := os.MkdirAll(workspacesDir, 0755); err != nil {
			return fmt.Errorf("workspaces directory is not accessible: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("workspaces directory is not accessible: %w", err)
	}

	// Check if directory is writable
	if err := c.validateDirectoryWritable(workspacesDir); err != nil {
		return fmt.Errorf("workspaces directory is not writable: %w", err)
	}

	return nil
}

// ValidateStatusFile validates the status file path for accessibility and permissions.
func (c *realManager) ValidateStatusFile(statusFile string) error {
	if statusFile == "" {
		return ErrStatusFileEmpty
	}

	// Get the directory containing the status file
	statusDir := filepath.Dir(statusFile)

	// Check if directory exists
	if _, err := os.Stat(statusDir); os.IsNotExist(err) {
		// Try to create the directory to validate permissions
		if err := os.MkdirAll(statusDir, 0755); err != nil {
			return fmt.Errorf("status file directory is not accessible: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("status file directory is not accessible: %w", err)
	}

	// Check if directory is writable
	if err := c.validateDirectoryWritable(statusDir); err != nil {
		return fmt.Errorf("status file directory is not writable: %w", err)
	}

	return nil
}

// validateDirectoryWritable checks if a directory is writable.
func (c *realManager) validateDirectoryWritable(path string) error {
	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(path, ".cm_test_write")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	// Clean up test file
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the error but don't fail the test
			fmt.Printf("Warning: failed to close test file: %v\n", closeErr)
		}
		if removeErr := os.Remove(testFile); removeErr != nil {
			// Log the error but don't fail the test
			fmt.Printf("Warning: failed to remove test file: %v\n", removeErr)
		}
	}()
	return nil
}
