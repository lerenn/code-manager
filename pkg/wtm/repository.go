package wtm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
)

const defaultRemote = "origin"

// repository represents a single Git repository and provides methods for repository operations.
type repository struct {
	fs            fs.FS
	git           git.Git
	config        *config.Config
	statusManager status.Manager
	logger        logger.Logger
	verbose       bool
}

// newRepository creates a new Repository instance.
func newRepository(
	fs fs.FS,
	git git.Git,
	config *config.Config,
	statusManager status.Manager,
	logger logger.Logger,
	verbose bool,
) *repository {
	return &repository{
		fs:            fs,
		git:           git,
		config:        config,
		statusManager: statusManager,
		logger:        logger,
		verbose:       verbose,
	}
}

// Validate validates that the current directory is a working Git repository.
func (r *repository) Validate() error {
	r.verbosePrint(fmt.Sprintf("Validating repository: %s", "."))

	// Check if .git directory exists and is a directory
	exists, err := r.CheckGitDirExists()
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
	return r.validateGitConfiguration(".")
}

// CreateWorktree creates a worktree for the repository with the specified branch.
func (r *repository) CreateWorktree(branch string) error {
	r.verbosePrint(fmt.Sprintf("Creating worktree for single repository with branch: %s", branch))

	// Validate and prepare repository
	repoURL, worktreePath, err := r.prepareWorktreeCreation(branch)
	if err != nil {
		return err
	}

	// Create the worktree
	if err := r.executeWorktreeCreation(repoURL, branch, worktreePath); err != nil {
		return err
	}

	r.verbosePrint(fmt.Sprintf("Successfully created worktree for branch %s at %s", branch, worktreePath))

	return nil
}

// ListWorktrees lists all worktrees for the current repository.
func (r *repository) ListWorktrees() ([]status.Repository, error) {
	r.verbosePrint("Listing worktrees for single repository mode")

	// Note: Repository validation is already done in mode detection, so we skip it here
	// to avoid duplicate validation calls

	// 1. Extract repository name from remote origin URL (fallback to local path if no remote)
	repoName, err := r.git.GetRepositoryName(".")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository name: %w", err)
	}

	r.verbosePrint(fmt.Sprintf("Repository name: %s", repoName))

	// 2. Load all worktrees from status file
	allWorktrees, err := r.statusManager.ListAllWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to load worktrees from status file: %w", err)
	}

	r.verbosePrint(fmt.Sprintf("Found %d total worktrees in status file", len(allWorktrees)))

	// 3. Filter worktrees to only include those for the current repository and add remote information
	var filteredWorktrees []status.Repository
	for _, worktree := range allWorktrees {
		if worktree.URL == repoName {
			// Get the remote for this branch
			remote, err := r.git.GetBranchRemote(".", worktree.Branch)
			if err != nil {
				// If we can't determine the remote, use "origin" as default
				remote = defaultRemote
			}

			// Create a copy with remote information
			worktreeWithRemote := worktree
			worktreeWithRemote.Remote = remote
			filteredWorktrees = append(filteredWorktrees, worktreeWithRemote)
		}
	}

	r.verbosePrint(fmt.Sprintf("Found %d worktrees for current repository", len(filteredWorktrees)))

	return filteredWorktrees, nil
}

// CheckGitDirExists checks if the current directory is a single Git repository.
func (r *repository) CheckGitDirExists() (bool, error) {
	r.verbosePrint("Checking for .git directory...")

	// Check if .git exists
	exists, err := r.fs.Exists(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		r.verbosePrint("No .git directory found")
		return false, nil
	}

	r.verbosePrint("Verifying .git is a directory...")

	// Check if .git is a directory
	isDir, err := r.fs.IsDir(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	if !isDir {
		r.verbosePrint(".git exists but is not a directory")
		return false, nil
	}

	r.verbosePrint("Git repository detected")
	return true, nil
}

// validateGitStatus validates that git status works in the repository.
func (r *repository) validateGitStatus() error {
	// Execute git status to ensure repository is working
	r.verbosePrint(fmt.Sprintf("Executing git status in: %s", "."))
	_, err := r.git.Status(".")
	if err != nil {
		r.verbosePrint(fmt.Sprintf("Error: %v", err))
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	return nil
}

// validateGitConfiguration validates that Git configuration is functional.
func (r *repository) validateGitConfiguration(workDir string) error {
	r.verbosePrint(fmt.Sprintf("Validating Git configuration in: %s", workDir))

	// Execute git status to ensure Git is working
	_, err := r.git.Status(workDir)
	if err != nil {
		r.verbosePrint(fmt.Sprintf("Error: %v", err))
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	return nil
}

// getBasePath returns the base path from configuration.
func (r *repository) getBasePath() (string, error) {
	if r.config == nil {
		return "", ErrConfigurationNotInitialized
	}

	if r.config.BasePath == "" {
		return "", fmt.Errorf("base path is not configured")
	}

	return r.config.BasePath, nil
}

// prepareWorktreeCreation validates the repository and prepares the worktree path.
func (r *repository) prepareWorktreeCreation(branch string) (string, string, error) {
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
func (r *repository) validateRepository(branch string) (string, error) {
	// Get current working directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Validate that we're in a Git repository
	isSingleRepo, err := r.CheckGitDirExists()
	if err != nil {
		return "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get repository URL from remote origin URL with fallback to local path
	repoURL, err := r.git.GetRepositoryName(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	r.verbosePrint(fmt.Sprintf("Repository URL: %s", repoURL))

	// Check if worktree already exists in status file
	existingWorktree, err := r.statusManager.GetWorktree(repoURL, branch)
	if err == nil && existingWorktree != nil {
		return "", fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, repoURL, branch)
	}

	// Validate repository state (placeholder for future validation)
	isClean, err := r.git.IsClean(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to check repository state: %w", err)
	}
	if !isClean {
		return "", fmt.Errorf("%w: repository is not in a clean state", ErrRepositoryNotClean)
	}

	return repoURL, nil
}

// prepareWorktreePath prepares the worktree directory path.
func (r *repository) prepareWorktreePath(repoURL, branch string) (string, error) {
	// Get base path from configuration
	basePath, err := r.getBasePath()
	if err != nil {
		return "", fmt.Errorf("failed to get base path: %w", err)
	}

	// Create worktree directory path
	worktreePath := filepath.Join(basePath, repoURL, branch)

	r.verbosePrint(fmt.Sprintf("Worktree path: %s", worktreePath))

	// Check if worktree directory already exists
	exists, err := r.fs.Exists(worktreePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return "", fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, worktreePath)
	}

	// Create worktree directory structure
	if err := r.fs.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return worktreePath, nil
}

// executeWorktreeCreation creates the branch and worktree.
func (r *repository) executeWorktreeCreation(repoURL, branch, worktreePath string) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Ensure branch exists
	if err := r.ensureBranchExists(currentDir, branch); err != nil {
		return err
	}

	// Create worktree with cleanup
	if err := r.createWorktreeWithCleanup(repoURL, branch, worktreePath, currentDir); err != nil {
		return err
	}

	return nil
}

// ensureBranchExists ensures the branch exists, creating it if necessary.
func (r *repository) ensureBranchExists(currentDir, branch string) error {
	branchExists, err := r.git.BranchExists(currentDir, branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if !branchExists {
		r.verbosePrint(fmt.Sprintf("Branch %s does not exist, creating from current branch", branch))
		if err := r.git.CreateBranch(currentDir, branch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}
	}

	return nil
}

// createWorktreeWithCleanup creates the worktree with proper cleanup on failure.
func (r *repository) createWorktreeWithCleanup(repoURL, branch, worktreePath, currentDir string) error {
	// Update status file with worktree entry (before creating the worktree for proper cleanup)
	// Store the original repository path, not the worktree path
	if err := r.statusManager.AddWorktree(repoURL, branch, currentDir, ""); err != nil {
		// Clean up created directory on status update failure
		if cleanupErr := r.cleanupWorktreeDirectory(worktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up directory after status update failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	// Create the Git worktree
	if err := r.git.CreateWorktree(currentDir, worktreePath, branch); err != nil {
		// Clean up on worktree creation failure
		if cleanupErr := r.statusManager.RemoveWorktree(repoURL, branch); cleanupErr != nil {
			r.logger.Logf("Warning: failed to remove worktree from status after creation failure: %v", cleanupErr)
		}
		if cleanupErr := r.cleanupWorktreeDirectory(worktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up directory after worktree creation failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	return nil
}

// cleanupWorktreeDirectory removes the worktree directory and parent directories if empty.
func (r *repository) cleanupWorktreeDirectory(worktreePath string) error {
	r.verbosePrint(fmt.Sprintf("Cleaning up worktree directory: %s", worktreePath))

	// Remove the worktree directory if it exists
	exists, err := r.fs.Exists(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}

	if exists {
		// Remove the worktree directory
		if err := r.fs.RemoveAll(worktreePath); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
	}

	return nil
}

// verbosePrint prints a message only in verbose mode.
func (r *repository) verbosePrint(message string) {
	if r.verbose {
		r.logger.Logf(message)
	}
}

// DeleteWorktree deletes a worktree for the repository with the specified branch.
func (r *repository) DeleteWorktree(branch string, force bool) error {
	r.verbosePrint(fmt.Sprintf("Deleting worktree for single repository with branch: %s", branch))

	// Validate and prepare worktree deletion
	repoURL, worktreePath, err := r.prepareWorktreeDeletion(branch)
	if err != nil {
		return err
	}

	// Execute the deletion
	if err := r.executeWorktreeDeletion(repoURL, branch, worktreePath, force); err != nil {
		return err
	}

	r.verbosePrint(fmt.Sprintf("Successfully deleted worktree for branch %s", branch))

	return nil
}

// prepareWorktreeDeletion validates the repository and prepares the worktree deletion.
func (r *repository) prepareWorktreeDeletion(branch string) (string, string, error) {
	// Get current working directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Validate that we're in a Git repository
	isSingleRepo, err := r.CheckGitDirExists()
	if err != nil {
		return "", "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return "", "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get repository URL from remote origin URL with fallback to local path
	repoURL, err := r.git.GetRepositoryName(currentDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	r.verbosePrint(fmt.Sprintf("Repository URL: %s", repoURL))

	// Check if worktree exists in status file
	existingWorktree, err := r.statusManager.GetWorktree(repoURL, branch)
	if err != nil || existingWorktree == nil {
		return "", "", fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotInStatus, repoURL, branch)
	}

	// Get worktree path from Git
	worktreePath, err := r.git.GetWorktreePath(currentDir, branch)
	if err != nil {
		return "", "", fmt.Errorf("failed to get worktree path: %w", err)
	}

	r.verbosePrint(fmt.Sprintf("Worktree path: %s", worktreePath))

	return repoURL, worktreePath, nil
}

// executeWorktreeDeletion deletes the worktree with proper cleanup.
func (r *repository) executeWorktreeDeletion(repoURL, branch, worktreePath string, force bool) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Prompt for confirmation unless force flag is used
	if !force {
		if err := r.promptForConfirmation(branch, worktreePath); err != nil {
			return err
		}
	}

	// Remove worktree from Git tracking first
	if err := r.git.RemoveWorktree(currentDir, worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree from Git: %w", err)
	}

	// Remove worktree directory
	if err := r.fs.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Remove entry from status file
	if err := r.statusManager.RemoveWorktree(repoURL, branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status: %w", err)
	}

	return nil
}

// promptForConfirmation prompts the user for confirmation before deletion.
func (r *repository) promptForConfirmation(branch, worktreePath string) error {
	fmt.Printf("You are about to delete the worktree for branch '%s'\n", branch)
	fmt.Printf("Worktree path: %s\n", worktreePath)
	fmt.Print("Are you sure you want to continue? (y/N): ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	switch input {
	case "y", "Y", "yes", "YES":
		return nil
	case "n", "N", "no", "NO", "":
		return ErrDeletionCancelled
	default:
		fmt.Print("Please enter 'y' for yes or 'n' for no: ")
		return r.promptForConfirmation(branch, worktreePath)
	}
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (r *repository) LoadWorktree(remoteSource, branchName string) error {
	r.verbosePrint(fmt.Sprintf("Loading branch: remote=%s, branch=%s", remoteSource, branchName))

	// 1. Validate current directory is a Git repository
	gitExists, err := r.CheckGitDirExists()
	if err != nil {
		return fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !gitExists {
		return ErrGitRepositoryNotFound
	}

	// 2. Validate origin remote exists and is a valid GitHub URL
	if err := r.validateOriginRemote(); err != nil {
		return err
	}

	// 3. Parse remote source (default to "origin" if not specified)
	if remoteSource == "" {
		remoteSource = "origin"
	}

	// 4. Handle remote management
	if err := r.handleRemoteManagement(remoteSource); err != nil {
		return err
	}

	// 5. Fetch from the remote
	r.verbosePrint(fmt.Sprintf("Fetching from remote '%s'", remoteSource))
	if err := r.git.FetchRemote(".", remoteSource); err != nil {
		return fmt.Errorf("%w: %w", git.ErrFetchFailed, err)
	}

	// 6. Validate branch exists on remote
	r.verbosePrint(fmt.Sprintf("Checking if branch '%s' exists on remote '%s'", branchName, remoteSource))
	exists, err := r.git.BranchExistsOnRemote(".", remoteSource, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: branch '%s' not found on remote '%s'", git.ErrBranchNotFoundOnRemote, branchName, remoteSource)
	}

	// 7. Create worktree for the branch (using existing worktree creation logic directly)
	r.verbosePrint(fmt.Sprintf("Creating worktree for branch '%s'", branchName))
	return r.CreateWorktree(branchName)
}

// validateOriginRemote validates that the origin remote exists and is a valid GitHub URL.
func (r *repository) validateOriginRemote() error {
	r.verbosePrint("Validating origin remote")

	// Check if origin remote exists
	exists, err := r.git.RemoteExists(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to check origin remote: %w", err)
	}
	if !exists {
		return ErrOriginRemoteNotFound
	}

	// Get origin remote URL
	originURL, err := r.git.GetRemoteURL(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	// Validate that it's a GitHub URL
	if !r.isGitHubURL(originURL) {
		return ErrOriginRemoteInvalidURL
	}

	return nil
}

// handleRemoteManagement handles remote addition if the remote doesn't exist.
func (r *repository) handleRemoteManagement(remoteSource string) error {
	// If remote source is "origin", no need to add it
	if remoteSource == "origin" {
		r.verbosePrint("Using existing origin remote")
		return nil
	}

	// Check if remote already exists
	exists, err := r.git.RemoteExists(".", remoteSource)
	if err != nil {
		return fmt.Errorf("failed to check if remote '%s' exists: %w", remoteSource, err)
	}

	if exists {
		// Validate that the existing remote URL matches expected format
		remoteURL, err := r.git.GetRemoteURL(".", remoteSource)
		if err != nil {
			return fmt.Errorf("failed to get remote URL: %w", err)
		}

		// For now, just log that we're using existing remote
		r.verbosePrint(fmt.Sprintf("Using existing remote '%s' with URL: %s", remoteSource, remoteURL))
		return nil
	}

	// Add new remote
	r.verbosePrint(fmt.Sprintf("Adding new remote '%s'", remoteSource))

	// Get repository name from origin remote
	repoName, err := r.git.GetRepositoryName(".")
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	// Determine protocol from origin remote
	originURL, err := r.git.GetRemoteURL(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	protocol := r.determineProtocol(originURL)

	// Construct remote URL
	var remoteURL string
	if protocol == "ssh" {
		remoteURL = fmt.Sprintf("git@github.com:%s/%s.git", remoteSource, r.extractRepoNameFromFullPath(repoName))
	} else {
		remoteURL = fmt.Sprintf("https://github.com/%s/%s.git", remoteSource, r.extractRepoNameFromFullPath(repoName))
	}

	r.verbosePrint(fmt.Sprintf("Constructed remote URL: %s", remoteURL))

	// Add the remote
	if err := r.git.AddRemote(".", remoteSource, remoteURL); err != nil {
		return fmt.Errorf("%w: %w", git.ErrRemoteAddFailed, err)
	}

	return nil
}

// isGitHubURL checks if the URL is a GitHub URL.
func (r *repository) isGitHubURL(url string) bool {
	return strings.Contains(url, "github.com")
}

// determineProtocol determines the protocol (https or ssh) from the origin URL.
func (r *repository) determineProtocol(originURL string) string {
	if strings.HasPrefix(originURL, "git@") || strings.HasPrefix(originURL, "ssh://") {
		return "ssh"
	}
	return "https"
}

// extractRepoNameFromFullPath extracts just the repository name from the full path.
func (r *repository) extractRepoNameFromFullPath(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1] // Return the last part (repo name)
	}
	return fullPath
}
