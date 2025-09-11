package workspace

import (
	"fmt"
)

// Load handles the complete workspace loading workflow.
// It detects workspace files, handles user selection if multiple files are found,
// and loads the workspace configuration for display.
func (w *realWorkspace) Load() error {
	// If already loaded, just parse and display the configuration
	if w.file != "" {
		workspaceConfig, err := w.ParseFile(w.file)
		if err != nil {
			return fmt.Errorf("failed to parse workspace file: %w", err)
		}

		w.logger.Logf("Workspace mode detected")

		workspaceName := w.GetName(workspaceConfig, w.file)
		w.logger.Logf("Found workspace: %s", workspaceName)

		w.logger.Logf("Workspace configuration:")
		w.logger.Logf("  Folders: %d", len(workspaceConfig.Folders))
		for _, folder := range workspaceConfig.Folders {
			w.logger.Logf("    - %s: %s", folder.Name, folder.Path)
		}

		return nil
	}

	// Workspace mode is now determined by explicit --workspace flag
	// This method is no longer used for automatic workspace detection
	w.file = ""
	return nil
}
