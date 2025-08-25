// Package git provides Git operations and error definitions.
package git

import "errors"

// Git-specific error types.
var (
	ErrGitCommandFailed       = errors.New("git command failed")
	ErrBranchNotFound         = errors.New("branch not found")
	ErrWorktreeExists         = errors.New("worktree already exists")
	ErrWorktreeNotFound       = errors.New("worktree not found")
	ErrWorktreeDeletionFailed = errors.New("failed to delete worktree")
	ErrWorktreePathNotFound   = errors.New("worktree path not found")
	ErrRepositoryNotClean     = errors.New("repository is not clean")
	ErrRemoteOriginNotFound   = errors.New("remote origin not found")
	ErrRemoteNotFound         = errors.New("remote not found")
	ErrRemoteAddFailed        = errors.New("failed to add remote")
	ErrFetchFailed            = errors.New("failed to fetch from remote")
	ErrBranchNotFoundOnRemote = errors.New("branch not found on remote")
	ErrReferenceConflict      = errors.New("reference conflict: cannot create branch due to existing reference")
)
