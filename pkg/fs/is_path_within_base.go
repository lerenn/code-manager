package fs

import (
	"fmt"
	"path/filepath"
	"strings"
)

// IsPathWithinBase checks if a target path is within the base path.
func (f *realFS) IsPathWithinBase(repositoriesDir, targetPath string) (bool, error) {
	// Handle empty paths
	if repositoriesDir == "" && targetPath == "" {
		return true, nil
	}
	if repositoriesDir == "" {
		return false, nil
	}

	// Normalize path separators - convert backslashes to forward slashes for cross-platform compatibility
	normalizedRepositoriesDir := strings.ReplaceAll(repositoriesDir, "\\", "/")
	normalizedTargetPath := strings.ReplaceAll(targetPath, "\\", "/")

	// Clean the paths
	cleanRepositoriesDir := filepath.Clean(normalizedRepositoriesDir)
	cleanTargetPath := filepath.Clean(normalizedTargetPath)

	// Convert both paths to absolute paths for comparison
	absRepositoriesDir, err := filepath.Abs(cleanRepositoriesDir)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for base path: %w", err)
	}

	absTargetPath, err := filepath.Abs(cleanTargetPath)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for target path: %w", err)
	}

	// Check if target path is within base path by comparing path components
	relPath, err := filepath.Rel(absRepositoriesDir, absTargetPath)
	if err != nil {
		return false, err // Return the error if we can't get relative path
	}

	// If relative path starts with "..", target is outside base path
	return !strings.HasPrefix(relPath, "..") && relPath != "..", nil
}
