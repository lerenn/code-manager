package codemanager

import (
	"errors"
	"fmt"

	branchpkg "github.com/lerenn/code-manager/pkg/branch"
	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/forge"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/prompt"
)

// CreateWorkTreeOpts contains optional parameters for CreateWorkTree.
type CreateWorkTreeOpts struct {
	IDEName        string
	IssueRef       string
	WorkspaceName  string
	RepositoryName string
	Force          bool
	Remote         string // Remote name to use (defaults to "origin" if empty)
}

// CreateWorkTree executes the main application logic.
func (c *realCodeManager) CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error {
	// Extract and validate options
	options := c.extractCreateWorkTreeOptions(opts)

	// Validate that workspace and repository are not both specified
	if options.WorkspaceName != "" && options.RepositoryName != "" {
		return fmt.Errorf("cannot specify both WorkspaceName and RepositoryName")
	}

	// Handle interactive selection if neither workspace nor repository is specified
	if options.WorkspaceName == "" && options.RepositoryName == "" {
		result, err := c.promptSelectTargetOnly()
		if err != nil {
			return fmt.Errorf("failed to select target: %w", err)
		}

		switch result.Type {
		case prompt.TargetWorkspace:
			options.WorkspaceName = result.Name
		case prompt.TargetRepository:
			options.RepositoryName = result.Name
		default:
			return fmt.Errorf("invalid target type selected: %s", result.Type)
		}
	}

	// Handle interactive branch name input if not provided
	if branch == "" && options.IssueRef == "" {
		branchName, err := c.deps.Prompt.PromptForBranchName()
		if err != nil {
			return fmt.Errorf("failed to get branch name: %w", err)
		}
		branch = branchName
	}

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"branch":         branch,
		"issueRef":       options.IssueRef,
		"workspaceName":  options.WorkspaceName,
		"repositoryName": options.RepositoryName,
		"force":          options.Force,
	}
	if options.IDEName != "" {
		params["ideName"] = options.IDEName
	}

	// Execute with hooks
	return c.executeWithHooks(consts.CreateWorkTree, params, func() error {
		var worktreePath string
		var err error

		// Sanitize branch name
		sanitizedBranch, err := c.sanitizeBranchNameForCreation(branch, options.IssueRef)
		if err != nil {
			return err
		}

		// Log if branch name was sanitized
		if sanitizedBranch != branch && c.deps.Logger != nil {
			c.deps.Logger.Logf("Branch name sanitized: %s -> %s", branch, sanitizedBranch)
		}

		c.VerbosePrint("Starting CM execution for branch: %s (sanitized: %s)", branch, sanitizedBranch)

		// 1. First determine the mode (workspace or repository)
		projectType, err := c.detectProjectMode(options.WorkspaceName, options.RepositoryName)
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		// 2. Then handle creation based on mode and flags
		worktreePath, err = c.handleWorktreeCreation(handleWorktreeCreationParams{
			ProjectType:     projectType,
			SanitizedBranch: sanitizedBranch,
			IssueRef:        options.IssueRef,
			WorkspaceName:   options.WorkspaceName,
			RepositoryName:  options.RepositoryName,
			Options:         options,
			Opts:            opts,
		})
		if err != nil {
			return err
		}

		// Set worktreePath in params for the IDE opening hook
		params["worktreePath"] = worktreePath

		return nil
	})
}

// extractCreateWorkTreeOptions extracts and merges options from the variadic parameter.
func (c *realCodeManager) extractCreateWorkTreeOptions(opts []CreateWorkTreeOpts) CreateWorkTreeOpts {
	var result CreateWorkTreeOpts

	// Merge all provided options, with later options overriding earlier ones
	for _, opt := range opts {
		if opt.IssueRef != "" {
			result.IssueRef = opt.IssueRef
		}
		if opt.IDEName != "" {
			result.IDEName = opt.IDEName
		}
		if opt.WorkspaceName != "" {
			result.WorkspaceName = opt.WorkspaceName
		}
		if opt.RepositoryName != "" {
			result.RepositoryName = opt.RepositoryName
		}
		if opt.Force {
			result.Force = opt.Force
		}
	}

	return result
}

// sanitizeBranchNameForCreation sanitizes the branch name for worktree creation.
// It skips sanitization when using --from-issue with an empty branch name.
func (c *realCodeManager) sanitizeBranchNameForCreation(branch, issueRef string) (string, error) {
	if issueRef != "" && branch == "" {
		// When using --from-issue with empty branch, skip sanitization
		// The branch name will be generated from the issue
		return branch, nil
	}
	return branchpkg.SanitizeBranchName(branch)
}

// handleWorktreeCreationParams contains parameters for handleWorktreeCreation.
type handleWorktreeCreationParams struct {
	ProjectType     mode.Mode
	SanitizedBranch string
	IssueRef        string
	WorkspaceName   string
	RepositoryName  string
	Options         CreateWorkTreeOpts
	Opts            []CreateWorkTreeOpts
}

// handleWorktreeCreation handles worktree creation based on project type and flags.
func (c *realCodeManager) handleWorktreeCreation(params handleWorktreeCreationParams) (string, error) {
	switch params.ProjectType {
	case mode.ModeWorkspace:
		if params.IssueRef != "" {
			// Workspace mode with issue-based creation
			return c.createWorkTreeFromIssueForWorkspace(&params.SanitizedBranch, params.IssueRef, params.RepositoryName)
		}
		// Workspace mode with specific workspace name
		return c.createWorkTreeFromWorkspace(params.WorkspaceName, params.SanitizedBranch, params.Opts...)
	case mode.ModeSingleRepo:
		if params.IssueRef != "" {
			// Repository mode with issue-based creation
			return c.createWorkTreeFromIssueForSingleRepo(createWorkTreeFromIssueForSingleRepoParams{
				BranchName:     &params.SanitizedBranch,
				IssueRef:       params.IssueRef,
				RepositoryName: params.RepositoryName,
				Remote:         params.Options.Remote,
			})
		}
		// Repository mode with regular creation
		return c.handleRepositoryMode(params.SanitizedBranch, params.RepositoryName, params.Options.Remote)
	case mode.ModeNone:
		return "", ErrNoGitRepositoryOrWorkspaceFound
	default:
		return "", fmt.Errorf("unknown project type")
	}
}

// createWorkTreeFromWorkspace creates worktrees from workspace definition in status.yaml.
func (c *realCodeManager) createWorkTreeFromWorkspace(
	workspaceName, branch string, opts ...CreateWorkTreeOpts) (string, error) {
	c.VerbosePrint("Creating worktrees from workspace status: %s", workspaceName)

	// Create workspace instance
	workspaceProvider := c.deps.WorkspaceProvider
	workspaceInstance := workspaceProvider(ws.NewWorkspaceParams{
		Dependencies: c.deps,
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
func (c *realCodeManager) handleRepositoryMode(branch, repositoryName, remote string) (string, error) {
	c.VerbosePrint("Handling repository mode")

	// Create repository instance - let repositoryProvider handle repository name resolution
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: repositoryName, // Pass repository name directly, let provider handle resolution
	})

	// 1. Validate repository
	if err := repoInstance.Validate(); err != nil {
		return "", c.translateRepositoryError(err)
	}

	// 2. Create worktree for single repository
	worktreePath, err := repoInstance.CreateWorktree(branch, repo.CreateWorktreeOpts{Remote: remote})
	if err != nil {
		return "", c.translateRepositoryError(err)
	}

	c.VerbosePrint("CM execution completed successfully")

	return worktreePath, nil
}

// translateRepositoryError translates repository package errors to CM package errors.
func (c *realCodeManager) translateRepositoryError(err error) error {
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

// createWorkTreeFromIssueForSingleRepoParams contains parameters for createWorkTreeFromIssueForSingleRepo.
type createWorkTreeFromIssueForSingleRepoParams struct {
	BranchName     *string
	IssueRef       string
	RepositoryName string
	Remote         string
}

// createWorkTreeFromIssueForSingleRepo creates a worktree from issue for single repository.
func (c *realCodeManager) createWorkTreeFromIssueForSingleRepo(
	params createWorkTreeFromIssueForSingleRepoParams,
) (string, error) {
	c.VerbosePrint("Creating worktree from issue for single repository mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.deps.Logger, c.deps.StatusManager)

	// Get the appropriate forge for the repository
	selectedForge, err := forgeManager.GetForgeForRepository(params.RepositoryName)
	if err != nil {
		return "", fmt.Errorf("failed to get forge for repository: %w", err)
	}

	// Get issue information
	issueInfo, err := selectedForge.GetIssueInfo(params.IssueRef)
	if err != nil {
		return "", c.translateIssueError(err)
	}

	// Generate branch name if not provided
	if params.BranchName == nil || *params.BranchName == "" {
		generatedBranchName := selectedForge.GenerateBranchName(issueInfo)
		params.BranchName = &generatedBranchName
	}

	// Create worktree using existing logic
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: params.RepositoryName,
	})
	worktreePath, err := repoInstance.CreateWorktree(
		*params.BranchName, repo.CreateWorktreeOpts{IssueInfo: issueInfo, Remote: params.Remote})
	if err != nil {
		return "", err
	}
	return worktreePath, nil
}

// createWorkTreeFromIssueForWorkspace creates worktrees from issue for workspace.
func (c *realCodeManager) createWorkTreeFromIssueForWorkspace(
	branchName *string,
	issueRef, repositoryName string,
) (string, error) {
	c.VerbosePrint("Creating worktree from issue for workspace mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.deps.Logger, c.deps.StatusManager)

	// Get the appropriate forge for the repository
	selectedForge, err := forgeManager.GetForgeForRepository(repositoryName)
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
	workspaceProvider := c.deps.WorkspaceProvider
	workspaceInstance := workspaceProvider(ws.NewWorkspaceParams{
		Dependencies: c.deps,
	})
	worktreePath, err := workspaceInstance.CreateWorktree(*branchName)
	if err != nil {
		return "", err
	}
	return worktreePath, nil
}

// translateIssueError translates issue-related errors to preserve the original error types.
func (c *realCodeManager) translateIssueError(err error) error {
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
