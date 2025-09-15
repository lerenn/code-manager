// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"path/filepath"
)

// BuildPath constructs a worktree path from repository URL, remote name, and branch.
func (w *realWorktree) BuildPath(repoURL, remoteName, branch string) string {
	// Use structure: $base_path/<repo_url>/<remote_name>/<branch>
	return filepath.Join(w.repositoriesDir, repoURL, remoteName, branch)
}
