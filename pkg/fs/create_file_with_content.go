package fs

import (
	"os"
	"path/filepath"
)

// CreateFileWithContent creates a file with content.
func (f *realFS) CreateFileWithContent(path string, content []byte, perm os.FileMode) error {
	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := f.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file atomically
	return f.WriteFileAtomic(path, content, perm)
}
