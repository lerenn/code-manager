package fs

import (
	"os"
)

// Remove removes a file or empty directory.
func (f *realFS) Remove(path string) error {
	return os.Remove(path)
}
