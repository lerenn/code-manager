package workspace

import (
	"fmt"
)

// DetectWorkspaceFiles checks if the current directory contains workspace files.
func (w *realWorkspace) DetectWorkspaceFiles() ([]string, error) {
	w.logger.Logf("Detecting workspace files in current directory")

	// Check for .code-workspace files
	workspaceFiles, err := w.fs.Glob("*.code-workspace")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToCheckWorkspaceFiles, err)
	}

	w.logger.Logf("Found %d workspace files: %v", len(workspaceFiles), workspaceFiles)
	return workspaceFiles, nil
}
