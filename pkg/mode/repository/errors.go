// Package repository provides repository management functionality and error definitions.
package repository

import "errors"

// Error definitions for repository package.
var (
	// Git repository errors.
	ErrGitRepositoryNotFound = errors.New("not a valid Git repository: .git directory not found")
	ErrGitRepositoryInvalid  = errors.New("git repository is in an invalid state")
	ErrNotAGitRepository     = errors.New("not a Git repository: .git directory not found")

	// Worktree errors.
	ErrWorktreeExists      = errors.New("worktree already exists")
	ErrWorktreeNotInStatus = errors.New("worktree not found in status file")

	// Repository state errors.
	ErrRepositoryNotClean = errors.New("repository is not clean")
	ErrDirectoryExists    = errors.New("directory already exists")

	// User interaction errors.
	ErrDeletionCancelled = errors.New("deletion cancelled by user")

	// Remote errors.
	ErrOriginRemoteNotFound   = errors.New("origin remote not found or invalid")
	ErrOriginRemoteInvalidURL = errors.New("origin remote URL is not a valid Git hosting service URL")
)
