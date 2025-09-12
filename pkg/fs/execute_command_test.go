//go:build integration

package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFS_ExecuteCommand(t *testing.T) {
	fs := NewFS()

	// Test executing a simple command that should succeed
	err := fs.ExecuteCommand("echo", "hello world")
	assert.NoError(t, err)

	// Test executing a command with multiple arguments
	err = fs.ExecuteCommand("echo", "hello", "world", "test")
	assert.NoError(t, err)

	// Test executing a non-existing command (should fail)
	err = fs.ExecuteCommand("non-existing-command-xyz123")
	assert.Error(t, err)

	// Test executing ls command
	err = fs.ExecuteCommand("ls", "-la")
	assert.NoError(t, err)
}
