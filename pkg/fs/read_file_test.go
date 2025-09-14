//go:build integration

package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_ReadFile(t *testing.T) {
	fs := NewFS()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write content to the file
	content := []byte("test content")
	err = os.WriteFile(tmpFile.Name(), content, 0644)
	require.NoError(t, err)

	// Test reading existing file
	readContent, err := fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)

	// Test reading non-existing file
	_, err = fs.ReadFile("non-existing-file.txt")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}
