// Package dependencies provides a centralized dependency container for the CM application.
// This package follows Go idioms for dependency injection by grouping related dependencies
// together and providing a fluent API for configuration.
package dependencies

import (
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
	repositoryinterfaces "github.com/lerenn/code-manager/pkg/mode/repository/interfaces"
	workspaceinterfaces "github.com/lerenn/code-manager/pkg/mode/workspace/interfaces"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	worktreeinterfaces "github.com/lerenn/code-manager/pkg/worktree/interfaces"
)

// Dependencies holds shared dependencies across the application.
// This follows the Go idiom of grouping related data together.
type Dependencies struct {
	FS                 fs.FS
	Git                git.Git
	Config             config.Manager
	StatusManager      status.Manager
	Logger             logger.Logger
	Prompt             prompt.Prompter
	HookManager        hooks.HookManagerInterface
	RepositoryProvider repositoryinterfaces.RepositoryProvider
	WorkspaceProvider  workspaceinterfaces.WorkspaceProvider
	WorktreeProvider   worktreeinterfaces.WorktreeProvider
}

// New creates a new Dependencies instance with sensible defaults.
// This follows Go's convention of New* functions for constructors.
func New() *Dependencies {
	return &Dependencies{
		FS:          fs.NewFS(),
		Git:         git.NewGit(),
		Logger:      logger.NewNoopLogger(),
		Prompt:      prompt.NewPrompt(),
		HookManager: hooks.NewHookManager(),
		// Note: Config, StatusManager, and Providers are intentionally left nil
		// as they require specific configuration or are set via With* methods
	}
}

// WithFS sets the filesystem and returns the instance for chaining.
func (d *Dependencies) WithFS(fs fs.FS) *Dependencies {
	d.FS = fs
	return d
}

// WithGit sets the git instance and returns the instance for chaining.
func (d *Dependencies) WithGit(git git.Git) *Dependencies {
	d.Git = git
	return d
}

// WithConfig sets the config manager and returns the instance for chaining.
func (d *Dependencies) WithConfig(cfg config.Manager) *Dependencies {
	d.Config = cfg
	return d
}

// WithStatusManager sets the status manager and returns the instance for chaining.
func (d *Dependencies) WithStatusManager(sm status.Manager) *Dependencies {
	d.StatusManager = sm
	return d
}

// WithLogger sets the logger and returns the instance for chaining.
func (d *Dependencies) WithLogger(logger logger.Logger) *Dependencies {
	d.Logger = logger
	return d
}

// WithPrompt sets the prompt and returns the instance for chaining.
func (d *Dependencies) WithPrompt(prompt prompt.Prompter) *Dependencies {
	d.Prompt = prompt
	return d
}

// WithHookManager sets the hook manager and returns the instance for chaining.
func (d *Dependencies) WithHookManager(hm hooks.HookManagerInterface) *Dependencies {
	d.HookManager = hm
	return d
}

// WithRepositoryProvider sets the repository provider and returns the instance for chaining.
func (d *Dependencies) WithRepositoryProvider(rp repositoryinterfaces.RepositoryProvider) *Dependencies {
	d.RepositoryProvider = rp
	return d
}

// WithWorkspaceProvider sets the workspace provider and returns the instance for chaining.
func (d *Dependencies) WithWorkspaceProvider(wp workspaceinterfaces.WorkspaceProvider) *Dependencies {
	d.WorkspaceProvider = wp
	return d
}

// WithWorktreeProvider sets the worktree provider and returns the instance for chaining.
func (d *Dependencies) WithWorktreeProvider(wp worktreeinterfaces.WorktreeProvider) *Dependencies {
	d.WorktreeProvider = wp
	return d
}
