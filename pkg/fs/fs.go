package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:generate mockgen -source=fs.go -destination=mocks/fs.gen.go -package=mocks

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

	// Ensure parent directory exists before creating temporary file
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if err != nil {
			if removeErr := os.Remove(tmpPath); removeErr != nil {
				// Log the error but don't fail for cleanup errors
				_ = removeErr
			}
		}
	}()

	// Write data to temporary file
	if _, err := tmpFile.Write(data); err != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			// Log the error but don't fail for cleanup errors
			_ = closeErr
		}
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

// CreateDirectory creates a directory with permissions.
func (f *realFS) CreateDirectory(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// CreateFileWithContent creates a file with content.
func (f *realFS) CreateFileWithContent(path string, content []byte, perm os.FileMode) error {
	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := f.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file atomically
	return f.WriteFileAtomic(path, content, perm)
}

// ExpandPath expands ~ to user's home directory.
func (f *realFS) ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := f.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}

	return filepath.Join(homeDir, strings.TrimPrefix(path, "~")), nil
}

// RemoveAll removes a file or directory and all its contents.
func (f *realFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Which finds the executable path for a command using the system's PATH.
func (f *realFS) Which(command string) (string, error) {
	// Use exec.LookPath to find the executable in PATH
	path, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}
	return path, nil
}

// ExecuteCommand executes a command with arguments in the background.
func (f *realFS) ExecuteCommand(command string, args ...string) error {
	// Create command
	cmd := exec.Command(command, args...)

	// Start command in background (don't wait for completion)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Don't wait for the command to finish, let it run in background
	return nil
}

// IsPathWithinBase checks if a target path is within the base path.
func (f *realFS) IsPathWithinBase(repositoriesDir, targetPath string) (bool, error) {
	// Handle empty paths
	if repositoriesDir == "" && targetPath == "" {
		return true, nil
	}
	if repositoriesDir == "" {
		return false, nil
	}

	// Normalize path separators - convert backslashes to forward slashes for cross-platform compatibility
	normalizedRepositoriesDir := strings.ReplaceAll(repositoriesDir, "\\", "/")
	normalizedTargetPath := strings.ReplaceAll(targetPath, "\\", "/")

	// Clean the paths
	cleanRepositoriesDir := filepath.Clean(normalizedRepositoriesDir)
	cleanTargetPath := filepath.Clean(normalizedTargetPath)

	// Convert both paths to absolute paths for comparison
	absRepositoriesDir, err := filepath.Abs(cleanRepositoriesDir)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for base path: %w", err)
	}

	absTargetPath, err := filepath.Abs(cleanTargetPath)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for target path: %w", err)
	}

	// Check if target path is within base path by comparing path components
	relPath, err := filepath.Rel(absRepositoriesDir, absTargetPath)
	if err != nil {
		return false, err // Return the error if we can't get relative path
	}

	// If relative path starts with "..", target is outside base path
	return !strings.HasPrefix(relPath, "..") && relPath != "..", nil
}

// ResolvePath resolves relative paths from base directory.
func (f *realFS) ResolvePath(repositoriesDir, relativePath string) (string, error) {
	// Handle empty paths
	if repositoriesDir == "" {
		return "", fmt.Errorf("%w: base path cannot be empty", ErrPathResolution)
	}
	if relativePath == "" {
		return "", fmt.Errorf("%w: relative path cannot be empty", ErrPathResolution)
	}

	// If relativePath is already absolute, return it as-is
	if filepath.IsAbs(relativePath) {
		return filepath.Clean(relativePath), nil
	}

	// Resolve relative path from base directory
	resolvedPath := filepath.Join(repositoriesDir, relativePath)

	// Clean the resolved path to remove any ".." or "." components
	cleanPath := filepath.Clean(resolvedPath)

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("%w: failed to get absolute path for %s: %w", ErrPathResolution, cleanPath, err)
	}

	return absPath, nil
}

// ValidateRepositoryPath validates that path contains a Git repository.
func (f *realFS) ValidateRepositoryPath(path string) (bool, error) {
	// Handle empty path
	if path == "" {
		return false, fmt.Errorf("%w: path cannot be empty", ErrInvalidRepository)
	}

	// Check if path exists
	exists, err := f.Exists(path)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if path exists: %w", ErrInvalidRepository, err)
	}
	if !exists {
		return false, fmt.Errorf("%w: path does not exist: %s", ErrInvalidRepository, path)
	}

	// Check if path is a directory
	isDir, err := f.IsDir(path)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if path is directory: %w", ErrInvalidRepository, err)
	}
	if !isDir {
		return false, fmt.Errorf("%w: path is not a directory: %s", ErrInvalidRepository, path)
	}

	// Check if .git directory exists (indicating a Git repository)
	gitPath := filepath.Join(path, ".git")
	gitExists, err := f.Exists(gitPath)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if .git directory exists: %w", ErrInvalidRepository, err)
	}
	if !gitExists {
		return false, fmt.Errorf("%w: path does not contain a Git repository (.git directory not found): %s",
			ErrInvalidRepository, path)
	}

	// Check if .git is a directory (not a file for submodules)
	gitIsDir, err := f.IsDir(gitPath)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check if .git is a directory: %w", ErrInvalidRepository, err)
	}
	if !gitIsDir {
		// .git exists but is not a directory - this could be a submodule
		// For now, we'll consider this valid as it's still a Git repository
		return true, nil
	}

	return true, nil
}
