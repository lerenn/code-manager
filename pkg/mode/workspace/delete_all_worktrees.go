package workspace

import (
	"fmt"
)

// DeleteAllWorktrees deletes all worktrees for the workspace.
// This method works with workspace names from the status file, not workspace files.
func (w *realWorkspace) DeleteAllWorktrees(workspaceName string, force bool) error {
	w.logger.Logf("Deleting all worktrees for workspace: %s", workspaceName)

	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("workspace '%s' not found in status.yaml: %w", workspaceName, err)
	}

	if len(workspace.Worktrees) == 0 {
		w.logger.Logf("No worktrees found to delete")
		return nil
	}

	w.logger.Logf("Found %d worktrees to delete", len(workspace.Worktrees))

	var errors []error
	for _, branch := range workspace.Worktrees {
		w.logger.Logf("Deleting worktrees for branch: %s", branch)

		if err := w.DeleteWorktree(workspaceName, branch, force); err != nil {
			w.logger.Logf("Failed to delete worktrees for branch %s: %v", branch, err)
			errors = append(errors, fmt.Errorf("failed to delete worktrees for branch %s: %w", branch, err))
		} else {
			w.logger.Logf("Successfully deleted worktrees for branch %s", branch)
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(workspace.Worktrees) {
			// All deletions failed
			return fmt.Errorf("failed to delete all worktrees: %v", errors)
		}
		// Some deletions failed
		w.logger.Logf("Some worktrees failed to delete: %v", errors)
		return fmt.Errorf("some worktrees failed to delete: %v", errors)
	}

	w.logger.Logf("Successfully deleted all %d worktrees", len(workspace.Worktrees))
	return nil
}
