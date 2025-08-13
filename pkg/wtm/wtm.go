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
	CreateWorkTree() error
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
func (c *realWTM) CreateWorkTree() error {
	if c.verbose {
		c.logger.Logf("Starting WTM execution")
	}

	// Detect project mode once and store results
	projectType, workspaceFiles, err := c.detectProjectType()
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return err
	}

	// Handle detection output (Features 001, 002, 003)
	if err := c.handleProjectDetection(projectType, workspaceFiles); err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return err
	}

	// Validation logic (Feature 004)
	if err := c.validateProjectStructureWithResults(projectType, workspaceFiles); err != nil {
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
	fmt.Println("Single repository mode detected")
}

// handleWorkspaceMode handles the output for workspace mode.
func (c *realWTM) handleWorkspaceMode(workspaceFile string) error {
	workspaceConfig, err := c.parseWorkspaceFile(workspaceFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	c.verbosePrint("Workspace mode detected")

	workspaceName := c.getWorkspaceName(workspaceConfig, workspaceFile)
	fmt.Printf("Found workspace: %s\n", workspaceName)

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
	fmt.Println("No Git repository or workspace found")
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

// createReposDirectoryStructure creates the directory structure for repository worktrees.
//
//nolint:unused // This method will be used in the Run() method in future features
func (c *realWTM) createReposDirectoryStructure(repoName, branchName string) (string, error) {
	if c.verbose {
		c.logger.Logf("Creating directory structure for repo: %s, branch: %s", repoName, branchName)
	}

	// Sanitize repository and branch names
	sanitizedRepo, err := c.sanitizeRepositoryName(repoName)
	if err != nil {
		return "", fmt.Errorf("failed to sanitize repository name: %w", err)
	}

	sanitizedBranch, err := c.sanitizeBranchName(branchName)
	if err != nil {
		return "", fmt.Errorf("failed to sanitize branch name: %w", err)
	}

	// Get base path from config
	basePath, err := c.getBasePath()
	if err != nil {
		return "", fmt.Errorf("failed to get base path: %w", err)
	}

	// Construct full path
	fullPath := filepath.Join(basePath, "repos", sanitizedRepo, sanitizedBranch)

	if c.verbose {
		c.logger.Logf("Creating directory structure: %s", fullPath)
	}

	// Create directory structure
	err = c.fs.MkdirAll(fullPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory structure: %w", err)
	}

	if c.verbose {
		c.logger.Logf("Successfully created directory structure: %s", fullPath)
	}

	return fullPath, nil
}

// sanitizeRepositoryName extracts and sanitizes repository name from Git remote URL.
//
//nolint:unused // This method will be used in the Run() method in future features
func (c *realWTM) sanitizeRepositoryName(remoteURL string) (string, error) {
	if remoteURL == "" {
		return "", ErrRepositoryURLEmpty
	}

	// Remove .git suffix if present
	repoName := strings.TrimSuffix(remoteURL, ".git")

	// Extract repository path from URL
	repoName = c.extractRepoPathFromURL(repoName)

	// Sanitize repository name for filesystem
	// Replace invalid characters with underscores (but preserve forward slashes)
	invalidChars := regexp.MustCompile(`[<>:"\\|?*]`)
	sanitized := invalidChars.ReplaceAllString(repoName, "_")

	// Remove leading/trailing underscores and dots
	sanitized = strings.Trim(sanitized, "._")

	if sanitized == "" {
		return "", ErrRepositoryNameEmptyAfterSanitization
	}

	return sanitized, nil
}

// extractRepoPathFromURL extracts the repository path from various URL formats.
//
//nolint:unused // This method will be used in the Run() method in future features
func (c *realWTM) extractRepoPathFromURL(repoName string) string {
	// Handle SSH format: git@github.com:user/repo.git
	if strings.Contains(repoName, "git@") {
		parts := strings.Split(repoName, ":")
		if len(parts) == 2 {
			return parts[1]
		}
		return repoName
	}

	// Handle HTTPS format: https://github.com/user/repo
	// Extract everything after the domain
	urlParts := strings.Split(repoName, "://")
	if len(urlParts) != 2 {
		return repoName
	}

	pathParts := strings.Split(urlParts[1], "/")
	if len(pathParts) >= 3 {
		// Remove domain and take user/repo parts
		return strings.Join(pathParts[1:], "/")
	}

	return repoName
}

// sanitizeBranchName validates and sanitizes branch name for safe directory creation.
//
//nolint:unused // This method will be used in the Run() method in future features
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
//
//nolint:unused // This method will be used in the Run() method in future features
func (c *realWTM) getBasePath() (string, error) {
	if c.config == nil {
		return "", ErrConfigurationNotInitialized
	}

	if c.config.BasePath == "" {
		return "", fmt.Errorf("base path is not configured")
	}

	return c.config.BasePath, nil
}

// addWorktreeToStatus adds a worktree entry to the status file.
//
//nolint:unused // This method is used internally
func (c *realWTM) addWorktreeToStatus(repoName, branch, worktreePath, workspacePath string) error {
	if c.verbose {
		c.logger.Logf("Adding worktree to status: repo=%s, branch=%s, path=%s, workspace=%s",
			repoName, branch, worktreePath, workspacePath)
	}

	err := c.statusManager.AddWorktree(repoName, branch, worktreePath, workspacePath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAddWorktreeToStatus, err)
	}

	if c.verbose {
		c.logger.Logf("Successfully added worktree to status")
	}

	return nil
}

// removeWorktreeFromStatus removes a worktree entry from the status file.
//
//nolint:unused // This method is used internally
func (c *realWTM) removeWorktreeFromStatus(repoName, branch string) error {
	if c.verbose {
		c.logger.Logf("Removing worktree from status: repo=%s, branch=%s", repoName, branch)
	}

	err := c.statusManager.RemoveWorktree(repoName, branch)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRemoveWorktreeFromStatus, err)
	}

	if c.verbose {
		c.logger.Logf("Successfully removed worktree from status")
	}

	return nil
}

// getWorktreeStatus retrieves the status of a specific worktree.
//
//nolint:unused // This method is used internally
func (c *realWTM) getWorktreeStatus(repoName, branch string) (*status.Repository, error) {
	if c.verbose {
		c.logger.Logf("Getting worktree status: repo=%s, branch=%s", repoName, branch)
	}

	repo, err := c.statusManager.GetWorktree(repoName, branch)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetWorktreeStatus, err)
	}

	return repo, nil
}

// listAllWorktrees lists all tracked worktrees.
//
//nolint:unused // This method is used internally
func (c *realWTM) listAllWorktrees() ([]status.Repository, error) {
	if c.verbose {
		c.logger.Logf("Listing all worktrees")
	}

	repos, err := c.statusManager.ListAllWorktrees()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrListWorktrees, err)
	}

	if c.verbose {
		c.logger.Logf("Found %d worktrees", len(repos))
	}

	return repos, nil
}
