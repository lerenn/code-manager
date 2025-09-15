// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"github.com/lerenn/code-manager/pkg/dependencies"
	"github.com/lerenn/code-manager/pkg/mode/workspace/interfaces"
)

// Workspace interface provides workspace management capabilities.
// This interface is now defined in pkg/mode/workspace/interfaces to avoid circular imports.
type Workspace = interfaces.Workspace

// CreateWorktreeOpts contains optional parameters for worktree creation in workspace mode.
type CreateWorktreeOpts = interfaces.CreateWorktreeOpts

// Config represents the configuration of a workspace.
type Config = interfaces.Config

// Folder represents a folder in a workspace.
type Folder = interfaces.Folder

// NewWorkspaceParams contains parameters for creating a new Workspace instance.
type NewWorkspaceParams = interfaces.NewWorkspaceParams

// realWorkspace represents a workspace and provides methods for workspace operations.
type realWorkspace struct {
	deps *dependencies.Dependencies
	file string
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace(params NewWorkspaceParams) Workspace {
	// Cast interface{} to concrete type
	deps := params.Dependencies.(*dependencies.Dependencies)
	if deps == nil {
		deps = dependencies.New()
	}

	return &realWorkspace{
		deps: deps,
		file: params.File,
	}
}
