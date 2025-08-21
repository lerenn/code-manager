// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"
	"path/filepath"
)

// buildWorktreePath builds the worktree path for a repository, remote name, and branch.
func (w *realWorkspace) buildWorktreePath(repoURL, remoteName, branch string) string {
	// Use new structure: $base_path/<repo_url>/<remote_name>/<branch>
	return filepath.Join(w.config.BasePath, repoURL, remoteName, branch)
}

// cleanupWorktreeDirectory cleans up the worktree directory.
func (w *realWorkspace) cleanupWorktreeDirectory(worktreePath string) error {
	if err := w.fs.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to cleanup worktree directory: %w", err)
	}
	return nil
}
