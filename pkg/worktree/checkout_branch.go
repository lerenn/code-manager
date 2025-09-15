// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
)

// CheckoutBranch checks out the branch in the worktree after hooks have been executed.
func (w *realWorktree) CheckoutBranch(worktreePath, branch string) error {
	w.logger.Logf("Checking out branch %s in worktree at %s", branch, worktreePath)

	if err := w.git.CheckoutBranch(worktreePath, branch); err != nil {
		return fmt.Errorf("failed to checkout branch in worktree: %w", err)
	}

	w.logger.Logf("✓ Branch %s checked out successfully in worktree", branch)
	return nil
}
