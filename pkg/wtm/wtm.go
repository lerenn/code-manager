package wtm

import (
	"fmt"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/ide"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
)

// WTM interface provides Git repository detection functionality.
type WTM interface {
	// CreateWorkTree executes the main application logic.
	CreateWorkTree(branch string, ideName *string) error

	// OpenWorktree opens an existing worktree in the specified IDE.
	OpenWorktree(worktreeName, ideName string) error
	// SetVerbose enables or disables verbose mode.
	SetVerbose(verbose bool)
}

type realWTM struct {
	fs            fs.FS
	git           git.Git
	config        *config.Config
	statusManager status.Manager
	ideManager    ide.ManagerInterface
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
		ideManager:    ide.NewManager(fsInstance, loggerInstance),
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

	// Update the IDE manager with the new logger
	c.ideManager = ide.NewManager(c.fs, c.logger)
}

// CreateWorkTree executes the main application logic.
func (c *realWTM) CreateWorkTree(branch string, ideName *string) error {
	// Sanitize branch name first
	sanitizedBranch, err := c.sanitizeBranchName(branch)
	if err != nil {
		return err
	}

	// Log if branch name was sanitized (appears in normal and verbose modes, but not quiet)
	if sanitizedBranch != branch {
		c.logger.Logf("Branch name sanitized: %s -> %s", branch, sanitizedBranch)
	}

	c.verbosePrint(fmt.Sprintf("Starting WTM execution for branch: %s (sanitized: %s)", branch, sanitizedBranch))

	// 1. Detect project mode (repository or workspace)
	projectType, err := c.detectProjectMode()
	if err != nil {
		c.verbosePrint(fmt.Sprintf("Error: %v", err))
		return err
	}

	// 2. Handle based on project type
	var worktreeErr error
	switch projectType {
	case ProjectTypeSingleRepo:
		worktreeErr = c.handleRepositoryMode(sanitizedBranch)
	case ProjectTypeWorkspace:
		worktreeErr = c.handleWorkspaceMode(sanitizedBranch)
	case ProjectTypeNone:
		return fmt.Errorf("no Git repository or workspace found")
	default:
		return fmt.Errorf("unknown project type")
	}

	// 3. Open IDE if specified and worktree creation was successful
	if err := c.handleIDEOpening(worktreeErr, sanitizedBranch, ideName); err != nil {
		return err
	}

	return worktreeErr
}

// handleIDEOpening handles IDE opening if specified and worktree creation was successful.
func (c *realWTM) handleIDEOpening(worktreeErr error, branch string, ideName *string) error {
	if worktreeErr == nil && ideName != nil && *ideName != "" {
		if err := c.OpenWorktree(branch, *ideName); err != nil {
			return err
		}
	}
	return nil
}

// detectProjectMode detects if this is a repository or workspace mode.
func (c *realWTM) detectProjectMode() (ProjectType, error) {
	// First check for single repository mode
	repo := newRepository(c.fs, c.git, c.config, c.statusManager, c.logger, c.verbose)
	exists, err := repo.CheckGitDirExists()
	if err != nil {
		return ProjectTypeNone, fmt.Errorf("failed to detect repository mode: %w", err)
	}

	if exists {
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
	c.verbosePrint("Handling repository mode")

	// Create a single repository instance for all repository operations
	repo := newRepository(c.fs, c.git, c.config, c.statusManager, c.logger, c.verbose)

	// 1. Validate repository
	if err := repo.Validate(); err != nil {
		return err
	}

	// 2. Create worktree for single repository
	if err := repo.CreateWorktree(branch); err != nil {
		return err
	}

	c.verbosePrint("WTM execution completed successfully")

	return nil
}

// handleWorkspaceMode handles workspace mode: validation and placeholder for worktree creation.
func (c *realWTM) handleWorkspaceMode(_ string) error {
	c.verbosePrint("Handling workspace mode")

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
	c.verbosePrint("Workspace worktree creation not yet implemented")

	c.verbosePrint("WTM execution completed successfully")

	return nil
}

// OpenWorktree opens an existing worktree in the specified IDE.
func (c *realWTM) OpenWorktree(worktreeName, ideName string) error {
	// Get repository URL from local .git directory
	repoURL, err := c.git.GetRepositoryName(".")
	if err != nil {
		return fmt.Errorf("failed to get repository URL: %w", err)
	}

	// Get worktree from status.yaml using repository URL and branch name
	worktree, err := c.statusManager.GetWorktree(repoURL, worktreeName)
	if err != nil {
		return fmt.Errorf("%w: %s", ide.ErrWorktreeNotFound, worktreeName)
	}

	// Open IDE with the worktree path
	if err := c.ideManager.OpenIDE(ideName, worktree.Path, c.verbose); err != nil {
		return err
	}

	return nil
}

// verbosePrint prints a message only in verbose mode.
func (c *realWTM) verbosePrint(message string) {
	if c.verbose {
		c.logger.Logf(message)
	}
}
