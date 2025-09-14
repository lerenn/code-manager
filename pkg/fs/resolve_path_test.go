//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	repositoriesDir := tmpDir

	// Test simple relative path
	resolved, err := fs.ResolvePath(repositoriesDir, "level1")
	assert.NoError(t, err)
	expected := filepath.Join(tmpDir, "level1")
	assert.Equal(t, expected, resolved)

	// Test nested relative path
	resolved, err = fs.ResolvePath(repositoriesDir, "level1/level2")
	assert.NoError(t, err)
	expected = filepath.Join(tmpDir, "level1", "level2")
	assert.Equal(t, expected, resolved)

	// Test relative path with ".." components
	resolved, err = fs.ResolvePath(nestedDir, "../level1")
	assert.NoError(t, err)
	expected = filepath.Join(tmpDir, "level1", "level1")
	assert.Equal(t, expected, resolved)

	// Test relative path with "." components
	resolved, err = fs.ResolvePath(repositoriesDir, "./level1")
	assert.NoError(t, err)
	expected = filepath.Join(tmpDir, "level1")
	assert.Equal(t, expected, resolved)

	// Test absolute path (should return as-is)
	absPath := filepath.Join(tmpDir, "absolute")
	resolved, err = fs.ResolvePath(repositoriesDir, absPath)
	assert.NoError(t, err)
	assert.Equal(t, absPath, resolved)

	// Test empty base path
	_, err = fs.ResolvePath("", "relative")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPathResolution)

	// Test empty relative path
	_, err = fs.ResolvePath(repositoriesDir, "")
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
