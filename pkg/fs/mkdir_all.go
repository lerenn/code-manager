package fs

import "os"

// MkdirAll creates a directory and all parent directories.
func (f *realFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
