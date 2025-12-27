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
		rawRepoURL, err := c.deps.Git.GetRemoteURL(repo, "origin")
		if err != nil {
			// If no origin remote, use the path as the identifier
			finalRepos = append(finalRepos, repo)
			c.VerbosePrint("  ✓ %s (no remote, using path)", repo)
			continue
		}

		// Normalize the repository URL before checking status
		// This ensures consistent format (host/path) regardless of URL protocol (ssh://, git@, https://)
		normalizedRepoURL, err := c.normalizeRepositoryURL(rawRepoURL)
		if err != nil {
			// If normalization fails, fall back to using the path as the identifier
			c.VerbosePrint("  ⚠ Failed to normalize repository URL '%s': %v, using path as identifier", rawRepoURL, err)
			finalRepos = append(finalRepos, repo)
			continue
		}

		// Check for duplicate remote URLs within this workspace (using normalized URL)
		if seenURLs[normalizedRepoURL] {
			return nil, fmt.Errorf("%w: repository with URL '%s' already exists in this workspace",
				ErrDuplicateRepository, normalizedRepoURL)
		}
		seenURLs[normalizedRepoURL] = true

		// Check if repository already exists in status using the normalized URL
		if existingRepo, err := c.deps.StatusManager.GetRepository(normalizedRepoURL); err == nil && existingRepo != nil {
			finalRepos = append(finalRepos, normalizedRepoURL)
			c.VerbosePrint("  ✓ %s (already exists in status)", repo)
			continue
		}

		// Add new repository to status file
		finalRepoURL, err := c.addRepositoryToStatus(repo)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to add repository '%s': %w", ErrRepositoryAddition, repo, err)
		}

		finalRepos = append(finalRepos, finalRepoURL)
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

	// Detect default branch from remote, fallback to local repository if remote is not accessible
	defaultBranch := c.getDefaultBranchWithFallback(remoteURL, repoPath)

	// Determine target path (use local if valid, otherwise clone)
	targetPath, err := c.determineRepositoryTargetPath(remoteURL, normalizedURL, defaultBranch, repoPath)
	if err != nil {
		return "", err
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

// getDefaultBranchWithFallback gets the default branch from remote, falling back to local repository if needed.
func (c *realCodeManager) getDefaultBranchWithFallback(remoteURL, repoPath string) string {
	defaultBranch, err := c.deps.Git.GetDefaultBranch(remoteURL)
	if err != nil {
		// If remote is not accessible, try to get the current branch from the local repository
		c.VerbosePrint("Warning: failed to get default branch from remote '%s', trying local repository: %v",
			remoteURL, err)
		currentBranch, localErr := c.deps.Git.GetCurrentBranch(repoPath)
		if localErr != nil {
			// If we can't get the local branch either, fall back to "main"
			c.VerbosePrint("Warning: failed to get current branch from local repository, using 'main' as default: %v",
				localErr)
			return "main"
		}
		c.VerbosePrint("Using local repository's current branch '%s' as default branch", currentBranch)
		return currentBranch
	}
	return defaultBranch
}

// cloneOrUseLocalPath attempts to clone the repository, falling back to local path if cloning fails.
func (c *realCodeManager) cloneOrUseLocalPath(
	remoteURL, normalizedURL, defaultBranch, repoPath string,
) (string, error) {
	// Generate target path for cloning
	targetPath := c.generateClonePath(normalizedURL, defaultBranch)

	// Try to clone repository from remote URL to managed location
	cloneErr := func() error {
		// Create parent directories for the target path
		parentDir := filepath.Dir(targetPath)
		if err := c.deps.FS.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directories: %w", err)
		}

		// Clone repository from remote URL to managed location
		if err := c.deps.Git.Clone(git.CloneParams{
			RepoURL:    remoteURL,
			TargetPath: targetPath,
			Recursive:  true,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCloneRepository, err)
		}
		return nil
	}()

	// If cloning failed, check if the original path is a valid local repository
	if cloneErr != nil {
		c.VerbosePrint("Warning: failed to clone repository from remote '%s', checking if local path is valid: %v",
			remoteURL, cloneErr)
		// Validate that the original path is a valid Git repository
		isValid, err := c.deps.FS.ValidateRepositoryPath(repoPath)
		if err != nil || !isValid {
			// Original path is not a valid repository, return the clone error
			return "", cloneErr
		}
		// Use the original local path instead of the cloned path
		c.VerbosePrint("Using existing local repository path '%s' instead of cloning", repoPath)
		return repoPath, nil
	}

	return targetPath, nil
}

// determineRepositoryTargetPath determines the target path for a repository,
// using local path if valid, otherwise cloning.
func (c *realCodeManager) determineRepositoryTargetPath(
	remoteURL, normalizedURL, defaultBranch, repoPath string,
) (string, error) {
	// If repoPath looks like a local path, check if it's valid and use it
	if c.isLocalPath(repoPath) {
		if c.isValidLocalRepository(repoPath) {
			c.VerbosePrint("Using existing local repository path '%s' instead of cloning", repoPath)
			return repoPath, nil
		}
	}

	// Not a valid local path, proceed with cloning
	return c.cloneOrUseLocalPath(remoteURL, normalizedURL, defaultBranch, repoPath)
}

// isLocalPath checks if a path looks like a local file system path.
func (c *realCodeManager) isLocalPath(repoPath string) bool {
	return filepath.IsAbs(repoPath) || strings.Contains(repoPath, string(filepath.Separator))
}

// isValidLocalRepository checks if a local path is a valid Git repository.
func (c *realCodeManager) isValidLocalRepository(repoPath string) bool {
	isValid, err := c.deps.FS.ValidateRepositoryPath(repoPath)
	return err == nil && isValid
}
