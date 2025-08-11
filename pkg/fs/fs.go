package fs

import (
	"os"
	"path/filepath"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=fs.go -destination=mockfs.gen.go -package=fs

// FS interface provides file system operations for Git repository detection.
type FS interface {
	// Exists checks if a file or directory exists at the given path.
	Exists(path string) (bool, error)

	// IsDir checks if the path is a directory.
	IsDir(path string) (bool, error)

	// ReadFile reads the contents of a file.
	ReadFile(path string) ([]byte, error)

	// ReadDir reads the contents of a directory.
	ReadDir(path string) ([]os.DirEntry, error)

	// Glob finds files matching the pattern.
	Glob(pattern string) ([]string, error)
}

type fs struct {
	// No fields needed for basic file system operations
}

// NewFS creates a new FS instance.
func NewFS() FS {
	return &fs{}
}

// Exists checks if a file or directory exists at the given path.
func (f *fs) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// IsDir checks if the path is a directory.
func (f *fs) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ReadFile reads the contents of a file.
func (f *fs) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadDir reads the contents of a directory.
func (f *fs) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

// Glob finds files matching the pattern.
func (f *fs) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}
