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

	// Get worktree path from Git
	worktreePath, err := r.git.GetWorktreePath(validationResult.RepoPath, branch)
	if err != nil {
		return fmt.Errorf("failed to get worktree path: %w", err)
	}

	r.logger.Logf("Worktree path: %s", worktreePath)

	// Get current directory
	currentDir, err := filepath.Abs(r.repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create worktree instance using provider
	cfg, err := r.configManager.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: cfg.RepositoriesDir,
	})

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
