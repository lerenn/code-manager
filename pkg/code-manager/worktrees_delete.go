package codemanager

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/prompt"
)

// DeleteWorktreeOpts contains optional parameters for DeleteWorkTree.
type DeleteWorktreeOpts struct {
	WorkspaceName  string
	RepositoryName string
}

// DeleteWorkTree deletes a worktree for the specified branch.
func (c *realCodeManager) DeleteWorkTree(branch string, force bool, opts ...DeleteWorktreeOpts) error {
	// Parse options
	options := c.extractDeleteWorktreeOptions(opts)

	// Validate that workspace and repository are not both specified
	if options.WorkspaceName != "" && options.RepositoryName != "" {
		return fmt.Errorf("cannot specify both WorkspaceName and RepositoryName")
	}

	// Handle interactive selection if neither workspace nor repository is specified
	if options.WorkspaceName == "" && options.RepositoryName == "" {
		if branch == "" {
			// Two-step selection: first target, then worktree
			result, err := c.promptSelectTargetAndWorktree()
			if err != nil {
				return fmt.Errorf("failed to select target and worktree: %w", err)
			}

			switch result.Type {
			case prompt.TargetWorkspace:
				options.WorkspaceName = result.Name
			case prompt.TargetRepository:
				options.RepositoryName = result.Name
			default:
				return fmt.Errorf("invalid target type selected: %s", result.Type)
			}

			branch = result.Worktree
		} else {
			// Single-step selection: just target (branch already provided)
			result, err := c.promptSelectTargetOnly()
			if err != nil {
				return fmt.Errorf("failed to select target: %w", err)
			}

			switch result.Type {
			case prompt.TargetWorkspace:
				options.WorkspaceName = result.Name
			case prompt.TargetRepository:
				options.RepositoryName = result.Name
			default:
				return fmt.Errorf("invalid target type selected: %s", result.Type)
			}
		}
	} else if branch == "" {
		// Target specified but no branch - prompt for branch name
		branchName, err := c.deps.Prompt.PromptForBranchName()
		if err != nil {
			return fmt.Errorf("failed to get branch name: %w", err)
		}
		branch = branchName
	}

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"branch":          branch,
		"force":           force,
		"workspace_name":  options.WorkspaceName,
		"repository_name": options.RepositoryName,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.DeleteWorkTree, params, func() error {
		c.VerbosePrint("Deleting worktree for branch: %s (force: %t)", branch, force)

		// Detect project mode and delete accordingly
		projectType, err := c.detectProjectMode(options.WorkspaceName, options.RepositoryName)
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		switch projectType {
		case mode.ModeSingleRepo:
			if options.RepositoryName != "" {
				return c.deleteRepositoryWorktree(options.RepositoryName, branch, force)
			}
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
func (c *realCodeManager) handleRepositoryDeleteMode(branch string, force bool) error {
	c.VerbosePrint("Handling repository delete mode")

	// Create repository instance
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: ".",
	})

	// Delete worktree for single repository
	if err := repoInstance.DeleteWorktree(branch, force); err != nil {
		return c.translateRepositoryError(err)
	}

	c.VerbosePrint("CM delete execution completed successfully")

	return nil
}

// handleWorkspaceDeleteMode handles workspace mode: validation and worktree deletion.
func (c *realCodeManager) handleWorkspaceDeleteMode(branch string, force bool) error {
	c.VerbosePrint("Handling workspace delete mode")

	// Create workspace instance
	workspaceProvider := c.deps.WorkspaceProvider
	workspaceInstance := workspaceProvider(ws.NewWorkspaceParams{
		Dependencies: c.deps,
	})

	// Delete worktree for workspace
	if err := workspaceInstance.DeleteWorktree(branch, force); err != nil {
		return c.translateWorkspaceError(err)
	}

	c.VerbosePrint("Workspace worktree deletion completed successfully")
	return nil
}

// DeleteWorkTrees deletes multiple worktrees for the specified branches.
func (c *realCodeManager) DeleteWorkTrees(branches []string, force bool) error {
	if len(branches) == 0 {
		return fmt.Errorf("no branches specified for deletion")
	}

	c.VerbosePrint("Deleting %d worktrees: %v (force: %t)", len(branches), branches, force)

	var errors []error
	for _, branch := range branches {
		c.VerbosePrint("Deleting worktree for branch: %s", branch)
		if err := c.DeleteWorkTree(branch, force); err != nil {
			c.VerbosePrint("Failed to delete worktree for branch %s: %v", branch, err)
			errors = append(errors, fmt.Errorf("failed to delete worktree for branch %s: %w", branch, err))
		} else {
			c.VerbosePrint("Successfully deleted worktree for branch: %s", branch)
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(branches) {
			// All deletions failed
			return fmt.Errorf("failed to delete all worktrees: %v", errors)
		}
		// Some deletions failed
		c.VerbosePrint("Some worktrees failed to delete: %v", errors)
		return fmt.Errorf("some worktrees failed to delete: %v", errors)
	}

	c.VerbosePrint("All worktrees deleted successfully")
	return nil
}

// translateWorkspaceError translates workspace package errors to CM package errors.
func (c *realCodeManager) translateWorkspaceError(err error) error {
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

// deleteRepositoryWorktree deletes a worktree for a specific repository.
func (c *realCodeManager) deleteRepositoryWorktree(repositoryName, branch string, force bool) error {
	c.VerbosePrint("Deleting worktree for repository: %s, branch: %s", repositoryName, branch)

	// Create repository instance - let repositoryProvider handle repository name resolution
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: repositoryName, // Pass repository name directly, let provider handle resolution
	})

	// Delete the worktree
	if err := repoInstance.DeleteWorktree(branch, force); err != nil {
		return c.translateRepositoryError(err)
	}

	return nil
}

// extractDeleteWorktreeOptions extracts and merges options from the variadic parameter.
func (c *realCodeManager) extractDeleteWorktreeOptions(opts []DeleteWorktreeOpts) DeleteWorktreeOpts {
	var result DeleteWorktreeOpts

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
