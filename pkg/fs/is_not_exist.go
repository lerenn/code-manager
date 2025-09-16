package fs

import "os"

// IsNotExist checks if an error indicates that a file or directory doesn't exist.
func (f *realFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
