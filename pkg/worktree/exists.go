// Package worktree provides worktree management functionality for CM.
package worktree

// Exists checks if a worktree exists for the specified branch.
func (w *realWorktree) Exists(repoPath, branch string) (bool, error) {
	return w.git.WorktreeExists(repoPath, branch)
}
