package fs

import "os"

// ReadDir reads the contents of a directory.
func (f *realFS) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}
