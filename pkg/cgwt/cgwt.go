package cgwt

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/lerenn/cgwt/pkg/git"
	"github.com/lerenn/cgwt/pkg/logger"
)

// CGWT interface provides Git repository detection functionality.
type CGWT interface {
	// Run executes the main application logic.
	Run() error
	// SetVerbose enables or disables verbose mode.
	SetVerbose(verbose bool)
	// SetLogger sets a custom logger for verbose output.
	SetLogger(logger logger.Logger)
	// ValidateSingleRepository validates that the current directory is a working Git repository.
	ValidateSingleRepository() error
}

type realCGWT struct {
	fs      fs.FS
	git     git.Git
	verbose bool
	logger  logger.Logger
}

// NewCGWT creates a new CGWT instance.
func NewCGWT() CGWT {
	return &realCGWT{
		fs:      fs.NewFS(),
		git:     git.NewGit(),
		verbose: false,
		logger:  logger.NewNoopLogger(),
	}
}

func (c *realCGWT) SetVerbose(verbose bool) {
	c.verbose = verbose
	if verbose && c.logger == logger.NewNoopLogger() {
		c.logger = logger.NewDefaultLogger()
	} else if !verbose {
		c.logger = logger.NewNoopLogger()
	}
}

func (c *realCGWT) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// Run executes the main application logic.
func (c *realCGWT) Run() error {
	if c.verbose {
		c.logger.Logf("Starting CGWT execution")
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
		c.logger.Logf("CGWT execution completed successfully")
	}

	return nil
}

// verbosePrint prints a message only in verbose mode.
func (c *realCGWT) verbosePrint(message string) {
	if c.verbose {
		c.logger.Logf(message)
	}
}

// detectSingleRepoMode checks if the current directory is a single Git repository.
func (c *realCGWT) detectSingleRepoMode() (bool, error) {
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
func (c *realCGWT) handleSingleRepoMode() {
	fmt.Println("Single repository mode detected")
}

// handleWorkspaceMode handles the output for workspace mode.
func (c *realCGWT) handleWorkspaceMode(workspaceFile string) error {
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
func (c *realCGWT) handleNoProjectFound() {
	fmt.Println("No Git repository or workspace found")
}

// ValidateSingleRepository validates that the current directory is a working Git repository.
func (c *realCGWT) ValidateSingleRepository() error {
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
func (c *realCGWT) validateGitDirectory() error {
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
		return fmt.Errorf("not a valid Git repository: .git directory not found")
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
		return fmt.Errorf("not a valid Git repository: .git exists but is not a directory")
	}

	return nil
}

// validateGitStatus validates that git status works in the repository.
func (c *realCGWT) validateGitStatus() error {
	// Execute git status to ensure repository is working
	if c.verbose {
		c.logger.Logf("Executing git status in: %s", ".")
	}
	_, err := c.git.Status(".")
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("not a valid Git repository: %w", err)
	}

	return nil
}

// validateWorkspaceRepositories validates all repositories in a workspace.
func (c *realCGWT) validateWorkspaceRepositories(workspaceFiles []string) error {
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
		return fmt.Errorf("failed to parse workspace file: %w", err)
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
func (c *realCGWT) validateWorkspaceRepository(folder WorkspaceFolder, workspaceDir string) error {
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
		return fmt.Errorf("invalid repository in workspace: %s - %w", folder.Path, err)
	}

	return nil
}

// validateWorkspaceRepositoryPath validates that the repository path exists.
func (c *realCGWT) validateWorkspaceRepositoryPath(folder WorkspaceFolder, resolvedPath string) error {
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
		return fmt.Errorf("repository not found in workspace: %s", folder.Path)
	}

	return nil
}

// validateWorkspaceRepositoryGit validates that the repository has a .git directory and git status works.
func (c *realCGWT) validateWorkspaceRepositoryGit(folder WorkspaceFolder, resolvedPath string) error {
	// Verify path contains .git folder
	gitPath := filepath.Join(resolvedPath, ".git")
	exists, err := c.fs.Exists(gitPath)
	if err != nil {
		if c.verbose {
			c.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("invalid repository in workspace: %s - %w", folder.Path, err)
	}

	if !exists {
		if c.verbose {
			c.logger.Logf("Error: .git directory not found in repository")
		}
		return fmt.Errorf("invalid repository in workspace: %s - .git directory not found", folder.Path)
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
		return fmt.Errorf("invalid repository in workspace: %s - %w", folder.Path, err)
	}

	return nil
}

// validateGitConfiguration validates that Git is properly configured and working.
func (c *realCGWT) validateGitConfiguration(workDir string) error {
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

// ProjectType represents the type of project detected
type ProjectType int

const (
	ProjectTypeNone ProjectType = iota
	ProjectTypeSingleRepo
	ProjectTypeWorkspace
)

// detectProjectType detects the project type and returns the type and workspace files if applicable
func (c *realCGWT) detectProjectType() (ProjectType, []string, error) {
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
		return ProjectTypeNone, nil, fmt.Errorf("failed to detect workspace mode: %w", err)
	}

	if len(workspaceFiles) > 0 {
		return ProjectTypeWorkspace, workspaceFiles, nil
	}

	// No repository or workspace found
	return ProjectTypeNone, nil, nil
}

// handleProjectDetection handles the output for the detected project type
func (c *realCGWT) handleProjectDetection(projectType ProjectType, workspaceFiles []string) error {
	switch projectType {
	case ProjectTypeSingleRepo:
		c.handleSingleRepoMode()
		return nil
	case ProjectTypeWorkspace:
		if len(workspaceFiles) > 1 {
			selectedFile, err := c.handleMultipleWorkspaces(workspaceFiles)
			if err != nil {
				return fmt.Errorf("failed to handle multiple workspaces: %w", err)
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

// validateProjectStructureWithResults validates the project structure using pre-detected results
func (c *realCGWT) validateProjectStructureWithResults(projectType ProjectType, workspaceFiles []string) error {
	if c.verbose {
		c.logger.Logf("Starting project structure validation")
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		if c.verbose {
			c.logger.Logf("Validating single repository mode")
		}
		return c.ValidateSingleRepository()
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
