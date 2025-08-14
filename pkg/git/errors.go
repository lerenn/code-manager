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
)
