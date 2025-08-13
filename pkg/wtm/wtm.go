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
	return &realWTM{
		fs:            fsInstance,
		git:           git.NewGit(),
		config:        cfg,
		statusManager: status.NewManager(fsInstance, cfg),
		verbose:       false,
		logger:        logger.NewNoopLogger(),
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

	// Detect and validate project
	projectType, _, err := c.detectAndValidateProject()
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return err
	}

	// Create worktree for single repository (Feature 011)
	if err := c.handleWorktreeCreation(sanitizedBranch, projectType); err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return err
	}

	if c.verbose {
		c.logger.Logf("WTM execution completed successfully")
	}

	return nil
}

// detectAndValidateProject handles project detection and validation.
func (c *realWTM) detectAndValidateProject() (ProjectType, []string, error) {
	// Detect project mode once and store results
	projectType, workspaceFiles, err := c.detectProjectType()
	if err != nil {
		return ProjectTypeNone, nil, err
	}

	// Handle detection output (Features 001, 002, 003)
	if err := c.handleProjectDetection(projectType, workspaceFiles); err != nil {
		return ProjectTypeNone, nil, err
	}

	// Validation logic (Feature 004)
	if err := c.validateProjectStructureWithResults(projectType, workspaceFiles); err != nil {
		return ProjectTypeNone, nil, err
	}

	return projectType, workspaceFiles, nil
}

// handleWorktreeCreation handles worktree creation for single repositories.
func (c *realWTM) handleWorktreeCreation(branch string, projectType ProjectType) error {
	if projectType == ProjectTypeSingleRepo {
		if err := c.createWorktreeForSingleRepo(branch); err != nil {
			return err
		}
	}

	return nil
}

// verbosePrint prints a message only in verbose mode.
func (c *realWTM) verbosePrint(message string) {
	if c.verbose {
		c.logger.Logf(message)
	}
}

// detectSingleRepoMode checks if the current directory is a single Git repository.
func (c *realWTM) detectSingleRepoMode() (bool, error) {
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

// handleSingleRepoMode handles the output for single repository mode.
func (c *realWTM) handleSingleRepoMode() {
	c.verbosePrint("Single repository mode detected")
}

// handleWorkspaceMode handles the output for workspace mode.
func (c *realWTM) handleWorkspaceMode(workspaceFile string) error {
	workspaceConfig, err := c.parseWorkspaceFile(workspaceFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	c.verbosePrint("Workspace mode detected")

	workspaceName := c.getWorkspaceName(workspaceConfig, workspaceFile)
	c.verbosePrint(fmt.Sprintf("Found workspace: %s", workspaceName))

	if c.verbose {
		c.verbosePrint("Workspace configuration:")
		c.verbosePrint(fmt.Sprintf("  Folders: %d", len(workspaceConfig.Folders)))
		for _, folder := range workspaceConfig.Folders {
			c.verbosePrint(fmt.Sprintf("    - %s: %s", folder.Name, folder.Path))
		}
	}
	return nil
}

// handleNoProjectFound handles the output when no project is found.
func (c *realWTM) handleNoProjectFound() {
	c.logger.Logf("No Git repository or workspace found")
}

// validateSingleRepository validates that the current directory is a working Git repository.
func (c *realWTM) validateSingleRepository() error {
	if c.verbose {
		c.logger.Logf("Validating repository: %s", ".")
	}

	if err := c.validateGitDirectory(); err != nil {
		return err
	}

	if err := c.validateGitStatus(); err != nil {
		return err
	}

	// Validate Git configuration is functional
	return c.validateGitConfiguration(".")
}

// validateGitDirectory validates that .git directory exists and is a directory.
func (c *realWTM) validateGitDirectory() error {
	// Check current directory contains .git folder
	exists, err := c.fs.Exists(".git")
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		if c.verbose {
			c.logger.Logf("Error: .git directory not found")
		}
		return ErrGitRepositoryNotFound
	}

	// Verify .git is a directory
	isDir, err := c.fs.IsDir(".git")
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("failed to check .git directory: %w", err)
	}

	if !isDir {
		if c.verbose {
			c.logger.Logf("Error: .git exists but is not a directory")
		}
		return ErrGitRepositoryNotDirectory
	}

	return nil
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

// validateWorkspaceRepositories validates all repositories in a workspace.
func (c *realWTM) validateWorkspaceRepositories(workspaceFiles []string) error {
	// For now, validate the first workspace file
	// TODO: Handle multiple workspace files selection
	workspaceFile := workspaceFiles[0]

	if c.verbose {
		c.logger.Logf("Validating workspace: %s", workspaceFile)
	}

	workspaceConfig, err := c.parseWorkspaceFile(workspaceFile)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %w", ErrWorkspaceFileRead, err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(workspaceFile)

	for _, folder := range workspaceConfig.Folders {
		if err := c.validateWorkspaceRepository(folder, workspaceDir); err != nil {
			return err
		}
	}

	return nil
}

// validateWorkspaceRepository validates a single repository in a workspace.
func (c *realWTM) validateWorkspaceRepository(folder WorkspaceFolder, workspaceDir string) error {
	// Resolve relative path from workspace file location
	resolvedPath := filepath.Join(workspaceDir, folder.Path)

	if c.verbose {
		c.logger.Logf("Validating repository: %s", resolvedPath)
	}

	if err := c.validateWorkspaceRepositoryPath(folder, resolvedPath); err != nil {
		return err
	}

	if err := c.validateWorkspaceRepositoryGit(folder, resolvedPath); err != nil {
		return err
	}

	// Validate Git configuration is functional
	err := c.validateGitConfiguration(resolvedPath)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	return nil
}

// validateWorkspaceRepositoryPath validates that the repository path exists.
func (c *realWTM) validateWorkspaceRepositoryPath(folder WorkspaceFolder, resolvedPath string) error {
	// Check repository path exists
	exists, err := c.fs.Exists(resolvedPath)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("repository not found in workspace: %s - %w", folder.Path, err)
	}

	if !exists {
		if c.verbose {
			c.logger.Logf("Error: repository path does not exist")
		}
		return fmt.Errorf("%w: %s", ErrRepositoryNotFoundInWorkspace, folder.Path)
	}

	return nil
}

// validateWorkspaceRepositoryGit validates that the repository has a .git directory and git status works.
func (c *realWTM) validateWorkspaceRepositoryGit(folder WorkspaceFolder, resolvedPath string) error {
	// Verify path contains .git folder
	gitPath := filepath.Join(resolvedPath, ".git")
	exists, err := c.fs.Exists(gitPath)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	if !exists {
		if c.verbose {
			c.logger.Logf("Error: .git directory not found in repository")
		}
		return fmt.Errorf("%w: %s", ErrInvalidRepositoryInWorkspaceNoGit, folder.Path)
	}

	// Execute git status to ensure repository is working
	if c.verbose {
		c.logger.Logf("Executing git status in: %s", resolvedPath)
	}
	_, err = c.git.Status(resolvedPath)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	return nil
}

// validateGitConfiguration validates that Git is properly configured and working.
func (c *realWTM) validateGitConfiguration(workDir string) error {
	if c.verbose {
		c.logger.Logf("Validating Git configuration in: %s", workDir)
	}

	// Execute git status to ensure basic Git functionality
	if c.verbose {
		c.logger.Logf("Executing git status in: %s", workDir)
	}
	_, err := c.git.Status(workDir)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("git configuration error: %w", err)
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

// detectProjectType detects the project type and returns the type and workspace files if applicable.
func (c *realWTM) detectProjectType() (ProjectType, []string, error) {
	// First check for single repository mode
	isSingleRepo, err := c.detectSingleRepoMode()
	if err != nil {
		return ProjectTypeNone, nil, fmt.Errorf("failed to detect repository mode: %w", err)
	}

	if isSingleRepo {
		return ProjectTypeSingleRepo, nil, nil
	}

	// If no single repo found, check for workspace mode
	workspaceFiles, err := c.detectWorkspaceMode()
	if err != nil {
		return ProjectTypeNone, nil, fmt.Errorf("%w: %w", ErrWorkspaceDetection, err)
	}

	if len(workspaceFiles) > 0 {
		return ProjectTypeWorkspace, workspaceFiles, nil
	}

	// No repository or workspace found
	return ProjectTypeNone, nil, nil
}

// handleProjectDetection handles the output for the detected project type.
func (c *realWTM) handleProjectDetection(projectType ProjectType, workspaceFiles []string) error {
	switch projectType {
	case ProjectTypeSingleRepo:
		c.handleSingleRepoMode()
		return nil
	case ProjectTypeWorkspace:
		if len(workspaceFiles) > 1 {
			selectedFile, err := c.handleMultipleWorkspaces(workspaceFiles)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrMultipleWorkspaces, err)
			}
			return c.handleWorkspaceMode(selectedFile)
		}
		if len(workspaceFiles) == 1 {
			return c.handleWorkspaceMode(workspaceFiles[0])
		}
	case ProjectTypeNone:
		c.handleNoProjectFound()
		return nil
	}
	return nil
}

// validateProjectStructureWithResults validates the project structure using pre-detected results.
func (c *realWTM) validateProjectStructureWithResults(projectType ProjectType, workspaceFiles []string) error {
	if c.verbose {
		c.logger.Logf("Starting project structure validation")
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		if c.verbose {
			c.logger.Logf("Validating single repository mode")
		}
		return c.validateSingleRepository()
	case ProjectTypeWorkspace:
		if c.verbose {
			c.logger.Logf("Validating workspace mode with %d workspace files", len(workspaceFiles))
		}
		return c.validateWorkspaceRepositories(workspaceFiles)
	case ProjectTypeNone:
		return fmt.Errorf("no Git repository or workspace found")
	}
	return nil
}

// sanitizeBranchName validates and sanitizes branch name for safe directory creation.

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
	isSingleRepo, err := c.detectSingleRepoMode()
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
