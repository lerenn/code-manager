package repository

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/worktree"
)

// DeleteWorktree deletes a worktree for the repository with the specified branch.
func (r *realRepository) DeleteWorktree(branch string, force bool) error {
	r.logger.Logf("Deleting worktree for single repository with branch: %s", branch)

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
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: r.config.RepositoriesDir,
	})

	// Get worktree path from Git
	worktreePath, err := r.git.GetWorktreePath(validationResult.RepoPath, branch)
	if err != nil {
		r.logger.Logf("Failed to get worktree path for branch %s (worktree may not exist in Git): %v",
			branch, err)
		// If worktree doesn't exist in Git, just remove from status
		if err := worktreeInstance.RemoveFromStatus(validationResult.RepoURL, branch); err != nil {
			r.logger.Logf("Failed to remove worktree from status for branch %s: %v", branch, err)
			return fmt.Errorf("failed to remove worktree from status for branch %s: %w",
				branch, err)
		}
		r.logger.Logf("Successfully removed worktree from status for branch %s (worktree did not exist in Git)",
			branch)
		return nil
	}

	r.logger.Logf("Worktree path: %s", worktreePath)

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

	r.logger.Logf("Successfully deleted worktree for branch %s", branch)

	return nil
}
