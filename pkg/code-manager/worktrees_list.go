package codemanager

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// ListWorktreesOpts contains options for ListWorktrees.
type ListWorktreesOpts struct {
	WorkspaceName  string // Name of the workspace to list worktrees for (optional)
	RepositoryName string // Name of the repository to list worktrees for (optional)
}

// ListWorktrees lists worktrees for a workspace or repository.
func (c *realCodeManager) ListWorktrees(opts ...ListWorktreesOpts) ([]status.WorktreeInfo, error) {
	// Parse options
	options := c.extractListWorktreesOptions(opts)

	// Validate that workspace and repository are not both specified
	if options.WorkspaceName != "" && options.RepositoryName != "" {
		return nil, fmt.Errorf("cannot specify both WorkspaceName and RepositoryName")
	}

	// Prepare parameters for hooks
	params := map[string]interface{}{
		"workspace_name":  options.WorkspaceName,
		"repository_name": options.RepositoryName,
	}

	// Execute with hooks
	var result []status.WorktreeInfo
	err := c.executeWithHooks(consts.ListWorktrees, params, func() error {
		var err error
		c.VerbosePrint("Listing worktrees")

		// Detect project mode and list accordingly
		projectType, err := c.detectProjectMode(options.WorkspaceName, options.RepositoryName)
		if err != nil {
			return fmt.Errorf("failed to detect project mode: %w", err)
		}

		switch projectType {
		case mode.ModeSingleRepo:
			if options.RepositoryName != "" {
				result, err = c.listRepositoryWorktrees(options.RepositoryName)
				return err
			}
			// Create repository instance for current directory
			repoInstance := c.repositoryProvider(repo.NewRepositoryParams{
				FS:               c.fs,
				Git:              c.git,
				ConfigManager:    c.configManager,
				StatusManager:    c.statusManager,
				Logger:           c.logger,
				Prompt:           c.prompt,
				WorktreeProvider: worktree.NewWorktree,
				HookManager:      c.hookManager,
				RepositoryName:   ".",
			})
			worktrees, err := repoInstance.ListWorktrees()
			result = worktrees
			return c.translateListError(err)
		case mode.ModeWorkspace:
			result, err = c.listWorkspaceWorktrees(options.WorkspaceName)
			return err
		case mode.ModeNone:
			return ErrNoGitRepositoryOrWorkspaceFound
		default:
			return fmt.Errorf("unknown project type")
		}
	})
	return result, err
}

// listWorkspaceWorktrees lists all worktrees associated with a workspace.
func (c *realCodeManager) listWorkspaceWorktrees(workspaceName string) ([]status.WorktreeInfo, error) {
	c.VerbosePrint("Listing worktrees for workspace: %s", workspaceName)

	// Get workspace from status
	workspace, err := c.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return c.listWorkspaceWorktreesFromWorkspace(workspace)
}

// listWorkspaceWorktreesFromWorkspace lists all worktrees from a workspace object.
func (c *realCodeManager) listWorkspaceWorktreesFromWorkspace(
	workspace *status.Workspace) ([]status.WorktreeInfo, error) {
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
func (c *realCodeManager) translateListError(err error) error {
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

// listRepositoryWorktrees lists worktrees for a specific repository.
func (c *realCodeManager) listRepositoryWorktrees(repositoryName string) ([]status.WorktreeInfo, error) {
	c.VerbosePrint("Listing worktrees for repository: %s", repositoryName)

	// Create repository instance - let repositoryProvider handle repository name resolution
	repoInstance := c.repositoryProvider(repo.NewRepositoryParams{
		FS:               c.fs,
		Git:              c.git,
		ConfigManager:    c.configManager,
		StatusManager:    c.statusManager,
		Logger:           c.logger,
		Prompt:           c.prompt,
		WorktreeProvider: worktree.NewWorktree,
		HookManager:      c.hookManager,
		RepositoryName:   repositoryName, // Pass repository name directly, let provider handle resolution
	})

	// List worktrees
	worktrees, err := repoInstance.ListWorktrees()
	if err != nil {
		return nil, c.translateListError(err)
	}

	return worktrees, nil
}

// extractListWorktreesOptions extracts and merges options from the variadic parameter.
func (c *realCodeManager) extractListWorktreesOptions(opts []ListWorktreesOpts) ListWorktreesOpts {
	var result ListWorktreesOpts

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
