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

// NewWorktree creates a new Worktree instance.
func NewWorktree(params NewWorktreeParams) Worktree {
	// Cast interface{} types to concrete types
	fs := params.FS.(fs.FS)
	git := params.Git.(git.Git)
	statusManager := params.StatusManager.(status.Manager)
	prompt := params.Prompt.(prompt.Prompter)

	// Set default logger if not provided
	var log logger.Logger
	if params.Logger == nil {
		log = logger.NewNoopLogger()
	} else {
		log = params.Logger.(logger.Logger)
	}

	return &realWorktree{
		fs:              fs,
		git:             git,
		statusManager:   statusManager,
		logger:          log,
		prompt:          prompt,
		repositoriesDir: params.RepositoriesDir,
	}
}
