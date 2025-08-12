package fs

import (
	"os"
	"path/filepath"
	"syscall"
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

	// WriteFileAtomic writes data to a file atomically using a temporary file and rename.
	WriteFileAtomic(filename string, data []byte, perm os.FileMode) error

	// FileLock acquires a file lock and returns an unlock function.
	FileLock(filename string) (func(), error)

	// CreateFileIfNotExists creates a file with initial content if it doesn't exist.
	CreateFileIfNotExists(filename string, initialContent []byte, perm os.FileMode) error
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

// WriteFileAtomic writes data to a file atomically using a temporary file and rename.
func (f *realFS) WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	// Create temporary file in the same directory
	dir := filepath.Dir(filename)
	tmpFile, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// Write data to temporary file
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}

	// Close temporary file
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Set permissions on temporary file
	if err := os.Chmod(tmpPath, perm); err != nil {
		return err
	}

	// Atomically rename temporary file to target file
	if err := os.Rename(tmpPath, filename); err != nil {
		return err
	}

	return nil
}

// FileLock acquires a file lock and returns an unlock function.
func (f *realFS) FileLock(filename string) (func(), error) {
	// Create lock file path
	lockPath := filename + ".lock"

	// Create lock file
	lockFile, err := os.Create(lockPath)
	if err != nil {
		return nil, err
	}

	// Acquire file lock
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		lockFile.Close()
		os.Remove(lockPath)
		return nil, err
	}

	// Return unlock function
	unlock := func() {
		_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		os.Remove(lockPath)
	}

	return unlock, nil
}

// CreateFileIfNotExists creates a file with initial content if it doesn't exist.
func (f *realFS) CreateFileIfNotExists(filename string, initialContent []byte, perm os.FileMode) error {
	// Check if file already exists
	exists, err := f.Exists(filename)
	if err != nil {
		return err
	}

	if exists {
		return nil // File already exists, nothing to do
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(filename)
	if err := f.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file with initial content
	return f.WriteFileAtomic(filename, initialContent, perm)
}
