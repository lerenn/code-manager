package config

import "errors"

// Error definitions for config package.
var (
	// Configuration file errors.
	ErrConfigFileNotFound = errors.New("config file not found")
	ErrConfigFileParse    = errors.New("failed to parse config file")
	// Configuration validation errors.
	ErrBasePathEmpty = errors.New("base_path cannot be empty")
)
