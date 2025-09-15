// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
)

// Delete deletes a worktree with proper cleanup and confirmation.
func (w *realWorktree) Delete(params DeleteParams) error {
	w.logger.Logf("Deleting worktree for %s at %s", params.Branch, params.WorktreePath)

	// Validate deletion
	if err := w.ValidateDeletion(ValidateDeletionParams{
		RepoURL: params.RepoURL,
		Branch:  params.Branch,
	}); err != nil {
		return err
	}

	// Prompt for confirmation unless force flag is used
	if !params.Force {
		if err := w.promptForConfirmation(params.Branch, params.WorktreePath); err != nil {
			return err
		}
	}

	// Remove worktree from Git tracking first
	if err := w.git.RemoveWorktree(params.RepoPath, params.WorktreePath, params.Force); err != nil {
		w.logger.Logf("Failed to remove worktree from Git (worktree may not exist in Git): %v", err)
		// If worktree doesn't exist in Git, we'll still try to clean up the directory and status
	} else {
		w.logger.Logf("Successfully removed worktree from Git")
	}

	// Remove worktree directory (this may fail if directory doesn't exist, which is fine)
	if err := w.fs.RemoveAll(params.WorktreePath); err != nil {
		w.logger.Logf("Failed to remove worktree directory (directory may not exist): %v", err)
		// Don't return error here as the directory might not exist
	} else {
		w.logger.Logf("Successfully removed worktree directory")
	}

	// Remove entry from status file
	if err := w.RemoveFromStatus(params.RepoURL, params.Branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status: %w", err)
	}

	w.logger.Logf("✓ Worktree deleted successfully for %s", params.Branch)
	return nil
}
