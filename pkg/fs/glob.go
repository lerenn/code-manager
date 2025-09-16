package fs

import "path/filepath"

// Glob finds files matching the pattern.
func (f *realFS) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}
