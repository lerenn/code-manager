//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_IsPathWithinBase(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "test-path-within-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directories
	baseDir := filepath.Join(tmpDir, "base")
	nestedDir1 := filepath.Join(baseDir, "level1")
	nestedDir2 := filepath.Join(nestedDir1, "level2")
	outsideDir := filepath.Join(tmpDir, "outside")

	err = os.MkdirAll(nestedDir2, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(outsideDir, 0755)
	require.NoError(t, err)

	// Test positive cases - paths within base
	within, err := fs.IsPathWithinBase(baseDir, nestedDir1)
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase(baseDir, nestedDir2)
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase(baseDir, baseDir)
	assert.NoError(t, err)
	assert.True(t, within)

	// Test negative cases - paths outside base
	within, err = fs.IsPathWithinBase(baseDir, outsideDir)
	assert.NoError(t, err)
	assert.False(t, within)

	within, err = fs.IsPathWithinBase(baseDir, tmpDir)
	assert.NoError(t, err)
	assert.False(t, within)

	// Test with relative paths
	within, err = fs.IsPathWithinBase("base", "base/level1")
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase("base", "outside")
	assert.NoError(t, err)
	assert.False(t, within)

	// Test edge cases
	within, err = fs.IsPathWithinBase("", "")
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase("base", "")
	assert.NoError(t, err)
	assert.False(t, within)

	// Test with different path separators (should work cross-platform)
	within, err = fs.IsPathWithinBase("base", "base\\level1")
	assert.NoError(t, err)
	assert.True(t, within)

	within, err = fs.IsPathWithinBase("base", "base/level1")
	assert.NoError(t, err)
	assert.True(t, within)
}
