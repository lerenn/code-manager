package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=config.go -destination=mockconfig.gen.go -package=config

// Config represents the application configuration.
type Config struct {
	BasePath   string `yaml:"base_path"`
	StatusFile string `yaml:"status_file"`
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

	basePath := filepath.Join(homeDir, ".cursor", "wtm")
	statusFile := filepath.Join(basePath, "status.yaml")

	return &Config{
		BasePath:   basePath,
		StatusFile: statusFile,
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
		os.RemoveAll(dir)
	} else if err != nil {
		return fmt.Errorf("base_path parent directory is not accessible: %w", err)
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
