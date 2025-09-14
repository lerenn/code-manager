//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFileIfNotExists_NewFile(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "new_file.txt")
	initialContent := []byte("Initial content")

	// Create file that doesn't exist
	err := fs.CreateFileIfNotExists(testFile, initialContent, 0644)
	require.NoError(t, err)

	// Verify file exists and has correct content
	exists, err := fs.Exists(testFile)
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := fs.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, initialContent, content)
}

func TestCreateFileIfNotExists_ExistingFile(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "existing_file.txt")
	initialContent := []byte("Initial content")
	newContent := []byte("New content")

	// Create initial file
	err := fs.CreateFileIfNotExists(testFile, initialContent, 0644)
	require.NoError(t, err)

	// Try to create file again (should not overwrite)
	err = fs.CreateFileIfNotExists(testFile, newContent, 0600)
	require.NoError(t, err)

	// Verify original content is preserved
	content, err := fs.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, initialContent, content)

	// Verify original permissions are preserved
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestCreateFileIfNotExists_CreateDirectories(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "deep", "directory")
	testFile := filepath.Join(nestedDir, "test.txt")
	initialContent := []byte("Initial content")

	// Create file in non-existent directory structure
	err := fs.CreateFileIfNotExists(testFile, initialContent, 0644)
	require.NoError(t, err)

	// Verify file exists
	exists, err := fs.Exists(testFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify directories were created
	exists, err = fs.Exists(nestedDir)
	require.NoError(t, err)
	assert.True(t, exists)

	isDir, err := fs.IsDir(nestedDir)
	require.NoError(t, err)
	assert.True(t, isDir)
}
