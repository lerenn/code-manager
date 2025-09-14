package workspace

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/status"
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

	var errors []error
	for _, worktreeInfo := range worktrees {
		if err := w.deleteSingleWorkspaceWorktree(worktreeInfo, workspace, force); err != nil {
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

// deleteSingleWorkspaceWorktree deletes a single worktree in workspace mode using repositoryProvider.
func (w *realWorkspace) deleteSingleWorkspaceWorktree(
	worktreeInfo status.WorktreeInfo,
	workspace *status.Workspace,
	force bool,
) error {
	w.logger.Logf("Deleting worktree for branch: %s", worktreeInfo.Branch)

	// Find the repository URL for this worktree
	repoURL := w.findRepositoryURLForWorktree(worktreeInfo, workspace)
	if repoURL == "" {
		w.logger.Logf("Could not find repository URL for worktree branch %s, skipping", worktreeInfo.Branch)
		return nil
	}

	// Get repository path
	repoPath := filepath.Join(w.config.RepositoriesDir, repoURL)

	// Create repository instance using repositoryProvider
	repoInstance := w.repositoryProvider(repository.NewRepositoryParams{
		FS:               w.fs,
		Git:              w.git,
		Config:           w.config,
		StatusManager:    w.statusManager,
		Logger:           w.logger,
		Prompt:           w.prompt,
		WorktreeProvider: w.safeWorktreeProvider(),
		HookManager:      w.hookManager,
		RepositoryName:   repoPath,
	})

	// Use repository's DeleteWorktree method
	if err := repoInstance.DeleteWorktree(worktreeInfo.Branch, force); err != nil {
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
