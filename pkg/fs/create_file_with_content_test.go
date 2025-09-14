//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_CreateFileWithContent(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-create-file-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test creating a file with content
	testFile := filepath.Join(tmpDir, "test-file.txt")
	testContent := []byte("test content for file")
	err = fs.CreateFileWithContent(testFile, testContent, 0644)
	assert.NoError(t, err)

	// Verify file was created with correct content
	readContent, err := fs.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testContent, readContent)

	// Verify file permissions
	info, err := os.Stat(testFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())

	// Test creating file in nested directory
	nestedFile := filepath.Join(tmpDir, "level1", "level2", "nested-file.txt")
	nestedContent := []byte("nested file content")
	err = fs.CreateFileWithContent(nestedFile, nestedContent, 0755)
	assert.NoError(t, err)

	// Verify nested file was created
	readContent, err = fs.ReadFile(nestedFile)
	assert.NoError(t, err)
	assert.Equal(t, nestedContent, readContent)

	// Verify nested file permissions
	info, err = os.Stat(nestedFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())

	// Test overwriting existing file
	newContent := []byte("new content")
	err = fs.CreateFileWithContent(testFile, newContent, 0644)
	assert.NoError(t, err)

	// Verify file was updated
	readContent, err = fs.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, newContent, readContent)
}
