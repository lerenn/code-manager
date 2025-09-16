package fs

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to user's home directory.
func (f *realFS) ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := f.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}

	return filepath.Join(homeDir, strings.TrimPrefix(path, "~")), nil
}
