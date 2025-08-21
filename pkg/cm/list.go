package cm

import (
	"errors"
	"fmt"

	repo "github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/status"
	ws "github.com/lerenn/code-manager/pkg/workspace"
)

// ListWorktrees lists worktrees for the current project with mode detection.
func (c *realCM) ListWorktrees() ([]status.WorktreeInfo, ProjectType, error) {
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
		return worktrees, ProjectTypeSingleRepo, c.translateListError(err)
	case ProjectTypeWorkspace:
		workspaceInstance := ws.NewWorkspace(ws.NewWorkspaceParams{
			FS:            c.FS,
			Git:           c.Git,
			Config:        c.Config,
			StatusManager: c.StatusManager,
			Logger:        c.Logger,
			Prompt:        c.Prompt,
			Verbose:       c.IsVerbose(),
		})
		worktrees, err := workspaceInstance.ListWorktrees()
		return worktrees, ProjectTypeWorkspace, c.translateListError(err)
	case ProjectTypeNone:
		return nil, ProjectTypeNone, ErrNoGitRepositoryOrWorkspaceFound
	default:
		return nil, ProjectTypeNone, fmt.Errorf("unknown project type")
	}
}

// translateListError translates errors from list operations to CM package errors.
func (c *realCM) translateListError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific status errors and translate them
	if errors.Is(err, status.ErrConfigurationNotInitialized) {
		return ErrNotInitialized
	}

	// Return the original error if no translation is needed
	return err
}
