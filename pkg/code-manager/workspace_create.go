package codemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/status"
)

// CreateWorkspaceParams contains parameters for CreateWorkspace.
type CreateWorkspaceParams struct {
	WorkspaceName string   // Name of the workspace
	Repositories  []string // Repository identifiers (names, paths, URLs)
}

// DeleteWorkspaceParams contains parameters for DeleteWorkspace.
type DeleteWorkspaceParams struct {
	WorkspaceName string // Name of the workspace to delete
	Force         bool   // Skip confirmation prompts
}

// CreateWorkspace creates a new workspace with repository selection.
func (c *realCodeManager) CreateWorkspace(params CreateWorkspaceParams) error {
	return c.executeWithHooks("create_workspace", map[string]interface{}{
		"workspace_name": params.WorkspaceName,
		"repositories":   params.Repositories,
	}, func() error {
		return c.createWorkspace(params)
	})
}

// createWorkspace implements the workspace creation business logic.
func (c *realCodeManager) createWorkspace(params CreateWorkspaceParams) error {
	c.VerbosePrint("Creating workspace: %s", params.WorkspaceName)

	// Validate workspace name
	if err := c.validateWorkspaceName(params.WorkspaceName); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidWorkspaceName, err)
	}

	// Check if workspace already exists
	existingWorkspace, err := c.deps.StatusManager.GetWorkspace(params.WorkspaceName)
	if err == nil && existingWorkspace != nil {
		return fmt.Errorf("%w: workspace '%s' already exists", ErrWorkspaceAlreadyExists, params.WorkspaceName)
	}

	// Validate and resolve repositories
	resolvedRepos, err := c.validateAndResolveRepositories(params.Repositories)
	if err != nil {
		return fmt.Errorf("failed to validate repositories: %w", err)
	}

	// Add new repositories to status file if needed
	finalRepos, err := c.addRepositoriesToStatus(resolvedRepos)
	if err != nil {
		return fmt.Errorf("failed to add repositories to status: %w", err)
	}

	// Add workspace to status file
	workspaceParams := status.AddWorkspaceParams{
		Repositories: finalRepos,
	}
	if err := c.deps.StatusManager.AddWorkspace(params.WorkspaceName, workspaceParams); err != nil {
		return fmt.Errorf("%w: %w", ErrStatusUpdate, err)
	}

	c.VerbosePrint("Workspace created successfully")
	return nil
}

// validateWorkspaceName validates the workspace name.
func (c *realCodeManager) validateWorkspaceName(name string) error {
	if name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}
	// Check for invalid characters (basic validation)
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return fmt.Errorf("workspace name contains invalid characters")
	}
	return nil
}

// validateAndResolveRepositories validates and resolves repository paths/names.
func (c *realCodeManager) validateAndResolveRepositories(repositories []string) ([]string, error) {
	if len(repositories) == 0 {
		return nil, fmt.Errorf("at least one repository must be specified")
	}

	var resolvedRepos []string
	seenRepos := make(map[string]bool)

	for _, repo := range repositories {
		if repo == "" {
			return nil, fmt.Errorf("repository identifier cannot be empty")
		}

		// Check for duplicates
		if seenRepos[repo] {
			return nil, fmt.Errorf("%w: repository '%s' specified multiple times", ErrDuplicateRepository, repo)
		}
		seenRepos[repo] = true

		// Resolve repository path/name
		resolvedRepo, err := c.resolveRepository(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve repository '%s': %w", repo, err)
		}

		resolvedRepos = append(resolvedRepos, resolvedRepo)
	}

	return resolvedRepos, nil
}

// resolveRepository resolves a single repository identifier.
func (c *realCodeManager) resolveRepository(repo string) (string, error) {
	// First, check if it's a repository name from status.yaml
	if existingRepo, err := c.deps.StatusManager.GetRepository(repo); err == nil && existingRepo != nil {
		c.VerbosePrint("  ✓ %s (from status.yaml): %s", repo, existingRepo.Path)
		return repo, nil
	}

	// Check if it's an absolute path
	if filepath.IsAbs(repo) {
		return c.validateRepositoryPath(repo)
	}

	// Resolve relative path from current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("%w: failed to get current directory: %w", ErrPathResolution, err)
	}

	resolvedPath, err := c.deps.FS.ResolvePath(currentDir, repo)
	if err != nil {
		return "", fmt.Errorf("%w: failed to resolve path '%s': %w", ErrPathResolution, repo, err)
	}

	return c.validateRepositoryPath(resolvedPath)
}

// validateRepositoryPath validates that a path contains a Git repository.
func (c *realCodeManager) validateRepositoryPath(path string) (string, error) {
	// Check if path exists
	exists, err := c.deps.FS.Exists(path)
	if err != nil {
		return "", fmt.Errorf("%w: failed to check if path exists: %w", ErrRepositoryNotFound, err)
	}
	if !exists {
		return "", fmt.Errorf("%w: path '%s' does not exist", ErrRepositoryNotFound, path)
	}

	// Validate it's a Git repository
	isValid, err := c.deps.FS.ValidateRepositoryPath(path)
	if err != nil {
		return "", fmt.Errorf("%w: failed to validate repository: %w", ErrInvalidRepository, err)
	}
	if !isValid {
		return "", fmt.Errorf("%w: path '%s' is not a valid Git repository", ErrInvalidRepository, path)
	}

	c.VerbosePrint("  ✓ %s: %s", path, path)
	return path, nil
}

// addRepositoriesToStatus adds new repositories to status file and returns final repository URLs.
func (c *realCodeManager) addRepositoriesToStatus(repositories []string) ([]string, error) {
	var finalRepos []string
	seenURLs := make(map[string]bool)

	for _, repo := range repositories {
		// Get repository URL from Git remote origin
		repoURL, err := c.deps.Git.GetRemoteURL(repo, "origin")
		if err != nil {
			// If no origin remote, use the path as the identifier
			repoURL = repo
		}

		// Check for duplicate remote URLs within this workspace
		if seenURLs[repoURL] {
			return nil, fmt.Errorf("%w: repository with URL '%s' already exists in this workspace",
				ErrDuplicateRepository, repoURL)
		}
		seenURLs[repoURL] = true

		// Check if repository already exists in status using the remote URL
		if existingRepo, err := c.deps.StatusManager.GetRepository(repoURL); err == nil && existingRepo != nil {
			finalRepos = append(finalRepos, repoURL)
			c.VerbosePrint("  ✓ %s (already exists in status)", repo)
			continue
		}

		// Add new repository to status file
		repoURL, err = c.addRepositoryToStatus(repo)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to add repository '%s': %w", ErrRepositoryAddition, repo, err)
		}

		finalRepos = append(finalRepos, repoURL)
		c.VerbosePrint("  ✓ %s (added to status)", repo)
	}

	return finalRepos, nil
}

// addRepositoryToStatus adds a new repository to the status file and returns its URL.
// It clones the repository to the managed location if it's not already there.
func (c *realCodeManager) addRepositoryToStatus(repoPath string) (string, error) {
	// Get repository URL from Git remote origin
	remoteURL, err := c.deps.Git.GetRemoteURL(repoPath, "origin")
	if err != nil {
		// If no origin remote, use the path as the identifier
		// In this case, we can't clone from remote, so we'll use the local path
		repoParams := status.AddRepositoryParams{
			Path: repoPath,
		}
		if err := c.deps.StatusManager.AddRepository(repoPath, repoParams); err != nil {
			return "", fmt.Errorf("failed to add repository to status file: %w", err)
		}
		return repoPath, nil
	}

	// Normalize the repository URL
	normalizedURL, err := c.normalizeRepositoryURL(remoteURL)
	if err != nil {
		return "", fmt.Errorf("failed to normalize repository URL: %w", err)
	}

	// Check if repository already exists in status
	if existingRepo, err := c.deps.StatusManager.GetRepository(normalizedURL); err == nil && existingRepo != nil {
		// Repository already exists, return the normalized URL
		return normalizedURL, nil
	}

	// Detect default branch from remote
	defaultBranch, err := c.deps.Git.GetDefaultBranch(remoteURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToDetectDefaultBranch, err)
	}

	// Generate target path for cloning
	targetPath := c.generateClonePath(normalizedURL, defaultBranch)

	// Create parent directories for the target path
	parentDir := filepath.Dir(targetPath)
	if err := c.deps.FS.MkdirAll(parentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Clone repository from remote URL to managed location
	if err := c.deps.Git.Clone(git.CloneParams{
		RepoURL:    remoteURL,
		TargetPath: targetPath,
		Recursive:  true,
	}); err != nil {
		return "", fmt.Errorf("%w: %w", ErrFailedToCloneRepository, err)
	}

	// Add repository to status file with the managed path
	remotes := map[string]status.Remote{
		"origin": {
			DefaultBranch: defaultBranch,
		},
	}
	repoParams := status.AddRepositoryParams{
		Path:    targetPath,
		Remotes: remotes,
	}
	if err := c.deps.StatusManager.AddRepository(normalizedURL, repoParams); err != nil {
		return "", fmt.Errorf("failed to add repository to status file: %w", err)
	}

	// Note: We don't automatically add the default branch worktree to status here.
	// The worktree will be added when it's actually needed (e.g., when creating worktrees
	// for a workspace). This avoids adding unnecessary worktrees to status when they're
	// not going to be used.

	return normalizedURL, nil
}
