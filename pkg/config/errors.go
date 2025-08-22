package config

import "errors"

// Error definitions for config package.
var (
	// Configuration file errors.
	ErrConfigFileParse = errors.New("failed to parse config file")
	// Configuration validation errors.
	ErrBasePathEmpty = errors.New("base_path cannot be empty")
	// Configuration initialization errors.
	ErrConfigNotInitialized = errors.New("CM configuration not found. Run 'cm init' to initialize")
)
