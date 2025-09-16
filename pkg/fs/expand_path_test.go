//go:build integration

package fs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_ExpandPath(t *testing.T) {
	fs := NewFS()

	// Get home directory for testing
	homeDir, err := fs.GetHomeDir()
	require.NoError(t, err)

	// Test expanding ~ to home directory
	expanded, err := fs.ExpandPath("~")
	assert.NoError(t, err)
	assert.Equal(t, homeDir, expanded)

	// Test expanding ~/path to home directory + path
	expanded, err = fs.ExpandPath("~/test/path")
	assert.NoError(t, err)
	expected := filepath.Join(homeDir, "test", "path")
	assert.Equal(t, expected, expanded)

	// Test expanding ~/ with trailing slash
	expanded, err = fs.ExpandPath("~/")
	assert.NoError(t, err)
	assert.Equal(t, homeDir, expanded)

	// Test path without ~ (should return as-is)
	regularPath := "/some/regular/path"
	expanded, err = fs.ExpandPath(regularPath)
	assert.NoError(t, err)
	assert.Equal(t, regularPath, expanded)

	// Test relative path without ~ (should return as-is)
	relativePath := "relative/path"
	expanded, err = fs.ExpandPath(relativePath)
	assert.NoError(t, err)
	assert.Equal(t, relativePath, expanded)

	// Test empty path
	expanded, err = fs.ExpandPath("")
	assert.NoError(t, err)
	assert.Equal(t, "", expanded)

	// Test path with multiple ~ (should only expand the first one)
	multiTildePath := "~/test/~/path"
	expanded, err = fs.ExpandPath(multiTildePath)
	assert.NoError(t, err)
	expected = filepath.Join(homeDir, "test", "~/path")
	assert.Equal(t, expected, expanded)
}
