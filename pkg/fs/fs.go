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

	// MkdirAll creates a directory and all parent directories.
	MkdirAll(path string, perm os.FileMode) error

	// GetHomeDir returns the user's home directory path.
	GetHomeDir() (string, error)

	// IsNotExist checks if an error indicates that a file or directory doesn't exist.
	IsNotExist(err error) bool
}

type realFS struct {
	// No fields needed for basic file system operations
}

// NewFS creates a new FS instance.
func NewFS() FS {
	return &realFS{}
}

// Exists checks if a file or directory exists at the given path.
func (f *realFS) Exists(path string) (bool, error) {
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
func (f *realFS) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ReadFile reads the contents of a file.
func (f *realFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadDir reads the contents of a directory.
func (f *realFS) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

// Glob finds files matching the pattern.
func (f *realFS) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// MkdirAll creates a directory and all parent directories.
func (f *realFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// GetHomeDir returns the user's home directory path.
func (f *realFS) GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// IsNotExist checks if an error indicates that a file or directory doesn't exist.
func (f *realFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
