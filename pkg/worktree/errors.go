// Package worktree provides worktree management functionality and error definitions.
package worktree

import "errors"

// Error definitions for worktree package.
var (
	// Worktree errors.
	ErrWorktreeExists      = errors.New("worktree already exists")
	ErrWorktreeNotInStatus = errors.New("worktree not found in status file")

	// Directory errors.
	ErrDirectoryExists = errors.New("directory already exists")

	// User interaction errors.
	ErrDeletionCancelled = errors.New("deletion cancelled by user")
)
