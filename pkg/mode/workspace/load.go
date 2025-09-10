package workspace

import (
	"fmt"
)

// Load handles the complete workspace loading workflow.
// It detects workspace files, handles user selection if multiple files are found,
// and loads the workspace configuration for display.
func (w *realWorkspace) Load(force bool) error {
	// If already loaded, just parse and display the configuration
	if w.OriginalFile != "" {
		workspaceConfig, err := w.ParseFile(w.OriginalFile)
		if err != nil {
			return fmt.Errorf("failed to parse workspace file: %w", err)
		}

		w.logger.Logf("Workspace mode detected")

		workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
		w.logger.Logf("Found workspace: %s", workspaceName)

		w.logger.Logf("Workspace configuration:")
		w.logger.Logf("  Folders: %d", len(workspaceConfig.Folders))
		for _, folder := range workspaceConfig.Folders {
			w.logger.Logf("    - %s: %s", folder.Name, folder.Path)
		}

		return nil
	}

	// Detect workspace files
	workspaceFiles, err := w.DetectWorkspaceFiles()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWorkspaceFileNotFound, err)
	}

	if len(workspaceFiles) == 0 {
		w.OriginalFile = ""
		return nil
	}

	// If only one workspace file, store it directly
	if len(workspaceFiles) == 1 {
		w.OriginalFile = workspaceFiles[0]
	} else {
		// If multiple workspace files, handle user selection
		selectedFile, err := w.HandleMultipleFiles(workspaceFiles, force)
		if err != nil {
			return err
		}
		w.OriginalFile = selectedFile
	}

	// Load and display workspace configuration
	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	w.logger.Logf("Workspace mode detected")

	workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
	w.logger.Logf("Found workspace: %s", workspaceName)

	w.logger.Logf("Workspace configuration:")
	w.logger.Logf("  Folders: %d", len(workspaceConfig.Folders))
	for _, folder := range workspaceConfig.Folders {
		w.logger.Logf("    - %s: %s", folder.Name, folder.Path)
	}

	return nil
}
