package fs

import (
	"fmt"
	"path/filepath"
)

// ValidateRepositoryPath validates that path contains a Git repository.
func (f *realFS) ValidateRepositoryPath(path string) (bool, error) {
	// Handle empty path
	if path == "" {
		return false, fmt.Errorf("%w: path cannot be empty", ErrInvalidRepository)
	}

	// Check if path exists
	exists, err := f.Exists(path)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if path exists: %w", ErrInvalidRepository, err)
	}
	if !exists {
		return false, fmt.Errorf("%w: path does not exist: %s", ErrInvalidRepository, path)
	}

	// Check if path is a directory
	isDir, err := f.IsDir(path)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if path is directory: %w", ErrInvalidRepository, err)
	}
	if !isDir {
		return false, fmt.Errorf("%w: path is not a directory: %s", ErrInvalidRepository, path)
	}

	// Check if .git directory exists (indicating a Git repository)
	gitPath := filepath.Join(path, ".git")
	gitExists, err := f.Exists(gitPath)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if .git directory exists: %w", ErrInvalidRepository, err)
	}
	if !gitExists {
		return false, fmt.Errorf("%w: path does not contain a Git repository (.git directory not found): %s",
			ErrInvalidRepository, path)
	}

	// Check if .git is a directory (not a file for submodules)
	gitIsDir, err := f.IsDir(gitPath)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if .git is a directory: %w", ErrInvalidRepository, err)
	}
	if !gitIsDir {
		// .git exists but is not a directory - this could be a submodule
		// For now, we'll consider this valid as it's still a Git repository
		return true, nil
	}

	return true, nil
}
