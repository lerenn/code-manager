package cm

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// DeleteAllWorktrees deletes all worktrees for the current repository or workspace.
func (c *realCM) DeleteAllWorktrees(force bool) error {
	// Prepare parameters for hooks
	params := map[string]interface{}{
		"force": force,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.DeleteAllWorktrees, params, func() error {
		c.VerbosePrint("Deleting all worktrees (force: %t)", force)

		// Detect project mode
		projectType, err := c.detectProjectMode("")
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		switch projectType {
		case mode.ModeSingleRepo:
			return c.handleRepositoryDeleteAllMode(force)
		case mode.ModeWorkspace:
			return c.handleWorkspaceDeleteAllMode(force)
		case mode.ModeNone:
			return ErrNoGitRepositoryOrWorkspaceFound
		default:
			return fmt.Errorf("unknown project type")
		}
	})
}

// handleRepositoryDeleteAllMode handles repository mode: delete all worktrees.
func (c *realCM) handleRepositoryDeleteAllMode(force bool) error {
	c.VerbosePrint("Handling repository delete all mode")

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

	// Delete all worktrees for single repository
	if err := repoInstance.DeleteAllWorktrees(force); err != nil {
		return c.translateRepositoryError(err)
	}

	c.VerbosePrint("CM delete all execution completed successfully")

	return nil
}

// handleWorkspaceDeleteAllMode handles workspace mode: delete all worktrees.
func (c *realCM) handleWorkspaceDeleteAllMode(force bool) error {
	c.VerbosePrint("Handling workspace delete all mode")

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

	// Delete all worktrees for workspace
	if err := workspaceInstance.DeleteAllWorktrees(force); err != nil {
		return c.translateWorkspaceError(err)
	}

	c.VerbosePrint("Workspace delete all worktrees completed successfully")
	return nil
}
