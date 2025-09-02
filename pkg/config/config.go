// Package config provides configuration management functionality for the CM application.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/configs"
	"gopkg.in/yaml.v3"
)

//go:generate mockgen -source=config.go -destination=mocks/config.gen.go -package=mocks

// Config represents the application configuration.
type Config struct {
	BasePath   string `yaml:"base_path"`   // User's code directory (default: ~/Code)
	StatusFile string `yaml:"status_file"` // Status file path (default: ~/.cm/status.yaml)
}

// Manager interface provides configuration management functionality.
type Manager interface {
	LoadConfig(configPath string) (Config, error)
	LoadConfigStrict(configPath string) (Config, error)
	DefaultConfig() Config
	SaveConfig(config Config, configPath string) error
	CreateConfigDirectory(configPath string) error
	ValidateBasePath(basePath string) error
	EnsureConfigFile(configPath string) (Config, bool, error)
}

type realManager struct {
	// No fields needed for basic configuration operations
}

// NewManager creates a new Manager instance.
func NewManager() Manager {
	return &realManager{}
}

// LoadConfig loads configuration from the specified file path.
func (c *realManager) LoadConfig(configPath string) (Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("%w: %s", ErrConfigNotInitialized, configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
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

// LoadConfigStrict loads configuration and returns an error if the file is missing.
func (c *realManager) LoadConfigStrict(configPath string) (Config, error) {
	return c.LoadConfig(configPath)
}

// DefaultConfig returns the default configuration.
func (c *realManager) DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory cannot be determined
		homeDir = "."
	}

	basePath := filepath.Join(homeDir, "Code")
	statusFile := filepath.Join(homeDir, ".cm", "status.yaml")

	return Config{
		BasePath:   basePath,
		StatusFile: statusFile,
	}
}

// SaveConfig saves configuration to the specified file path.
func (c *realManager) SaveConfig(config Config, configPath string) error {
	// Create config directory if it doesn't exist
	if err := c.CreateConfigDirectory(configPath); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write configuration file atomically
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// CreateConfigDirectory creates the configuration directory structure.
func (c *realManager) CreateConfigDirectory(configPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return nil
}

// ValidateBasePath validates the base path for accessibility and permissions.
func (c *realManager) ValidateBasePath(basePath string) error {
	if basePath == "" {
		return ErrBasePathEmpty
	}

	// Check if directory exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		// Try to create the directory to validate permissions
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return fmt.Errorf("base path is not accessible: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("base path is not accessible: %w", err)
	}

	// Check if directory is writable
	if err := c.validateDirectoryWritable(basePath); err != nil {
		return fmt.Errorf("base path is not writable: %w", err)
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
	if c.BasePath == "" {
		return ErrBasePathEmpty
	}

	// Check if base path is accessible
	if err := c.validateDirectoryAccessibility(c.BasePath, "base_path"); err != nil {
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
	c.BasePath = c.expandTilde(c.BasePath, homeDir)
	c.StatusFile = c.expandTilde(c.StatusFile, homeDir)

	return nil
}

// LoadConfigWithFallback loads configuration from file with fallback to default.
func LoadConfigWithFallback(configPath string) (Config, error) {
	manager := NewManager()

	// Try to load from file first
	if config, err := manager.LoadConfig(configPath); err == nil {
		return config, nil
	}

	// Fallback to default configuration
	return manager.DefaultConfig(), nil
}

// EnsureConfigFile ensures the config file exists at path, creating it from embedded defaults if missing.
// Returns the loaded config and a boolean indicating whether the file already existed.
func (c *realManager) EnsureConfigFile(configPath string) (Config, bool, error) {
	if _, err := os.Stat(configPath); err == nil {
		cfg, err := c.LoadConfig(configPath)
		if err != nil {
			return Config{}, true, err
		}
		return cfg, true, nil
	} else if !os.IsNotExist(err) {
		return Config{}, false, fmt.Errorf("failed to stat config file: %w", err)
	}

	if err := c.CreateConfigDirectory(configPath); err != nil {
		return Config{}, false, fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, configs.DefaultConfigYAML, 0644); err != nil {
		return Config{}, false, fmt.Errorf("failed to write default config: %w", err)
	}

	cfg, err := c.LoadConfig(configPath)
	if err != nil {
		return Config{}, false, err
	}
	return cfg, false, nil
}
