package cm

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// RepositoryProvider is a function type that creates repository instances.
type RepositoryProvider func(params repository.NewRepositoryParams) repository.Repository

// WorkspaceProvider is a function type that creates workspace instances.
type WorkspaceProvider func(params workspace.NewWorkspaceParams) workspace.Workspace

// CM interface provides Git repository detection functionality.
type CM interface {
	// CreateWorkTree executes the main application logic.
	CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error
	// DeleteWorkTree deletes a worktree for the specified branch.
	DeleteWorkTree(branch string, force bool) error
	// OpenWorktree opens an existing worktree in the specified IDE.
	OpenWorktree(worktreeName, ideName string) error
	// ListWorktrees lists worktrees for the current project with mode detection.
	ListWorktrees(force bool) ([]status.WorktreeInfo, ProjectType, error)
	// LoadWorktree loads a branch from a remote source and creates a worktree.
	LoadWorktree(branchArg string, opts ...LoadWorktreeOpts) error
	// Init initializes CM configuration.
	Init(opts InitOpts) error
	// Clone clones a repository and initializes it in CM.
	Clone(repoURL string, opts ...CloneOpts) error
	// ListRepositories lists all repositories from the status file with base path validation.
	ListRepositories() ([]RepositoryInfo, error)
	// SetLogger sets the logger for this CM instance.
	SetLogger(logger logger.Logger)
	// Hook management methods
	RegisterHook(operation string, hook hooks.Hook) error
	UnregisterHook(operation, hookName string) error
}

// NewCMParams contains parameters for creating a new CM instance.
type NewCMParams struct {
	RepositoryProvider RepositoryProvider
	WorkspaceProvider  WorkspaceProvider
	Config             config.Config
	HookManager        hooks.HookManagerInterface
	Status             status.Manager
	FS                 fs.FS
	Git                git.Git
	Logger             logger.Logger
	Prompt             prompt.Prompter
}

type realCM struct {
	fs            fs.FS
	git           git.Git
	config        config.Config
	statusManager status.Manager
	logger        logger.Logger
	prompt        prompt.Prompter
	repository    repository.Repository
	workspace     workspace.Workspace
	hookManager   hooks.HookManagerInterface
}

// NewCM creates a new CM instance.
func NewCM(params NewCMParams) (CM, error) {
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

	// Create repository and workspace instances using providers
	repoInstance := createRepositoryInstance(
		repoProvider, fsInstance, gitInstance, params.Config,
		statusInstance, loggerInstance, promptInstance,
	)
	workspaceInstance := createWorkspaceInstance(
		workspaceProvider, fsInstance, gitInstance, params.Config,
		statusInstance, loggerInstance, promptInstance,
	)

	// Use provided hook manager or create a new one
	hookManager := createHookManager(params.HookManager)
	cmInstance := &realCM{
		fs:            fsInstance,
		git:           gitInstance,
		config:        params.Config,
		statusManager: statusInstance,
		logger:        loggerInstance,
		prompt:        promptInstance,
		repository:    repoInstance,
		workspace:     workspaceInstance,
		hookManager:   hookManager,
	}
	return cmInstance, nil
}

// createHookManager creates a hook manager instance.
func createHookManager(providedHookManager hooks.HookManagerInterface) hooks.HookManagerInterface {
	if providedHookManager != nil {
		return providedHookManager
	}
	return hooks.NewHookManager()
}

// createRepositoryInstance creates a repository instance using the provided provider.
func createRepositoryInstance(
	repoProvider RepositoryProvider,
	fsInstance fs.FS,
	gitInstance git.Git,
	config config.Config,
	statusInstance status.Manager,
	loggerInstance logger.Logger,
	promptInstance prompt.Prompter,
) repository.Repository {
	return repoProvider(repository.NewRepositoryParams{
		FS:               fsInstance,
		Git:              gitInstance,
		Config:           config,
		StatusManager:    statusInstance,
		Logger:           loggerInstance,
		Prompt:           promptInstance,
		WorktreeProvider: worktree.NewWorktree,
	})
}

// createWorkspaceInstance creates a workspace instance using the provided provider.
func createWorkspaceInstance(
	workspaceProvider WorkspaceProvider,
	fsInstance fs.FS,
	gitInstance git.Git,
	config config.Config,
	statusInstance status.Manager,
	loggerInstance logger.Logger,
	promptInstance prompt.Prompter,
) workspace.Workspace {
	return workspaceProvider(workspace.NewWorkspaceParams{
		FS:               fsInstance,
		Git:              gitInstance,
		Config:           config,
		StatusManager:    statusInstance,
		Logger:           loggerInstance,
		Prompt:           promptInstance,
		WorktreeProvider: worktree.NewWorktree,
	})
}

// VerbosePrint logs a formatted message using the current logger.
func (c *realCM) VerbosePrint(msg string, args ...interface{}) {
	c.logger.Logf(fmt.Sprintf(msg, args...))
}

// SetLogger sets the logger for this CM instance.
func (c *realCM) SetLogger(logger logger.Logger) {
	c.logger = logger
	// Propagate logger setting to components
	c.repository.SetLogger(logger)
	c.workspace.SetLogger(logger)
}

// BuildWorktreePath constructs a worktree path from repository URL, remote name, and branch.
func (c *realCM) BuildWorktreePath(repoURL, _, branch string) string {
	// For now, construct the path manually since we don't have direct access to the worktree
	// This should be refactored to use the worktree component properly
	return fmt.Sprintf("%s/worktrees/%s/%s", c.config.BasePath, repoURL, branch)
}

// RegisterHook registers a hook for a specific operation.
func (c *realCM) RegisterHook(operation string, hook hooks.Hook) error {
	// This is a simplified implementation - in practice, you'd want to determine
	// the hook type and register it appropriately.
	switch h := hook.(type) {
	case hooks.PostHook:
		return c.hookManager.RegisterPostHook(operation, h)
	case hooks.PreHook:
		return c.hookManager.RegisterPreHook(operation, h)
	case hooks.ErrorHook:
		return c.hookManager.RegisterErrorHook(operation, h)
	default:
		return fmt.Errorf("unsupported hook type")
	}
}

// UnregisterHook removes a hook by name from a specific operation.
func (c *realCM) UnregisterHook(operation, hookName string) error {
	return c.hookManager.RemoveHook(operation, hookName)
}

// executeWithHooks executes an operation with pre and post hooks.
func (c *realCM) executeWithHooks(operationName string, params map[string]interface{}, operation func() error) error {
	ctx := &hooks.HookContext{
		OperationName: operationName,
		Parameters:    params,
		Results:       make(map[string]interface{}),
		CM:            c,
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

// executeWithHooksAndReturnListWorktrees executes an operation with pre and post hooks
// that returns worktrees and project type.
func (c *realCM) executeWithHooksAndReturnListWorktrees(
	operationName string,
	params map[string]interface{},
	operation func() ([]status.WorktreeInfo, ProjectType, error),
) ([]status.WorktreeInfo, ProjectType, error) {
	ctx := &hooks.HookContext{
		OperationName: operationName,
		Parameters:    params,
		Results:       make(map[string]interface{}),
		CM:            c,
		Metadata:      make(map[string]interface{}),
	}
	// Execute pre-hooks (if hook manager is available)
	if err := c.executePreHooks(operationName, ctx); err != nil {
		return nil, ProjectTypeNone, err
	}
	// Execute operation
	var worktrees []status.WorktreeInfo
	var projectType ProjectType
	var resultErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				resultErr = fmt.Errorf("panic in %s: %v", operationName, r)
			}
		}()
		worktrees, projectType, resultErr = operation()
	}()
	// Update context with results
	ctx.Error = resultErr
	if resultErr == nil {
		ctx.Results["worktrees"] = worktrees
		ctx.Results["projectType"] = projectType
		ctx.Results["success"] = true
	}
	// Execute post-hooks or error-hooks (if hook manager is available)
	if hookErr := c.executeHooks(operationName, ctx, resultErr); hookErr != nil {
		return nil, ProjectTypeNone, hookErr
	}
	return worktrees, projectType, resultErr
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
		CM:            c,
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
func (c *realCM) detectProjectMode() (ProjectType, error) {
	c.VerbosePrint("Detecting project mode...")
	// First, check if we're in a Git repository
	exists, err := c.repository.IsGitRepository()
	if err != nil {
		return ProjectTypeNone, fmt.Errorf("failed to check Git repository: %w", err)
	}
	if exists {
		c.VerbosePrint("Single repository mode detected")
		return ProjectTypeSingleRepo, nil
	}
	// If not a Git repository, check for workspace files
	workspaceFiles, err := c.fs.Glob("*.code-workspace")
	if err != nil {
		return ProjectTypeNone, fmt.Errorf("failed to detect workspace files: %w", err)
	}
	if len(workspaceFiles) > 0 {
		c.VerbosePrint("Workspace mode detected")
		return ProjectTypeWorkspace, nil
	}
	c.VerbosePrint("No project mode detected")
	return ProjectTypeNone, nil
}
