package cm

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// ListWorktrees lists worktrees for the current project with mode detection.
func (c *realCM) ListWorktrees(force bool) ([]status.WorktreeInfo, mode.Mode, error) {
	// Prepare parameters for hooks
	params := map[string]interface{}{
		"force": force,
	}

	// Execute with hooks
	return c.executeWithHooksAndReturnListWorktrees(consts.ListWorktrees, params, func() (
		[]status.WorktreeInfo, mode.Mode, error,
	) {
		c.VerbosePrint("Listing worktrees with mode detection")

		// Detect project mode
		projectType, err := c.detectProjectMode()
		if err != nil {
			return nil, mode.ModeNone, fmt.Errorf("failed to detect project mode: %w", err)
		}

		switch projectType {
		case mode.ModeSingleRepo:
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
			worktrees, err := repoInstance.ListWorktrees()
			return worktrees, mode.ModeSingleRepo, c.translateListError(err)
		case mode.ModeWorkspace:
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
			worktrees, err := workspaceInstance.ListWorktrees()
			return worktrees, mode.ModeWorkspace, c.translateListError(err)
		case mode.ModeNone:
			return nil, mode.ModeNone, ErrNoGitRepositoryOrWorkspaceFound
		default:
			return nil, mode.ModeNone, fmt.Errorf("unknown project type")
		}
	})
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
