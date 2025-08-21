package cm

import (
	"errors"
	"fmt"

	repo "github.com/lerenn/code-manager/pkg/repository"
	ws "github.com/lerenn/code-manager/pkg/workspace"
)

// DeleteWorkTree deletes a worktree for the specified branch.
func (c *realCM) DeleteWorkTree(branch string, force bool) error {
	c.VerbosePrint("Deleting worktree for branch: %s (force: %t)", branch, force)

	// Detect project mode
	projectType, err := c.detectProjectMode()
	if err != nil {
		return fmt.Errorf("failed to detect project mode: %w", err)
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		return c.handleRepositoryDeleteMode(branch, force)
	case ProjectTypeWorkspace:
		return c.handleWorkspaceDeleteMode(branch, force)
	case ProjectTypeNone:
		return ErrNoGitRepositoryOrWorkspaceFound
	default:
		return fmt.Errorf("unknown project type")
	}
}

// handleRepositoryDeleteMode handles repository mode: validation and worktree deletion.
func (c *realCM) handleRepositoryDeleteMode(branch string, force bool) error {
	c.VerbosePrint("Handling repository delete mode")

	// Create a single repository instance for all repository operations
	repoInstance := repo.NewRepository(repo.NewRepositoryParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})

	// Delete worktree for single repository
	if err := repoInstance.DeleteWorktree(branch, force); err != nil {
		return c.translateRepositoryError(err)
	}

	c.VerbosePrint("CM delete execution completed successfully")

	return nil
}

// handleWorkspaceDeleteMode handles workspace mode: validation and worktree deletion.
func (c *realCM) handleWorkspaceDeleteMode(branch string, force bool) error {
	c.VerbosePrint("Handling workspace delete mode")

	// Create workspace instance
	workspace := ws.NewWorkspace(ws.NewWorkspaceParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})

	// Delete worktree for workspace
	if err := workspace.DeleteWorktree(branch, force); err != nil {
		return c.translateWorkspaceError(err)
	}

	c.VerbosePrint("Workspace worktree deletion completed successfully")
	return nil
}

// translateWorkspaceError translates workspace package errors to CM package errors.
func (c *realCM) translateWorkspaceError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific workspace errors and translate them
	if errors.Is(err, ws.ErrWorktreeExists) {
		return ErrWorktreeExists
	}
	if errors.Is(err, ws.ErrWorktreeNotInStatus) {
		return ErrWorktreeNotInStatus
	}
	if errors.Is(err, ws.ErrRepositoryNotClean) {
		return ErrRepositoryNotClean
	}
	if errors.Is(err, ws.ErrDirectoryExists) {
		return ErrDirectoryExists
	}

	// Return the original error if no translation is needed
	return err
}
