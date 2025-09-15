package workspace

import (
	"errors"
)

// Error definitions for workspace package.
var (
	// Worktree errors.
	ErrWorktreeExists      = errors.New("worktree already exists")
	ErrWorktreeNotInStatus = errors.New("worktree not found in status")
	ErrRepositoryNotClean  = errors.New("repository is not clean")
	ErrDirectoryExists     = errors.New("directory already exists")

	// User interaction errors.
	ErrDeletionCancelled = errors.New("deletion cancelled by user")
)
