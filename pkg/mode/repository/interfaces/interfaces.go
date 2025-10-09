// Package interfaces defines the repository interface for dependency injection.
// This package is separate to avoid circular imports.
//
//nolint:revive // interfaces package name is intentional to avoid circular dependencies
package interfaces

//go:generate go run go.uber.org/mock/mockgen@latest -source=interfaces.go -destination=../mocks/repository.gen.go -package=mocks

import (
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/status"
)

// CreateWorktreeOpts contains optional parameters for worktree creation in repository mode.
type CreateWorktreeOpts struct {
	IDEName       string
	IssueInfo     *issue.Info
	WorkspaceName string
	Remote        string // Remote name to use (defaults to DefaultRemote if empty)
}

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IssueInfo *issue.Info
}

// ValidationParams contains parameters for repository validation.
type ValidationParams struct {
	CurrentDir string
	Branch     string
}

// ValidationResult contains the result of repository validation.
type ValidationResult struct {
	RepoURL  string
	RepoPath string
}

// StatusParams contains parameters for status operations.
type StatusParams struct {
	RepoURL       string
	Branch        string
	WorktreePath  string
	WorkspacePath string
	Remote        string
	IssueInfo     *issue.Info
	Detached      bool
}

// Repository defines the interface for repository operations.
// This interface is implemented by the concrete repository types.
type Repository interface {
	Validate() error
	CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error)
	DeleteWorktree(branch string, force bool) error
	DeleteAllWorktrees(force bool) error
	ListWorktrees() ([]status.WorktreeInfo, error)
	LoadWorktree(remoteSource, branchName string) (string, error)
	IsGitRepository() (bool, error)
	ValidateGitConfiguration(workDir string) error
	ValidateGitStatus() error
	ValidateRepository(params ValidationParams) (*ValidationResult, error)
	ValidateWorktreeExists(repoURL, branch string) error
	ValidateOriginRemote() error
	HandleRemoteManagement(repoURL string) error
	ExtractHostFromURL(url string) string
	DetermineProtocol(url string) string
	ExtractRepoNameFromFullPath(fullPath string) string
	ConstructRemoteURL(originURL, remoteSource, repoName string) (string, error)
	AddWorktreeToStatus(params StatusParams) error
	AutoAddRepositoryToStatus(repoURL, repoPath string) error
}

// RepositoryProvider defines the function signature for creating repository instances.
type RepositoryProvider func(params NewRepositoryParams) Repository

// NewRepositoryParams contains parameters for creating a new Repository instance.
type NewRepositoryParams struct {
	Dependencies   interface{} // *dependencies.Dependencies
	RepositoryName string      // Name/Path of the repository (optional, defaults to current directory)
}
