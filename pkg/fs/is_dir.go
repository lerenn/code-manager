package fs

import "os"

// IsDir checks if the path is a directory.
func (f *realFS) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
