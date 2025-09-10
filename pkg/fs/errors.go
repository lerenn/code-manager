// Package fs provides file system operations and error definitions.
package fs

import "errors"

// Error definitions for fs package.
var (
	// File lock errors.
	ErrFileLock = errors.New("lock")

	// Path resolution errors.
	ErrPathResolution = errors.New("path resolution failed")

	// Repository validation errors.
	ErrInvalidRepository = errors.New("invalid repository path")
)
