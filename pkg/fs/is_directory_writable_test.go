//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_IsDirectoryWritable(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-writable-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test writable directory
	writable, err := fs.IsDirectoryWritable(tmpDir)
	assert.NoError(t, err)
	assert.True(t, writable)

	// Test non-existent directory
	nonExistentDir := filepath.Join(tmpDir, "non-existent")
	writable, err = fs.IsDirectoryWritable(nonExistentDir)
	assert.Error(t, err)
	assert.False(t, writable)

	// Test file instead of directory
	testFile := filepath.Join(tmpDir, "test-file.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	writable, err = fs.IsDirectoryWritable(testFile)
	assert.Error(t, err)
	assert.False(t, writable)

	// Test nested directory
	nestedDir := filepath.Join(tmpDir, "level1", "level2")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	writable, err = fs.IsDirectoryWritable(nestedDir)
	assert.NoError(t, err)
	assert.True(t, writable)
}
