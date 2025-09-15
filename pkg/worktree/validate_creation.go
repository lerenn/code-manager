// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
	"path/filepath"
)

// ValidateCreation validates that worktree creation is possible.
func (w *realWorktree) ValidateCreation(params ValidateCreationParams) error {
	// Check if worktree directory already exists
	exists, err := w.fs.Exists(params.WorktreePath)
	if err != nil {
		return fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, params.WorktreePath)
	}

	// Check if worktree already exists in status file
	existingWorktree, err := w.statusManager.GetWorktree(params.RepoURL, params.Branch)
	if err == nil && existingWorktree != nil {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, params.RepoURL, params.Branch)
	}

	// Create worktree directory structure
	if err := w.fs.MkdirAll(filepath.Dir(params.WorktreePath), 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return nil
}
