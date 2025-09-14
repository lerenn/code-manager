//go:build !windows

package fs

import (
	"os"
	"path/filepath"
	"syscall"
)

// FileLock acquires a file lock and returns an unlock function.
// This implementation uses syscall.Flock which is available on Unix systems.
func (f *realFS) FileLock(filename string) (func(), error) {
	// Create lock file path
	lockPath := filename + ".lock"

	// Ensure parent directory exists before creating lock file
	lockDir := filepath.Dir(lockPath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, err
	}

	// Create lock file
	lockFile, err := os.Create(lockPath)
	if err != nil {
		return nil, err
	}

	// Acquire file lock (non-blocking)
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if closeErr := lockFile.Close(); closeErr != nil {
			// Log the error but don't fail for cleanup errors
			_ = closeErr
		}
		if removeErr := os.Remove(lockPath); removeErr != nil {
			// Log the error but don't fail for cleanup errors
			_ = removeErr
		}
		return nil, err
	}

	// Return unlock function
	unlock := func() {
		_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
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
