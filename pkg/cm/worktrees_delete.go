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

// DeleteWorktreeOpts contains optional parameters for DeleteWorkTree.
type DeleteWorktreeOpts struct {
	WorkspaceName string
}

// DeleteWorkTree deletes a worktree for the specified branch.
func (c *realCM) DeleteWorkTree(branch string, force bool, opts ...DeleteWorktreeOpts) error {
	// Parse options
	var options DeleteWorktreeOpts
	if len(opts) > 0 {
		options = opts[0]
	}

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"branch":         branch,
		"force":          force,
		"workspace_name": options.WorkspaceName,
	}

	// Execute with hooks
	return c.executeWithHooks(consts.DeleteWorkTree, params, func() error {
		c.VerbosePrint("Deleting worktree for branch: %s (force: %t)", branch, force)

		// If workspace name is provided, delete workspace worktree
		if options.WorkspaceName != "" {
			return c.deleteWorkspaceWorktree(options.WorkspaceName, branch, force)
		}

		// Otherwise, detect project mode and delete accordingly
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

// deleteWorkspaceWorktree deletes a worktree from a specific workspace.
func (c *realCM) deleteWorkspaceWorktree(workspaceName, branch string, force bool) error {
	c.VerbosePrint("Deleting worktree '%s' from workspace '%s'", branch, workspaceName)

	// Get workspace from status
	workspace, err := c.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if the worktree exists in the workspace
	worktreeExists := false
	for _, worktreeRef := range workspace.Worktrees {
		if worktreeRef == branch {
			worktreeExists = true
			break
		}
	}

	if !worktreeExists {
		return fmt.Errorf("worktree '%s' not found in workspace '%s'", branch, workspaceName)
	}

	// Find the worktree in the workspace's repositories and delete it
	found := false
	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := c.statusManager.GetRepository(repoURL)
		if err != nil {
			c.VerbosePrint("  ⚠ Skipping repository %s: %v", repoURL, err)
			continue // Skip if repository not found
		}

		// Look for the worktree reference in this repository's worktrees
		for worktreeKey, worktree := range repo.Worktrees {
			if worktree.Branch == branch {
				// Found the worktree, now delete it
				if err := c.deleteWorktreeFromRepository(repoURL, worktreeKey, worktree, force); err != nil {
					return fmt.Errorf("failed to delete worktree from repository %s: %w", repoURL, err)
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("worktree '%s' not found in any repository of workspace '%s'", branch, workspaceName)
	}

	// Remove the worktree from the workspace's worktrees list
	if err := c.removeWorktreeFromWorkspace(workspaceName, branch); err != nil {
		return fmt.Errorf("failed to remove worktree from workspace: %w", err)
	}

	c.VerbosePrint("Successfully deleted worktree '%s' from workspace '%s'", branch, workspaceName)
	return nil
}

// deleteWorktreeFromRepository deletes a worktree from a specific repository.
func (c *realCM) deleteWorktreeFromRepository(repoURL, _ string, worktreeInfo status.WorktreeInfo, force bool) error {
	c.VerbosePrint("Deleting worktree from repository: %s", repoURL)

	// Get repository path from status
	repo, err := c.statusManager.GetRepository(repoURL)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// Create worktree instance for deletion
	worktreeInstance := c.worktreeProvider(worktree.NewWorktreeParams{
		FS:              c.fs,
		Git:             c.git,
		StatusManager:   c.statusManager,
		Logger:          c.logger,
		Prompt:          c.prompt,
		RepositoriesDir: c.config.RepositoriesDir,
	})

	// Build worktree path
	worktreePath := worktreeInstance.BuildPath(repoURL, worktreeInfo.Remote, worktreeInfo.Branch)

	// Delete the worktree
	if err := worktreeInstance.Delete(worktree.DeleteParams{
		RepoURL:      repoURL,
		Branch:       worktreeInfo.Branch,
		WorktreePath: worktreePath,
		RepoPath:     repo.Path,
		Force:        force,
	}); err != nil {
		return fmt.Errorf("failed to delete worktree: %w", err)
	}

	return nil
}

// removeWorktreeFromWorkspace removes a worktree reference from a workspace.
func (c *realCM) removeWorktreeFromWorkspace(workspaceName, branch string) error {
	// Get current workspace
	workspace, err := c.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Remove the worktree from the list
	newWorktrees := make([]string, 0, len(workspace.Worktrees))
	for _, worktreeRef := range workspace.Worktrees {
		if worktreeRef != branch {
			newWorktrees = append(newWorktrees, worktreeRef)
		}
	}
	workspace.Worktrees = newWorktrees

	// Update workspace in status file
	if err := c.statusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	return nil
}
