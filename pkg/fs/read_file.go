package fs

import "os"

// ReadFile reads the contents of a file.
func (f *realFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
