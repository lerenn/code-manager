// Package repository provides Git repository management functionality for CM.
package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// DefaultRemote is the default remote name used for Git operations.
const DefaultRemote = "origin"

//go:generate mockgen -source=repository.go -destination=mocks/repository.gen.go -package=mocks

// CreateWorktreeOpts contains optional parameters for worktree creation in repository mode.
type CreateWorktreeOpts struct {
	IDEName       string
	IssueInfo     *issue.Info
	WorkspaceName string
	Remote        string // Remote name to use (defaults to DefaultRemote if empty)
}

// WorktreeProvider is a function type that creates worktree instances.
type WorktreeProvider func(params worktree.NewWorktreeParams) worktree.Worktree

// Repository interface provides repository management capabilities.
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

// realRepository represents a single Git repository and provides methods for repository operations.
type realRepository struct {
	fs               fs.FS
	git              git.Git
	config           config.Config
	statusManager    status.Manager
	logger           logger.Logger
	prompt           prompt.Prompter
	worktreeProvider WorktreeProvider
	hookManager      hooks.HookManagerInterface
	repositoryPath   string
}

// NewRepositoryParams contains parameters for creating a new Repository instance.
type NewRepositoryParams struct {
	FS               fs.FS
	Git              git.Git
	Config           config.Config
	StatusManager    status.Manager
	Logger           logger.Logger
	Prompt           prompt.Prompter
	WorktreeProvider WorktreeProvider
	HookManager      hooks.HookManagerInterface
	RepositoryName   string // Name/Path of the repository (optional, defaults to current directory)
}

// NewRepository creates a new Repository instance.
func NewRepository(params NewRepositoryParams) Repository {
	l := params.Logger
	if l == nil {
		l = logger.NewNoopLogger()
	}

	// Resolve repository path from repository name/path
	repoPath, err := resolveRepositoryPath(params.RepositoryName, params.StatusManager, l)
	if err != nil {
		l.Logf("Warning: failed to resolve repository path '%s': %v", params.RepositoryName, err)
		repoPath = "." // Fallback to current directory
	}

	return &realRepository{
		fs:               params.FS,
		git:              params.Git,
		config:           params.Config,
		statusManager:    params.StatusManager,
		logger:           l,
		prompt:           params.Prompt,
		worktreeProvider: params.WorktreeProvider,
		hookManager:      params.HookManager,
		repositoryPath:   repoPath,
	}
}

// CreateWorktreeOpts contains optional parameters for CreateWorktree.
// This is now defined in the mode package for consistency.

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IssueInfo *issue.Info
}

// resolveRepositoryPath resolves a repository name/path to an actual path, checking status file first.
func resolveRepositoryPath(repoName string, statusManager status.Manager, logger logger.Logger) (string, error) {
	// If empty, use current directory
	if repoName == "" {
		return ".", nil
	}

	// First, check if it's a repository name from status.yaml
	if existingRepo, err := statusManager.GetRepository(repoName); err == nil && existingRepo != nil {
		logger.Logf("Resolved repository '%s' from status.yaml: %s", repoName, existingRepo.Path)
		return existingRepo.Path, nil
	}

	// Check if it's an absolute path
	if filepath.IsAbs(repoName) {
		return repoName, nil
	}

	// Resolve relative path from current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	resolvedPath := filepath.Join(currentDir, repoName)
	return resolvedPath, nil
}
