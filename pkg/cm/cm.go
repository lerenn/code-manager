package cm

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	defaulthooks "github.com/lerenn/code-manager/pkg/hooks/default"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// RepositoryProvider is a function type that creates repository instances.
type RepositoryProvider func(params repository.NewRepositoryParams) repository.Repository

// WorkspaceProvider is a function type that creates workspace instances.
type WorkspaceProvider func(params workspace.NewWorkspaceParams) workspace.Workspace

// WorktreeProvider is a function type that creates worktree instances.
type WorktreeProvider func(params worktree.NewWorktreeParams) worktree.Worktree

// CM interface provides Git repository detection functionality.
type CM interface {
	// CreateWorkTree executes the main application logic.
	CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error
	// DeleteWorkTree deletes a worktree for the specified branch.
	DeleteWorkTree(branch string, force bool, opts ...DeleteWorktreeOpts) error
	// DeleteWorkTrees deletes multiple worktrees for the specified branches.
	DeleteWorkTrees(branches []string, force bool) error
	// OpenWorktree opens an existing worktree in the specified IDE.
	OpenWorktree(worktreeName, ideName string) error
	// ListWorktrees lists worktrees for a workspace or repository.
	ListWorktrees(opts ...ListWorktreesOpts) ([]status.WorktreeInfo, error)
	// LoadWorktree loads a branch from a remote source and creates a worktree.
	LoadWorktree(branchArg string, opts ...LoadWorktreeOpts) error
	// Init initializes CM configuration.
	Init(opts InitOpts) error
	// Clone clones a repository and initializes it in CM.
	Clone(repoURL string, opts ...CloneOpts) error
	// ListRepositories lists all repositories from the status file with base path validation.
	ListRepositories() ([]RepositoryInfo, error)
	// CreateWorkspace creates a new workspace with repository selection.
	CreateWorkspace(params CreateWorkspaceParams) error
	// DeleteWorkspace deletes a workspace and all associated resources.
	DeleteWorkspace(params DeleteWorkspaceParams) error
	// ListWorkspaces lists all workspaces from the status file.
	ListWorkspaces() ([]WorkspaceInfo, error)
	// SetLogger sets the logger for this CM instance.
	SetLogger(logger logger.Logger)
}

// NewCMParams contains parameters for creating a new CM instance.
type NewCMParams struct {
	RepositoryProvider RepositoryProvider
	WorkspaceProvider  WorkspaceProvider
	WorktreeProvider   WorktreeProvider
	Config             config.Config
	Hooks              hooks.HookManagerInterface
	Status             status.Manager
	FS                 fs.FS
	Git                git.Git
	Logger             logger.Logger
	Prompt             prompt.Prompter
}

type realCM struct {
	fs                 fs.FS
	git                git.Git
	config             config.Config
	statusManager      status.Manager
	logger             logger.Logger
	prompt             prompt.Prompter
	repositoryProvider RepositoryProvider
	workspaceProvider  WorkspaceProvider
	worktreeProvider   WorktreeProvider
	hookManager        hooks.HookManagerInterface
}

// NewCM creates a new CM instance.
func NewCM(params NewCMParams) (CM, error) {
	instances := createInstances(params)
	hookManager := createHookManager(params.Hooks)

	return &realCM{
		fs:                 instances.fs,
		git:                instances.git,
		config:             params.Config,
		statusManager:      instances.status,
		logger:             instances.logger,
		prompt:             instances.prompt,
		repositoryProvider: instances.repoProvider,
		workspaceProvider:  instances.workspaceProvider,
		worktreeProvider:   instances.worktreeProvider,
		hookManager:        hookManager,
	}, nil
}

// createHookManager creates a hook manager instance.
func createHookManager(providedHookManager hooks.HookManagerInterface) hooks.HookManagerInterface {
	if providedHookManager != nil {
		return providedHookManager
	}

	// Use default hooks manager which includes IDE opening hooks
	defaultHookManager, err := defaulthooks.NewDefaultHooksManager()
	if err != nil {
		// Fallback to basic hook manager if default hooks fail to initialize
		return hooks.NewHookManager()
	}
	return defaultHookManager
}

// cmInstances holds the created instances for CM.
type cmInstances struct {
	fs                fs.FS
	git               git.Git
	logger            logger.Logger
	prompt            prompt.Prompter
	status            status.Manager
	repoProvider      RepositoryProvider
	workspaceProvider WorkspaceProvider
	worktreeProvider  WorktreeProvider
}

// createInstances creates and initializes all required instances for CM.
func createInstances(params NewCMParams) cmInstances {
	fsInstance := params.FS
	if fsInstance == nil {
		fsInstance = fs.NewFS()
	}

	gitInstance := params.Git
	if gitInstance == nil {
		gitInstance = git.NewGit()
	}

	loggerInstance := params.Logger
	if loggerInstance == nil {
		loggerInstance = logger.NewNoopLogger()
	}

	promptInstance := params.Prompt
	if promptInstance == nil {
		promptInstance = prompt.NewPrompt()
	}

	statusInstance := params.Status
	if statusInstance == nil {
		statusInstance = status.NewManager(fsInstance, params.Config)
	}

	repoProvider := params.RepositoryProvider
	if repoProvider == nil {
		repoProvider = repository.NewRepository
	}

	workspaceProvider := params.WorkspaceProvider
	if workspaceProvider == nil {
		workspaceProvider = workspace.NewWorkspace
	}

	worktreeProvider := params.WorktreeProvider
	if worktreeProvider == nil {
		worktreeProvider = worktree.NewWorktree
	}

	return cmInstances{
		fs:                fsInstance,
		git:               gitInstance,
		logger:            loggerInstance,
		prompt:            promptInstance,
		status:            statusInstance,
		repoProvider:      repoProvider,
		workspaceProvider: workspaceProvider,
		worktreeProvider:  worktreeProvider,
	}
}

// VerbosePrint logs a formatted message using the current logger.
func (c *realCM) VerbosePrint(msg string, args ...interface{}) {
	if c.logger != nil {
		c.logger.Logf(fmt.Sprintf(msg, args...))
	}
}

// SetLogger sets the logger for this CM instance.
func (c *realCM) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// BuildWorktreePath constructs a worktree path from repository URL, remote name, and branch.
func (c *realCM) BuildWorktreePath(repoURL, remoteName, branch string) string {
	// Use the same path format as the worktree component
	return fmt.Sprintf("%s/%s/%s/%s", c.config.RepositoriesDir, repoURL, remoteName, branch)
}

// executeWithHooks executes an operation with pre and post hooks.
func (c *realCM) executeWithHooks(operationName string, params map[string]interface{}, operation func() error) error {
	ctx := &hooks.HookContext{
		OperationName: operationName,
		Parameters:    params,
		Results:       make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
	}
	// Execute pre-hooks (if hook manager is available)
	if err := c.executePreHooks(operationName, ctx); err != nil {
		return err
	}
	// Execute operation
	var resultErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				resultErr = fmt.Errorf("panic in %s: %v", operationName, r)
			}
		}()
		resultErr = operation()
	}()
	// Update context with results
	ctx.Error = resultErr
	if resultErr == nil {
		ctx.Results["success"] = true
	}
	// Execute post-hooks or error-hooks (if hook manager is available)
	if hookErr := c.executeHooks(operationName, ctx, resultErr); hookErr != nil {
		return hookErr
	}
	return resultErr
}

// executeWithHooksAndReturnRepositories executes an operation with pre and post hooks that returns repositories.
func (c *realCM) executeWithHooksAndReturnRepositories(
	operationName string,
	params map[string]interface{},
	operation func() ([]RepositoryInfo, error),
) ([]RepositoryInfo, error) {
	ctx := &hooks.HookContext{
		OperationName: operationName,
		Parameters:    params,
		Results:       make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
	}
	// Execute pre-hooks (if hook manager is available)
	if err := c.executePreHooks(operationName, ctx); err != nil {
		return nil, err
	}
	// Execute operation
	var repositories []RepositoryInfo
	var resultErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				resultErr = fmt.Errorf("panic in %s: %v", operationName, r)
			}
		}()
		repositories, resultErr = operation()
	}()
	// Update context with results
	ctx.Error = resultErr
	if resultErr == nil {
		ctx.Results["repositories"] = repositories
		ctx.Results["success"] = true
	}
	// Execute post-hooks or error-hooks (if hook manager is available)
	if hookErr := c.executeHooks(operationName, ctx, resultErr); hookErr != nil {
		return nil, hookErr
	}
	return repositories, resultErr
}

// executeWithHooksAndReturnWorkspaces executes an operation with pre and post hooks that returns workspaces.
func (c *realCM) executeWithHooksAndReturnWorkspaces(
	operationName string,
	params map[string]interface{},
	operation func() ([]WorkspaceInfo, error),
) ([]WorkspaceInfo, error) {
	ctx := &hooks.HookContext{
		OperationName: operationName,
		Parameters:    params,
		Results:       make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
	}
	// Execute pre-hooks (if hook manager is available)
	if err := c.executePreHooks(operationName, ctx); err != nil {
		return nil, err
	}
	// Execute operation
	var workspaces []WorkspaceInfo
	var resultErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				resultErr = fmt.Errorf("panic in %s: %v", operationName, r)
			}
		}()
		workspaces, resultErr = operation()
	}()
	// Update context with results
	ctx.Error = resultErr
	if resultErr == nil {
		ctx.Results["workspaces"] = workspaces
		ctx.Results["success"] = true
	}
	// Execute post-hooks or error-hooks (if hook manager is available)
	if hookErr := c.executeHooks(operationName, ctx, resultErr); hookErr != nil {
		return nil, hookErr
	}
	return workspaces, resultErr
}

// executeHooks executes pre-hooks, post-hooks, or error-hooks based on the operation result.
func (c *realCM) executeHooks(operationName string, ctx *hooks.HookContext, resultErr error) error {
	if c.hookManager == nil {
		return nil
	}

	if resultErr != nil {
		return c.hookManager.ExecuteErrorHooks(operationName, ctx)
	}
	return c.hookManager.ExecutePostHooks(operationName, ctx)
}

// executePreHooks executes pre-hooks if hook manager is available.
func (c *realCM) executePreHooks(operationName string, ctx *hooks.HookContext) error {
	if c.hookManager == nil {
		return nil
	}
	return c.hookManager.ExecutePreHooks(operationName, ctx)
}

// detectProjectMode detects the type of project (single repository or workspace).
func (c *realCM) detectProjectMode(workspaceName string) (mode.Mode, error) {
	c.VerbosePrint("Detecting project mode...")

	// If workspaceName is provided, return workspace mode
	if workspaceName != "" {
		c.VerbosePrint("Workspace mode detected (workspace: %s)", workspaceName)
		return mode.ModeWorkspace, nil
	}

	// Create repository instance to check if we're in a Git repository
	repoInstance := c.repositoryProvider(repository.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
	})

	// Check if we're in a Git repository
	exists, err := repoInstance.IsGitRepository()
	if err != nil {
		return mode.ModeNone, fmt.Errorf("failed to check Git repository: %w", err)
	}
	if exists {
		c.VerbosePrint("Single repository mode detected")
		return mode.ModeSingleRepo, nil
	}

	c.VerbosePrint("No project mode detected")
	return mode.ModeNone, nil
}
