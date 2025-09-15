// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/status"
)

// AddToStatus adds the worktree to the status file.
func (w *realWorktree) AddToStatus(params AddToStatusParams) error {
	if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:       params.RepoURL,
		Branch:        params.Branch,
		WorktreePath:  params.WorktreePath,
		WorkspacePath: params.WorkspacePath,
		Remote:        params.Remote,
		IssueInfo:     params.IssueInfo,
	}); err != nil {
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}
	return nil
}
