// Package interfaces defines the workspace interface for dependency injection.
// This package is separate to avoid circular imports.
//
//nolint:revive // interfaces package name is intentional to avoid circular dependencies
package interfaces

//go:generate go run go.uber.org/mock/mockgen@latest -source=interfaces.go -destination=../mocks/workspace.gen.go -package=mocks

import (
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
)

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

// Workspace defines the interface for workspace operations.
// This interface is implemented by the concrete workspace types.
type Workspace interface {
	Validate() error
	CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error)
	DeleteWorktree(branch string, force bool) error
	DeleteAllWorktrees(force bool) error
	ListWorktrees() ([]status.WorktreeInfo, error)
	OpenWorktree(workspaceName, branch string) (string, error)
	SetLogger(logger logger.Logger)
	Load() error
	ParseFile(filename string) (Config, error)
	GetName(config Config, filename string) string
	ValidateWorkspaceReferences() error
}

// WorkspaceProvider defines the function signature for creating workspace instances.
type WorkspaceProvider func(params NewWorkspaceParams) Workspace

// NewWorkspaceParams contains parameters for creating a new Workspace instance.
type NewWorkspaceParams struct {
	Dependencies interface{} // *dependencies.Dependencies
	File         string      // workspace file path
}
