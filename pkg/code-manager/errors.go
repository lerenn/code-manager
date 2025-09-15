// Package codemanager provides worktree management functionality and error definitions.
package codemanager

import "errors"

// Error definitions for cm package.
var (
	// Git repository errors.
	ErrGitRepositoryNotFound = errors.New("not a valid Git repository: .git directory not found")
	ErrGitRepositoryInvalid  = errors.New("not a valid Git repository")

	// Repository and branch errors.
	ErrRepositoryURLEmpty = errors.New("repository URL cannot be empty")

	// Workspace errors.

	// Worktree creation errors.
	ErrWorktreeExists     = errors.New("worktree already exists for this branch")
	ErrRepositoryNotClean = errors.New("repository is not in a clean state")
	ErrDirectoryExists    = errors.New("worktree directory already exists")

	// Worktree deletion errors.
	ErrWorktreeNotInStatus = errors.New("worktree not found in status file")
	ErrDeletionCancelled   = errors.New("deletion cancelled by user")

	// Load branch errors.
	ErrBranchNameContainsColon  = errors.New("branch name contains invalid character ':'")
	ErrArgumentEmpty            = errors.New("argument cannot be empty")
	ErrOriginRemoteNotFound     = errors.New("origin remote not found or invalid")
	ErrOriginRemoteInvalidURL   = errors.New("origin remote URL is not a valid Git hosting service URL")
	ErrFailedToLoadRepositories = errors.New("failed to load repositories from status file")

	// Initialization errors.
	ErrNotInitialized                = errors.New("CM is not initialized")
	ErrFailedToExpandRepositoriesDir = errors.New("failed to expand repositories directory")

	// Project detection errors.
	ErrNoGitRepositoryOrWorkspaceFound = errors.New("no Git repository or workspace found")
	ErrWorkspaceModeNotSupported       = errors.New("workspace mode not yet supported for load command")

	// Clone errors.
	ErrRepositoryExists               = errors.New("repository already exists")
	ErrUnsupportedRepositoryURLFormat = errors.New("unsupported repository URL format")

	// Clone operation errors.
	ErrFailedToDetectDefaultBranch  = errors.New("failed to detect default branch")
	ErrFailedToCloneRepository      = errors.New("failed to clone repository")
	ErrFailedToInitializeRepository = errors.New("failed to initialize repository in CM")

	// Workspace creation errors.
	ErrInvalidWorkspaceName   = errors.New("invalid workspace name")
	ErrRepositoryNotFound     = errors.New("repository not found")
	ErrInvalidRepository      = errors.New("invalid repository")
	ErrDuplicateRepository    = errors.New("duplicate repository")
	ErrWorkspaceAlreadyExists = errors.New("workspace already exists")
	ErrStatusUpdate           = errors.New("status file update failed")
	ErrRepositoryAddition     = errors.New("failed to add repository to status file")
	ErrPathResolution         = errors.New("path resolution failed")

	// Repository deletion errors.
	ErrInvalidRepositoryName = errors.New("invalid repository name")

	// Workspace deletion errors.
	ErrWorkspaceNotFound = errors.New("workspace not found")
)
