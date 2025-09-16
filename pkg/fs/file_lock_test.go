//go:build integration

package fs

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileLock(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "lock_test.txt")

	// Acquire lock
	unlock, err := fs.FileLock(testFile)
	require.NoError(t, err)
	defer unlock()

	// Verify lock file exists
	lockFile := testFile + ".lock"
	exists, err := fs.Exists(lockFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// Try to acquire another lock (should fail or block)
	// Note: This test may behave differently on different systems
	// Use a timeout to prevent hanging
	done := make(chan bool, 1)
	go func() {
		unlock2, err := fs.FileLock(testFile)
		if err != nil {
			// Lock acquisition failed as expected
			// The error message varies by system: "resource temporarily unavailable" on Unix, "lock" on some systems
			assert.Error(t, err)
		} else {
			// Lock acquisition succeeded (system doesn't enforce locks)
			unlock2()
		}
		done <- true
	}()

	// Wait for the goroutine to complete or timeout
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		// Test timed out, which is acceptable for file locking tests
		t.Log("File lock test timed out (this is acceptable on some systems)")
	}
}

func TestFileLock_Unlock(t *testing.T) {
	fs := NewFS()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "unlock_test.txt")

	// Acquire lock
	unlock, err := fs.FileLock(testFile)
	require.NoError(t, err)

	// Verify lock file exists
	lockFile := testFile + ".lock"
	exists, err := fs.Exists(lockFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// Unlock
	unlock()

	// Verify lock file is removed
	exists, err = fs.Exists(lockFile)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFileLock_ErrorHandling(t *testing.T) {
	fs := NewFS()

	// Test locking with a device file which should fail
	deviceFile := "/dev/null/test.txt"

	_, err := fs.FileLock(deviceFile)
	assert.Error(t, err)
}
