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

	// Test writing to non-existent directory (should fail)
	nonExistentFile := filepath.Join("/non/existent/path", "test.txt")
	err = fs.WriteFileAtomic(nonExistentFile, testContent, 0644)
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

	// Test acquiring lock on non-existent directory (should fail)
	nonExistentFile := filepath.Join("/non/existent/path", "test.txt")
	_, err = fs.FileLock(nonExistentFile)
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
