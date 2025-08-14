package wtm

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
)

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
	if err := r.statusManager.AddWorktree(repoURL, branch, worktreePath, ""); err != nil {
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
