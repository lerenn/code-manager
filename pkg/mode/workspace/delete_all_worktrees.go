package workspace

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/status"
)

// DeleteAllWorktrees deletes all worktrees for the workspace.
func (w *realWorkspace) DeleteAllWorktrees(force bool) error {
	w.logger.Logf("Deleting all worktrees for workspace")

	// Load workspace configuration (only if not already loaded)
	if w.file == "" {
		if err := w.Load(); err != nil {
			return fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// Get all worktrees for this workspace
	allWorktrees, err := w.ListWorktrees()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	if len(allWorktrees) == 0 {
		w.logger.Logf("No worktrees found to delete")
		return nil
	}

	w.logger.Logf("Found %d worktrees to delete", len(allWorktrees))

	// Group worktrees by branch
	branchGroups := make(map[string][]status.WorktreeInfo)
	for _, worktree := range allWorktrees {
		branchGroups[worktree.Branch] = append(branchGroups[worktree.Branch], worktree)
	}

	var errors []error
	for branch := range branchGroups {
		w.logger.Logf("Deleting worktrees for branch: %s", branch)

		if err := w.DeleteWorktree(branch, force); err != nil {
			w.logger.Logf("Failed to delete worktrees for branch %s: %v", branch, err)
			errors = append(errors, fmt.Errorf("failed to delete worktrees for branch %s: %w", branch, err))
		} else {
			w.logger.Logf("Successfully deleted worktrees for branch %s", branch)
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(branchGroups) {
			// All deletions failed
			return fmt.Errorf("failed to delete all worktrees: %v", errors)
		}
		// Some deletions failed
		w.logger.Logf("Some worktrees failed to delete: %v", errors)
		return fmt.Errorf("some worktrees failed to delete: %v", errors)
	}

	w.logger.Logf("Successfully deleted all %d worktrees across %d branches", len(allWorktrees), len(branchGroups))
	return nil
}
