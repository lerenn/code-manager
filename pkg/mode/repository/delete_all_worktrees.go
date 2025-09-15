// Package repository provides Git repository management functionality for CM.
package repository

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// DeleteAllWorktrees deletes all worktrees for the repository.
func (r *realRepository) DeleteAllWorktrees(force bool) error {
	r.deps.Logger.Logf("Deleting all worktrees for single repository")

	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{})
	if err != nil {
		return err
	}

	// Get all worktrees for this repository
	worktrees, err := r.ListWorktrees()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		r.deps.Logger.Logf("No worktrees found to delete")
		return nil
	}

	r.deps.Logger.Logf("Found %d worktrees to delete", len(worktrees))

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

	var errors []error
	for _, worktreeInfo := range worktrees {
		if err := r.deleteSingleWorktree(worktreeInfo, validationResult, worktreeInstance, currentDir, force); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(worktrees) {
			// All deletions failed
			return fmt.Errorf("failed to delete all worktrees: %v", errors)
		}
		// Some deletions failed
		r.deps.Logger.Logf("Some worktrees failed to delete: %v", errors)
		return fmt.Errorf("some worktrees failed to delete: %v", errors)
	}

	r.deps.Logger.Logf("Successfully deleted all %d worktrees", len(worktrees))
	return nil
}

// deleteSingleWorktree deletes a single worktree, handling cases where it doesn't exist in Git.
func (r *realRepository) deleteSingleWorktree(
	worktreeInfo status.WorktreeInfo,
	validationResult *ValidationResult,
	worktreeInstance worktree.Worktree,
	currentDir string,
	force bool,
) error {
	r.deps.Logger.Logf("Deleting worktree for branch: %s", worktreeInfo.Branch)

	// Get worktree path from Git
	worktreePath, err := r.deps.Git.GetWorktreePath(validationResult.RepoPath, worktreeInfo.Branch)
	if err != nil {
		r.deps.Logger.Logf("Failed to get worktree path for branch %s (worktree may not exist in Git): %v",
			worktreeInfo.Branch, err)
		// If worktree doesn't exist in Git, just remove from status
		if err := worktreeInstance.RemoveFromStatus(validationResult.RepoURL, worktreeInfo.Branch); err != nil {
			r.deps.Logger.Logf("Failed to remove worktree from status for branch %s: %v", worktreeInfo.Branch, err)
			return fmt.Errorf("failed to remove worktree from status for branch %s: %w",
				worktreeInfo.Branch, err)
		}
		r.deps.Logger.Logf("Successfully removed worktree from status for branch %s (worktree did not exist in Git)",
			worktreeInfo.Branch)
		return nil
	}

	// Delete the worktree
	if err := worktreeInstance.Delete(worktree.DeleteParams{
		RepoURL:      validationResult.RepoURL,
		Branch:       worktreeInfo.Branch,
		WorktreePath: worktreePath,
		RepoPath:     currentDir,
		Force:        force,
	}); err != nil {
		r.deps.Logger.Logf("Failed to delete worktree for branch %s: %v", worktreeInfo.Branch, err)
		return fmt.Errorf("failed to delete worktree for branch %s: %w", worktreeInfo.Branch, err)
	}

	r.deps.Logger.Logf("Successfully deleted worktree for branch %s", worktreeInfo.Branch)
	return nil
}
