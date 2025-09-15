// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
)

// CleanupDirectory removes the worktree directory.
func (w *realWorktree) CleanupDirectory(worktreePath string) error {
	if err := w.fs.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}
	return nil
}
