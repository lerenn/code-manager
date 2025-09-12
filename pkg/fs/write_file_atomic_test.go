//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFileAtomic(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("Hello, World!")

	// Test atomic write
	err := fs.WriteFileAtomic(testFile, testData, 0644)
	require.NoError(t, err)

	// Verify file exists and has correct content
	exists, err := fs.Exists(testFile)
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := fs.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, testData, content)

	// Verify file permissions
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestWriteFileAtomic_Overwrite(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	initialData := []byte("Initial content")
	newData := []byte("New content")

	// Create initial file
	err := fs.WriteFileAtomic(testFile, initialData, 0644)
	require.NoError(t, err)

	// Overwrite with new data
	err = fs.WriteFileAtomic(testFile, newData, 0600)
	require.NoError(t, err)

	// Verify new content
	content, err := fs.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, newData, content)

	// Verify new permissions
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestWriteFileAtomic_ConcurrentAccess(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "concurrent.txt")

	// Test concurrent writes (this should not cause corruption)
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		data := []byte("Data from goroutine 1")
		fs.WriteFileAtomic(testFile, data, 0644)
	}()

	go func() {
		defer func() { done <- true }()
		data := []byte("Data from goroutine 2")
		fs.WriteFileAtomic(testFile, data, 0644)
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify file exists and has content (one of the writes should succeed)
	exists, err := fs.Exists(testFile)
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := fs.ReadFile(testFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestWriteFileAtomic_ErrorHandling(t *testing.T) {
	fs := NewFS()

	// Test writing to a device file which should fail
	deviceFile := "/dev/null/test.txt"
	testData := []byte("Test data")

	err := fs.WriteFileAtomic(deviceFile, testData, 0644)
	assert.Error(t, err)

	// Verify file was not created (this might fail due to the error, but that's expected)
	exists, _ := fs.Exists(deviceFile)
	assert.False(t, exists)
}
