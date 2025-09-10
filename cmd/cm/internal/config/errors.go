// Package config provides common configuration and utility functions for the CM CLI.
package config

import "errors"

// Error definitions for config package.
var (
	// Configuration loading errors.
	ErrFailedToLoadConfig = errors.New("failed to load configuration")
)
