package codemanager

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
)

// DeleteAllWorktreesOpts contains options for DeleteAllWorktrees.
type DeleteAllWorktreesOpts struct {
	WorkspaceName  string // Name of the workspace to delete all worktrees for (optional)
	RepositoryName string // Name of the repository to delete all worktrees for (optional)
}

// DeleteAllWorktrees deletes all worktrees for the current repository or workspace.
func (c *realCodeManager) DeleteAllWorktrees(force bool, opts ...DeleteAllWorktreesOpts) error {
	// Parse options
	options := c.extractDeleteAllWorktreesOptions(opts)

	// Validate that workspace and repository are not both specified
	if options.WorkspaceName != "" && options.RepositoryName != "" {
		return fmt.Errorf("cannot specify both WorkspaceName and RepositoryName")
	}

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"force":           force,
		"workspace_name":  options.WorkspaceName,
		"repository_name": options.RepositoryName,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.DeleteAllWorktrees, params, func() error {
		c.VerbosePrint("Deleting all worktrees (force: %t)", force)

		// Detect project mode and delete accordingly
		projectType, err := c.detectProjectMode(options.WorkspaceName, options.RepositoryName)
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		switch projectType {
		case mode.ModeSingleRepo:
			if options.RepositoryName != "" {
				return c.deleteAllRepositoryWorktrees(options.RepositoryName, force)
			}
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
func (c *realCodeManager) handleRepositoryDeleteAllMode(force bool) error {
	c.VerbosePrint("Handling repository delete all mode")

	// Create repository instance
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: ".",
	})

	// Delete all worktrees for single repository
	if err := repoInstance.DeleteAllWorktrees(force); err != nil {
		return c.translateRepositoryError(err)
	}

	c.VerbosePrint("CM delete all execution completed successfully")

	return nil
}

// handleWorkspaceDeleteAllMode handles workspace mode: delete all worktrees.
func (c *realCodeManager) handleWorkspaceDeleteAllMode(force bool) error {
	c.VerbosePrint("Handling workspace delete all mode")

	// Create workspace instance
	workspaceProvider := c.deps.WorkspaceProvider
	workspaceInstance := workspaceProvider(ws.NewWorkspaceParams{
		Dependencies: c.deps,
	})

	// Delete all worktrees for workspace
	if err := workspaceInstance.DeleteAllWorktrees(force); err != nil {
		return c.translateWorkspaceError(err)
	}

	c.VerbosePrint("Workspace delete all worktrees completed successfully")
	return nil
}

// deleteAllRepositoryWorktrees deletes all worktrees for a specific repository.
func (c *realCodeManager) deleteAllRepositoryWorktrees(repositoryName string, force bool) error {
	c.VerbosePrint("Deleting all worktrees for repository: %s", repositoryName)

	// Create repository instance - let repositoryProvider handle repository name resolution
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: repositoryName, // Pass repository name directly, let provider handle resolution
	})

	// Delete all worktrees
	if err := repoInstance.DeleteAllWorktrees(force); err != nil {
		return c.translateRepositoryError(err)
	}

	c.VerbosePrint("Repository delete all worktrees completed successfully")
	return nil
}

// extractDeleteAllWorktreesOptions extracts and merges options from the variadic parameter.
func (c *realCodeManager) extractDeleteAllWorktreesOptions(opts []DeleteAllWorktreesOpts) DeleteAllWorktreesOpts {
	var result DeleteAllWorktreesOpts

	// Merge all provided options, with later options overriding earlier ones
	for _, opt := range opts {
		if opt.WorkspaceName != "" {
			result.WorkspaceName = opt.WorkspaceName
		}
		if opt.RepositoryName != "" {
			result.RepositoryName = opt.RepositoryName
		}
	}

	return result
}
