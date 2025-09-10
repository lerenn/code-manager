package cm

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/mode"
)

// OpenWorktree opens an existing worktree in the specified IDE.
func (c *realCM) OpenWorktree(worktreeName, ideName string) error {
	// Prepare parameters for hooks
	params := map[string]interface{}{
		"worktreeName": worktreeName,
		"ideName":      ideName,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.OpenWorktree, params, func() error {
		c.VerbosePrint("Opening worktree: %s in IDE: %s", worktreeName, ideName)

		// Detect project mode
		projectType, err := c.detectProjectMode()
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		switch projectType {
		case mode.ModeSingleRepo:
			// For single repository, worktreeName is the branch name
			// Get repository URL from local .git directory
			repoURL, err := c.git.GetRepositoryName(".")
			if err != nil {
				return fmt.Errorf("failed to get repository URL: %w", err)
			}

			// Check if the worktree exists in the status file
			worktreeInfo, err := c.statusManager.GetWorktree(repoURL, worktreeName)
			if err != nil {
				return ErrWorktreeNotInStatus
			}

			// Build the worktree path using the remote from status
			worktreePath := c.BuildWorktreePath(repoURL, worktreeInfo.Remote, worktreeName)

			// Store the worktree path in parameters for the hook to access
			// Note: This modifies the params map that was passed to executeWithHooks
			// The hook context will have access to this updated parameter
			params["worktreePath"] = worktreePath
			return nil
		case mode.ModeWorkspace:
			// For workspace mode, we need to find the worktree in the workspace
			// For now, return an error indicating this needs to be implemented
			return fmt.Errorf("workspace mode open worktree not yet implemented")
		case mode.ModeNone:
			return ErrNoGitRepositoryOrWorkspaceFound
		default:
			return fmt.Errorf("unknown project type")
		}
	})
}
