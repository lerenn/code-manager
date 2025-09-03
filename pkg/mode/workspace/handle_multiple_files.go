package workspace

import (
	"fmt"
)

// HandleMultipleFiles handles the selection of workspace files when multiple are found.
func (w *realWorkspace) HandleMultipleFiles(workspaceFiles []string, force bool) (string, error) {
	w.logger.Logf("Handling multiple workspace files: %v", workspaceFiles)

	if len(workspaceFiles) == 0 {
		return "", fmt.Errorf("no workspace files provided")
	}

	if len(workspaceFiles) == 1 {
		return workspaceFiles[0], nil
	}

	// If force is true, use the first file
	if force {
		w.logger.Logf("Force mode enabled, using first workspace file: %s", workspaceFiles[0])
		return workspaceFiles[0], nil
	}

	// For now, return an error indicating this needs user interaction
	// In a real implementation, this would prompt the user to select a file
	return "", fmt.Errorf("multiple workspace files found, user selection not yet implemented")
}
