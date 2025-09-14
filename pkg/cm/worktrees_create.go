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
	IDEName        string
	IssueRef       string
	WorkspaceName  string
	RepositoryName string
	Force          bool
	Remote         string // Remote name to use (defaults to "origin" if empty)
}

// CreateWorkTree executes the main application logic.
func (c *realCM) CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error {
	// Extract and validate options
	options := c.extractCreateWorkTreeOptions(opts)

	// Validate that workspace and repository are not both specified
	if options.WorkspaceName != "" && options.RepositoryName != "" {
		return fmt.Errorf("cannot specify both WorkspaceName and RepositoryName")
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
func (c *realCM) extractCreateWorkTreeOptions(opts []CreateWorkTreeOpts) CreateWorkTreeOpts {
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
func (c *realCM) handleWorktreeCreation(params handleWorktreeCreationParams) (string, error) {
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
func (c *realCM) handleRepositoryMode(branch, repositoryName, remote string) (string, error) {
	c.VerbosePrint("Handling repository mode")

	// Create repository instance - let repositoryProvider handle repository name resolution
	repoInstance := c.repositoryProvider(repo.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
		RepositoryName:   repositoryName, // Pass repository name directly, let provider handle resolution
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

// createWorkTreeFromIssueForSingleRepoParams contains parameters for createWorkTreeFromIssueForSingleRepo.
type createWorkTreeFromIssueForSingleRepoParams struct {
	BranchName     *string
	IssueRef       string
	RepositoryName string
	Remote         string
}

// createWorkTreeFromIssueForSingleRepo creates a worktree from issue for single repository.
func (c *realCM) createWorkTreeFromIssueForSingleRepo(
	params createWorkTreeFromIssueForSingleRepoParams,
) (string, error) {
	c.VerbosePrint("Creating worktree from issue for single repository mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.logger, c.statusManager)

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
	repoInstance := repo.NewRepository(repo.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
		RepositoryName:   params.RepositoryName,
	})
	worktreePath, err := repoInstance.CreateWorktree(
		*params.BranchName, repo.CreateWorktreeOpts{IssueInfo: issueInfo, Remote: params.Remote})
	if err != nil {
		return "", err
	}
	return worktreePath, nil
}

// createWorkTreeFromIssueForWorkspace creates worktrees from issue for workspace.
func (c *realCM) createWorkTreeFromIssueForWorkspace(
	branchName *string,
	issueRef, repositoryName string,
) (string, error) {
	c.VerbosePrint("Creating worktree from issue for workspace mode")

	// Create forge manager
	forgeManager := forge.NewManager(c.logger, c.statusManager)

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
