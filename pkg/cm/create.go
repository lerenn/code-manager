package cm

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/forge"
	"github.com/lerenn/code-manager/pkg/issue"
	repo "github.com/lerenn/code-manager/pkg/repository"
	ws "github.com/lerenn/code-manager/pkg/workspace"
)

// CreateWorkTreeOpts contains optional parameters for CreateWorkTree.
type CreateWorkTreeOpts struct {
	IDEName  string
	IssueRef string
}

// CreateWorkTree executes the main application logic.
func (c *realCM) CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error {
	// Extract and validate options
	issueRef, ideName := c.extractCreateWorkTreeOptions(opts)

	// Handle issue-based worktree creation
	if issueRef != "" {
		return c.createWorkTreeFromIssue(branch, issueRef, ideName)
	}

	// Handle regular worktree creation
	return c.createRegularWorkTree(branch, ideName)
}

// extractCreateWorkTreeOptions extracts options from the variadic parameter.
func (c *realCM) extractCreateWorkTreeOptions(opts []CreateWorkTreeOpts) (string, *string) {
	var issueRef string
	var ideName *string

	if len(opts) > 0 {
		if opts[0].IssueRef != "" {
			issueRef = opts[0].IssueRef
		}
		if opts[0].IDEName != "" {
			ideName = &opts[0].IDEName
		}
	}

	return issueRef, ideName
}

// createRegularWorkTree handles regular worktree creation (non-issue based).
func (c *realCM) createRegularWorkTree(branch string, ideName *string) error {
	// Sanitize branch name first
	sanitizedBranch, err := c.sanitizeBranchName(branch)
	if err != nil {
		return err
	}

	// Log if branch name was sanitized
	if sanitizedBranch != branch {
		c.Logger.Logf("Branch name sanitized: %s -> %s", branch, sanitizedBranch)
	}

	c.VerbosePrint("Starting CM execution for branch: %s (sanitized: %s)", branch, sanitizedBranch)

	// Detect project mode and handle accordingly
	worktreeErr := c.handleProjectMode(sanitizedBranch)

	// Open IDE if specified and worktree creation was successful
	if err := c.handleIDEOpening(worktreeErr, sanitizedBranch, ideName); err != nil {
		return err
	}

	return worktreeErr
}

// handleProjectMode detects project mode and handles worktree creation accordingly.
func (c *realCM) handleProjectMode(sanitizedBranch string) error {
	projectType, err := c.detectProjectMode()
	if err != nil {
		c.VerbosePrint("Error: %v", err)
		return err
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		return c.handleRepositoryMode(sanitizedBranch)
	case ProjectTypeWorkspace:
		return c.handleWorkspaceMode(sanitizedBranch)
	case ProjectTypeNone:
		return ErrNoGitRepositoryOrWorkspaceFound
	default:
		return fmt.Errorf("unknown project type")
	}
}

// handleRepositoryMode handles repository mode: validation and worktree creation.
func (c *realCM) handleRepositoryMode(branch string) error {
	c.VerbosePrint("Handling repository mode")

	// Create a single repository instance for all repository operations
	repoInstance := repo.NewRepository(repo.NewRepositoryParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})

	// 1. Validate repository
	if err := repoInstance.Validate(); err != nil {
		return c.translateRepositoryError(err)
	}

	// 2. Create worktree for single repository
	if err := repoInstance.CreateWorktree(branch); err != nil {
		return c.translateRepositoryError(err)
	}

	c.VerbosePrint("CM execution completed successfully")

	return nil
}

// translateRepositoryError translates repository package errors to CM package errors.
func (c *realCM) translateRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific repository errors and translate them
	if errors.Is(err, repo.ErrWorktreeExists) {
		return ErrWorktreeExists
	}
	if errors.Is(err, repo.ErrWorktreeNotInStatus) {
		return ErrWorktreeNotInStatus
	}
	if errors.Is(err, repo.ErrRepositoryNotClean) {
		return ErrRepositoryNotClean
	}
	if errors.Is(err, repo.ErrDirectoryExists) {
		return ErrDirectoryExists
	}
	if errors.Is(err, repo.ErrGitRepositoryNotFound) {
		return ErrGitRepositoryNotFound
	}

	// Return the original error if no translation is needed
	return err
}

// handleWorkspaceMode handles workspace mode: validation and worktree creation.
func (c *realCM) handleWorkspaceMode(branch string) error {
	c.VerbosePrint("Handling workspace mode")

	// Create workspace instance
	workspace := ws.NewWorkspace(ws.NewWorkspaceParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})

	// Create worktree for workspace
	if err := workspace.CreateWorktree(branch); err != nil {
		return err
	}

	c.VerbosePrint("Workspace worktree creation completed successfully")
	return nil
}

// createWorkTreeFromIssue creates a worktree from a forge issue.
func (c *realCM) createWorkTreeFromIssue(branch string, issueRef string, ideName *string) error {
	c.VerbosePrint("Starting worktree creation from issue: %s", issueRef)

	// 1. Detect project mode (repository or workspace)
	projectType, err := c.detectProjectMode()
	if err != nil {
		c.VerbosePrint("Error: %v", err)
		return fmt.Errorf("failed to detect project mode: %w", err)
	}

	// 2. Handle based on project type
	var createErr error
	var branchName *string
	if branch != "" {
		branchName = &branch
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		createErr = c.createWorkTreeFromIssueForSingleRepo(branchName, issueRef)
	case ProjectTypeWorkspace:
		createErr = c.createWorkTreeFromIssueForWorkspace(branchName, issueRef)
	case ProjectTypeNone:
		return ErrNoGitRepositoryOrWorkspaceFound
	default:
		return fmt.Errorf("unknown project type")
	}

	// 3. Open IDE if specified and worktree creation was successful
	if branchName != nil {
		if err := c.handleIDEOpening(createErr, *branchName, ideName); err != nil {
			return err
		}
	}

	return createErr
}

// createWorkTreeFromIssueForSingleRepo creates a worktree from issue for single repository.
func (c *realCM) createWorkTreeFromIssueForSingleRepo(branchName *string, issueRef string) error {
	c.VerbosePrint("Creating worktree from issue for single repository mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.Logger)

	// Get the appropriate forge for the repository
	selectedForge, err := forgeManager.GetForgeForRepository(".")
	if err != nil {
		return fmt.Errorf("failed to get forge for repository: %w", err)
	}

	// Get issue information
	issueInfo, err := selectedForge.GetIssueInfo(issueRef)
	if err != nil {
		return c.translateIssueError(err)
	}

	// Generate branch name if not provided
	if branchName == nil || *branchName == "" {
		generatedBranchName := selectedForge.GenerateBranchName(issueInfo)
		branchName = &generatedBranchName
	}

	// Create worktree using existing logic
	repoInstance := repo.NewRepository(repo.NewRepositoryParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})
	return repoInstance.CreateWorktree(*branchName, repo.CreateWorktreeOpts{IssueInfo: issueInfo})
}

// createWorkTreeFromIssueForWorkspace creates worktrees from issue for workspace.
func (c *realCM) createWorkTreeFromIssueForWorkspace(branchName *string, issueRef string) error {
	c.VerbosePrint("Creating worktree from issue for workspace mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.Logger)

	// Get the appropriate forge for the repository
	selectedForge, err := forgeManager.GetForgeForRepository(".")
	if err != nil {
		return fmt.Errorf("failed to get forge for repository: %w", err)
	}

	// Get issue information
	issueInfo, err := selectedForge.GetIssueInfo(issueRef)
	if err != nil {
		return c.translateIssueError(err)
	}

	// Generate branch name if not provided
	if branchName == nil || *branchName == "" {
		generatedBranchName := selectedForge.GenerateBranchName(issueInfo)
		branchName = &generatedBranchName
	}

	// Create workspace instance
	workspace := ws.NewWorkspace(ws.NewWorkspaceParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})
	return workspace.CreateWorktree(*branchName)
}

// translateIssueError translates issue-related errors to preserve the original error types.
func (c *realCM) translateIssueError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific issue errors and preserve them
	if errors.Is(err, issue.ErrIssueNumberRequiresContext) {
		return issue.ErrIssueNumberRequiresContext
	}

	// Check for specific forge errors and preserve them
	if errors.Is(err, forge.ErrInvalidIssueRef) {
		return forge.ErrInvalidIssueRef
	}

	// Return the original error if no translation is needed
	return err
}
