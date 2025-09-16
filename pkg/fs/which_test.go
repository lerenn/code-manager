//go:build integration

package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFS_Which(t *testing.T) {
	fs := NewFS()

	// Test finding an existing command (ls should be available on Unix-like systems)
	path, err := fs.Which("ls")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "ls")

	// Test finding a non-existing command
	_, err = fs.Which("non-existing-command-xyz123")
	assert.Error(t, err)

	// Test finding echo command
	path, err = fs.Which("echo")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "echo")
}
