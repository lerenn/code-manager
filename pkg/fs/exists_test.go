//go:build integration

package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
