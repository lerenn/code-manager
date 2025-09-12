package config

import "errors"

// Error definitions for config package.
var (
	// Configuration file errors.
	ErrConfigFileParse = errors.New("failed to parse config file")
	// Configuration validation errors.
	ErrRepositoriesDirEmpty = errors.New("repositories_dir cannot be empty")
	ErrWorkspacesDirEmpty   = errors.New("workspaces_dir cannot be empty")
	ErrStatusFileEmpty      = errors.New("status_file cannot be empty")
	// Configuration initialization errors.
	ErrConfigNotInitialized = errors.New("CM configuration not found. Run 'cm init' to initialize")
)
