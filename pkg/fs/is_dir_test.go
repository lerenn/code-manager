//go:build integration

package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
