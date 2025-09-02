package cm

import (
	"errors"
	"fmt"

	branchpkg "github.com/lerenn/code-manager/pkg/branch"
	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/forge"
	"github.com/lerenn/code-manager/pkg/issue"
	repo "github.com/lerenn/code-manager/pkg/repository"
	ws "github.com/lerenn/code-manager/pkg/workspace"
)

// CreateWorkTreeOpts contains optional parameters for CreateWorkTree.
type CreateWorkTreeOpts struct {
	IDEName  string
	IssueRef string
	Force    bool
}

// CreateWorkTree executes the main application logic.
func (c *realCM) CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error {
	// Extract and validate options
	issueRef, ideName, force := c.extractCreateWorkTreeOptions(opts)

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"branch":   branch,
		"issueRef": issueRef,
		"force":    force,
	}
	if ideName != nil {
		params["ideName"] = *ideName
	}

	// Execute with hooks
	return c.executeWithHooks(consts.CreateWorkTree, params, func() error {
		var worktreePath string
		var err error

		// Handle issue-based worktree creation
		if issueRef != "" {
			worktreePath, err = c.createWorkTreeFromIssue(branch, issueRef)
		} else {
			// Handle regular worktree creation
			worktreePath, err = c.createRegularWorkTree(branch, force)
		}

		if err != nil {
			return err
		}

		// Set worktreePath in params for the IDE opening hook
		params["worktreePath"] = worktreePath

		return nil
	})
}

// extractCreateWorkTreeOptions extracts options from the variadic parameter.
func (c *realCM) extractCreateWorkTreeOptions(opts []CreateWorkTreeOpts) (string, *string, bool) {
	var issueRef string
	var ideName *string
	var force bool

	if len(opts) > 0 {
		if opts[0].IssueRef != "" {
			issueRef = opts[0].IssueRef
		}
		if opts[0].IDEName != "" {
			ideName = &opts[0].IDEName
		}
		force = opts[0].Force
	}

	return issueRef, ideName, force
}

// createRegularWorkTree handles regular worktree creation (non-issue based).
func (c *realCM) createRegularWorkTree(branch string, force bool) (string, error) {
	// Sanitize branch name first
	sanitizedBranch, err := branchpkg.SanitizeBranchName(branch)
	if err != nil {
		return "", err
	}

	// Log if branch name was sanitized
	if sanitizedBranch != branch {
		c.logger.Logf("Branch name sanitized: %s -> %s", branch, sanitizedBranch)
	}

	c.VerbosePrint("Starting CM execution for branch: %s (sanitized: %s)", branch, sanitizedBranch)

	// Detect project mode and handle accordingly
	worktreePath, err := c.handleProjectMode(sanitizedBranch, force)
	if err != nil {
		return "", err
	}

	return worktreePath, nil
}

// handleProjectMode detects project mode and handles worktree creation accordingly.
func (c *realCM) handleProjectMode(sanitizedBranch string, force bool) (string, error) {
	projectType, err := c.detectProjectMode()
	if err != nil {
		c.VerbosePrint("Error: %v", err)
		return "", err
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		return c.handleRepositoryMode(sanitizedBranch)
	case ProjectTypeWorkspace:
		return c.handleWorkspaceMode(sanitizedBranch, force)
	case ProjectTypeNone:
		return "", ErrNoGitRepositoryOrWorkspaceFound
	default:
		return "", fmt.Errorf("unknown project type")
	}
}

// handleRepositoryMode handles repository mode: validation and worktree creation.
func (c *realCM) handleRepositoryMode(branch string) (string, error) {
	c.VerbosePrint("Handling repository mode")

	// 1. Validate repository
	if err := c.repository.Validate(); err != nil {
		return "", c.translateRepositoryError(err)
	}

	// 2. Create worktree for single repository
	worktreePath, err := c.repository.CreateWorktree(branch)
	if err != nil {
		return "", c.translateRepositoryError(err)
	}

	c.VerbosePrint("CM execution completed successfully")

	return worktreePath, nil
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
func (c *realCM) handleWorkspaceMode(branch string, force bool) (string, error) {
	c.VerbosePrint("Handling workspace mode")

	// Create worktree for workspace
	worktreePath, err := c.workspace.CreateWorktree(branch, force)
	if err != nil {
		return "", err
	}

	c.VerbosePrint("Workspace worktree creation completed successfully")
	return worktreePath, nil
}

// createWorkTreeFromIssue creates a worktree from a forge issue.
func (c *realCM) createWorkTreeFromIssue(branch string, issueRef string) (string, error) {
	c.VerbosePrint("Starting worktree creation from issue: %s", issueRef)

	// 1. Detect project mode (repository or workspace)
	projectType, err := c.detectProjectMode()
	if err != nil {
		c.VerbosePrint("Error: %v", err)
		return "", fmt.Errorf("failed to detect project mode: %w", err)
	}

	// 2. Handle based on project type
	var createErr error
	var worktreePath string

	switch projectType {
	case ProjectTypeSingleRepo:
		worktreePath, createErr = c.createWorkTreeFromIssueForSingleRepo(&branch, issueRef)
	case ProjectTypeWorkspace:
		worktreePath, createErr = c.createWorkTreeFromIssueForWorkspace(&branch, issueRef)
	case ProjectTypeNone:
		return "", ErrNoGitRepositoryOrWorkspaceFound
	default:
		return "", fmt.Errorf("unknown project type")
	}

	// 3. Check for worktree creation errors first
	if createErr != nil {
		return "", createErr
	}

	return worktreePath, nil
}

// createWorkTreeFromIssueForSingleRepo creates a worktree from issue for single repository.
func (c *realCM) createWorkTreeFromIssueForSingleRepo(branchName *string, issueRef string) (string, error) {
	c.VerbosePrint("Creating worktree from issue for single repository mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.logger)

	// Get the appropriate forge for the repository
	selectedForge, err := forgeManager.GetForgeForRepository(".")
	if err != nil {
		return "", fmt.Errorf("failed to get forge for repository: %w", err)
	}

	// Get issue information
	issueInfo, err := selectedForge.GetIssueInfo(issueRef)
	if err != nil {
		return "", c.translateIssueError(err)
	}

	// Generate branch name if not provided
	if branchName == nil || *branchName == "" {
		generatedBranchName := selectedForge.GenerateBranchName(issueInfo)
		branchName = &generatedBranchName
	}

	// Create worktree using existing logic
	repoInstance := repo.NewRepository(repo.NewRepositoryParams{
		FS:            c.fs,
		Git:           c.git,
		Config:        c.config,
		StatusManager: c.statusManager,
		Logger:        c.logger,
		Prompt:        c.prompt,
	})
	worktreePath, err := repoInstance.CreateWorktree(*branchName, repo.CreateWorktreeOpts{IssueInfo: issueInfo})
	if err != nil {
		return "", err
	}
	return worktreePath, nil
}

// createWorkTreeFromIssueForWorkspace creates worktrees from issue for workspace.
func (c *realCM) createWorkTreeFromIssueForWorkspace(branchName *string, issueRef string) (string, error) {
	c.VerbosePrint("Creating worktree from issue for workspace mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.logger)

	// Get the appropriate forge for the repository
	selectedForge, err := forgeManager.GetForgeForRepository(".")
	if err != nil {
		return "", fmt.Errorf("failed to get forge for repository: %w", err)
	}

	// Get issue information
	issueInfo, err := selectedForge.GetIssueInfo(issueRef)
	if err != nil {
		return "", c.translateIssueError(err)
	}

	// Generate branch name if not provided
	if branchName == nil || *branchName == "" {
		generatedBranchName := selectedForge.GenerateBranchName(issueInfo)
		branchName = &generatedBranchName
	}

	// Create workspace instance
	workspace := ws.NewWorkspace(ws.NewWorkspaceParams{
		FS:            c.fs,
		Git:           c.git,
		Config:        c.config,
		StatusManager: c.statusManager,
		Logger:        c.logger,
		Prompt:        c.prompt,
	})
	worktreePath, err := workspace.CreateWorktree(*branchName, false)
	if err != nil {
		return "", err
	}
	return worktreePath, nil
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
