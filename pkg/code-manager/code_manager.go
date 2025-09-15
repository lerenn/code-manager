package codemanager

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/status"
)

// CodeManager interface provides Git repository detection functionality.
type CodeManager interface {
	// CreateWorkTree executes the main application logic.
	CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error
	// DeleteWorkTree deletes a worktree for the specified branch.
	DeleteWorkTree(branch string, force bool, opts ...DeleteWorktreeOpts) error
	// DeleteWorkTrees deletes multiple worktrees for the specified branches.
	DeleteWorkTrees(branches []string, force bool) error
	// DeleteAllWorktrees deletes all worktrees for the current repository or workspace.
	DeleteAllWorktrees(force bool, opts ...DeleteAllWorktreesOpts) error
	// OpenWorktree opens an existing worktree in the specified IDE.
	OpenWorktree(worktreeName, ideName string, opts ...OpenWorktreeOpts) error
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
	// DeleteRepository deletes a repository and all associated resources.
	DeleteRepository(params DeleteRepositoryParams) error
	// CreateWorkspace creates a new workspace with repository selection.
	CreateWorkspace(params CreateWorkspaceParams) error
	// DeleteWorkspace deletes a workspace and all associated resources.
	DeleteWorkspace(params DeleteWorkspaceParams) error
	// ListWorkspaces lists all workspaces from the status file.
	ListWorkspaces() ([]WorkspaceInfo, error)
	// SetLogger sets the logger for this CM instance.
	SetLogger(logger logger.Logger)
}

// NewCodeManagerParams contains parameters for creating a new CodeManager instance.
type NewCodeManagerParams struct {
	Dependencies *dependencies.Dependencies
}

type realCodeManager struct {
	deps *dependencies.Dependencies
}

// NewCodeManager creates a new CodeManager instance.
func NewCodeManager(params NewCodeManagerParams) (CodeManager, error) {
	deps := params.Dependencies
	if deps == nil {
		deps = dependencies.New()
	}

	return &realCodeManager{
		deps: deps,
	}, nil
}

// VerbosePrint logs a formatted message using the current logger.
func (c *realCodeManager) VerbosePrint(msg string, args ...interface{}) {
	if c.deps.Logger != nil {
		c.deps.Logger.Logf(fmt.Sprintf(msg, args...))
	}
}

// SetLogger sets the logger for this CodeManager instance.
func (c *realCodeManager) SetLogger(logger logger.Logger) {
	c.deps.Logger = logger
}

// getConfig gets the configuration from the ConfigManager with fallback.
func (c *realCodeManager) getConfig() (config.Config, error) {
	return c.deps.Config.GetConfigWithFallback()
}

// BuildWorktreePath constructs a worktree path from repository URL, remote name, and branch.
func (c *realCodeManager) BuildWorktreePath(repoURL, remoteName, branch string) string {
	// Get config from ConfigManager
	cfg, err := c.getConfig()
	if err != nil {
		// Fallback to a default path if config cannot be loaded
		return fmt.Sprintf("~/Code/repos/%s/%s/%s", repoURL, remoteName, branch)
	}
	// Use the same path format as the worktree component
	return fmt.Sprintf("%s/%s/%s/%s", cfg.RepositoriesDir, repoURL, remoteName, branch)
}

// executeWithHooks executes an operation with pre and post hooks.
func (c *realCodeManager) executeWithHooks(
	operationName string, params map[string]interface{}, operation func() error) error {
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
func (c *realCodeManager) executeWithHooksAndReturnRepositories(
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
func (c *realCodeManager) executeWithHooksAndReturnWorkspaces(
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
func (c *realCodeManager) executeHooks(operationName string, ctx *hooks.HookContext, resultErr error) error {
	if c.deps.HookManager == nil {
		return nil
	}

	if resultErr != nil {
		return c.deps.HookManager.ExecuteErrorHooks(operationName, ctx)
	}
	return c.deps.HookManager.ExecutePostHooks(operationName, ctx)
}

// executePreHooks executes pre-hooks if hook manager is available.
func (c *realCodeManager) executePreHooks(operationName string, ctx *hooks.HookContext) error {
	if c.deps.HookManager == nil {
		return nil
	}
	return c.deps.HookManager.ExecutePreHooks(operationName, ctx)
}

// detectProjectMode detects the type of project (single repository or workspace).
func (c *realCodeManager) detectProjectMode(workspaceName, repositoryName string) (mode.Mode, error) {
	c.VerbosePrint("Detecting project mode...")

	// If workspaceName is provided, return workspace mode
	if workspaceName != "" {
		c.VerbosePrint("Workspace mode detected (workspace: %s)", workspaceName)
		return mode.ModeWorkspace, nil
	}

	// If repositoryName is provided, return single repository mode
	if repositoryName != "" {
		c.VerbosePrint("Repository mode detected (repository: %s)", repositoryName)
		return mode.ModeSingleRepo, nil
	}

	// Neither workspace nor repository specified, detect from current directory
	c.VerbosePrint("No specific workspace or repository provided, detecting from current directory")

	// Create repository instance to check if we're in a Git repository
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repository.NewRepositoryParams{
		Dependencies: c.deps,
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
