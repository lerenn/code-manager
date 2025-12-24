// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
	"path/filepath"
)

// ValidateCreation validates that worktree creation is possible.
func (w *realWorktree) ValidateCreation(params ValidateCreationParams) error {
	// Check if worktree already exists in status file first
	// If it exists in status, the worktree is already managed and we should not create it again
	existingWorktree, err := w.statusManager.GetWorktree(params.RepoURL, params.Branch)
	if err == nil && existingWorktree != nil {
		// Worktree exists in status - this is a valid state (e.g., after cloning)
		// Check if the directory also exists - if so, this is expected and valid
		exists, dirErr := w.fs.Exists(params.WorktreePath)
		if dirErr == nil && exists {
			// Both status entry and directory exist - this is the expected state after cloning
			// Return error to indicate worktree already exists (caller should handle this)
			return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, params.RepoURL, params.Branch)
		}
		// Status entry exists but directory doesn't - this is an inconsistent state
		// Still return error to indicate worktree exists in status
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, params.RepoURL, params.Branch)
	}

	// Check if worktree directory already exists (only if not in status)
	exists, err := w.fs.Exists(params.WorktreePath)
	if err != nil {
		return fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, params.WorktreePath)
	}

	// Create worktree directory structure
	if err := w.fs.MkdirAll(filepath.Dir(params.WorktreePath), 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return nil
}
