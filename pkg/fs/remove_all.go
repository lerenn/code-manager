package fs

import "os"

// RemoveAll removes a file or directory and all its contents.
func (f *realFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
