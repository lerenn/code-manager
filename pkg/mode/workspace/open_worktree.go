package workspace

import (
	"fmt"
)

// OpenWorktree opens an existing worktree in workspace mode.
func (w *realWorkspace) OpenWorktree(workspaceName, branch string) (string, error) {
	w.deps.Logger.Logf("Opening worktree in workspace mode: %s", branch)

	// Get workspace from status
	workspace, err := w.deps.StatusManager.GetWorkspace(workspaceName)
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

	// Get config to access WorkspacesDir
	cfg, err := w.deps.Config.GetConfigWithFallback()
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	// Use shared utility to build workspace file path
	workspaceFilePath := BuildWorkspaceFilePath(cfg.WorkspacesDir, workspaceName, branch)

	w.deps.Logger.Logf("Opening workspace file: %s", workspaceFilePath)
	return workspaceFilePath, nil
}
