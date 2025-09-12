package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsDirectoryWritable checks if a directory is writable.
func (f *realFS) IsDirectoryWritable(path string) (bool, error) {
	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(path, ".cm_test_write")
	file, err := os.Create(testFile)
	if err != nil {
		return false, err
	}
	// Clean up test file
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the error but don't fail the test
			fmt.Printf("Warning: failed to close test file: %v\n", closeErr)
		}
		if removeErr := os.Remove(testFile); removeErr != nil {
			// Log the error but don't fail the test
			fmt.Printf("Warning: failed to remove test file: %v\n", removeErr)
		}
	}()
	return true, nil
}
