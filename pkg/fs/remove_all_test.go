//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
