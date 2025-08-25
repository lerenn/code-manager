package cm

import (
	"fmt"
)

// OpenWorktree opens an existing worktree in the specified IDE.
func (c *realCM) OpenWorktree(worktreeName, ideName string) error {
	c.VerbosePrint("Opening worktree: %s in IDE: %s", worktreeName, ideName)

	// Detect project mode
	projectType, err := c.detectProjectMode()
	if err != nil {
		return fmt.Errorf("failed to detect project mode: %w", err)
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		// For single repository, worktreeName is the branch name
		// Get repository URL from local .git directory
		repoURL, err := c.Git.GetRepositoryName(".")
		if err != nil {
			return fmt.Errorf("failed to get repository URL: %w", err)
		}

		// Check if the worktree exists
		worktreePath := c.BuildWorktreePath(repoURL, "origin", worktreeName)
		exists, err := c.FS.Exists(worktreePath)
		if err != nil {
			return fmt.Errorf("failed to check if worktree exists: %w", err)
		}
		if !exists {
			return ErrWorktreeNotInStatus
		}

		return c.ideManager.OpenIDE(ideName, worktreePath, c.IsVerbose())
	case ProjectTypeWorkspace:
		// For workspace mode, we need to find the worktree in the workspace
		// For now, return an error indicating this needs to be implemented
		return fmt.Errorf("workspace mode open worktree not yet implemented")
	case ProjectTypeNone:
		return ErrNoGitRepositoryOrWorkspaceFound
	default:
		return fmt.Errorf("unknown project type")
	}
}
