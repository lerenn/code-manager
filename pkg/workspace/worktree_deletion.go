// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/status"
)

// WorktreeWithRepo represents a worktree with its associated repository information.
type WorktreeWithRepo struct {
	status.WorktreeInfo
	RepoURL  string
	RepoPath string
}

// getWorkspaceWorktrees gets all worktrees for this workspace and branch.
func (w *realWorkspace) getWorkspaceWorktrees(branch string) ([]WorktreeWithRepo, error) {
	// Get workspace path
	workspacePath, err := filepath.Abs(w.OriginalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspacePath)
	if err != nil {
		// If workspace not found, return empty list with no error
		if errors.Is(err, status.ErrWorkspaceNotFound) {
			return []WorktreeWithRepo{}, nil
		}
		return nil, err
	}

	w.verboseLogf("Looking for worktrees with workspace path: %s", workspacePath)
	w.verboseLogf("Workspace repositories: %v", workspace.Repositories)

	// Get worktrees for each repository in the workspace that match the branch
	var workspaceWorktrees []WorktreeWithRepo

	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := w.statusManager.GetRepository(repoURL)
		if err != nil {
			continue // Skip if repository not found
		}

		// Get worktrees for this repository that match the branch
		for _, worktree := range repo.Worktrees {
			if worktree.Branch == branch {
				workspaceWorktrees = append(workspaceWorktrees, WorktreeWithRepo{
					WorktreeInfo: worktree,
					RepoURL:      repoURL,
					RepoPath:     repo.Path,
				})
				w.verboseLogf("✓ Found matching worktree: %s:%s for repository %s", worktree.Remote, worktree.Branch, repoURL)
			}
		}
	}

	return workspaceWorktrees, nil
}

// deleteWorktreeRepositories deletes worktrees for all repositories.
func (w *realWorkspace) deleteWorktreeRepositories(workspaceWorktrees []WorktreeWithRepo, force bool) error {
	for i, worktreeWithRepo := range workspaceWorktrees {
		w.verboseLogf("Deleting worktree %d/%d: %s:%s for repository %s", i+1, len(workspaceWorktrees),
			worktreeWithRepo.Remote, worktreeWithRepo.Branch, worktreeWithRepo.RepoURL)

		// Delete Git worktree
		worktreePath := w.buildWorktreePath(worktreeWithRepo.RepoURL, worktreeWithRepo.Remote, worktreeWithRepo.Branch)
		if err := w.git.RemoveWorktree(worktreeWithRepo.RepoPath, worktreePath); err != nil {
			if !force {
				return fmt.Errorf("failed to delete Git worktree for %s:%s: %w",
					worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
			}
			w.verboseLogf("Warning: failed to delete Git worktree for %s:%s: %v",
				worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
		}

		// Remove worktree directory
		if err := w.fs.RemoveAll(worktreePath); err != nil {
			if !force {
				return fmt.Errorf("failed to remove worktree directory %s: %w", worktreePath, err)
			}
			w.verboseLogf("Warning: failed to remove worktree directory %s: %v", worktreePath, err)
		}

		w.verboseLogf("✓ Worktree deleted successfully for %s:%s", worktreeWithRepo.Remote, worktreeWithRepo.Branch)
	}

	return nil
}

// removeWorktreeStatusEntries removes worktree entries from status file.
func (w *realWorkspace) removeWorktreeStatusEntries(workspaceWorktrees []WorktreeWithRepo, force bool) error {
	for _, worktreeWithRepo := range workspaceWorktrees {
		if err := w.statusManager.RemoveWorktree(worktreeWithRepo.RepoURL, worktreeWithRepo.Branch); err != nil {
			if !force {
				return fmt.Errorf("failed to remove worktree from status file: %w", err)
			}
			w.verboseLogf("Warning: failed to remove worktree from status file: %v", err)
		}
	}

	return nil
}
