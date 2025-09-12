package fs

import "os"

// GetHomeDir returns the user's home directory path.
func (f *realFS) GetHomeDir() (string, error) {
	return os.UserHomeDir()
}
