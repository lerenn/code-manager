// Package gitcrypt provides git-crypt functionality as a hook for worktree operations.
package gitcrypt

import "errors"

// Git-crypt specific errors.
var (
	// ErrKeyFileNotFound indicates that the git-crypt key file was not found.
	ErrKeyFileNotFound = errors.New("git-crypt key file not found")

	// ErrKeyFileInvalid indicates that the git-crypt key file is invalid or corrupted.
	ErrKeyFileInvalid = errors.New("git-crypt key file is invalid or corrupted")

	// ErrRepositoryPathNotFound indicates that the repository path was not found in the hook context.
	ErrRepositoryPathNotFound = errors.New("repository path not found in hook context")

	// ErrWorktreePathNotFound indicates that the worktree path was not found in the hook context.
	ErrWorktreePathNotFound = errors.New("worktree path not found in hook context")

	// ErrBranchNotFound indicates that the branch was not found in the hook context.
	ErrBranchNotFound = errors.New("branch not found in hook context")
)
