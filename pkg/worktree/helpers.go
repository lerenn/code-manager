// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
)

// createWorktreeDirectory creates the worktree directory.
func (w *realWorktree) createWorktreeDirectory(worktreePath string) error {
	if err := w.fs.MkdirAll(worktreePath, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}
	return nil
}

// cleanupWorktreeDirectory removes the worktree directory.
func (w *realWorktree) cleanupWorktreeDirectory(worktreePath string) error {
	if err := w.fs.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to cleanup worktree directory: %w", err)
	}
	return nil
}

// promptForConfirmation prompts the user for confirmation before deletion.
func (w *realWorktree) promptForConfirmation(branch, worktreePath string) error {
	message := fmt.Sprintf(
		"You are about to delete the worktree for branch '%s'\nWorktree path: %s\nAre you sure you want to continue?",
		branch, worktreePath,
	)

	result, err := w.prompt.PromptForConfirmation(message, false)
	if err != nil {
		return err
	}

	if !result {
		return ErrDeletionCancelled
	}

	return nil
}
