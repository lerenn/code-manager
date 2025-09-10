// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

//go:generate mockgen -source=workspace.go -destination=mocks/workspace.gen.go -package=mocks

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
// It implements the common Interface and adds workspace-specific methods.
type Workspace interface {
	mode.Interface

	// Workspace-specific methods
	Load(force bool) error
	DetectWorkspaceFiles() ([]string, error)
	ParseFile(filename string) (Config, error)
	GetName(config Config, filename string) string
	HandleMultipleFiles(workspaceFiles []string, force bool) (string, error)
	ValidateWorkspaceReferences() error
}

// WorktreeProvider is a function type that creates worktree instances.
type WorktreeProvider func(params worktree.NewWorktreeParams) worktree.Worktree

// realWorkspace represents a workspace and provides methods for workspace operations.
type realWorkspace struct {
	fs               fs.FS
	git              git.Git
	config           config.Config
	statusManager    status.Manager
	logger           logger.Logger
	prompt           prompt.Prompter
	worktreeProvider WorktreeProvider
	hookManager      hooks.HookManagerInterface
	OriginalFile     string
}

// NewWorkspaceParams contains parameters for creating a new Workspace instance.
type NewWorkspaceParams struct {
	FS               fs.FS
	Git              git.Git
	Config           config.Config
	StatusManager    status.Manager
	Logger           logger.Logger
	Prompt           prompt.Prompter
	WorktreeProvider WorktreeProvider
	HookManager      hooks.HookManagerInterface
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace(params NewWorkspaceParams) Workspace {
	l := params.Logger
	if l == nil {
		l = logger.NewNoopLogger()
	}

	return &realWorkspace{
		fs:               params.FS,
		git:              params.Git,
		config:           params.Config,
		statusManager:    params.StatusManager,
		logger:           l,
		prompt:           params.Prompt,
		worktreeProvider: params.WorktreeProvider,
		hookManager:      params.HookManager,
	}
}
