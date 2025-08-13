// Package fs provides file system operations and error definitions.
package fs

import "errors"

// Error definitions for fs package.
var (
	// File lock errors.
	ErrFileLock = errors.New("lock")
)
