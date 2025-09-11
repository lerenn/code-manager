package cm

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// DeleteWorkTree deletes a worktree for the specified branch.
func (c *realCM) DeleteWorkTree(branch string, force bool) error {
	// Prepare parameters for hooks
	params := map[string]interface{}{
		"branch": branch,
		"force":  force,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.DeleteWorkTree, params, func() error {
		c.VerbosePrint("Deleting worktree for branch: %s (force: %t)", branch, force)

		// Detect project mode
		projectType, err := c.detectProjectMode("")
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		switch projectType {
		case mode.ModeSingleRepo:
			return c.handleRepositoryDeleteMode(branch, force)
		case mode.ModeWorkspace:
			return c.handleWorkspaceDeleteMode(branch, force)
		case mode.ModeNone:
			return ErrNoGitRepositoryOrWorkspaceFound
		default:
			return fmt.Errorf("unknown project type")
		}
	})
}

// handleRepositoryDeleteMode handles repository mode: validation and worktree deletion.
func (c *realCM) handleRepositoryDeleteMode(branch string, force bool) error {
	c.VerbosePrint("Handling repository delete mode")

	// Create repository instance
	repoInstance := c.repositoryProvider(repo.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
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
	workspaceInstance := c.workspaceProvider(ws.NewWorkspaceParams{
		FS:               c.fs,
		Git:              c.git,
		Config:           c.config,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
	})

	// Delete worktree for workspace
	if err := workspaceInstance.DeleteWorktree(branch, force); err != nil {
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
