package cm

import (
	"fmt"

	ws "github.com/lerenn/cm/pkg/workspace"
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
		worktreePath := c.BuildWorktreePath(repoURL, worktreeName)
		exists, err := c.FS.Exists(worktreePath)
		if err != nil {
			return fmt.Errorf("failed to check if worktree exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("worktree not found for branch: %s", worktreeName)
		}

		return c.ideManager.OpenIDE(ideName, worktreePath, c.IsVerbose())
	case ProjectTypeWorkspace:
		// For workspace, we need to find the worktree path from the workspace
		workspace := ws.NewWorkspace(ws.NewWorkspaceParams{
			FS:            c.FS,
			Git:           c.Git,
			Config:        c.Config,
			StatusManager: c.StatusManager,
			Logger:        c.Logger,
			Prompt:        c.Prompt,
			Verbose:       c.IsVerbose(),
		})

		// Load workspace to get worktree paths
		if err := workspace.Load(); err != nil {
			return fmt.Errorf("failed to load workspace: %w", err)
		}

		// For now, we'll use a simple approach - open the workspace file
		// In the future, this could be enhanced to open specific worktree directories
		return c.ideManager.OpenIDE(ideName, ".", c.IsVerbose())
	case ProjectTypeNone:
		return fmt.Errorf("no Git repository or workspace found")
	default:
		return fmt.Errorf("unknown project type")
	}
}
