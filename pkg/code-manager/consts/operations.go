// Package consts provides operation name constants for the hook system.
package consts

// Operation names for the hook system.
const (
	// Worktree operations.
	CreateWorkTree     = "CreateWorkTree"
	DeleteWorkTree     = "DeleteWorkTree"
	DeleteAllWorktrees = "DeleteAllWorktrees"
	LoadWorktree       = "LoadWorktree"
	ListWorktrees      = "ListWorktrees"
	OpenWorktree       = "OpenWorktree"

	// Repository operations.
	CloneRepository  = "CloneRepository"
	ListRepositories = "ListRepositories"
	DeleteRepository = "DeleteRepository"
	Clone            = "Clone" // Legacy name for backward compatibility

	// Workspace operations.
	ListWorkspaces                = "ListWorkspaces"
	AddRepositoryToWorkspace      = "AddRepositoryToWorkspace"
	RemoveRepositoryFromWorkspace = "RemoveRepositoryFromWorkspace"

	// Prompt operations.
	PromptSelectTarget = "PromptSelectTarget"

	// Initialization operations.
	Init = "Init"
)
