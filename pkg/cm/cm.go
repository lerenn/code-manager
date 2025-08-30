package cm

import (
	"fmt"

	basepkg "github.com/lerenn/code-manager/internal/base"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/lerenn/code-manager/pkg/worktree"
)

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
	// SetVerbose enables or disables verbose mode.
	SetVerbose(verbose bool)
	// Hook management methods
	RegisterHook(operation string, hook hooks.Hook) error
	UnregisterHook(operation, hookName string) error
}

// NewCMParams contains parameters for creating a new CM instance.
type NewCMParams struct {
	Repository  repository.Repository
	Workspace   workspace.Workspace
	Config      *config.Config
	HookManager hooks.HookManagerInterface // Optional: for testing with mocked hooks
}
type realCM struct {
	*basepkg.Base
	repository  repository.Repository
	workspace   workspace.Workspace
	hookManager hooks.HookManagerInterface
}

// NewCM creates a new CM instance.
func NewCM(cfg *config.Config) (CM, error) {
	fsInstance := fs.NewFS()
	gitInstance := git.NewGit()
	loggerInstance := logger.NewNoopLogger()
	promptInstance := prompt.NewPrompt()
	worktreeInstance := worktree.NewWorktree(worktree.NewWorktreeParams{
		FS:            fsInstance,
		Git:           gitInstance,
		StatusManager: status.NewManager(fsInstance, cfg),
		Logger:        loggerInstance,
		Prompt:        promptInstance,
		BasePath:      cfg.BasePath,
		Verbose:       false,
	})
	// Create repository and workspace instances
	repoInstance := repository.NewRepository(repository.NewRepositoryParams{
		FS:            fsInstance,
		Git:           gitInstance,
		Config:        cfg,
		StatusManager: status.NewManager(fsInstance, cfg),
		Logger:        loggerInstance,
		Prompt:        promptInstance,
		Worktree:      worktreeInstance,
		Verbose:       false,
	})
	workspaceInstance := workspace.NewWorkspace(workspace.NewWorkspaceParams{
		FS:            fsInstance,
		Git:           gitInstance,
		Config:        cfg,
		StatusManager: status.NewManager(fsInstance, cfg),
		Logger:        loggerInstance,
		Prompt:        promptInstance,
		Worktree:      worktreeInstance,
		Verbose:       false,
	})
	cmInstance := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            fsInstance,
			Git:           gitInstance,
			Config:        cfg,
			StatusManager: status.NewManager(fsInstance, cfg),
			Logger:        loggerInstance,
			Prompt:        promptInstance,
			Verbose:       false,
		}),
		repository:  repoInstance,
		workspace:   workspaceInstance,
		hookManager: hooks.NewHookManager(),
	}
	// Setup hooks for the CM instance
	if err := cmInstance.setupHooks(); err != nil {
		return nil, err
	}
	return cmInstance, nil
}

// setupHooks configures and registers all hooks for the CM instance.
func (c *realCM) setupHooks() error {
	// Register IDE opening hook
	if err := ide.NewOpeningHook().RegisterForOperations(c.RegisterHook); err != nil {
		return err
	}
	return nil
}

// NewCMWithDependencies creates a new CM instance with custom repository and workspace dependencies.
// This is primarily used for testing with mocked dependencies.
func NewCMWithDependencies(params NewCMParams) CM {
	fsInstance := fs.NewFS()
	gitInstance := git.NewGit()
	loggerInstance := logger.NewNoopLogger()
	statusInstance := status.NewManager(fsInstance, params.Config)
	// Use provided hook manager or create a new one
	var hookManager hooks.HookManagerInterface
	if params.HookManager != nil {
		hookManager = params.HookManager
	} else {
		hookManager = hooks.NewHookManager()
	}

	return &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            fsInstance,
			Git:           gitInstance,
			Config:        params.Config,
			StatusManager: statusInstance,
			Logger:        loggerInstance,
			Prompt:        prompt.NewPrompt(),
			Verbose:       false,
		}),
		repository:  params.Repository,
		workspace:   params.Workspace,
		hookManager: hookManager,
	}
}
func (c *realCM) SetVerbose(verbose bool) {
	// Create a new Base with the updated verbose setting
	newBase := basepkg.NewBase(basepkg.NewBaseParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       verbose,
	})
	c.Base = newBase
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
	if c.hookManager != nil {
		if err := c.hookManager.ExecutePreHooks(operationName, ctx); err != nil {
			return err
		}
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
	if c.hookManager != nil {
		if resultErr != nil {
			if hookErr := c.hookManager.ExecuteErrorHooks(operationName, ctx); hookErr != nil {
				return hookErr
			}
		} else {
			if hookErr := c.hookManager.ExecutePostHooks(operationName, ctx); hookErr != nil {
				return hookErr
			}
		}
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
	if c.hookManager != nil {
		if err := c.hookManager.ExecutePreHooks(operationName, ctx); err != nil {
			return nil, ProjectTypeNone, err
		}
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
	if c.hookManager != nil {
		if resultErr != nil {
			if hookErr := c.hookManager.ExecuteErrorHooks(operationName, ctx); hookErr != nil {
				return nil, ProjectTypeNone, hookErr
			}
		} else {
			if hookErr := c.hookManager.ExecutePostHooks(operationName, ctx); hookErr != nil {
				return nil, ProjectTypeNone, hookErr
			}
		}
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
	if c.hookManager != nil {
		if err := c.hookManager.ExecutePreHooks(operationName, ctx); err != nil {
			return nil, err
		}
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
	if c.hookManager != nil {
		if resultErr != nil {
			if hookErr := c.hookManager.ExecuteErrorHooks(operationName, ctx); hookErr != nil {
				return nil, hookErr
			}
		} else {
			if hookErr := c.hookManager.ExecutePostHooks(operationName, ctx); hookErr != nil {
				return nil, hookErr
			}
		}
	}
	return repositories, resultErr
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
	workspaceFiles, err := c.FS.Glob("*.code-workspace")
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
