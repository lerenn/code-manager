package fs

import (
	"fmt"
	"path/filepath"
)

// ResolvePath resolves relative paths from base directory.
func (f *realFS) ResolvePath(repositoriesDir, relativePath string) (string, error) {
	// Handle empty paths
	if repositoriesDir == "" {
		return "", fmt.Errorf("%w: base path cannot be empty", ErrPathResolution)
	}
	if relativePath == "" {
		return "", fmt.Errorf("%w: relative path cannot be empty", ErrPathResolution)
	}

	// If relativePath is already absolute, return it as-is
	if filepath.IsAbs(relativePath) {
		return filepath.Clean(relativePath), nil
	}

	// Resolve relative path from base directory
	resolvedPath := filepath.Join(repositoriesDir, relativePath)

	// Clean the resolved path to remove any ".." or "." components
	cleanPath := filepath.Clean(resolvedPath)

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("%w: failed to get absolute path for %s: %w", ErrPathResolution, cleanPath, err)
	}

	return absPath, nil
}
