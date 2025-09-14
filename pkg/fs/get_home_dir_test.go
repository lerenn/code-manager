//go:build integration

package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFS_GetHomeDir(t *testing.T) {
	fs := NewFS()

	homeDir, err := fs.GetHomeDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, homeDir)

	// Verify it's a directory
	exists, err := fs.Exists(homeDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	isDir, err := fs.IsDir(homeDir)
	assert.NoError(t, err)
	assert.True(t, isDir)
}
