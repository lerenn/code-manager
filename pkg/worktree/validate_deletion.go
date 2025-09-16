// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
)

// ValidateDeletion validates that worktree deletion is possible.
func (w *realWorktree) ValidateDeletion(params ValidateDeletionParams) error {
	// Check if worktree exists in status file
	existingWorktree, err := w.statusManager.GetWorktree(params.RepoURL, params.Branch)
	if err != nil || existingWorktree == nil {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotInStatus, params.RepoURL, params.Branch)
	}

	return nil
}
