// Package workspace provides workspace management functionality for CM.
package workspace

// cleanupOnFailure performs cleanup operations when worktree creation fails.
func (w *realWorkspace) cleanupOnFailure(
	createdWorktrees []struct {
		repoURL string
		branch  string
		path    string
	},
	worktreeWorkspacePath string,
) {
	w.cleanupFailedWorktrees(createdWorktrees)
	w.cleanupWorktreeWorkspaceFile(worktreeWorkspacePath)
}

// cleanupFailedWorktrees removes worktree entries from status file.
func (w *realWorkspace) cleanupFailedWorktrees(createdWorktrees []struct {
	repoURL string
	branch  string
	path    string
}) {
	w.verboseLogf("Cleaning up failed worktrees from status file")
	for _, worktree := range createdWorktrees {
		if err := w.statusManager.RemoveWorktree(worktree.repoURL, worktree.branch); err != nil {
			w.verboseLogf("Warning: failed to remove worktree from status file: %v", err)
		}
	}
}

// cleanupWorktreeWorkspaceFile removes the worktree-specific workspace file.
func (w *realWorkspace) cleanupWorktreeWorkspaceFile(worktreeWorkspacePath string) {
	w.verboseLogf("Cleaning up worktree workspace file")
	if err := w.fs.RemoveAll(worktreeWorkspacePath); err != nil {
		w.verboseLogf("Warning: failed to remove worktree workspace file: %v", err)
	}
}
