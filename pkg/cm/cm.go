package cm

import (
	basepkg "github.com/lerenn/cm/internal/base"
	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/ide"
	"github.com/lerenn/cm/pkg/logger"
	"github.com/lerenn/cm/pkg/prompt"
	"github.com/lerenn/cm/pkg/status"
)

// CM interface provides Git repository detection functionality.
type CM interface {
	// CreateWorkTree executes the main application logic.
	CreateWorkTree(branch string, opts ...CreateWorkTreeOpts) error

	// DeleteWorkTree deletes a worktree for the specified branch.
	DeleteWorkTree(branch string, force bool) error

	// OpenWorktree opens an existing worktree in the specified IDE.
	OpenWorktree(worktreeName, ideName string) error

	// ListWorktrees lists worktrees for the current project with mode detection.
	ListWorktrees() ([]status.Repository, ProjectType, error)

	// LoadWorktree loads a branch from a remote source and creates a worktree.
	LoadWorktree(branchArg string, opts ...LoadWorktreeOpts) error

	// Init initializes CM configuration.
	Init(opts InitOpts) error

	// IsInitialized checks if CM is initialized.
	IsInitialized() (bool, error)

	// SetVerbose enables or disables verbose mode.
	SetVerbose(verbose bool)
}

type realCM struct {
	*basepkg.Base
	ideManager ide.ManagerInterface
}

// NewCM creates a new CM instance.
func NewCM(cfg *config.Config) CM {
	fsInstance := fs.NewFS()
	gitInstance := git.NewGit()
	loggerInstance := logger.NewNoopLogger()
	promptInstance := prompt.NewPrompt()

	return &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            fsInstance,
			Git:           gitInstance,
			Config:        cfg,
			StatusManager: status.NewManager(fsInstance, cfg),
			Logger:        loggerInstance,
			Prompt:        promptInstance,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(fsInstance, loggerInstance),
	}
}

func (c *realCM) SetVerbose(verbose bool) {
	// Create a new Base with the updated verbose setting
	newBase := basepkg.NewBase(basepkg.NewBaseParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       verbose,
	})
	c.Base = newBase

	// Update the IDE manager with the new logger
	c.ideManager = ide.NewManager(c.FS, c.Logger)
}
