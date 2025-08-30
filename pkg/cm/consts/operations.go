// Package consts provides operation name constants for the hook system.
package consts

// Operation names for the hook system.
const (
	// Worktree operations.
	CreateWorkTree = "CreateWorkTree"
	DeleteWorkTree = "DeleteWorkTree"
	LoadWorktree   = "LoadWorktree"
	ListWorktrees  = "ListWorktrees"
	OpenWorktree   = "OpenWorktree"

	// Repository operations.
	CloneRepository  = "CloneRepository"
	ListRepositories = "ListRepositories"
	Clone            = "Clone" // Legacy name for backward compatibility

	// Initialization operations.
	Init = "Init"
)
