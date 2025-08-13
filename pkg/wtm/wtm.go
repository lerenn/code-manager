package wtm

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
)

// WTM interface provides Git repository detection functionality.
type WTM interface {
	// CreateWorkTree executes the main application logic.
	CreateWorkTree(branch string) error
	// SetVerbose enables or disables verbose mode.
	SetVerbose(verbose bool)
}

type realWTM struct {
	fs            fs.FS
	git           git.Git
	config        *config.Config
	statusManager status.Manager
	verbose       bool
	logger        logger.Logger
}

// NewWTM creates a new WTM instance.
func NewWTM(cfg *config.Config) WTM {
	fsInstance := fs.NewFS()
	gitInstance := git.NewGit()
	loggerInstance := logger.NewNoopLogger()

	return &realWTM{
		fs:            fsInstance,
		git:           gitInstance,
		config:        cfg,
		statusManager: status.NewManager(fsInstance, cfg),
		verbose:       false,
		logger:        loggerInstance,
	}
}

func (c *realWTM) SetVerbose(verbose bool) {
	c.verbose = verbose
	if verbose {
		c.logger = logger.NewDefaultLogger()
	} else {
		c.logger = logger.NewNoopLogger()
	}
}

// CreateWorkTree executes the main application logic.
func (c *realWTM) CreateWorkTree(branch string) error {
	// Sanitize branch name first
	sanitizedBranch, err := c.sanitizeBranchName(branch)
	if err != nil {
		return err
	}

	// Log if branch name was sanitized (appears in normal and verbose modes, but not quiet)
	if sanitizedBranch != branch {
		c.logger.Logf("Branch name sanitized: %s -> %s", branch, sanitizedBranch)
	}

	if c.verbose {
		c.logger.Logf("Starting WTM execution for branch: %s (sanitized: %s)", branch, sanitizedBranch)
	}

	// 1. Detect project mode (repository or workspace)
	projectType, err := c.detectProjectMode()
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return err
	}

	// 2. Handle based on project type
	switch projectType {
	case ProjectTypeSingleRepo:
		return c.handleRepositoryMode(sanitizedBranch)
	case ProjectTypeWorkspace:
		return c.handleWorkspaceMode(sanitizedBranch)
	case ProjectTypeNone:
		return fmt.Errorf("no Git repository or workspace found")
	default:
		return fmt.Errorf("unknown project type")
	}
}

// detectProjectMode detects if this is a repository or workspace mode.
func (c *realWTM) detectProjectMode() (ProjectType, error) {
	// First check for single repository mode
	isSingleRepo, err := c.checkGitDirExists()
	if err != nil {
		return ProjectTypeNone, fmt.Errorf("failed to detect repository mode: %w", err)
	}

	if isSingleRepo {
		return ProjectTypeSingleRepo, nil
	}

	// If no single repo found, check for workspace mode
	workspaceFiles, err := c.fs.Glob("*.code-workspace")
	if err != nil {
		return ProjectTypeNone, fmt.Errorf("failed to check for workspace files: %w", err)
	}

	if len(workspaceFiles) > 0 {
		return ProjectTypeWorkspace, nil
	}

	// No repository or workspace found
	return ProjectTypeNone, nil
}

// handleRepositoryMode handles repository mode: validation and worktree creation.
func (c *realWTM) handleRepositoryMode(branch string) error {
	if c.verbose {
		c.logger.Logf("Handling repository mode")
	}

	// 1. Validate repository
	if err := c.validateCurrentDirIsGitRepository(); err != nil {
		return err
	}

	// 2. Create worktree for single repository
	if err := c.createWorktreeForSingleRepo(branch); err != nil {
		return err
	}

	if c.verbose {
		c.logger.Logf("WTM execution completed successfully")
	}

	return nil
}

// handleWorkspaceMode handles workspace mode: validation and placeholder for worktree creation.
func (c *realWTM) handleWorkspaceMode(_ string) error {
	if c.verbose {
		c.logger.Logf("Handling workspace mode")
	}

	// Create a single workspace instance for all workspace operations
	workspace := newWorkspace(c.fs, c.git, c.logger, c.verbose)

	// 1. Load workspace (detection, selection, and display)
	if err := workspace.Load(); err != nil {
		return err
	}

	// 2. Validate all repositories in workspace
	if err := workspace.Validate(); err != nil {
		return err
	}

	// 4. TODO: Create worktrees for workspace repositories (placeholder)
	if c.verbose {
		c.logger.Logf("Workspace worktree creation not yet implemented")
	}

	if c.verbose {
		c.logger.Logf("WTM execution completed successfully")
	}

	return nil
}

// verbosePrint prints a message only in verbose mode.
func (c *realWTM) verbosePrint(message string) {
	if c.verbose {
		c.logger.Logf(message)
	}
}

// checkGitDirExists checks if the current directory is a single Git repository.
func (c *realWTM) checkGitDirExists() (bool, error) {
	c.verbosePrint("Checking for .git directory...")

	// Check if .git exists
	exists, err := c.fs.Exists(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		c.verbosePrint("No .git directory found")
		return false, nil
	}

	c.verbosePrint("Verifying .git is a directory...")

	// Check if .git is a directory
	isDir, err := c.fs.IsDir(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	if !isDir {
		c.verbosePrint(".git exists but is not a directory")
		return false, nil
	}

	c.verbosePrint("Git repository detected")
	return true, nil
}

// validateCurrentDirIsGitRepository validates that the current directory is a working Git repository.
func (c *realWTM) validateCurrentDirIsGitRepository() error {
	if c.verbose {
		c.logger.Logf("Validating repository: %s", ".")
	}

	// Check if .git directory exists and is a directory
	exists, err := c.checkGitDirExists()
	if err != nil {
		return err
	}
	if !exists {
		return ErrGitRepositoryNotFound
	}

	if err := c.validateGitStatus(); err != nil {
		return err
	}

	// Validate Git configuration is functional
	return c.validateGitConfiguration(".")
}

// validateGitStatus validates that git status works in the repository.
func (c *realWTM) validateGitStatus() error {
	// Execute git status to ensure repository is working
	if c.verbose {
		c.logger.Logf("Executing git status in: %s", ".")
	}
	_, err := c.git.Status(".")
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	return nil
}

// validateGitConfiguration validates that Git configuration is functional.
func (c *realWTM) validateGitConfiguration(workDir string) error {
	if c.verbose {
		c.logger.Logf("Validating Git configuration in: %s", workDir)
	}

	// Execute git status to ensure Git is working
	_, err := c.git.Status(workDir)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	return nil
}

// ProjectType represents the type of project detected.
type ProjectType int

// Project type constants.
const (
	ProjectTypeNone ProjectType = iota
	ProjectTypeSingleRepo
	ProjectTypeWorkspace
)

// sanitizeBranchName validates and sanitizes branch name for safe directory creation.
func (c *realWTM) sanitizeBranchName(branchName string) (string, error) {
	if branchName == "" {
		return "", ErrBranchNameEmpty
	}

	// Replace invalid characters with underscores
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*#]`)
	sanitized := invalidChars.ReplaceAllString(branchName, "_")

	// Remove leading/trailing underscores and dots
	sanitized = strings.Trim(sanitized, "._")

	// Limit length to 255 characters (filesystem limit)
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
		// Ensure we don't end with a dot or underscore
		sanitized = strings.TrimRight(sanitized, "._")
	}

	if sanitized == "" {
		return "", ErrBranchNameEmptyAfterSanitization
	}

	return sanitized, nil
}

// getBasePath returns the base path from configuration.
func (c *realWTM) getBasePath() (string, error) {
	if c.config == nil {
		return "", ErrConfigurationNotInitialized
	}

	if c.config.BasePath == "" {
		return "", fmt.Errorf("base path is not configured")
	}

	return c.config.BasePath, nil
}

// createWorktreeForSingleRepo creates a worktree for a single repository.
func (c *realWTM) createWorktreeForSingleRepo(branch string) error {
	if c.verbose {
		c.logger.Logf("Creating worktree for single repository with branch: %s", branch)
	}

	// Validate and prepare repository
	repoURL, worktreePath, err := c.prepareWorktreeCreation(branch)
	if err != nil {
		return err
	}

	// Create the worktree
	if err := c.executeWorktreeCreation(repoURL, branch, worktreePath); err != nil {
		return err
	}

	if c.verbose {
		c.logger.Logf("Successfully created worktree for branch %s at %s", branch, worktreePath)
	}

	return nil
}

// prepareWorktreeCreation validates the repository and prepares the worktree path.
func (c *realWTM) prepareWorktreeCreation(branch string) (string, string, error) {
	// Validate repository
	repoURL, err := c.validateRepository(branch)
	if err != nil {
		return "", "", err
	}

	// Prepare worktree path
	worktreePath, err := c.prepareWorktreePath(repoURL, branch)
	if err != nil {
		return "", "", err
	}

	return repoURL, worktreePath, nil
}

// validateRepository validates the repository and gets the repository name.
func (c *realWTM) validateRepository(branch string) (string, error) {
	// Get current working directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Validate that we're in a Git repository
	isSingleRepo, err := c.checkGitDirExists()
	if err != nil {
		return "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get repository URL from remote origin URL with fallback to local path
	repoURL, err := c.git.GetRepositoryName(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	if c.verbose {
		c.logger.Logf("Repository URL: %s", repoURL)
	}

	// Check if worktree already exists in status file
	existingWorktree, err := c.statusManager.GetWorktree(repoURL, branch)
	if err == nil && existingWorktree != nil {
		return "", fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, repoURL, branch)
	}

	// Validate repository state (placeholder for future validation)
	isClean, err := c.git.IsClean(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to check repository state: %w", err)
	}
	if !isClean {
		return "", fmt.Errorf("%w: repository is not in a clean state", ErrRepositoryNotClean)
	}

	return repoURL, nil
}

// prepareWorktreePath prepares the worktree directory path.
func (c *realWTM) prepareWorktreePath(repoURL, branch string) (string, error) {
	// Get base path from configuration
	basePath, err := c.getBasePath()
	if err != nil {
		return "", fmt.Errorf("failed to get base path: %w", err)
	}

	// Create worktree directory path
	worktreePath := filepath.Join(basePath, repoURL, branch)

	if c.verbose {
		c.logger.Logf("Worktree path: %s", worktreePath)
	}

	// Check if worktree directory already exists
	exists, err := c.fs.Exists(worktreePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return "", fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, worktreePath)
	}

	// Create worktree directory structure
	if err := c.fs.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return worktreePath, nil
}

// executeWorktreeCreation creates the branch and worktree.
func (c *realWTM) executeWorktreeCreation(repoURL, branch, worktreePath string) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Ensure branch exists
	if err := c.ensureBranchExists(currentDir, branch); err != nil {
		return err
	}

	// Create worktree with cleanup
	if err := c.createWorktreeWithCleanup(repoURL, branch, worktreePath, currentDir); err != nil {
		return err
	}

	return nil
}

// ensureBranchExists ensures the branch exists, creating it if necessary.
func (c *realWTM) ensureBranchExists(currentDir, branch string) error {
	branchExists, err := c.git.BranchExists(currentDir, branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if !branchExists {
		if c.verbose {
			c.logger.Logf("Branch %s does not exist, creating from current branch", branch)
		}
		if err := c.git.CreateBranch(currentDir, branch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}
	}

	return nil
}

// createWorktreeWithCleanup creates the worktree with proper cleanup on failure.
func (c *realWTM) createWorktreeWithCleanup(repoURL, branch, worktreePath, currentDir string) error {
	// Update status file with worktree entry (before creating the worktree for proper cleanup)
	if err := c.statusManager.AddWorktree(repoURL, branch, currentDir, ""); err != nil {
		// Clean up created directory on status update failure
		if cleanupErr := c.cleanupWorktreeDirectory(worktreePath); cleanupErr != nil {
			c.logger.Logf("Warning: failed to clean up directory after status update failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	// Create the Git worktree
	if err := c.git.CreateWorktree(currentDir, worktreePath, branch); err != nil {
		// Clean up on worktree creation failure
		if cleanupErr := c.statusManager.RemoveWorktree(repoURL, branch); cleanupErr != nil {
			c.logger.Logf("Warning: failed to remove worktree from status after creation failure: %v", cleanupErr)
		}
		if cleanupErr := c.cleanupWorktreeDirectory(worktreePath); cleanupErr != nil {
			c.logger.Logf("Warning: failed to clean up directory after worktree creation failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	return nil
}

// cleanupWorktreeDirectory removes the worktree directory and parent directories if empty.
func (c *realWTM) cleanupWorktreeDirectory(worktreePath string) error {
	if c.verbose {
		c.logger.Logf("Cleaning up worktree directory: %s", worktreePath)
	}

	// Remove the worktree directory if it exists
	exists, err := c.fs.Exists(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}

	if exists {
		// Remove the worktree directory
		if err := c.fs.RemoveAll(worktreePath); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
	}

	return nil
}
