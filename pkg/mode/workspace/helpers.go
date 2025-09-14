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
func (w *realWorkspace) getWorkspaceWorktrees(branch string) ([]status.WorktreeInfo, error) {
	// Get all worktrees for the workspace
	allWorktrees, err := w.ListWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Filter by branch if specified
	if branch != "" {
		var filteredWorktrees []status.WorktreeInfo
		for _, worktree := range allWorktrees {
			if worktree.Branch == branch {
				filteredWorktrees = append(filteredWorktrees, worktree)
			}
		}
		return filteredWorktrees, nil
	}

	return allWorktrees, nil
}

// deleteWorktreeRepositories deletes worktrees for all repositories in the workspace.
func (w *realWorkspace) deleteWorktreeRepositories(worktrees []status.WorktreeInfo, force bool) error {
	w.logger.Logf("Deleting worktrees: %d worktrees, force: %v", len(worktrees), force)

	if len(worktrees) == 0 {
		w.logger.Logf("No worktrees to delete")
		return nil
	}

	// Get workspace path
	workspacePath := w.getWorkspacePath()

	// Get workspace from status to get repository URLs
	workspace, err := w.statusManager.GetWorkspace(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to get workspace from status: %w", err)
	}

	// Create worktree instance using provider
	worktreeInstance := w.worktreeProvider(worktree.NewWorktreeParams{
		FS:              w.fs,
		Git:             w.git,
		StatusManager:   w.statusManager,
		Logger:          w.logger,
		Prompt:          w.prompt,
		RepositoriesDir: w.config.RepositoriesDir,
	})

	var errors []error
	for _, worktreeInfo := range worktrees {
		if err := w.deleteSingleWorkspaceWorktree(worktreeInfo, workspace, worktreeInstance, force); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(worktrees) {
			// All deletions failed
			return fmt.Errorf("failed to delete all worktrees: %v", errors)
		}
		// Some deletions failed
		w.logger.Logf("Some worktrees failed to delete: %v", errors)
		return fmt.Errorf("some worktrees failed to delete: %v", errors)
	}

	w.logger.Logf("Successfully deleted all %d worktrees", len(worktrees))
	return nil
}

// removeWorktreeStatusEntries removes worktree entries from status.
func (w *realWorkspace) removeWorktreeStatusEntries(worktrees []status.WorktreeInfo, force bool) error {
	w.logger.Logf("Removing worktree status entries: %d worktrees, force: %v", len(worktrees), force)

	if len(worktrees) == 0 {
		w.logger.Logf("No worktree status entries to remove")
		return nil
	}

	// Get workspace path
	workspacePath := w.getWorkspacePath()

	// Get workspace from status to get repository URLs
	workspace, err := w.statusManager.GetWorkspace(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to get workspace from status: %w", err)
	}

	var errors []error
	for _, worktreeInfo := range worktrees {
		if err := w.removeSingleWorktreeStatusEntry(worktreeInfo, workspace); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(worktrees) {
			// All removals failed
			return fmt.Errorf("failed to remove all worktree status entries: %v", errors)
		}
		// Some removals failed
		w.logger.Logf("Some worktree status entries failed to remove: %v", errors)
		return fmt.Errorf("some worktree status entries failed to remove: %v", errors)
	}

	w.logger.Logf("Successfully removed all %d worktree status entries", len(worktrees))
	return nil
}

// getWorkspacePath gets the workspace path.
func (w *realWorkspace) getWorkspacePath() string {
	// This is a simplified implementation
	// In practice, you might want to implement proper workspace path resolution
	return filepath.Dir(w.file)
}

// deleteSingleWorkspaceWorktree deletes a single worktree in workspace mode.
func (w *realWorkspace) deleteSingleWorkspaceWorktree(
	worktreeInfo status.WorktreeInfo,
	workspace *status.Workspace,
	worktreeInstance worktree.Worktree,
	force bool,
) error {
	w.logger.Logf("Deleting worktree for branch: %s", worktreeInfo.Branch)

	// Find the repository URL for this worktree
	repoURL := w.findRepositoryURLForWorktree(worktreeInfo, workspace)
	if repoURL == "" {
		w.logger.Logf("Could not find repository URL for worktree branch %s, skipping", worktreeInfo.Branch)
		return nil
	}

	// Try to get worktree path from Git
	repoPath := filepath.Join(w.config.RepositoriesDir, repoURL)
	gitWorktreePath, err := w.git.GetWorktreePath(repoPath, worktreeInfo.Branch)
	if err != nil {
		w.logger.Logf("Failed to get worktree path for branch %s (worktree may not exist in Git): %v",
			worktreeInfo.Branch, err)
		// If worktree doesn't exist in Git, just remove from status
		if err := worktreeInstance.RemoveFromStatus(repoURL, worktreeInfo.Branch); err != nil {
			w.logger.Logf("Failed to remove worktree from status for branch %s: %v", worktreeInfo.Branch, err)
			return fmt.Errorf("failed to remove worktree from status for branch %s: %w",
				worktreeInfo.Branch, err)
		}
		w.logger.Logf("Successfully removed worktree from status for branch %s (worktree did not exist in Git)",
			worktreeInfo.Branch)
		return nil
	}

	// Delete the worktree
	if err := worktreeInstance.Delete(worktree.DeleteParams{
		RepoURL:      repoURL,
		Branch:       worktreeInfo.Branch,
		WorktreePath: gitWorktreePath,
		RepoPath:     repoPath,
		Force:        force,
	}); err != nil {
		w.logger.Logf("Failed to delete worktree for branch %s: %v", worktreeInfo.Branch, err)
		return fmt.Errorf("failed to delete worktree for branch %s: %w", worktreeInfo.Branch, err)
	}

	w.logger.Logf("Successfully deleted worktree for branch %s", worktreeInfo.Branch)
	return nil
}

// findRepositoryURLForWorktree finds the repository URL for a given worktree.
func (w *realWorkspace) findRepositoryURLForWorktree(
	worktreeInfo status.WorktreeInfo,
	workspace *status.Workspace,
) string {
	for _, repoURLCandidate := range workspace.Repositories {
		repo, err := w.statusManager.GetRepository(repoURLCandidate)
		if err != nil {
			continue
		}
		// Check if this repository has the worktree
		for _, repoWorktree := range repo.Worktrees {
			if repoWorktree.Branch == worktreeInfo.Branch && repoWorktree.Remote == worktreeInfo.Remote {
				return repoURLCandidate
			}
		}
	}
	return ""
}

// removeSingleWorktreeStatusEntry removes a single worktree status entry.
func (w *realWorkspace) removeSingleWorktreeStatusEntry(
	worktreeInfo status.WorktreeInfo,
	workspace *status.Workspace,
) error {
	w.logger.Logf("Removing worktree status entry for branch: %s", worktreeInfo.Branch)

	// Find the repository URL for this worktree
	repoURL := w.findRepositoryURLForWorktree(worktreeInfo, workspace)
	if repoURL == "" {
		w.logger.Logf("Could not find repository URL for worktree branch %s, skipping status removal", worktreeInfo.Branch)
		return nil
	}

	if err := w.statusManager.RemoveWorktree(repoURL, worktreeInfo.Branch); err != nil {
		w.logger.Logf("Failed to remove worktree status entry for branch %s: %v", worktreeInfo.Branch, err)
		return fmt.Errorf("failed to remove worktree status entry for branch %s: %w",
			worktreeInfo.Branch, err)
	}

	w.logger.Logf("Successfully removed worktree status entry for branch %s", worktreeInfo.Branch)
	return nil
}
