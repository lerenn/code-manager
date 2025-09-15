// Package interfaces defines the worktree interface for dependency injection.
// This package is separate to avoid circular imports.
//
//nolint:revive // interfaces package name is intentional to avoid circular dependencies
package interfaces

//go:generate go run go.uber.org/mock/mockgen@latest -source=interfaces.go -destination=../mocks/worktree.gen.go -package=mocks

import (
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
)

// Worktree defines the interface for worktree operations.
// This interface is implemented by the concrete worktree types.
type Worktree interface {
	// BuildPath constructs a worktree path from repository URL, remote name, and branch.
	BuildPath(repoURL, remoteName, branch string) string

	// Create creates a new worktree with proper validation and cleanup.
	Create(params CreateParams) error

	// CheckoutBranch checks out the branch in the worktree after hooks have been executed.
	CheckoutBranch(worktreePath, branch string) error

	// Delete deletes a worktree with proper cleanup and confirmation.
	Delete(params DeleteParams) error

	// ValidateCreation validates that worktree creation is possible.
	ValidateCreation(params ValidateCreationParams) error

	// ValidateDeletion validates that worktree deletion is possible.
	ValidateDeletion(params ValidateDeletionParams) error

	// EnsureBranchExists ensures the specified branch exists, creating it if necessary.
	EnsureBranchExists(repoPath, branch string) error

	// AddToStatus adds the worktree to the status file.
	AddToStatus(params AddToStatusParams) error

	// RemoveFromStatus removes the worktree from the status file.
	RemoveFromStatus(repoURL, branch string) error

	// CleanupDirectory removes the worktree directory.
	CleanupDirectory(worktreePath string) error

	// Exists checks if a worktree exists for the specified branch.
	Exists(repoPath, branch string) (bool, error)

	// SetLogger sets the logger for this worktree instance.
	SetLogger(logger logger.Logger)
}

// CreateParams contains parameters for worktree creation.
type CreateParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	RepoPath     string
	Remote       string
	IssueInfo    *issue.Info
	Force        bool
}

// DeleteParams contains parameters for worktree deletion.
type DeleteParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	RepoPath     string
	Force        bool
}

// ValidateCreationParams contains parameters for worktree creation validation.
type ValidateCreationParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	RepoPath     string
}

// ValidateDeletionParams contains parameters for worktree deletion validation.
type ValidateDeletionParams struct {
	RepoURL string
	Branch  string
}

// AddToStatusParams contains parameters for adding worktree to status.
type AddToStatusParams struct {
	RepoURL       string
	Branch        string
	WorktreePath  string
	WorkspacePath string
	Remote        string
	IssueInfo     *issue.Info
}

// WorktreeProvider defines the function signature for creating worktree instances.
type WorktreeProvider func(params NewWorktreeParams) Worktree

// NewWorktreeParams contains parameters for creating a new Worktree instance.
type NewWorktreeParams struct {
	FS              interface{} // fs.FS
	Git             interface{} // git.Git
	StatusManager   interface{} // status.Manager
	Logger          interface{} // logger.Logger
	Prompt          interface{} // prompt.Prompter
	RepositoriesDir string
}
