//go:build integration

package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_IsNotExist(t *testing.T) {
	fs := NewFS()

	// Test with non-existent file error
	_, err := fs.ReadFile("non-existent-file.txt")
	assert.Error(t, err)
	assert.True(t, fs.IsNotExist(err))

	// Test with existing file (should not be IsNotExist)
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.False(t, fs.IsNotExist(err))
}
