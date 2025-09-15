// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/logger"
)

// Create creates a new worktree with proper validation and cleanup.
func (w *realWorktree) Create(params CreateParams) error {
	// Create logger if not already created
	if w.logger == nil {
		w.logger = logger.NewNoopLogger()
	}

	w.logger.Logf("Creating worktree for %s:%s at %s", params.Remote, params.Branch, params.WorktreePath)

	// Validate creation
	if err := w.ValidateCreation(ValidateCreationParams{
		RepoURL:      params.RepoURL,
		Branch:       params.Branch,
		WorktreePath: params.WorktreePath,
		RepoPath:     params.RepoPath,
	}); err != nil {
		return err
	}

	// Ensure branch exists
	if err := w.EnsureBranchExists(params.RepoPath, params.Branch); err != nil {
		return err
	}

	// Create worktree directory
	if err := w.createWorktreeDirectory(params.WorktreePath); err != nil {
		return err
	}

	// Create Git worktree with --no-checkout to allow hooks to prepare
	if err := w.git.CreateWorktreeWithNoCheckout(params.RepoPath, params.WorktreePath, params.Branch); err != nil {
		// Clean up directory on failure
		if cleanupErr := w.cleanupWorktreeDirectory(params.WorktreePath); cleanupErr != nil {
			w.logger.Logf("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	w.logger.Logf("✓ Worktree created successfully for %s:%s", params.Remote, params.Branch)
	return nil
}
