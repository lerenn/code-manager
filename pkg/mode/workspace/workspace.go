// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/branch"
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
	DeleteWorktree(workspaceName, branch string, force bool) error
	DeleteAllWorktrees(workspaceName string, force bool) error
	OpenWorktree(workspaceName, branch string) (string, error)
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
	config             config.Config
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
	Config             config.Config
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
		config:             params.Config,
		statusManager:      params.StatusManager,
		logger:             l,
		prompt:             params.Prompt,
		worktreeProvider:   params.WorktreeProvider,
		repositoryProvider: params.RepositoryProvider,
		hookManager:        params.HookManager,
	}
}

// buildWorkspaceFilePath constructs the workspace file path for a given workspace name and branch.
// This is a shared utility function used by create, delete, and open operations.
// The workspace file is named: {workspaceName}/{sanitizedBranchName}.code-workspace.
func buildWorkspaceFilePath(workspacesDir, workspaceName, branchName string) string {
	// Sanitize branch name for filename (replace / with -)
	sanitizedBranchForFilename := branch.SanitizeBranchNameForFilename(branchName)

	// Create workspace file path
	workspaceFileName := fmt.Sprintf("%s/%s.code-workspace", workspaceName, sanitizedBranchForFilename)
	return filepath.Join(workspacesDir, workspaceFileName)
}
