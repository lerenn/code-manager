// Package status provides status management functionality and error definitions.
package status

import "errors"

// Error definitions for status package.
var (
	// Worktree management errors.
	ErrWorktreeAlreadyExists       = errors.New("worktree already exists")
	ErrWorktreeNotFound            = errors.New("worktree not found")
	ErrConfigurationNotInitialized = errors.New("configuration is not initialized")
	ErrNotInitialized              = errors.New("CM is not initialized. Run 'cm init' to initialize")

	// Repository management errors.
	ErrRepositoryNotFound      = errors.New("repository not found in status")
	ErrRepositoryAlreadyExists = errors.New("repository already exists")

	// Workspace management errors.
	ErrWorkspaceNotFound      = errors.New("workspace not found in status")
	ErrWorkspaceAlreadyExists = errors.New("workspace already exists")

	// Status file errors.
	ErrStatusFileParse = errors.New("failed to parse status file")
)
