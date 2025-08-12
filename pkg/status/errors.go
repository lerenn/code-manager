package status

import "errors"

// Error definitions for status package.
var (
	// Worktree management errors.
	ErrWorktreeAlreadyExists       = errors.New("worktree already exists")
	ErrWorktreeNotFound            = errors.New("worktree not found")
	ErrConfigurationNotInitialized = errors.New("configuration is not initialized")
)
