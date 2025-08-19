// Package repository provides Git repository management functionality for CM.
package repository

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/cm/internal/base"
	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/issue"
	"github.com/lerenn/cm/pkg/logger"
	"github.com/lerenn/cm/pkg/prompt"
	"github.com/lerenn/cm/pkg/status"
)

// DefaultRemote is the default remote name used for Git operations.
const DefaultRemote = "origin"

// Error definitions.
var (
	ErrGitRepositoryNotFound  = errors.New("not a valid Git repository: .git directory not found")
	ErrGitRepositoryInvalid   = errors.New("git repository is in an invalid state")
	ErrWorktreeExists         = errors.New("worktree already exists")
	ErrRepositoryNotClean     = errors.New("repository is not clean")
	ErrDirectoryExists        = errors.New("directory already exists")
	ErrWorktreeNotInStatus    = errors.New("worktree not found in status file")
	ErrDeletionCancelled      = errors.New("deletion cancelled by user")
	ErrOriginRemoteNotFound   = errors.New("origin remote not found or invalid")
	ErrOriginRemoteInvalidURL = errors.New("origin remote URL is not a valid Git hosting service URL")
)

// Repository represents a single Git repository and provides methods for repository operations.
type Repository struct {
	*base.Base
}

// NewRepositoryParams contains parameters for creating a new Repository instance.
type NewRepositoryParams struct {
	FS            fs.FS
	Git           git.Git
	Config        *config.Config
	StatusManager status.Manager
	Logger        logger.Logger
	Prompt        prompt.Prompt
	Verbose       bool
}

// NewRepository creates a new Repository instance.
func NewRepository(params NewRepositoryParams) *Repository {
	return &Repository{
		Base: base.NewBase(base.NewBaseParams{
			FS:            params.FS,
			Git:           params.Git,
			Config:        params.Config,
			StatusManager: params.StatusManager,
			Logger:        params.Logger,
			Prompt:        params.Prompt,
			Verbose:       params.Verbose,
		}),
	}
}

// Validate validates that the current directory is a working Git repository.
func (r *Repository) Validate() error {
	r.VerbosePrint("Validating repository: %s", ".")

	// Check if we're in a Git repository
	exists, err := r.IsGitRepository()
	if err != nil {
		return err
	}
	if !exists {
		return ErrGitRepositoryNotFound
	}

	if err := r.validateGitStatus(); err != nil {
		return err
	}

	// Validate Git configuration is functional
	return r.ValidateGitConfiguration(".")
}

// CreateWorktreeOpts contains optional parameters for CreateWorktree.
type CreateWorktreeOpts struct {
	IssueInfo *issue.Info
}

// CreateWorktree creates a worktree for the repository with the specified branch.
func (r *Repository) CreateWorktree(branch string, opts ...CreateWorktreeOpts) error {
	r.VerbosePrint("Creating worktree for single repository with branch: %s", branch)

	// Validate and prepare repository
	repoURL, worktreePath, err := r.prepareWorktreeCreation(branch)
	if err != nil {
		return err
	}

	// Create the worktree
	var issueInfo *issue.Info
	if len(opts) > 0 && opts[0].IssueInfo != nil {
		issueInfo = opts[0].IssueInfo
	}
	if err := r.executeWorktreeCreation(ExecuteWorktreeCreationParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		IssueInfo:    issueInfo,
	}); err != nil {
		return err
	}

	// Issue information is now stored in status file instead of creating initial commits

	r.VerbosePrint("Successfully created worktree for branch %s at %s", branch, worktreePath)

	return nil
}

// ListWorktrees lists all worktrees for the current repository.
func (r *Repository) ListWorktrees() ([]status.Repository, error) {
	r.VerbosePrint("Listing worktrees for single repository mode")

	// Note: Repository validation is already done in mode detection, so we skip it here
	// to avoid duplicate validation calls

	// 1. Extract repository name from remote origin URL (fallback to local path if no remote)
	repoName, err := r.Git.GetRepositoryName(".")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository name: %w", err)
	}

	r.VerbosePrint("Repository name: %s", repoName)

	// 2. Load all worktrees from status file
	allWorktrees, err := r.StatusManager.ListAllWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to load worktrees from status file: %w", err)
	}

	r.VerbosePrint("Found %d total worktrees in status file", len(allWorktrees))

	// 3. Filter worktrees to only include those for the current repository and add remote information
	var filteredWorktrees []status.Repository
	for _, worktree := range allWorktrees {
		if worktree.URL == repoName {
			// Get the remote for this branch
			remote, err := r.Git.GetBranchRemote(".", worktree.Branch)
			if err != nil {
				// If we can't determine the remote, use "origin" as default
				remote = DefaultRemote
			}

			// Create a copy with remote information
			worktreeWithRemote := worktree
			worktreeWithRemote.Remote = remote
			filteredWorktrees = append(filteredWorktrees, worktreeWithRemote)
		}
	}

	r.VerbosePrint("Found %d worktrees for current repository", len(filteredWorktrees))

	return filteredWorktrees, nil
}

// IsGitRepository checks if the current directory is a Git repository (including worktrees).
func (r *Repository) IsGitRepository() (bool, error) {
	r.VerbosePrint("Checking if current directory is a Git repository...")

	// Check if .git exists
	exists, err := r.FS.Exists(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		r.VerbosePrint("No .git found")
		return false, nil
	}

	// Check if .git is a directory (regular repository)
	isDir, err := r.FS.IsDir(".git")
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

	content, err := r.FS.ReadFile(".git")
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
func (r *Repository) IsWorkspaceFile() (bool, error) {
	r.VerbosePrint("Checking for workspace files...")

	// Check for .code-workspace files
	workspaceFiles, err := r.FS.Glob("*.code-workspace")
	if err != nil {
		return false, fmt.Errorf("failed to check for workspace files: %w", err)
	}

	if len(workspaceFiles) > 0 {
		r.VerbosePrint("Workspace files found: %v", workspaceFiles)
		return true, nil
	}

	r.VerbosePrint("No workspace files found")
	return false, nil
}

// validateGitStatus validates that git status works in the repository.
func (r *Repository) validateGitStatus() error {
	// Execute git status to ensure repository is working
	r.VerbosePrint("Executing git status in: %s", ".")
	_, err := r.Git.Status(".")
	if err != nil {
		r.VerbosePrint("Error: %v", err)
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	return nil
}

// prepareWorktreeCreation validates the repository and prepares the worktree path.
func (r *Repository) prepareWorktreeCreation(branch string) (string, string, error) {
	// Validate repository
	repoURL, err := r.validateRepository(branch)
	if err != nil {
		return "", "", err
	}

	// Prepare worktree path
	worktreePath, err := r.prepareWorktreePath(repoURL, branch)
	if err != nil {
		return "", "", err
	}

	return repoURL, worktreePath, nil
}

// validateRepository validates the repository and gets the repository name.
func (r *Repository) validateRepository(branch string) (string, error) {
	// Get current working directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Validate that we're in a Git repository
	isSingleRepo, err := r.IsGitRepository()
	if err != nil {
		return "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get repository URL from remote origin URL with fallback to local path
	repoURL, err := r.Git.GetRepositoryName(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	r.VerbosePrint("Repository URL: %s", repoURL)

	// Check if worktree already exists in status file
	existingWorktree, err := r.StatusManager.GetWorktree(repoURL, branch)
	if err == nil && existingWorktree != nil {
		return "", fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, repoURL, branch)
	}

	// Validate repository state (placeholder for future validation)
	isClean, err := r.Git.IsClean(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to check repository state: %w", err)
	}
	if !isClean {
		return "", fmt.Errorf("%w: repository is not in a clean state", ErrRepositoryNotClean)
	}

	return repoURL, nil
}

// prepareWorktreePath prepares the worktree directory path.
func (r *Repository) prepareWorktreePath(repoURL, branch string) (string, error) {
	// Create worktree directory path
	worktreePath := r.BuildWorktreePath(repoURL, branch)

	r.VerbosePrint("Worktree path: %s", worktreePath)

	// Check if worktree directory already exists
	exists, err := r.FS.Exists(worktreePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return "", fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, worktreePath)
	}

	// Create worktree directory structure
	if err := r.FS.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return worktreePath, nil
}

// ExecuteWorktreeCreationParams contains parameters for executing worktree creation.
type ExecuteWorktreeCreationParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	IssueInfo    *issue.Info
}

// executeWorktreeCreation creates the branch and worktree.
func (r *Repository) executeWorktreeCreation(params ExecuteWorktreeCreationParams) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Ensure branch exists
	if err := r.ensureBranchExists(currentDir, params.Branch); err != nil {
		return err
	}

	// Create worktree with cleanup
	if err := r.createWorktreeWithCleanup(createWorktreeWithCleanupParams{
		RepoURL:      params.RepoURL,
		Branch:       params.Branch,
		WorktreePath: params.WorktreePath,
		CurrentDir:   currentDir,
		IssueInfo:    params.IssueInfo,
	}); err != nil {
		return err
	}

	return nil
}

// ensureBranchExists ensures the branch exists, creating it if necessary.
func (r *Repository) ensureBranchExists(currentDir, branch string) error {
	branchExists, err := r.Git.BranchExists(currentDir, branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if !branchExists {
		r.VerbosePrint("Branch %s does not exist, creating from current branch", branch)
		if err := r.Git.CreateBranch(currentDir, branch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}
	}

	return nil
}

// createWorktreeWithCleanupParams contains parameters for createWorktreeWithCleanup.
type createWorktreeWithCleanupParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	CurrentDir   string
	IssueInfo    *issue.Info
}

// createWorktreeWithCleanup creates the worktree with proper cleanup on failure.
func (r *Repository) createWorktreeWithCleanup(params createWorktreeWithCleanupParams) error {
	// Update status file with worktree entry (before creating the worktree for proper cleanup)
	// Store the original repository path, not the worktree path
	if err := r.StatusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:       params.RepoURL,
		Branch:        params.Branch,
		WorktreePath:  params.CurrentDir,
		WorkspacePath: "",
		IssueInfo:     params.IssueInfo,
	}); err != nil {
		// Clean up created directory on status update failure
		if cleanupErr := r.CleanupWorktreeDirectory(params.WorktreePath); cleanupErr != nil {
			r.Logger.Logf("Warning: failed to clean up directory after status update failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	// Create the Git worktree
	if err := r.Git.CreateWorktree(params.CurrentDir, params.WorktreePath, params.Branch); err != nil {
		// Clean up on worktree creation failure
		if cleanupErr := r.StatusManager.RemoveWorktree(params.RepoURL, params.Branch); cleanupErr != nil {
			r.Logger.Logf("Warning: failed to remove worktree from status after creation failure: %v", cleanupErr)
		}
		if cleanupErr := r.CleanupWorktreeDirectory(params.WorktreePath); cleanupErr != nil {
			r.Logger.Logf("Warning: failed to clean up directory after worktree creation failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	return nil
}

// DeleteWorktree deletes a worktree for the repository with the specified branch.
func (r *Repository) DeleteWorktree(branch string, force bool) error {
	r.VerbosePrint("Deleting worktree for single repository with branch: %s", branch)

	// Validate and prepare worktree deletion
	repoURL, worktreePath, err := r.prepareWorktreeDeletion(branch)
	if err != nil {
		return err
	}

	// Execute the deletion
	if err := r.executeWorktreeDeletion(executeWorktreeDeletionParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		Force:        force,
	}); err != nil {
		return err
	}

	r.VerbosePrint("Successfully deleted worktree for branch %s", branch)

	return nil
}

// prepareWorktreeDeletion validates the repository and prepares the worktree deletion.
func (r *Repository) prepareWorktreeDeletion(branch string) (string, string, error) {
	// Get current working directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Validate that we're in a Git repository
	isSingleRepo, err := r.IsGitRepository()
	if err != nil {
		return "", "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return "", "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get repository URL from remote origin URL with fallback to local path
	repoURL, err := r.Git.GetRepositoryName(currentDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	r.VerbosePrint("Repository URL: %s", repoURL)

	// Check if worktree exists in status file
	existingWorktree, err := r.StatusManager.GetWorktree(repoURL, branch)
	if err != nil || existingWorktree == nil {
		return "", "", fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotInStatus, repoURL, branch)
	}

	// Get worktree path from Git
	worktreePath, err := r.Git.GetWorktreePath(currentDir, branch)
	if err != nil {
		return "", "", fmt.Errorf("failed to get worktree path: %w", err)
	}

	r.VerbosePrint("Worktree path: %s", worktreePath)

	return repoURL, worktreePath, nil
}

// executeWorktreeDeletionParams contains parameters for executeWorktreeDeletion.
type executeWorktreeDeletionParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	Force        bool
}

// executeWorktreeDeletion deletes the worktree with proper cleanup.
func (r *Repository) executeWorktreeDeletion(params executeWorktreeDeletionParams) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Prompt for confirmation unless force flag is used
	if !params.Force {
		if err := r.promptForConfirmation(params.Branch, params.WorktreePath); err != nil {
			return err
		}
	}

	// Remove worktree from Git tracking first
	if err := r.Git.RemoveWorktree(currentDir, params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree from Git: %w", err)
	}

	// Remove worktree directory
	if err := r.FS.RemoveAll(params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Remove entry from status file
	if err := r.StatusManager.RemoveWorktree(params.RepoURL, params.Branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status: %w", err)
	}

	return nil
}

// promptForConfirmation prompts the user for confirmation before deletion.
func (r *Repository) promptForConfirmation(branch, worktreePath string) error {
	fmt.Printf("You are about to delete the worktree for branch '%s'\n", branch)
	fmt.Printf("Worktree path: %s\n", worktreePath)
	fmt.Print("Are you sure you want to continue? (y/N): ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	result, err := r.ParseConfirmationInput(input)
	if err != nil {
		return err
	}

	if !result {
		return ErrDeletionCancelled
	}

	return nil
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (r *Repository) LoadWorktree(remoteSource, branchName string) error {
	r.VerbosePrint("Loading branch: remote=%s, branch=%s", remoteSource, branchName)

	// 1. Validate current directory is a Git repository
	gitExists, err := r.IsGitRepository()
	if err != nil {
		return fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !gitExists {
		return ErrGitRepositoryNotFound
	}

	// 2. Validate origin remote exists and is a valid Git hosting service URL
	if err := r.validateOriginRemote(); err != nil {
		return err
	}

	// 3. Parse remote source (default to "origin" if not specified)
	if remoteSource == "" {
		remoteSource = DefaultRemote
	}

	// 4. Handle remote management
	if err := r.handleRemoteManagement(remoteSource); err != nil {
		return err
	}

	// 5. Fetch from the remote
	r.VerbosePrint("Fetching from remote '%s'", remoteSource)
	if err := r.Git.FetchRemote(".", remoteSource); err != nil {
		return fmt.Errorf("%w: %w", git.ErrFetchFailed, err)
	}

	// 6. Validate branch exists on remote
	r.VerbosePrint("Checking if branch '%s' exists on remote '%s'", branchName, remoteSource)
	exists, err := r.Git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: remoteSource,
		Branch:     branchName,
	})
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: branch '%s' not found on remote '%s'", git.ErrBranchNotFoundOnRemote, branchName, remoteSource)
	}

	// 7. Create worktree for the branch (using existing worktree creation logic directly)
	r.VerbosePrint("Creating worktree for branch '%s'", branchName)
	return r.CreateWorktree(branchName)
}

// validateOriginRemote validates that the origin remote exists and is a valid Git hosting service URL.
func (r *Repository) validateOriginRemote() error {
	r.VerbosePrint("Validating origin remote")

	// Check if origin remote exists
	exists, err := r.Git.RemoteExists(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to check origin remote: %w", err)
	}
	if !exists {
		return ErrOriginRemoteNotFound
	}

	// Get origin remote URL
	originURL, err := r.Git.GetRemoteURL(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	// Validate that it's a valid Git hosting service URL
	if r.extractHostFromURL(originURL) == "" {
		return ErrOriginRemoteInvalidURL
	}

	return nil
}

// extractHostFromURL extracts the host from a Git remote URL.
func (r *Repository) extractHostFromURL(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@host:user/repo
	if strings.Contains(url, "@") && strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			hostParts := strings.Split(parts[0], "@")
			if len(hostParts) == 2 {
				return hostParts[1] // host
			}
		}
	}

	// Handle HTTPS format: https://host/user/repo
	if strings.HasPrefix(url, "http") {
		parts := strings.Split(url, "/")
		if len(parts) >= 3 {
			return parts[2] // host
		}
	}

	return ""
}

// handleRemoteManagement handles remote addition if the remote doesn't exist.
func (r *Repository) handleRemoteManagement(remoteSource string) error {
	// If remote source is "origin", no need to add it
	if remoteSource == "origin" {
		r.VerbosePrint("Using existing origin remote")
		return nil
	}

	// Check if remote already exists and handle existing remote
	if err := r.handleExistingRemote(remoteSource); err != nil {
		return err
	}

	// Add new remote
	return r.addNewRemote(remoteSource)
}

// handleExistingRemote checks if remote exists and handles it appropriately.
func (r *Repository) handleExistingRemote(remoteSource string) error {
	exists, err := r.Git.RemoteExists(".", remoteSource)
	if err != nil {
		return fmt.Errorf("failed to check if remote '%s' exists: %w", remoteSource, err)
	}

	if exists {
		remoteURL, err := r.Git.GetRemoteURL(".", remoteSource)
		if err != nil {
			return fmt.Errorf("failed to get remote URL: %w", err)
		}

		r.VerbosePrint("Using existing remote '%s' with URL: %s", remoteSource, remoteURL)
		return nil
	}

	return nil
}

// addNewRemote adds a new remote for the given remote source.
func (r *Repository) addNewRemote(remoteSource string) error {
	r.VerbosePrint("Adding new remote '%s'", remoteSource)

	// Get repository information
	repoName, err := r.Git.GetRepositoryName(".")
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	originURL, err := r.Git.GetRemoteURL(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	// Construct remote URL
	remoteURL, err := r.constructRemoteURL(originURL, remoteSource, repoName)
	if err != nil {
		return err
	}

	r.VerbosePrint("Constructed remote URL: %s", remoteURL)

	// Add the remote
	if err := r.Git.AddRemote(".", remoteSource, remoteURL); err != nil {
		return fmt.Errorf("%w: %w", git.ErrRemoteAddFailed, err)
	}

	return nil
}

// constructRemoteURL constructs the remote URL based on origin URL and remote source.
func (r *Repository) constructRemoteURL(originURL, remoteSource, repoName string) (string, error) {
	protocol := r.determineProtocol(originURL)
	host := r.extractHostFromURL(originURL)

	if host == "" {
		return "", fmt.Errorf("failed to extract host from origin URL: %s", originURL)
	}

	repoNameShort := r.extractRepoNameFromFullPath(repoName)

	if protocol == "ssh" {
		return fmt.Sprintf("git@%s:%s/%s.git", host, remoteSource, repoNameShort), nil
	}
	return fmt.Sprintf("https://%s/%s/%s.git", host, remoteSource, repoNameShort), nil
}

// determineProtocol determines the protocol (https or ssh) from the origin URL.
func (r *Repository) determineProtocol(originURL string) string {
	if strings.HasPrefix(originURL, "git@") || strings.HasPrefix(originURL, "ssh://") {
		return "ssh"
	}
	return "https"
}

// extractRepoNameFromFullPath extracts just the repository name from the full path.
func (r *Repository) extractRepoNameFromFullPath(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1] // Return the last part (repo name)
	}
	return fullPath
}

// ParseConfirmationInput parses confirmation input from user.
func (r *Repository) ParseConfirmationInput(input string) (bool, error) {
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
