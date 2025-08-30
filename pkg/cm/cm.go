package cm

import (
	"fmt"

	basepkg "github.com/lerenn/code-manager/internal/base"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/ide"
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
	Repository repository.Repository
	Workspace  workspace.Workspace
	Config     *config.Config
}

type realCM struct {
	*basepkg.Base
	ideManager  ide.ManagerInterface
	repository  repository.Repository
	workspace   workspace.Workspace
	hookManager hooks.HookManagerInterface
}

// NewCM creates a new CM instance.
func NewCM(cfg *config.Config) CM {
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

	return &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            fsInstance,
			Git:           gitInstance,
			Config:        cfg,
			StatusManager: status.NewManager(fsInstance, cfg),
			Logger:        loggerInstance,
			Prompt:        promptInstance,
			Verbose:       false,
		}),
		ideManager:  ide.NewManager(fsInstance, loggerInstance),
		repository:  repoInstance,
		workspace:   workspaceInstance,
		hookManager: hooks.NewHookManager(),
	}
}

// NewCMWithDependencies creates a new CM instance with custom repository and workspace dependencies.
// This is primarily used for testing with mocked dependencies.
func NewCMWithDependencies(params NewCMParams) CM {
	fsInstance := fs.NewFS()
	gitInstance := git.NewGit()
	loggerInstance := logger.NewNoopLogger()
	statusInstance := status.NewManager(fsInstance, params.Config)

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
		ideManager:  ide.NewManager(fsInstance, loggerInstance),
		repository:  params.Repository,
		workspace:   params.Workspace,
		hookManager: hooks.NewHookManager(),
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

	// Update the IDE manager with the new logger
	c.ideManager = ide.NewManager(c.FS, c.Logger)
}

// RegisterHook registers a hook for a specific operation.
func (c *realCM) RegisterHook(operation string, hook hooks.Hook) error {
	// This is a simplified implementation - in practice, you'd want to determine
	// the hook type and register it appropriately.
	switch h := hook.(type) {
	case hooks.GlobalHook:
		return c.hookManager.RegisterGlobalHook(h)
	case hooks.PreHook:
		return c.hookManager.RegisterPreHook(operation, h)
	case hooks.PostHook:
		return c.hookManager.RegisterPostHook(operation, h)
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
