//go:build windows

package fs

import (
	"os"
	"path/filepath"
)

// FileLock acquires a file lock and returns an unlock function.
// This is a no-op implementation for Windows since flock is not available.
// On Windows, file locking is typically handled differently and may not be needed
// for the use cases in this application.
func (f *realFS) FileLock(filename string) (func(), error) {
	// Create lock file path
	lockPath := filename + ".lock"

	// Ensure parent directory exists before creating lock file
	lockDir := filepath.Dir(lockPath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, err
	}

	// Create lock file (this serves as a simple indicator that the file is "locked")
	lockFile, err := os.Create(lockPath)
	if err != nil {
		return nil, err
	}

	// Return unlock function that removes the lock file
	unlock := func() {
		if closeErr := lockFile.Close(); closeErr != nil {
			// Log the error but don't fail for cleanup errors
			_ = closeErr
		}
		if removeErr := os.Remove(lockPath); removeErr != nil {
			// Log the error but don't fail for cleanup errors
			_ = removeErr
		}
	}

	return unlock, nil
}
