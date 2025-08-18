// Package cm provides worktree management functionality and error definitions.
package cm

import "errors"

// Error definitions for CM package.
var (
	// Git repository errors.
	ErrGitRepositoryNotFound     = errors.New("not a valid Git repository: .git directory not found")
	ErrGitRepositoryNotDirectory = errors.New("not a valid Git repository: .git exists but is not a directory")
	ErrGitRepositoryInvalid      = errors.New("not a valid Git repository")

	// Repository and branch errors.
	ErrRepositoryURLEmpty                   = errors.New("repository URL cannot be empty")
	ErrBranchNameEmpty                      = errors.New("branch name cannot be empty")
	ErrRepositoryNameEmptyAfterSanitization = errors.New("repository name is empty after sanitization")
	ErrBranchNameEmptyAfterSanitization     = errors.New("branch name is empty after sanitization")

	// Configuration errors.
	ErrConfigurationNotInitialized = errors.New("configuration is not initialized")

	// Workspace errors.
	ErrWorkspaceFileMalformed            = errors.New("invalid .code-workspace file: malformed JSON")
	ErrRepositoryNotFoundInWorkspace     = errors.New("repository not found in workspace")
	ErrInvalidRepositoryInWorkspace      = errors.New("invalid repository in workspace")
	ErrInvalidRepositoryInWorkspaceNoGit = errors.New("invalid repository in workspace - .git directory not found")
	ErrMultipleWorkspaces                = errors.New("failed to handle multiple workspaces")
	ErrWorkspaceFileRead                 = errors.New("failed to parse workspace file")
	ErrWorkspaceFileReadError            = errors.New("failed to read workspace file")
	ErrWorkspaceDetection                = errors.New("failed to detect workspace mode")
	ErrWorkspaceEmptyFolders             = errors.New("workspace file must contain non-empty folders array")

	// Worktree creation errors.
	ErrWorktreeExists     = errors.New("worktree already exists for this branch")
	ErrRepositoryNotClean = errors.New("repository is not in a clean state")
	ErrDirectoryExists    = errors.New("worktree directory already exists")

	// Worktree deletion errors.
	ErrWorktreeNotInStatus      = errors.New("worktree not found in status file")
	ErrDeletionCancelled        = errors.New("deletion cancelled by user")
	ErrWorktreeValidationFailed = errors.New("worktree validation failed")

	// Load branch errors.
	ErrInvalidArgumentFormat   = errors.New("invalid argument format")
	ErrEmptyRemoteSource       = errors.New("empty remote source")
	ErrEmptyBranchName         = errors.New("empty branch name")
	ErrBranchNameContainsColon = errors.New("branch name contains invalid character ':'")
	ErrOriginRemoteNotFound    = errors.New("origin remote not found or invalid")
	ErrOriginRemoteInvalidURL  = errors.New("origin remote URL is not a valid Git hosting service URL")
)
