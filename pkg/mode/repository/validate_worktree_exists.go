package repository

import "fmt"

// ValidateWorktreeExists validates that a worktree exists in the status file.
func (r *realRepository) ValidateWorktreeExists(repoURL, branch string) error {
	existingWorktree, err := r.deps.StatusManager.GetWorktree(repoURL, branch)
	if err != nil || existingWorktree == nil {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotInStatus, repoURL, branch)
	}
	return nil
}
