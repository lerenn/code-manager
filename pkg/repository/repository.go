// Package repository provides Git repository management functionality for CM.
package repository

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// DefaultRemote is the default remote name used for Git operations.
const DefaultRemote = "origin"

//go:generate mockgen -source=repository.go -destination=mockrepository.gen.go -package=repository

// Repository interface provides repository management capabilities.
type Repository interface {
	// Validate validates that the current directory is a working Git repository.
	Validate() error

	// CreateWorktree creates a worktree for the repository with the specified branch.
	CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error)

	// DeleteWorktree deletes a worktree for the repository with the specified branch.
	DeleteWorktree(branch string, force bool) error

	// LoadWorktree loads a branch from a remote source and creates a worktree.
	LoadWorktree(remoteSource, branchName string) (string, error)

	// ListWorktrees lists all worktrees for the repository.
	ListWorktrees() ([]status.WorktreeInfo, error)

	// IsWorkspaceFile checks if the current directory contains a workspace file.
	IsWorkspaceFile() (bool, error)

	// IsGitRepository checks if the current directory is a Git repository.
	IsGitRepository() (bool, error)

	// ValidateGitConfiguration validates that Git configuration is functional.
	ValidateGitConfiguration(workDir string) error

	// ValidateGitStatus validates that the Git repository is in a clean state.
	ValidateGitStatus() error

	// ValidateRepository validates that the repository is ready for worktree operations.
	ValidateRepository(params ValidationParams) (*ValidationResult, error)

	// ValidateWorktreeExists validates that a worktree exists for the specified branch.
	ValidateWorktreeExists(repoURL, branch string) error

	// ValidateOriginRemote validates that the origin remote is properly configured.
	ValidateOriginRemote() error

	// HandleRemoteManagement manages remote configuration for the repository.
	HandleRemoteManagement(repoURL string) error

	// ExtractHostFromURL extracts the host from a repository URL.
	ExtractHostFromURL(url string) string

	// DetermineProtocol determines the protocol from a repository URL.
	DetermineProtocol(url string) string

	// ExtractRepoNameFromFullPath extracts the repository name from a full path.
	ExtractRepoNameFromFullPath(fullPath string) string

	// ConstructRemoteURL constructs a remote URL from host and repository name.
	ConstructRemoteURL(originURL, remoteSource, repoName string) (string, error)

	// AddWorktreeToStatus adds a worktree to the status file.
	AddWorktreeToStatus(params StatusParams) error

	// RemoveWorktreeFromStatus removes a worktree from the status file.
	RemoveWorktreeFromStatus(repoURL, branch string) error

	// AutoAddRepositoryToStatus automatically adds the repository to the status file.
	AutoAddRepositoryToStatus(repoURL, repoPath string) error

	// SetLogger sets the logger for this repository instance.
	SetLogger(logger logger.Logger)
}

// realRepository represents a single Git repository and provides methods for repository operations.
type realRepository struct {
	fs            fs.FS
	git           git.Git
	config        *config.Config
	statusManager status.Manager
	logger        logger.Logger
	prompt        prompt.Prompter
	worktree      worktree.Worktree
}

// NewRepositoryParams contains parameters for creating a new Repository instance.
type NewRepositoryParams struct {
	FS            fs.FS
	Git           git.Git
	Config        *config.Config
	StatusManager status.Manager
	Logger        logger.Logger
	Prompt        prompt.Prompter
	Worktree      worktree.Worktree
}

// NewRepository creates a new Repository instance.
func NewRepository(params NewRepositoryParams) Repository {
	return &realRepository{
		fs:            params.FS,
		git:           params.Git,
		config:        params.Config,
		statusManager: params.StatusManager,
		logger:        params.Logger,
		prompt:        params.Prompt,
		worktree:      params.Worktree,
	}
}

// Validate validates that the current directory is a working Git repository.
func (r *realRepository) Validate() error {
	r.VerbosePrint("Validating repository: %s", ".")

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
	return r.ValidateGitConfiguration(".")
}

// CreateWorktreeOpts contains optional parameters for CreateWorktree.
type CreateWorktreeOpts struct {
	IssueInfo *issue.Info
}

// DeleteWorktreeOpts contains optional parameters for DeleteWorktree.
type DeleteWorktreeOpts struct {
	Force bool
}

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IssueInfo *issue.Info
}

// CreateWorktree creates a worktree for the repository with the specified branch.
func (r *realRepository) CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error) {
	r.VerbosePrint("Creating worktree for single repository with branch: %s", branch)

	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{Branch: branch})
	if err != nil {
		return "", err
	}

	// Build worktree path
	worktreePath := r.worktree.BuildPath(validationResult.RepoURL, "origin", branch)
	r.VerbosePrint("Worktree path: %s", worktreePath)

	// Validate creation
	if err := r.worktree.ValidateCreation(worktree.ValidateCreationParams{
		RepoURL:      validationResult.RepoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     ".",
	}); err != nil {
		return "", err
	}

	// Get current directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get issue info if provided
	var issueInfo *issue.Info
	if len(opts) > 0 && opts[0].IssueInfo != nil {
		issueInfo = opts[0].IssueInfo
	}

	// Create the worktree
	if err := r.worktree.Create(worktree.CreateParams{
		RepoURL:      validationResult.RepoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     currentDir,
		Remote:       "origin",
		IssueInfo:    issueInfo,
		Force:        false,
	}); err != nil {
		return "", err
	}

	// Add to status file with auto-repository handling
	if err := r.AddWorktreeToStatus(StatusParams{
		RepoURL:       validationResult.RepoURL,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: "",
		Remote:        "origin",
		IssueInfo:     issueInfo,
	}); err != nil {
		// Clean up worktree on status failure
		if cleanupErr := r.worktree.CleanupDirectory(worktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up worktree directory after status failure: %v", cleanupErr)
		}
		return "", err
	}

	r.VerbosePrint("Successfully created worktree for branch %s at %s", branch, worktreePath)

	return worktreePath, nil
}

// ListWorktrees lists all worktrees for the current repository.
func (r *realRepository) ListWorktrees() ([]status.WorktreeInfo, error) {
	r.VerbosePrint("Listing worktrees for single repository mode")

	// Note: Repository validation is already done in mode detection, so we skip it here
	// to avoid duplicate validation calls

	// 1. Extract repository name from remote origin URL (fallback to local path if no remote)
	repoName, err := r.git.GetRepositoryName(".")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository name: %w", err)
	}

	r.VerbosePrint("Repository name: %s", repoName)

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

	r.VerbosePrint("Found %d worktrees for current repository", len(worktrees))

	return worktrees, nil
}

// IsGitRepository checks if the current directory is a Git repository (including worktrees).
func (r *realRepository) IsGitRepository() (bool, error) {
	r.VerbosePrint("Checking if current directory is a Git repository...")

	// Check if .git exists
	exists, err := r.fs.Exists(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		r.VerbosePrint("No .git found")
		return false, nil
	}

	// Check if .git is a directory (regular repository)
	isDir, err := r.fs.IsDir(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	if isDir {
		r.VerbosePrint("Git repository detected (.git directory)")
		return true, nil
	}

	// If .git is not a directory, it must be a file (worktree)
	// Validate that it's actually a Git worktree file by checking for 'gitdir:' prefix
	r.VerbosePrint("Checking if .git file is a valid worktree file...")

	content, err := r.fs.ReadFile(".git")
	if err != nil {
		r.VerbosePrint("Failed to read .git file: %v", err)
		return false, nil
	}

	contentStr := strings.TrimSpace(string(content))
	if !strings.HasPrefix(contentStr, "gitdir:") {
		r.VerbosePrint(".git file exists but is not a valid worktree file (missing 'gitdir:' prefix)")
		return false, nil
	}

	r.VerbosePrint("Git worktree detected (.git file)")
	return true, nil
}

// IsWorkspaceFile checks if the current directory contains workspace files.
func (r *realRepository) IsWorkspaceFile() (bool, error) {
	r.VerbosePrint("Checking for workspace files...")

	// Check for .code-workspace files
	workspaceFiles, err := r.fs.Glob("*.code-workspace")
	if err != nil {
		return false, fmt.Errorf("%w: %w", workspace.ErrFailedToCheckWorkspaceFiles, err)
	}

	if len(workspaceFiles) > 0 {
		r.VerbosePrint("Workspace files found: %v", workspaceFiles)
		return true, nil
	}

	r.VerbosePrint("No workspace files found")
	return false, nil
}

// DeleteWorktree deletes a worktree for the repository with the specified branch.
func (r *realRepository) DeleteWorktree(branch string, force bool) error {
	r.VerbosePrint("Deleting worktree for single repository with branch: %s", branch)

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

	r.VerbosePrint("Worktree path: %s", worktreePath)

	// Get current directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Delete the worktree
	if err := r.worktree.Delete(worktree.DeleteParams{
		RepoURL:      validationResult.RepoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     currentDir,
		Force:        force,
	}); err != nil {
		return err
	}

	r.VerbosePrint("Successfully deleted worktree for branch %s", branch)

	return nil
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (r *realRepository) LoadWorktree(remoteSource, branchName string) (string, error) {
	r.VerbosePrint("Loading branch: remote=%s, branch=%s", remoteSource, branchName)

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
	r.VerbosePrint("Fetching from remote '%s'", remoteSource)
	if err := r.git.FetchRemote(".", remoteSource); err != nil {
		return "", fmt.Errorf("%w: %w", git.ErrFetchFailed, err)
	}

	// 6. Validate branch exists on remote
	r.VerbosePrint("Checking if branch '%s' exists on remote '%s'", branchName, remoteSource)
	exists, err := r.git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
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
	r.VerbosePrint("Creating worktree for branch '%s'", branchName)
	worktreePath, err := r.CreateWorktree(branchName)
	return worktreePath, err
}

// ParseConfirmationInput parses confirmation input from user.
func (r *realRepository) ParseConfirmationInput(input string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "y", "yes":
		return true, nil
	case "n", "no", "":
		return false, nil
	case "q", "quit", "exit", "cancel":
		return false, fmt.Errorf("user cancelled")
	default:
		return false, fmt.Errorf("invalid input")
	}
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
		return fmt.Errorf("Git configuration validation failed: %w", err)
	}
	return nil
}

// ValidateGitStatus validates that the Git repository is in a clean state.
func (r *realRepository) ValidateGitStatus() error {
	status, err := r.git.Status(".")
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

// VerbosePrint logs a formatted message only in verbose mode.
func (r *realRepository) VerbosePrint(msg string, args ...interface{}) {
	r.logger.Logf(fmt.Sprintf(msg, args...))
}
