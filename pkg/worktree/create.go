// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/git"
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

	// Create worktree directory
	if err := w.createWorktreeDirectory(params.WorktreePath); err != nil {
		return err
	}

	// Create worktree or clone based on detached mode
	if params.Detached {
		return w.createDetachedClone(params)
	}

	// Ensure branch exists (only for regular worktrees)
	if err := w.EnsureBranchExists(params.RepoPath, params.Branch); err != nil {
		return err
	}

	return w.createRegularWorktree(params)
}

// createDetachedClone creates a standalone clone for detached worktrees.
func (w *realWorktree) createDetachedClone(params CreateParams) error {
	// Check if branch exists locally - if so, clone from local path
	// Otherwise, clone from remote URL
	branchExists, err := w.git.BranchExists(params.RepoPath, params.Branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if branchExists {
		return w.createDetachedCloneFromLocal(params)
	}
	return w.createDetachedCloneFromRemote(params)
}

// createDetachedCloneFromLocal creates a detached clone from a local repository path.
func (w *realWorktree) createDetachedCloneFromLocal(params CreateParams) error {
	w.logger.Logf("Branch exists locally, cloning from local path: %s", params.RepoPath)
	if err := w.git.CloneToPath(params.RepoPath, params.WorktreePath, params.Branch); err != nil {
		w.cleanupOnError(params.WorktreePath)
		return fmt.Errorf("failed to create detached clone: %w", err)
	}
	w.logger.Logf("✓ Detached clone created successfully for %s:%s", params.Remote, params.Branch)
	return nil
}

// createDetachedCloneFromRemote creates a detached clone from a remote URL.
func (w *realWorktree) createDetachedCloneFromRemote(params CreateParams) error {
	remoteURL, err := w.git.GetRemoteURL(params.RepoPath, params.Remote)
	if err != nil {
		return fmt.Errorf("failed to get remote URL for %s: %w", params.Remote, err)
	}

	w.logger.Logf("Branch doesn't exist locally, cloning from remote URL: %s", remoteURL)

	// Clone from the remote URL (this will fetch all branches)
	if err := w.git.Clone(git.CloneParams{
		RepoURL:    remoteURL,
		TargetPath: params.WorktreePath,
		Recursive:  true,
	}); err != nil {
		w.cleanupOnError(params.WorktreePath)
		return fmt.Errorf("failed to clone from remote: %w", err)
	}

	// Checkout the specific branch
	if err := w.git.CheckoutBranch(params.WorktreePath, params.Branch); err != nil {
		w.cleanupOnError(params.WorktreePath)
		return fmt.Errorf("failed to checkout branch %s: %w", params.Branch, err)
	}

	w.logger.Logf("✓ Detached clone created successfully for %s:%s", params.Remote, params.Branch)
	return nil
}

// cleanupOnError attempts to clean up the worktree directory on error.
func (w *realWorktree) cleanupOnError(worktreePath string) {
	if cleanupErr := w.cleanupWorktreeDirectory(worktreePath); cleanupErr != nil {
		w.logger.Logf("Warning: failed to clean up worktree directory: %v", cleanupErr)
	}
}

// createRegularWorktree creates a regular Git worktree.
func (w *realWorktree) createRegularWorktree(params CreateParams) error {
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
