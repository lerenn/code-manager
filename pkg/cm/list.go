package cm

import (
	"fmt"

	repo "github.com/lerenn/cm/pkg/repository"
	"github.com/lerenn/cm/pkg/status"
	ws "github.com/lerenn/cm/pkg/workspace"
)

// ListWorktrees lists worktrees for the current project with mode detection.
func (c *realCM) ListWorktrees() ([]status.Repository, ProjectType, error) {
	c.VerbosePrint("Listing worktrees with mode detection")

	// Detect project mode
	projectType, err := c.detectProjectMode()
	if err != nil {
		return nil, ProjectTypeNone, fmt.Errorf("failed to detect project mode: %w", err)
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		repoInstance := repo.NewRepository(repo.NewRepositoryParams{
			FS:            c.FS,
			Git:           c.Git,
			Config:        c.Config,
			StatusManager: c.StatusManager,
			Logger:        c.Logger,
			Prompt:        c.Prompt,
			Verbose:       c.IsVerbose(),
		})
		worktrees, err := repoInstance.ListWorktrees()
		return worktrees, ProjectTypeSingleRepo, err
	case ProjectTypeWorkspace:
		workspace := ws.NewWorkspace(ws.NewWorkspaceParams{
			FS:            c.FS,
			Git:           c.Git,
			Config:        c.Config,
			StatusManager: c.StatusManager,
			Logger:        c.Logger,
			Prompt:        c.Prompt,
			Verbose:       c.IsVerbose(),
		})
		worktrees, err := workspace.ListWorktrees()
		return worktrees, ProjectTypeWorkspace, err
	case ProjectTypeNone:
		return nil, ProjectTypeNone, fmt.Errorf("no Git repository or workspace found")
	default:
		return nil, ProjectTypeNone, fmt.Errorf("unknown project type")
	}
}
