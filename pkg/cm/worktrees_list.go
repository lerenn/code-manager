package cm

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// ListWorktrees lists worktrees for a workspace or repository.
func (c *realCM) ListWorktrees(opts ...ListWorktreesOpts) ([]status.WorktreeInfo, error) {
	// Parse options
	var options ListWorktreesOpts
	if len(opts) > 0 {
		options = opts[0]
	}

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"workspace_name": options.WorkspaceName,
	}

	// Execute with hooks
	var result []status.WorktreeInfo
	err := c.executeWithHooks(consts.ListWorktrees, params, func() error {
		var err error
		c.VerbosePrint("Listing worktrees")

		// If workspace name is provided, list workspace worktrees
		if options.WorkspaceName != "" {
			result, err = c.listWorkspaceWorktrees(options.WorkspaceName)
			return err
		}

		// Otherwise, detect project mode and list accordingly
		projectType, err := c.detectProjectMode("")
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
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
			result = worktrees
			return c.translateListError(err)
		case mode.ModeWorkspace:
			return fmt.Errorf("workspace mode detected but no workspace name provided")
		case mode.ModeNone:
			return ErrNoGitRepositoryOrWorkspaceFound
		default:
			return fmt.Errorf("unknown project type")
		}
	})
	return result, err
}

// listWorkspaceWorktrees lists all worktrees associated with a workspace.
func (c *realCM) listWorkspaceWorktrees(workspaceName string) ([]status.WorktreeInfo, error) {
	c.VerbosePrint("Listing worktrees for workspace: %s", workspaceName)

	// Get workspace from status
	workspace, err := c.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return c.listWorkspaceWorktreesFromWorkspace(workspace)
}

// listWorkspaceWorktreesFromWorkspace lists all worktrees from a workspace object.
func (c *realCM) listWorkspaceWorktreesFromWorkspace(workspace *status.Workspace) ([]status.WorktreeInfo, error) {
	// Get worktrees that are specifically associated with this workspace
	workspaceWorktrees := make([]status.WorktreeInfo, 0)

	c.VerbosePrint("Listing worktrees for workspace with %d worktree references: %v",
		len(workspace.Worktrees), workspace.Worktrees)
	c.VerbosePrint("Workspace has %d repositories: %v", len(workspace.Repositories), workspace.Repositories)

	// For workspace deletion, we need to find worktrees in ALL repositories that match the workspace's worktree references
	// This is because workspace creation creates worktrees in all repositories but only tracks one reference
	for _, worktreeRef := range workspace.Worktrees {
		c.VerbosePrint("  Looking for worktree reference: %s", worktreeRef)
		for _, repoURL := range workspace.Repositories {
			// Get repository to check its worktrees
			repo, err := c.statusManager.GetRepository(repoURL)
			if err != nil {
				c.VerbosePrint("    ⚠ Skipping repository %s: %v", repoURL, err)
				continue // Skip if repository not found
			}

			c.VerbosePrint("    Checking repository %s with %d worktrees: %v", repoURL, len(repo.Worktrees), repo.Worktrees)
			// Look for the worktree reference in this repository's worktrees
			// The worktrees are stored with "remote:branch" as the key, but workspace stores just "branch"
			// So we need to find worktrees where the branch matches
			for worktreeKey, worktree := range repo.Worktrees {
				if worktree.Branch == worktreeRef {
					c.VerbosePrint("    ✓ Found worktree %s (key: %s) in repository %s", worktreeRef, worktreeKey, repoURL)
					workspaceWorktrees = append(workspaceWorktrees, worktree)
					break // Found in this repository, continue to next repository
				}
			}
		}
	}

	c.VerbosePrint("Found %d worktrees for workspace", len(workspaceWorktrees))
	return workspaceWorktrees, nil
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
