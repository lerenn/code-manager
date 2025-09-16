package fs

import (
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to a file atomically using a temporary file and rename.
func (f *realFS) WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	// Create temporary file in the same directory
	dir := filepath.Dir(filename)

	// Ensure parent directory exists before creating temporary file
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if err != nil {
			if removeErr := os.Remove(tmpPath); removeErr != nil {
				// Log the error but don't fail for cleanup errors
				_ = removeErr
			}
		}
	}()

	// Write data to temporary file
	if _, err := tmpFile.Write(data); err != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			// Log the error but don't fail for cleanup errors
			_ = closeErr
		}
		return err
	}

	// Close temporary file
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Set permissions on temporary file
	if err := os.Chmod(tmpPath, perm); err != nil {
		return err
	}

	// Atomically rename temporary file to target file
	if err := os.Rename(tmpPath, filename); err != nil {
		return err
	}

	return nil
}
