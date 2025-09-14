//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
