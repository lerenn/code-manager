package fs

import (
	"os"
	"path/filepath"
)

// CreateFileIfNotExists creates a file with initial content if it doesn't exist.
func (f *realFS) CreateFileIfNotExists(filename string, initialContent []byte, perm os.FileMode) error {
	// Check if file already exists
	exists, err := f.Exists(filename)
	if err != nil {
		return err
	}

	if exists {
		return nil // File already exists, nothing to do
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(filename)
	if err := f.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file with initial content
	return f.WriteFileAtomic(filename, initialContent, perm)
}
