// Package repository provides Git repository management functionality for CM.
package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	SetLogger(logger logger.Logger)
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

// Validate validates that the current directory is a working Git repository.
func (r *realRepository) Validate() error {
	r.logger.Logf("Validating repository: %s", r.repositoryPath)

	// Check if we're in a Git repository
	exists, err := r.IsGitRepository()
	if err != nil {
		return err
	}
	if !exists {
		return ErrGitRepositoryNotFound
	}

	if err := r.ValidateGitStatus(); err != nil {
		return err
	}

	// Validate Git configuration is functional
	return r.ValidateGitConfiguration(r.repositoryPath)
}

// CreateWorktreeOpts contains optional parameters for CreateWorktree.
// This is now defined in the mode package for consistency.

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IssueInfo *issue.Info
}

// CreateWorktree creates a worktree for the repository with the specified branch.
func (r *realRepository) CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error) {
	r.logger.Logf("Creating worktree for single repository with branch: %s", branch)

	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{Branch: branch})
	if err != nil {
		return "", err
	}

	// Create and validate worktree instance
	worktreeInstance, worktreePath, err := r.createAndValidateWorktreeInstance(validationResult.RepoURL, branch)
	if err != nil {
		return "", err
	}

	// Get current directory
	currentDir, err := filepath.Abs(r.repositoryPath)
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get issue info if provided
	var issueInfo *issue.Info
	if len(opts) > 0 && opts[0].IssueInfo != nil {
		issueInfo = opts[0].IssueInfo
	}

	// Create the worktree (with --no-checkout)
	if err := r.createWorktreeWithNoCheckout(
		worktreeInstance, validationResult.RepoURL, branch, worktreePath, currentDir, issueInfo,
	); err != nil {
		return "", err
	}

	// Execute worktree checkout hooks (for git-crypt setup, etc.)
	if err := r.executeWorktreeCheckoutHooks(
		worktreeInstance, worktreePath, branch, currentDir, validationResult.RepoURL,
	); err != nil {
		return "", err
	}

	// Now checkout the branch
	if err := r.checkoutBranchInWorktree(worktreeInstance, worktreePath, branch); err != nil {
		return "", err
	}

	// Add to status file with auto-repository handling
	if err := r.addWorktreeToStatusAndHandleCleanup(
		worktreeInstance, validationResult.RepoURL, branch, worktreePath, issueInfo,
	); err != nil {
		return "", err
	}

	r.logger.Logf("Successfully created worktree for branch %s at %s", branch, worktreePath)

	return worktreePath, nil
}

// executeWorktreeCheckoutHooks executes worktree checkout hooks with proper error handling.
func (r *realRepository) executeWorktreeCheckoutHooks(
	worktreeInstance worktree.Worktree,
	worktreePath, branch, currentDir, repoURL string,
) error {
	if r.hookManager == nil {
		return nil
	}

	ctx := &hooks.HookContext{
		OperationName: "CreateWorkTree",
		Parameters: map[string]interface{}{
			"worktreePath": worktreePath,
			"branch":       branch,
			"repoPath":     currentDir,
			"repoURL":      repoURL,
		},
		Results:  make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	if err := r.hookManager.ExecuteWorktreeCheckoutHooks("CreateWorkTree", ctx); err != nil {
		// Cleanup failed worktree
		if cleanupErr := worktreeInstance.CleanupDirectory(worktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up worktree directory after hook failure: %v", cleanupErr)
		}
		return fmt.Errorf("worktree checkout hooks failed: %w", err)
	}

	return nil
}

// createAndValidateWorktreeInstance creates and validates a worktree instance.
func (r *realRepository) createAndValidateWorktreeInstance(repoURL, branch string) (worktree.Worktree, string, error) {
	// Create worktree instance using provider
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: r.config.RepositoriesDir,
	})

	// Build worktree path
	worktreePath := worktreeInstance.BuildPath(repoURL, "origin", branch)
	r.logger.Logf("Worktree path: %s", worktreePath)

	// Validate creation
	if err := worktreeInstance.ValidateCreation(worktree.ValidateCreationParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     r.repositoryPath,
	}); err != nil {
		return nil, "", err
	}

	return worktreeInstance, worktreePath, nil
}

// createWorktreeWithNoCheckout creates the worktree with --no-checkout flag.
func (r *realRepository) createWorktreeWithNoCheckout(
	worktreeInstance worktree.Worktree,
	repoURL, branch, worktreePath, currentDir string,
	issueInfo *issue.Info,
) error {
	return worktreeInstance.Create(worktree.CreateParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     currentDir,
		Remote:       "origin",
		IssueInfo:    issueInfo,
		Force:        false,
	})
}

// checkoutBranchInWorktree checks out the branch in the worktree with proper error handling.
func (r *realRepository) checkoutBranchInWorktree(
	worktreeInstance worktree.Worktree,
	worktreePath, branch string,
) error {
	if err := worktreeInstance.CheckoutBranch(worktreePath, branch); err != nil {
		// Cleanup failed worktree
		if cleanupErr := worktreeInstance.CleanupDirectory(worktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up worktree directory after checkout failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to checkout branch in worktree: %w", err)
	}
	return nil
}

// addWorktreeToStatusAndHandleCleanup adds the worktree to status and handles cleanup on failure.
func (r *realRepository) addWorktreeToStatusAndHandleCleanup(
	worktreeInstance worktree.Worktree,
	repoURL string,
	branch string,
	worktreePath string,
	issueInfo *issue.Info,
) error {
	if err := r.AddWorktreeToStatus(StatusParams{
		RepoURL:       repoURL,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: "",
		Remote:        "origin",
		IssueInfo:     issueInfo,
	}); err != nil {
		// Clean up worktree on status failure
		if cleanupErr := worktreeInstance.CleanupDirectory(worktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up worktree directory after status failure: %v", cleanupErr)
		}
		return err
	}
	return nil
}

// ListWorktrees lists all worktrees for the current repository.
func (r *realRepository) ListWorktrees() ([]status.WorktreeInfo, error) {
	r.logger.Logf("Listing worktrees for single repository mode")

	// Note: Repository validation is already done in mode detection, so we skip it here
	// to avoid duplicate validation calls

	// 1. Extract repository name from remote origin URL (fallback to local path if no remote)
	repoName, err := r.git.GetRepositoryName(r.repositoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository name: %w", err)
	}

	r.logger.Logf("Repository name: %s", repoName)

	// 2. Get repository from status file
	repo, err := r.statusManager.GetRepository(repoName)
	if err != nil {
		// If repository not found, return empty list
		// But propagate other errors (like status file corruption)
		if errors.Is(err, status.ErrRepositoryNotFound) {
			return []status.WorktreeInfo{}, nil
		}
		return nil, err
	}

	// 3. Convert repository worktrees to WorktreeInfo slice
	var worktrees []status.WorktreeInfo
	for _, worktree := range repo.Worktrees {
		worktrees = append(worktrees, worktree)
	}

	r.logger.Logf("Found %d worktrees for current repository", len(worktrees))

	return worktrees, nil
}

// IsGitRepository checks if the current directory is a Git repository (including worktrees).
func (r *realRepository) IsGitRepository() (bool, error) {
	r.logger.Logf("Checking if directory %s is a Git repository...", r.repositoryPath)

	// Check if .git exists
	gitPath := filepath.Join(r.repositoryPath, ".git")
	exists, err := r.fs.Exists(gitPath)
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		r.logger.Logf("No .git found")
		return false, nil
	}

	// Check if .git is a directory (regular repository)
	isDir, err := r.fs.IsDir(gitPath)
	if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	if isDir {
		r.logger.Logf("Git repository detected (.git directory)")
		return true, nil
	}

	// If .git is not a directory, it must be a file (worktree)
	// Validate that it's actually a Git worktree file by checking for 'gitdir:' prefix
	r.logger.Logf("Checking if .git file is a valid worktree file...")

	content, err := r.fs.ReadFile(gitPath)
	if err != nil {
		r.logger.Logf("Failed to read .git file: %v", err)
		return false, nil
	}

	contentStr := strings.TrimSpace(string(content))
	if !strings.HasPrefix(contentStr, "gitdir:") {
		r.logger.Logf(".git file exists but is not a valid worktree file (missing 'gitdir:' prefix)")
		return false, nil
	}

	r.logger.Logf("Git worktree detected (.git file)")
	return true, nil
}

// DeleteWorktree deletes a worktree for the repository with the specified branch.
func (r *realRepository) DeleteWorktree(branch string, force bool) error {
	r.logger.Logf("Deleting worktree for single repository with branch: %s", branch)

	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{})
	if err != nil {
		return err
	}

	// Check if worktree exists in status file
	if err := r.ValidateWorktreeExists(validationResult.RepoURL, branch); err != nil {
		return err
	}

	// Get worktree path from Git
	worktreePath, err := r.git.GetWorktreePath(validationResult.RepoPath, branch)
	if err != nil {
		return fmt.Errorf("failed to get worktree path: %w", err)
	}

	r.logger.Logf("Worktree path: %s", worktreePath)

	// Get current directory
	currentDir, err := filepath.Abs(r.repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create worktree instance using provider
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: r.config.RepositoriesDir,
	})

	// Delete the worktree
	if err := worktreeInstance.Delete(worktree.DeleteParams{
		RepoURL:      validationResult.RepoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     currentDir,
		Force:        force,
	}); err != nil {
		return err
	}

	r.logger.Logf("Successfully deleted worktree for branch %s", branch)

	return nil
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (r *realRepository) LoadWorktree(remoteSource, branchName string) (string, error) {
	r.logger.Logf("Loading branch: remote=%s, branch=%s", remoteSource, branchName)

	// 1. Validate current directory is a Git repository
	gitExists, err := r.IsGitRepository()
	if err != nil {
		return "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !gitExists {
		return "", ErrGitRepositoryNotFound
	}

	// 2. Validate origin remote exists and is a valid Git hosting service URL
	if err := r.ValidateOriginRemote(); err != nil {
		return "", err
	}

	// 3. Parse remote source (default to "origin" if not specified)
	if remoteSource == "" {
		remoteSource = DefaultRemote
	}

	// 4. Handle remote management
	if err := r.HandleRemoteManagement(remoteSource); err != nil {
		return "", err
	}

	// 5. Fetch from the remote
	r.logger.Logf("Fetching from remote '%s'", remoteSource)
	if err := r.git.FetchRemote(r.repositoryPath, remoteSource); err != nil {
		return "", fmt.Errorf("%w: %w", git.ErrFetchFailed, err)
	}

	// 6. Validate branch exists on remote
	r.logger.Logf("Checking if branch '%s' exists on remote '%s'", branchName, remoteSource)
	exists, err := r.git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   r.repositoryPath,
		RemoteName: remoteSource,
		Branch:     branchName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf(
			"%w: branch '%s' not found on remote '%s'",
			git.ErrBranchNotFoundOnRemote,
			branchName,
			remoteSource,
		)
	}

	// 7. Create worktree for the branch (using existing worktree creation logic directly)
	r.logger.Logf("Creating worktree for branch '%s'", branchName)
	worktreePath, err := r.CreateWorktree(branchName)
	return worktreePath, err
}

// SetLogger sets the logger for this repository instance.
func (r *realRepository) SetLogger(logger logger.Logger) {
	r.logger = logger
}

// ValidateGitConfiguration validates that Git configuration is functional.
func (r *realRepository) ValidateGitConfiguration(workDir string) error {
	// Check if Git is available and functional by running a simple command
	_, err := r.git.GetCurrentBranch(workDir)
	if err != nil {
		return fmt.Errorf("git configuration validation failed: %w", err)
	}
	return nil
}

// ValidateGitStatus validates that the Git repository is in a clean state.
func (r *realRepository) ValidateGitStatus() error {
	status, err := r.git.Status(r.repositoryPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	// Check if the status indicates a clean repository
	// This is a simple check - in practice you might want more sophisticated parsing
	if status == "" {
		return fmt.Errorf("%w: empty git status", ErrGitRepositoryInvalid)
	}

	return nil
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
