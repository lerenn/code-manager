// Package worktree provides worktree management functionality for CM.
package worktree

// RemoveFromStatus removes the worktree from the status file.
func (w *realWorktree) RemoveFromStatus(repoURL, branch string) error {
	return w.statusManager.RemoveWorktree(repoURL, branch)
}
