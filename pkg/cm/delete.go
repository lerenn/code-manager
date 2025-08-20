package cm

import (
	"fmt"

	repo "github.com/lerenn/cm/pkg/repository"
	ws "github.com/lerenn/cm/pkg/workspace"
)

// DeleteWorkTree deletes a worktree for the specified branch.
func (c *realCM) DeleteWorkTree(branch string, force bool) error {
	c.VerbosePrint("Starting worktree deletion for branch: %s", branch)

	// Detect project mode
	projectType, err := c.detectProjectMode()
	if err != nil {
		return fmt.Errorf("failed to detect project mode: %w", err)
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
		return repoInstance.DeleteWorktree(branch, force)
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
		return workspace.DeleteWorktree(branch, force)
	case ProjectTypeNone:
		return fmt.Errorf("no Git repository or workspace found")
	default:
		return fmt.Errorf("unknown project type")
	}
}
