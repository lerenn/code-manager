package mode

import (
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
)

// Mode represents the type of project detected.
type Mode int

// Mode constants.
const (
	ModeNone Mode = iota
	ModeSingleRepo
	ModeWorkspace
)

// CreateWorktreeOpts contains unified optional parameters for worktree creation.
// This combines options from both workspace and repository modes.
type CreateWorktreeOpts struct {
	IDEName   string
	IssueInfo *issue.Info
}

// ModeInterface provides the common interface for both workspace and repository modes.
type ModeInterface interface {
	// Common methods with harmonized signatures
	Validate() error
	CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error)
	DeleteWorktree(branch string, force bool) error
	ListWorktrees() ([]status.WorktreeInfo, error)
	SetLogger(logger logger.Logger)
}
