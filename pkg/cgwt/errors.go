package cgwt

import "errors"

// Error definitions for CGWT package.
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

	// Status management errors.
	ErrAddWorktreeToStatus      = errors.New("failed to add worktree to status")
	ErrRemoveWorktreeFromStatus = errors.New("failed to remove worktree from status")
	ErrGetWorktreeStatus        = errors.New("failed to get worktree status")
	ErrListWorktrees            = errors.New("failed to list worktrees")
)
