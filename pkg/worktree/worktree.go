// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree/interfaces"
)

// Worktree interface provides worktree management capabilities.
// This interface is now defined in pkg/worktree/interfaces to avoid circular imports.
type Worktree = interfaces.Worktree

// CreateParams contains parameters for worktree creation.
type CreateParams = interfaces.CreateParams

// DeleteParams contains parameters for worktree deletion.
type DeleteParams = interfaces.DeleteParams

// ValidateCreationParams contains parameters for worktree creation validation.
type ValidateCreationParams = interfaces.ValidateCreationParams

// ValidateDeletionParams contains parameters for worktree deletion validation.
type ValidateDeletionParams = interfaces.ValidateDeletionParams

// AddToStatusParams contains parameters for adding worktree to status.
type AddToStatusParams = interfaces.AddToStatusParams

// NewWorktreeParams contains parameters for creating a new Worktree instance.
type NewWorktreeParams = interfaces.NewWorktreeParams

// realWorktree provides the real implementation of the Worktree interface.
type realWorktree struct {
	fs              fs.FS
	git             git.Git
	statusManager   status.Manager
	logger          logger.Logger
	prompt          prompt.Prompter
	repositoriesDir string
}
