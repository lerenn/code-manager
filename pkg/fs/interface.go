package fs

import (
	"os"
)

//go:generate mockgen -source=interface.go -destination=mocks/fs.gen.go -package=mocks

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

	// Remove removes a file or empty directory.
	Remove(path string) error

	// RemoveAll removes a file or directory and all its contents.
	RemoveAll(path string) error

	// Which finds the executable path for a command using the system's PATH.
	Which(command string) (string, error)

	// ExecuteCommand executes a command with arguments in the background.
	ExecuteCommand(command string, args ...string) error

	// CreateDirectory creates a directory with permissions.
	CreateDirectory(path string, perm os.FileMode) error

	// CreateFileWithContent creates a file with content.
	CreateFileWithContent(path string, content []byte, perm os.FileMode) error

	// IsDirectoryWritable checks if a directory is writable.
	IsDirectoryWritable(path string) (bool, error)

	// ExpandPath expands ~ to user's home directory.
	ExpandPath(path string) (string, error)

	// IsPathWithinBase checks if a target path is within the base path.
	IsPathWithinBase(repositoriesDir, targetPath string) (bool, error)

	// ResolvePath resolves relative paths from base directory.
	ResolvePath(repositoriesDir, relativePath string) (string, error)

	// ValidateRepositoryPath validates that path contains a Git repository.
	ValidateRepositoryPath(path string) (bool, error)
}

type realFS struct {
	// No fields needed for basic file system operations
}

// NewFS creates a new FS instance.
func NewFS() FS {
	return &realFS{}
}
