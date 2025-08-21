// Package workspace provides workspace management functionality and error definitions.
package workspace

import "errors"

// Error definitions for workspace package.
var (
	// Workspace file errors.
	ErrFailedToCheckWorkspaceFiles = errors.New("failed to check for workspace files")
	ErrWorkspaceFileNotFound       = errors.New("workspace file not found")
	ErrInvalidWorkspaceFile        = errors.New("invalid workspace file")
	ErrNoRepositoriesFound         = errors.New("no repositories found in workspace")

	// Repository errors.
	ErrRepositoryNotFound = errors.New("repository not found in workspace")
	ErrRepositoryNotClean = errors.New("repository is not clean")

	// Worktree errors.
	ErrWorktreeExists      = errors.New("worktree already exists")
	ErrWorktreeNotInStatus = errors.New("worktree not found in status file")

	// Directory and file system errors.
	ErrDirectoryExists = errors.New("directory already exists")

	// User interaction errors.
	ErrDeletionCancelled = errors.New("deletion cancelled by user")
)
