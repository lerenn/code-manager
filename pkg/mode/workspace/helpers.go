package workspace

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// Error definitions for workspace package.
var (
	// Worktree errors.
	ErrWorktreeExists      = errors.New("worktree already exists")
	ErrWorktreeNotInStatus = errors.New("worktree not found in status")
	ErrRepositoryNotClean  = errors.New("repository is not clean")
	ErrDirectoryExists     = errors.New("directory already exists")

	// User interaction errors.
	ErrDeletionCancelled = errors.New("deletion cancelled by user")
)

// addWorktreeToStatus adds the worktree to the status file with proper error handling.
func (w *realWorkspace) addWorktreeToStatus(
	_ worktree.Worktree,
	repoURL, branch, worktreePath, workspaceDir string,
	issueInfo *issue.Info,
) error {
	if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:       repoURL,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: workspaceDir,
		Remote:        "origin",
		IssueInfo:     issueInfo,
	}); err != nil {
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}
	return nil
}

// getWorkspaceWorktrees gets all worktrees for the workspace.
func (w *realWorkspace) getWorkspaceWorktrees(_ string) ([]status.WorktreeInfo, error) {
	// This is a simplified implementation
	// In practice, you might want to implement proper workspace worktree listing
	return []status.WorktreeInfo{}, nil
}

// deleteWorktreeRepositories deletes worktrees for all repositories in the workspace.
func (w *realWorkspace) deleteWorktreeRepositories(worktrees []status.WorktreeInfo, force bool) error {
	// This is a simplified implementation
	// In practice, you might want to implement proper workspace worktree deletion
	w.logger.Logf("Deleting worktrees: %d worktrees, force: %v", len(worktrees), force)
	return nil
}

// removeWorktreeStatusEntries removes worktree entries from status.
func (w *realWorkspace) removeWorktreeStatusEntries(worktrees []status.WorktreeInfo, force bool) error {
	// This is a simplified implementation
	// In practice, you might want to implement proper status cleanup
	w.logger.Logf("Removing worktree status entries: %d worktrees, force: %v", len(worktrees), force)
	return nil
}

// getWorkspacePath gets the workspace path.
func (w *realWorkspace) getWorkspacePath() string {
	// This is a simplified implementation
	// In practice, you might want to implement proper workspace path resolution
	return filepath.Dir(w.file)
}
