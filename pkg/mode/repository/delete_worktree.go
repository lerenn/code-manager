package repository

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/worktree"
)

// DeleteWorktree deletes a worktree for the repository with the specified branch.
func (r *realRepository) DeleteWorktree(branch string, force bool) error {
	r.deps.Logger.Logf("Deleting worktree for single repository with branch: %s", branch)

	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{})
	if err != nil {
		return err
	}

	// Check if worktree exists in status file
	if err := r.ValidateWorktreeExists(validationResult.RepoURL, branch); err != nil {
		return err
	}

	// Get current directory
	currentDir, err := filepath.Abs(r.repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create worktree instance using provider
	cfg, err := r.deps.Config.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	worktreeProvider := r.deps.WorktreeProvider
	worktreeInstance := worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.deps.FS,
		Git:             r.deps.Git,
		StatusManager:   r.deps.StatusManager,
		Logger:          r.deps.Logger,
		Prompt:          r.deps.Prompt,
		RepositoriesDir: cfg.RepositoriesDir,
	})

	// Get worktree path from Git
	worktreePath, err := r.deps.Git.GetWorktreePath(validationResult.RepoPath, branch)
	if err != nil {
		r.deps.Logger.Logf("Failed to get worktree path for branch %s (worktree may not exist in Git): %v",
			branch, err)
		// If worktree doesn't exist in Git, just remove from status
		if err := worktreeInstance.RemoveFromStatus(validationResult.RepoURL, branch); err != nil {
			r.deps.Logger.Logf("Failed to remove worktree from status for branch %s: %v", branch, err)
			return fmt.Errorf("failed to remove worktree from status for branch %s: %w",
				branch, err)
		}
		r.deps.Logger.Logf("Successfully removed worktree from status for branch %s (worktree did not exist in Git)",
			branch)
		return nil
	}

	r.deps.Logger.Logf("Worktree path: %s", worktreePath)

	// Delete the worktree
	if err := worktreeInstance.Delete(worktree.DeleteParams{
		RepoURL:      validationResult.RepoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     currentDir,
		Force:        force,
	}); err != nil {
		return err
	}

	r.deps.Logger.Logf("Successfully deleted worktree for branch %s", branch)

	return nil
}
