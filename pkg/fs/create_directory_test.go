//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_CreateDirectory(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-create-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test creating a single directory
	testDir := filepath.Join(tmpDir, "test-dir")
	err = fs.CreateDirectory(testDir, 0755)
	assert.NoError(t, err)

	// Verify directory was created
	exists, err := fs.Exists(testDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	isDir, err := fs.IsDir(testDir)
	assert.NoError(t, err)
	assert.True(t, isDir)

	// Test creating nested directories
	nestedDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	err = fs.CreateDirectory(nestedDir, 0755)
	assert.NoError(t, err)

	// Verify nested directories were created
	exists, err = fs.Exists(nestedDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	isDir, err = fs.IsDir(nestedDir)
	assert.NoError(t, err)
	assert.True(t, isDir)

	// Test creating existing directory (should not error)
	err = fs.CreateDirectory(testDir, 0755)
	assert.NoError(t, err)

	// Test creating directory with different permissions
	permDir := filepath.Join(tmpDir, "perm-dir")
	err = fs.CreateDirectory(permDir, 0700)
	assert.NoError(t, err)

	// Verify directory permissions
	info, err := os.Stat(permDir)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}
