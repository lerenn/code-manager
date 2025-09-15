package workspace

import (
	"fmt"
)

// OpenWorktree opens an existing worktree in workspace mode.
// This method works with workspace names from the status file, not workspace files.
func (w *realWorkspace) OpenWorktree(workspaceName, branch string) (string, error) {
	w.logger.Logf("Opening worktree in workspace mode: %s", branch)

	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return "", fmt.Errorf("workspace '%s' not found in status.yaml: %w", workspaceName, err)
	}

	// Check if the worktree exists in the workspace worktrees list
	var found bool
	for _, worktree := range workspace.Worktrees {
		if worktree == branch {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("worktree '%s' not found in workspace '%s'", branch, workspaceName)
	}

	// Use shared utility to build workspace file path
	workspaceFilePath := buildWorkspaceFilePath(w.config.WorkspacesDir, workspaceName, branch)

	w.logger.Logf("Opening workspace file: %s", workspaceFilePath)
	return workspaceFilePath, nil
}
