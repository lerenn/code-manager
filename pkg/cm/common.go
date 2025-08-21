package cm

import (
	"fmt"

	repo "github.com/lerenn/code-manager/pkg/repository"
)

// detectProjectMode detects the type of project (single repository or workspace).
func (c *realCM) detectProjectMode() (ProjectType, error) {
	c.VerbosePrint("Detecting project mode...")

	// First, check if we're in a Git repository
	repoInstance := repo.NewRepository(repo.NewRepositoryParams{
		FS:            c.FS,
		Git:           c.Git,
		Config:        c.Config,
		StatusManager: c.StatusManager,
		Logger:        c.Logger,
		Prompt:        c.Prompt,
		Verbose:       c.IsVerbose(),
	})

	// Check if we're in a Git repository
	exists, err := repoInstance.IsGitRepository()
	if err != nil {
		return ProjectTypeNone, fmt.Errorf("failed to check Git repository: %w", err)
	}

	if exists {
		c.VerbosePrint("Single repository mode detected")
		return ProjectTypeSingleRepo, nil
	}

	// If not a Git repository, check for workspace files
	workspaceFiles, err := c.FS.Glob("*.code-workspace")
	if err != nil {
		return ProjectTypeNone, fmt.Errorf("failed to detect workspace files: %w", err)
	}

	if len(workspaceFiles) > 0 {
		c.VerbosePrint("Workspace mode detected")
		return ProjectTypeWorkspace, nil
	}

	c.VerbosePrint("No project mode detected")
	return ProjectTypeNone, nil
}

// handleIDEOpening handles IDE opening if specified and worktree creation was successful.
func (c *realCM) handleIDEOpening(worktreeErr error, branch string, ideName *string) error {
	if worktreeErr == nil && ideName != nil && *ideName != "" {
		if err := c.OpenWorktree(branch, *ideName); err != nil {
			return err
		}
	}
	return nil
}
