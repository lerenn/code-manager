// Package git provides Git operations and error definitions.
package git

import "errors"

// Git-specific error types.
var (
	ErrBranchNotFound         = errors.New("branch not found")
	ErrWorktreeExists         = errors.New("worktree already exists")
	ErrWorktreeNotFound       = errors.New("worktree not found")
	ErrWorktreePathNotFound   = errors.New("worktree path not found")
	ErrRepositoryNotClean     = errors.New("repository is not clean")
	ErrRemoteAddFailed        = errors.New("failed to add remote")
	ErrFetchFailed            = errors.New("failed to fetch from remote")
	ErrBranchNotFoundOnRemote = errors.New("branch not found on remote")

	// Specific reference conflict error types for testing.
	ErrBranchParentExists = errors.New("cannot create branch: reference already exists")
	ErrTagParentExists    = errors.New("cannot create branch: tag already exists")
)
