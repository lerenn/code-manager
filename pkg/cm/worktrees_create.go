package cm

import (
	"errors"
	"fmt"

	branchpkg "github.com/lerenn/code-manager/pkg/branch"
	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/forge"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// CreateWorkTreeOpts contains optional parameters for CreateWorkTree.
type CreateWorkTreeOpts struct {
	IDEName       string
	IssueRef      string
	WorkspaceName string
	Force         bool
}

// CreateWorkTree executes the main application logic.
func (c *realCM) CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error {
	// Extract and validate options
	issueRef, ideName, workspaceName, force := c.extractCreateWorkTreeOptions(opts)

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"branch":        branch,
		"issueRef":      issueRef,
		"workspaceName": workspaceName,
		"force":         force,
	}
	if ideName != nil {
		params["ideName"] = *ideName
	}

	// Execute with hooks
	return c.executeWithHooks(consts.CreateWorkTree, params, func() error {
		var worktreePath string
		var err error

		// Sanitize branch name first
		sanitizedBranch, err := branchpkg.SanitizeBranchName(branch)
		if err != nil {
			return err
		}

		// Log if branch name was sanitized
		if sanitizedBranch != branch && c.logger != nil {
			c.logger.Logf("Branch name sanitized: %s -> %s", branch, sanitizedBranch)
		}

		c.VerbosePrint("Starting CM execution for branch: %s (sanitized: %s)", branch, sanitizedBranch)

		// 1. First determine the mode (workspace or repository)
		projectType, err := c.detectProjectMode()
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		// 2. Then handle creation based on mode and flags
		worktreePath, err = c.handleWorktreeCreation(projectType, sanitizedBranch, issueRef, workspaceName, opts...)
		if err != nil {
			return err
		}

		// Set worktreePath in params for the IDE opening hook
		params["worktreePath"] = worktreePath

		return nil
	})
}

// extractCreateWorkTreeOptions extracts options from the variadic parameter.
func (c *realCM) extractCreateWorkTreeOptions(opts []CreateWorkTreeOpts) (string, *string, string, bool) {
	var issueRef string
	var ideName *string
	var workspaceName string
	var force bool

	if len(opts) > 0 {
		if opts[0].IssueRef != "" {
			issueRef = opts[0].IssueRef
		}
		if opts[0].IDEName != "" {
			ideName = &opts[0].IDEName
		}
		if opts[0].WorkspaceName != "" {
			workspaceName = opts[0].WorkspaceName
		}
		force = opts[0].Force
	}

	return issueRef, ideName, workspaceName, force
}

// handleWorktreeCreation handles worktree creation based on project type and flags.
func (c *realCM) handleWorktreeCreation(
	projectType mode.Mode,
	sanitizedBranch, issueRef, workspaceName string,
	opts ...CreateWorkTreeOpts,
) (string, error) {
	switch projectType {
	case mode.ModeWorkspace:
		return c.handleWorktreeCreationInWorkspace(sanitizedBranch, issueRef, workspaceName, opts...)
	case mode.ModeSingleRepo:
		return c.handleWorktreeCreationInRepository(sanitizedBranch, issueRef, opts...)
	case mode.ModeNone:
		return "", ErrNoGitRepositoryOrWorkspaceFound
	default:
		return "", fmt.Errorf("unknown project type")
	}
}

// handleWorktreeCreationInWorkspace handles worktree creation in workspace mode.
func (c *realCM) handleWorktreeCreationInWorkspace(
	sanitizedBranch, issueRef, workspaceName string,
	opts ...CreateWorkTreeOpts,
) (string, error) {
	if issueRef != "" {
		// Workspace mode with issue-based creation
		// TODO: Implement workspace issue-based creation
		return "", fmt.Errorf("workspace issue-based creation not yet implemented")
	}
	// Workspace mode with specific workspace name
	return c.createWorkTreeFromWorkspace(workspaceName, sanitizedBranch, opts...)
}

// handleWorktreeCreationInRepository handles worktree creation in repository mode.
func (c *realCM) handleWorktreeCreationInRepository(
	sanitizedBranch, issueRef string,
	_ ...CreateWorkTreeOpts,
) (string, error) {
	if issueRef != "" {
		// Repository mode with issue-based creation
		return c.createWorkTreeFromIssue(sanitizedBranch, issueRef)
	}
	// Repository mode with regular creation
	return c.handleRepositoryMode(sanitizedBranch)
}

// createWorkTreeFromWorkspace creates worktrees from workspace definition in status.yaml.
func (c *realCM) createWorkTreeFromWorkspace(workspaceName, branch string, opts ...CreateWorkTreeOpts) (string, error) {
	c.VerbosePrint("Creating worktrees from workspace status: %s", workspaceName)

	// Create workspace instance
	workspaceInstance := c.workspaceProvider(ws.NewWorkspaceParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
	})

	// Convert CM options to workspace options
	var workspaceOpts []ws.CreateWorktreeOpts
	if len(opts) > 0 {
		workspaceOpts = append(workspaceOpts, ws.CreateWorktreeOpts{
			IDEName:       opts[0].IDEName,
			IssueInfo:     nil, // TODO: Handle issue info if needed
			WorkspaceName: workspaceName,
		})
	} else {
		workspaceOpts = append(workspaceOpts, ws.CreateWorktreeOpts{
			WorkspaceName: workspaceName,
		})
	}

	// Use workspace package method
	return workspaceInstance.CreateWorktree(branch, workspaceOpts...)
}

// handleRepositoryMode handles repository mode: validation and worktree creation.
func (c *realCM) handleRepositoryMode(branch string) (string, error) {
	c.VerbosePrint("Handling repository mode")

	// Create repository instance
	repoInstance := c.repositoryProvider(repo.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
	})

	// 1. Validate repository
	if err := repoInstance.Validate(); err != nil {
		return "", c.translateRepositoryError(err)
	}

	// 2. Create worktree for single repository
	worktreePath, err := repoInstance.CreateWorktree(branch, repo.CreateWorktreeOpts{})
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
	case mode.ModeSingleRepo:
		worktreePath, createErr = c.createWorkTreeFromIssueForSingleRepo(&branch, issueRef)
	case mode.ModeWorkspace:
		worktreePath, createErr = c.createWorkTreeFromIssueForWorkspace(&branch, issueRef)
	case mode.ModeNone:
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
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
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
	workspace := c.workspaceProvider(ws.NewWorkspaceParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
	})
	worktreePath, err := workspace.CreateWorktree(*branchName)
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
