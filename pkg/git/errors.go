// Package git provides Git operations and error definitions.
package git

import "errors"

// Git-specific error types.
var (
	ErrGitCommandFailed     = errors.New("git command failed")
	ErrBranchNotFound       = errors.New("branch not found")
	ErrWorktreeExists       = errors.New("worktree already exists")
	ErrRepositoryNotClean   = errors.New("repository is not clean")
	ErrRemoteOriginNotFound = errors.New("remote origin not found")
)
