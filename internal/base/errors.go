// Package base provides base functionality and error definitions.
package base

import "errors"

// Error definitions for base package.
var (
	// Git configuration errors.
	ErrGitConfiguration = errors.New("git configuration error")

	// Worktree directory errors.
	ErrFailedToCheckWorktreeDirectoryExists = errors.New("failed to check if worktree directory exists")
	ErrFailedToRemoveWorktreeDirectory      = errors.New("failed to remove worktree directory")
)
