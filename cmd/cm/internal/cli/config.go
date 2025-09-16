// Package cli provides common configuration and utility functions for the CM CLI.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/status"
)

var (
	// Quiet suppresses all output except errors.
	Quiet bool
	// Verbose enables verbose output.
	Verbose bool
	// ConfigPath specifies a custom config file path.
	ConfigPath string
)

// LoadConfig loads the configuration and returns an error if not found.
func LoadConfig() (config.Config, error) {
	configManager := NewConfigManager()
	return configManager.GetConfigStrict()
}

// NewConfigManager creates a new Manager with the appropriate config path.
func NewConfigManager() config.Manager {
	var path string
	if ConfigPath != "" {
		path = ConfigPath
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		path = filepath.Join(homeDir, ".cm", "config.yaml")
	}

	return config.NewConfigManager(path)
}

// GetConfigPath returns the config file path that would be used by LoadConfig.
func GetConfigPath() string {
	if ConfigPath != "" {
		return ConfigPath
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".cm", "config.yaml")
}

// CheckInitialization checks if CM is initialized and returns an error if not.
func CheckInitialization() error {
	configManager := NewConfigManager()
	cfg, err := configManager.GetConfigStrict()
	if err != nil {
		return err
	}

	fsInstance := fs.NewFS()

	// Check if status file exists
	exists, err := fsInstance.Exists(cfg.StatusFile)
	if err != nil {
		return fmt.Errorf("failed to check status file existence: %w", err)
	}

	if !exists {
		return status.ErrNotInitialized
	}

	return nil
}
