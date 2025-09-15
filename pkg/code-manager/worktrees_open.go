package codemanager

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/mode"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
)

// OpenWorktreeOpts contains optional parameters for OpenWorktree.
type OpenWorktreeOpts struct {
	WorkspaceName  string // Name of the workspace to open worktree for (optional)
	RepositoryName string // Name of the repository to open worktree for (optional)
}

// OpenWorktree opens an existing worktree in the specified IDE.
func (c *realCodeManager) OpenWorktree(worktreeName, ideName string, opts ...OpenWorktreeOpts) error {
	// Parse options
	options := c.extractOpenWorktreeOptions(opts)

	// Validate that workspace and repository are not both specified
	if options.WorkspaceName != "" && options.RepositoryName != "" {
		return fmt.Errorf("cannot specify both WorkspaceName and RepositoryName")
	}

	// Prepare parameters for hooks
	params := c.prepareOpenWorktreeParams(worktreeName, ideName, options)

	// Execute with hooks
	return c.executeWithHooks(consts.OpenWorktree, params, func() error {
		return c.executeOpenWorktreeLogic(worktreeName, ideName, options, params)
	})
}

func (c *realCodeManager) prepareOpenWorktreeParams(
	worktreeName, ideName string,
	options OpenWorktreeOpts,
) map[string]interface{} {
	return map[string]interface{}{
		"worktreeName":    worktreeName,
		"ideName":         ideName,
		"workspace_name":  options.WorkspaceName,
		"repository_name": options.RepositoryName,
	}
}

func (c *realCodeManager) executeOpenWorktreeLogic(
	worktreeName, ideName string,
	options OpenWorktreeOpts,
	params map[string]interface{},
) error {
	c.VerbosePrint("Opening worktree: %s in IDE: %s", worktreeName, ideName)

	// Detect project mode
	projectType, err := c.detectProjectMode(options.WorkspaceName, options.RepositoryName)
	if err != nil {
		return fmt.Errorf("failed to detect project mode: %w", err)
	}

	return c.handleOpenWorktreeByProjectType(projectType, worktreeName, options, params)
}

func (c *realCodeManager) handleOpenWorktreeByProjectType(
	projectType mode.Mode,
	worktreeName string,
	options OpenWorktreeOpts,
	params map[string]interface{},
) error {
	switch projectType {
	case mode.ModeSingleRepo:
		return c.handleSingleRepoOpenWorktree(worktreeName, options, params)
	case mode.ModeWorkspace:
		return fmt.Errorf("workspace mode open worktree not yet implemented")
	case mode.ModeNone:
		return ErrNoGitRepositoryOrWorkspaceFound
	default:
		return fmt.Errorf("unknown project type")
	}
}

func (c *realCodeManager) handleSingleRepoOpenWorktree(
	worktreeName string,
	options OpenWorktreeOpts,
	params map[string]interface{},
) error {
	if options.RepositoryName != "" {
		return c.handleRepositorySpecificOpenWorktree(options.RepositoryName, worktreeName, params)
	}
	return c.handleDefaultSingleRepoOpenWorktree(worktreeName, params)
}

func (c *realCodeManager) handleRepositorySpecificOpenWorktree(
	repositoryName, worktreeName string,
	params map[string]interface{},
) error {
	worktreePath, err := c.openWorktreeForRepository(repositoryName, worktreeName)
	if err != nil {
		return err
	}
	params["worktreePath"] = worktreePath
	return nil
}

func (c *realCodeManager) handleDefaultSingleRepoOpenWorktree(
	worktreeName string, params map[string]interface{}) error {
	// For single repository, worktreeName is the branch name
	// Get repository URL from local .git directory
	repoURL, err := c.deps.Git.GetRepositoryName(".")
	if err != nil {
		return fmt.Errorf("failed to get repository URL: %w", err)
	}

	// Check if the worktree exists in the status file
	worktreeInfo, err := c.deps.StatusManager.GetWorktree(repoURL, worktreeName)
	if err != nil {
		return ErrWorktreeNotInStatus
	}

	// Build the worktree path using the remote from status
	worktreePath := c.BuildWorktreePath(repoURL, worktreeInfo.Remote, worktreeName)

	// Store the worktree path in parameters for the hook to access
	params["worktreePath"] = worktreePath
	return nil
}

// openWorktreeForRepository opens a worktree for a specific repository.
func (c *realCodeManager) openWorktreeForRepository(repositoryName, worktreeName string) (string, error) {
	c.VerbosePrint("Opening worktree for repository: %s, worktree: %s", repositoryName, worktreeName)

	// Create repository instance - let repositoryProvider handle repository name resolution
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: repositoryName, // Pass repository name directly, let provider handle resolution
	})

	// Validate repository and get repository URL
	validationResult, err := repoInstance.ValidateRepository(repo.ValidationParams{})
	if err != nil {
		return "", fmt.Errorf("failed to validate repository: %w", err)
	}
	repoURL := validationResult.RepoURL

	// Check if the worktree exists in the status file
	worktreeInfo, err := c.deps.StatusManager.GetWorktree(repoURL, worktreeName)
	if err != nil {
		return "", ErrWorktreeNotInStatus
	}

	// Build the worktree path using the remote from status
	worktreePath := c.BuildWorktreePath(repoURL, worktreeInfo.Remote, worktreeName)

	return worktreePath, nil
}

// extractOpenWorktreeOptions extracts and merges options from the variadic parameter.
func (c *realCodeManager) extractOpenWorktreeOptions(opts []OpenWorktreeOpts) OpenWorktreeOpts {
	var result OpenWorktreeOpts

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
