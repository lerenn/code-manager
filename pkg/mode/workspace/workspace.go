// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

//go:generate mockgen -source=workspace.go -destination=mocks/workspace.gen.go -package=mocks

// CreateWorktreeOpts contains optional parameters for worktree creation in workspace mode.
type CreateWorktreeOpts struct {
	IDEName       string
	IssueInfo     *issue.Info
	WorkspaceName string
}

// Config represents the configuration of a workspace.
type Config struct {
	Name    string   `json:"name,omitempty"`
	Folders []Folder `json:"folders"`
}

// Folder represents a folder in a workspace.
type Folder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Workspace interface provides workspace management capabilities.
type Workspace interface {
	Validate() error
	CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error)
	DeleteWorktree(branch string, force bool) error
	DeleteAllWorktrees(force bool) error
	ListWorktrees() ([]status.WorktreeInfo, error)
	SetLogger(logger logger.Logger)
	Load() error
	ParseFile(filename string) (Config, error)
	GetName(config Config, filename string) string
	ValidateWorkspaceReferences() error
}

// WorktreeProvider is a function type that creates worktree instances.
type WorktreeProvider func(params worktree.NewWorktreeParams) worktree.Worktree

// RepositoryProvider is a function type that creates repository instances.
type RepositoryProvider func(params repository.NewRepositoryParams) repository.Repository

// realWorkspace represents a workspace and provides methods for workspace operations.
type realWorkspace struct {
	fs                 fs.FS
	git                git.Git
	configManager      config.Manager
	statusManager      status.Manager
	logger             logger.Logger
	prompt             prompt.Prompter
	worktreeProvider   WorktreeProvider
	repositoryProvider RepositoryProvider
	hookManager        hooks.HookManagerInterface
	file               string
}

// NewWorkspaceParams contains parameters for creating a new Workspace instance.
type NewWorkspaceParams struct {
	FS                 fs.FS
	Git                git.Git
	ConfigManager      config.Manager
	StatusManager      status.Manager
	Logger             logger.Logger
	Prompt             prompt.Prompter
	WorktreeProvider   WorktreeProvider
	RepositoryProvider RepositoryProvider
	HookManager        hooks.HookManagerInterface
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace(params NewWorkspaceParams) Workspace {
	l := params.Logger
	if l == nil {
		l = logger.NewNoopLogger()
	}

	return &realWorkspace{
		fs:                 params.FS,
		git:                params.Git,
		configManager:      params.ConfigManager,
		statusManager:      params.StatusManager,
		logger:             l,
		prompt:             params.Prompt,
		worktreeProvider:   params.WorktreeProvider,
		repositoryProvider: params.RepositoryProvider,
		hookManager:        params.HookManager,
	}
}
