package fs

import "os"

// CreateDirectory creates a directory with permissions.
func (f *realFS) CreateDirectory(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
