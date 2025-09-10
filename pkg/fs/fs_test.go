//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_ReadFile(t *testing.T) {
	fs := NewFS()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write content to the file
	content := []byte("test content")
	err = os.WriteFile(tmpFile.Name(), content, 0644)
	require.NoError(t, err)

	// Test reading existing file
	readContent, err := fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)

	// Test reading non-existing file
	_, err = fs.ReadFile("non-existing-file.txt")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestFS_ReadDir(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create some files in the directory
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err = os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Test reading directory contents
	entries, err := fs.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	// Verify file names
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	assert.Contains(t, names, "file1.txt")
	assert.Contains(t, names, "file2.txt")

	// Test reading non-existing directory
	_, err = fs.ReadDir("non-existing-dir")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestFS_Glob(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-glob-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create some .code-workspace files
	workspace1 := filepath.Join(tmpDir, "project.code-workspace")
	workspace2 := filepath.Join(tmpDir, "dev.code-workspace")
	otherFile := filepath.Join(tmpDir, "other.txt")

	err = os.WriteFile(workspace1, []byte("{}"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(workspace2, []byte("{}"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(otherFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Test glob pattern matching
	pattern := filepath.Join(tmpDir, "*.code-workspace")
	matches, err := fs.Glob(pattern)
	assert.NoError(t, err)
	assert.Len(t, matches, 2)
	assert.Contains(t, matches, workspace1)
	assert.Contains(t, matches, workspace2)

	// Test glob with no matches
	noMatchPattern := filepath.Join(tmpDir, "*.nonexistent")
	noMatches, err := fs.Glob(noMatchPattern)
	assert.NoError(t, err)
	assert.Len(t, noMatches, 0)

	// Test glob with invalid pattern
	_, err = fs.Glob("[invalid")
	assert.Error(t, err)
}

func TestFS_MkdirAll(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-mkdir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test creating nested directories
	nestedPath := filepath.Join(tmpDir, "level1", "level2", "level3")
	err = fs.MkdirAll(nestedPath, 0755)
	assert.NoError(t, err)

	// Verify directories were created
	exists, err := fs.Exists(nestedPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	isDir, err := fs.IsDir(nestedPath)
	assert.NoError(t, err)
	assert.True(t, isDir)

	// Test creating existing directory (should not error)
	err = fs.MkdirAll(nestedPath, 0755)
	assert.NoError(t, err)
}

func TestFS_GetHomeDir(t *testing.T) {
	fs := NewFS()

	homeDir, err := fs.GetHomeDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, homeDir)

	// Verify it's a directory
	exists, err := fs.Exists(homeDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	isDir, err := fs.IsDir(homeDir)
	assert.NoError(t, err)
	assert.True(t, isDir)
}

func TestFS_IsNotExist(t *testing.T) {
	fs := NewFS()

	// Test with non-existent file error
	_, err := fs.ReadFile("non-existent-file.txt")
	assert.Error(t, err)
	assert.True(t, fs.IsNotExist(err))

	// Test with existing file (should not be IsNotExist)
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.False(t, fs.IsNotExist(err))
}

func TestFS_WriteFileAtomic(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-atomic-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test writing file atomically
	testFile := filepath.Join(tmpDir, "atomic-test.txt")
	testContent := []byte("atomic test content")

	err = fs.WriteFileAtomic(testFile, testContent, 0644)
	assert.NoError(t, err)

	// Verify file was created with correct content
	readContent, err := fs.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testContent, readContent)

	// Verify file permissions
	info, err := os.Stat(testFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())

	// Test overwriting existing file
	newContent := []byte("new atomic content")
	err = fs.WriteFileAtomic(testFile, newContent, 0644)
	assert.NoError(t, err)

	// Verify file was updated
	readContent, err = fs.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, newContent, readContent)

	// Test writing to a device file which should fail
	deviceFile := "/dev/null/test.txt"
	err = fs.WriteFileAtomic(deviceFile, testContent, 0644)
	assert.Error(t, err)
}

func TestFS_FileLock(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-lock-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test acquiring a file lock
	testFile := filepath.Join(tmpDir, "test-file.txt")
	unlock, err := fs.FileLock(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, unlock)

	// Verify lock file was created
	lockFile := testFile + ".lock"
	exists, err := fs.Exists(lockFile)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test that we can't acquire the same lock again
	_, err = fs.FileLock(testFile)
	assert.Error(t, err)

	// Release the lock
	unlock()

	// Verify lock file was removed
	exists, err = fs.Exists(lockFile)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test that we can acquire the lock again after release
	unlock2, err := fs.FileLock(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, unlock2)
	unlock2()

	// Test acquiring lock on a device file which should fail
	deviceFile := "/dev/null/test.txt"
	_, err = fs.FileLock(deviceFile)
	assert.Error(t, err)
}

func TestFS_CreateFileIfNotExists(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-create-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test creating a new file
	testFile := filepath.Join(tmpDir, "new-file.txt")
	initialContent := []byte("initial content")

	err = fs.CreateFileIfNotExists(testFile, initialContent, 0644)
	assert.NoError(t, err)

	// Verify file was created with correct content
	readContent, err := fs.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, initialContent, readContent)

	// Test creating the same file again (should not error and not overwrite)
	newContent := []byte("new content")
	err = fs.CreateFileIfNotExists(testFile, newContent, 0644)
	assert.NoError(t, err)

	// Verify content was not changed
	readContent, err = fs.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, initialContent, readContent)

	// Test creating file in nested directory
	nestedFile := filepath.Join(tmpDir, "level1", "level2", "nested-file.txt")
	err = fs.CreateFileIfNotExists(nestedFile, initialContent, 0644)
	assert.NoError(t, err)

	// Verify nested file was created
	readContent, err = fs.ReadFile(nestedFile)
	assert.NoError(t, err)
	assert.Equal(t, initialContent, readContent)

	// Test creating file with different permissions
	permFile := filepath.Join(tmpDir, "perm-file.txt")
	err = fs.CreateFileIfNotExists(permFile, initialContent, 0755)
	assert.NoError(t, err)

	// Verify file permissions
	info, err := os.Stat(permFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestFS_RemoveAll(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-remove-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a nested directory structure with files
	nestedDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	err = fs.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	// Create some files in the nested structure
	file1 := filepath.Join(nestedDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "level1", "file2.txt")

	err = os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Test removing a file
	err = fs.RemoveAll(file1)
	assert.NoError(t, err)

	// Verify file was removed
	exists, err := fs.Exists(file1)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test removing a directory and all its contents
	level1Dir := filepath.Join(tmpDir, "level1")
	err = fs.RemoveAll(level1Dir)
	assert.NoError(t, err)

	// Verify directory and all its contents were removed
	exists, err = fs.Exists(level1Dir)
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = fs.Exists(file2)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test removing non-existent path (should not error)
	err = fs.RemoveAll(filepath.Join(tmpDir, "non-existent"))
	assert.NoError(t, err)

	// Test removing the entire temp directory
	err = fs.RemoveAll(tmpDir)
	assert.NoError(t, err)

	// Verify temp directory was removed
	exists, err = fs.Exists(tmpDir)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestFS_Exists(t *testing.T) {
	fs := NewFS()

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-exists-*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Test existing file
	exists, err := fs.Exists(tmpFile.Name())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing file
	exists, err = fs.Exists("non-existing-file.txt")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test existing directory
	tmpDir, err := os.MkdirTemp("", "test-exists-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exists, err = fs.Exists(tmpDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing directory
	exists, err = fs.Exists("non-existing-directory")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestFS_IsDir(t *testing.T) {
	fs := NewFS()

	// Create a temporary file and directory for testing
	tmpFile, err := os.CreateTemp("", "test-isdir-*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	tmpDir, err := os.MkdirTemp("", "test-isdir-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test file (should not be a directory)
	isDir, err := fs.IsDir(tmpFile.Name())
	assert.NoError(t, err)
	assert.False(t, isDir)

	// Test directory (should be a directory)
	isDir, err = fs.IsDir(tmpDir)
	assert.NoError(t, err)
	assert.True(t, isDir)

	// Test non-existing path
	_, err = fs.IsDir("non-existing-path")
	assert.Error(t, err)
}

func TestFS_Which(t *testing.T) {
	fs := NewFS()

	// Test finding an existing command (ls should be available on Unix-like systems)
	path, err := fs.Which("ls")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "ls")

	// Test finding a non-existing command
	_, err = fs.Which("non-existing-command-xyz123")
	assert.Error(t, err)

	// Test finding echo command
	path, err = fs.Which("echo")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "echo")
}

func TestFS_ExecuteCommand(t *testing.T) {
	fs := NewFS()

	// Test executing a simple command that should succeed
	err := fs.ExecuteCommand("echo", "hello world")
	assert.NoError(t, err)

	// Test executing a command with multiple arguments
	err = fs.ExecuteCommand("echo", "hello", "world", "test")
	assert.NoError(t, err)

	// Test executing a non-existing command (should fail)
	err = fs.ExecuteCommand("non-existing-command-xyz123")
	assert.Error(t, err)

	// Test executing ls command
	err = fs.ExecuteCommand("ls", "-la")
	assert.NoError(t, err)
}

func TestFS_IsPathWithinBase(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "test-path-within-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directories
	baseDir := filepath.Join(tmpDir, "base")
	nestedDir1 := filepath.Join(baseDir, "level1")
	nestedDir2 := filepath.Join(nestedDir1, "level2")
	outsideDir := filepath.Join(tmpDir, "outside")

	err = os.MkdirAll(nestedDir2, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(outsideDir, 0755)
	require.NoError(t, err)

	// Test positive cases - paths within base
	within, err := fs.IsPathWithinBase(baseDir, nestedDir1)
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase(baseDir, nestedDir2)
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase(baseDir, baseDir)
	assert.NoError(t, err)
	assert.True(t, within)

	// Test negative cases - paths outside base
	within, err = fs.IsPathWithinBase(baseDir, outsideDir)
	assert.NoError(t, err)
	assert.False(t, within)

	within, err = fs.IsPathWithinBase(baseDir, tmpDir)
	assert.NoError(t, err)
	assert.False(t, within)

	// Test with relative paths
	within, err = fs.IsPathWithinBase("base", "base/level1")
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase("base", "outside")
	assert.NoError(t, err)
	assert.False(t, within)

	// Test edge cases
	within, err = fs.IsPathWithinBase("", "")
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase("base", "")
	assert.NoError(t, err)
	assert.False(t, within)

	// Test with different path separators (should work cross-platform)
	within, err = fs.IsPathWithinBase("base", "base\\level1")
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase("base", "base/level1")
	assert.NoError(t, err)
	assert.True(t, within)
}

func TestFS_ResolvePath(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-resolve-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "level1", "level2")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	// Test resolving relative paths from base directory
	basePath := tmpDir

	// Test simple relative path
	resolved, err := fs.ResolvePath(basePath, "level1")
	assert.NoError(t, err)
	expected := filepath.Join(tmpDir, "level1")
	assert.Equal(t, expected, resolved)

	// Test nested relative path
	resolved, err = fs.ResolvePath(basePath, "level1/level2")
	assert.NoError(t, err)
	expected = filepath.Join(tmpDir, "level1", "level2")
	assert.Equal(t, expected, resolved)

	// Test relative path with ".." components
	resolved, err = fs.ResolvePath(nestedDir, "../level1")
	assert.NoError(t, err)
	expected = filepath.Join(tmpDir, "level1", "level1")
	assert.Equal(t, expected, resolved)

	// Test relative path with "." components
	resolved, err = fs.ResolvePath(basePath, "./level1")
	assert.NoError(t, err)
	expected = filepath.Join(tmpDir, "level1")
	assert.Equal(t, expected, resolved)

	// Test absolute path (should return as-is)
	absPath := filepath.Join(tmpDir, "absolute")
	resolved, err = fs.ResolvePath(basePath, absPath)
	assert.NoError(t, err)
	assert.Equal(t, absPath, resolved)

	// Test empty base path
	_, err = fs.ResolvePath("", "relative")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPathResolution)

	// Test empty relative path
	_, err = fs.ResolvePath(basePath, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPathResolution)

	// Test complex relative path with multiple ".." and "."
	resolved, err = fs.ResolvePath(nestedDir, "../../level1/./level2")
	assert.NoError(t, err)
	expected = filepath.Join(tmpDir, "level1", "level2")
	assert.Equal(t, expected, resolved)

	// Test path that goes outside base directory
	resolved, err = fs.ResolvePath(nestedDir, "../../../outside")
	assert.NoError(t, err)
	// Should still resolve but go outside the base directory
	expected = filepath.Join(filepath.Dir(tmpDir), "outside")
	assert.Equal(t, expected, resolved)
}

func TestFS_ValidateRepositoryPath(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-validate-repo-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test empty path
	valid, err := fs.ValidateRepositoryPath("")
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test non-existent path
	nonExistentPath := filepath.Join(tmpDir, "non-existent")
	valid, err = fs.ValidateRepositoryPath(nonExistentPath)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test path that exists but is a file (not directory)
	testFile := filepath.Join(tmpDir, "test-file.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(testFile)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test directory without .git
	regularDir := filepath.Join(tmpDir, "regular-dir")
	err = os.MkdirAll(regularDir, 0755)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(regularDir)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test directory with .git as a file (submodule case)
	submoduleDir := filepath.Join(tmpDir, "submodule-dir")
	err = os.MkdirAll(submoduleDir, 0755)
	require.NoError(t, err)

	// Create .git as a file (submodule)
	gitFile := filepath.Join(submoduleDir, ".git")
	err = os.WriteFile(gitFile, []byte("gitdir: ../.git/modules/submodule"), 0644)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(submoduleDir)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test directory with .git as a directory (regular repository)
	repoDir := filepath.Join(tmpDir, "repo-dir")
	err = os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	// Create .git directory
	gitDir := filepath.Join(repoDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create some basic Git repository structure
	objectsDir := filepath.Join(gitDir, "objects")
	refsDir := filepath.Join(gitDir, "refs")
	err = os.MkdirAll(objectsDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(refsDir, 0755)
	require.NoError(t, err)

	// Create HEAD file
	headFile := filepath.Join(gitDir, "HEAD")
	err = os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0644)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(repoDir)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test nested repository
	nestedRepoDir := filepath.Join(tmpDir, "nested", "repo")
	err = os.MkdirAll(nestedRepoDir, 0755)
	require.NoError(t, err)

	// Create .git directory in nested location
	nestedGitDir := filepath.Join(nestedRepoDir, ".git")
	err = os.MkdirAll(nestedGitDir, 0755)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(nestedRepoDir)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test repository with absolute path
	absRepoDir, err := filepath.Abs(repoDir)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(absRepoDir)
	assert.NoError(t, err)
	assert.True(t, valid)
}
