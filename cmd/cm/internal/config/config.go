// Package config provides common configuration and utility functions for the CM CLI.
package config

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
	manager := config.NewManager()

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

	config, err := manager.LoadConfigStrict(path)
	if err != nil {
		return config, fmt.Errorf("%w: %w", ErrFailedToLoadConfig, err)
	}

	return config, nil
}

// CheckInitialization checks if CM is initialized and returns an error if not.
func CheckInitialization() error {
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
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
