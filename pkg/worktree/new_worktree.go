// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
)

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
