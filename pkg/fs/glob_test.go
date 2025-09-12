//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_Glob(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-glob-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create some .code-workspace files
	workspace1 := filepath.Join(tmpDir, "project.code-workspace")
	workspace2 := filepath.Join(tmpDir, "dev.code-workspace")
	otherFile := filepath.Join(tmpDir, "other.txt")

	err = os.WriteFile(workspace1, []byte("{}"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(workspace2, []byte("{}"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(otherFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Test glob pattern matching
	pattern := filepath.Join(tmpDir, "*.code-workspace")
	matches, err := fs.Glob(pattern)
	assert.NoError(t, err)
	assert.Len(t, matches, 2)
	assert.Contains(t, matches, workspace1)
	assert.Contains(t, matches, workspace2)

	// Test glob with no matches
	noMatchPattern := filepath.Join(tmpDir, "*.nonexistent")
	noMatches, err := fs.Glob(noMatchPattern)
	assert.NoError(t, err)
	assert.Len(t, noMatches, 0)

	// Test glob with invalid pattern
	_, err = fs.Glob("[invalid")
	assert.Error(t, err)
}
