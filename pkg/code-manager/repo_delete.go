package codemanager

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/status"
)

// DeleteRepositoryParams contains parameters for DeleteRepository.
type DeleteRepositoryParams struct {
	RepositoryName string // Name of the repository to delete
	Force          bool   // Skip confirmation prompts
}

// DeleteRepository deletes a repository and all associated resources.
func (c *realCodeManager) DeleteRepository(params DeleteRepositoryParams) error {
	return c.executeWithHooks(consts.DeleteRepository, map[string]interface{}{
		"repository_name": params.RepositoryName,
		"force":           params.Force,
	}, func() error {
		return c.deleteRepository(params)
	})
}

// deleteRepository implements the repository deletion business logic.
func (c *realCodeManager) deleteRepository(params DeleteRepositoryParams) error {
	c.VerbosePrint("Deleting repository: %s", params.RepositoryName)

	// Validate repository name
	if err := c.validateRepositoryName(params.RepositoryName); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRepositoryName, err)
	}

	// Check if repository is part of any workspace
	if err := c.validateRepositoryNotInWorkspace(params.RepositoryName); err != nil {
		return err
	}

	// Get repository and worktrees
	repository, worktrees, err := c.getRepositoryAndWorktrees(params.RepositoryName)
	if err != nil {
		return err
	}

	// Show confirmation prompt with deletion summary (unless --force)
	if !params.Force {
		if err := c.showRepositoryDeletionConfirmation(params.RepositoryName, repository, worktrees); err != nil {
			return fmt.Errorf("deletion cancelled: %w", err)
		}
	}

	// Perform repository deletion steps
	if err := c.performRepositoryDeletion(params.RepositoryName, repository, worktrees, params.Force); err != nil {
		return err
	}

	c.VerbosePrint("Repository '%s' deleted successfully", params.RepositoryName)
	return nil
}

// getRepositoryAndWorktrees retrieves repository and associated worktrees.
func (c *realCodeManager) getRepositoryAndWorktrees(repositoryName string) (
	*status.Repository, []status.WorktreeInfo, error) {
	// Get repository from status
	repository, err := c.statusManager.GetRepository(repositoryName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Collect all worktrees for this repository
	var worktrees []status.WorktreeInfo
	for _, worktree := range repository.Worktrees {
		worktrees = append(worktrees, worktree)
	}

	return repository, worktrees, nil
}

// showRepositoryDeletionConfirmation shows a confirmation prompt with deletion summary.
func (c *realCodeManager) showRepositoryDeletionConfirmation(
	repositoryName string,
	repository *status.Repository,
	worktrees []status.WorktreeInfo,
) error {
	// Build confirmation message
	var message strings.Builder
	message.WriteString(fmt.Sprintf("Are you sure you want to delete repository '%s'?\n\n", repositoryName))

	message.WriteString("This will delete:\n")
	message.WriteString(fmt.Sprintf("  • Repository: %s\n", repositoryName))
	message.WriteString(fmt.Sprintf("  • Repository path: %s\n", repository.Path))
	message.WriteString(fmt.Sprintf("  • %d worktree(s)\n", len(worktrees)))

	// Get config from ConfigManager
	cfg, err := c.configManager.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Check if repository directory will be deleted
	inBasePath, err := c.fs.IsPathWithinBase(cfg.RepositoriesDir, repository.Path)
	switch {
	case err != nil:
		// If we can't determine, show a warning
		message.WriteString("  • Repository directory (validation failed)\n")
	case inBasePath:
		message.WriteString("  • Repository directory\n")
	default:
		message.WriteString("  • Repository directory (outside base path - will be skipped)\n")
	}

	if len(worktrees) > 0 {
		message.WriteString("  • Worktrees:\n")
		for _, worktree := range worktrees {
			message.WriteString(fmt.Sprintf("    - %s/%s\n", worktree.Remote, worktree.Branch))
		}
	}

	message.WriteString("\nThis action cannot be undone. Type 'yes' to confirm:")

	// Show confirmation prompt
	confirmed, err := c.prompt.PromptForConfirmation(message.String(), false)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		return fmt.Errorf("deletion cancelled by user")
	}

	return nil
}

// performRepositoryDeletion performs the actual repository deletion steps.
func (c *realCodeManager) performRepositoryDeletion(
	repositoryName string,
	repository *status.Repository,
	worktrees []status.WorktreeInfo,
	force bool,
) error {
	c.VerbosePrint("Performing repository deletion steps")

	// Step 1: Delete all worktrees
	if err := c.deleteRepositoryWorktrees(repositoryName, worktrees, force); err != nil {
		return fmt.Errorf("failed to delete worktrees: %w", err)
	}

	// Step 2: Remove repository from status
	if err := c.statusManager.RemoveRepository(repositoryName); err != nil {
		return fmt.Errorf("failed to remove repository from status: %w", err)
	}

	// Step 3: Delete repository directory (if it exists and is within base path)
	if err := c.deleteRepositoryDirectory(repository.Path); err != nil {
		c.VerbosePrint("Warning: failed to delete repository directory: %v", err)
		// Don't fail the entire operation if directory deletion fails
	}

	return nil
}

// deleteRepositoryWorktrees deletes all worktrees for the repository.
func (c *realCodeManager) deleteRepositoryWorktrees(
	repositoryName string, worktrees []status.WorktreeInfo, force bool) error {
	if len(worktrees) == 0 {
		c.VerbosePrint("No worktrees to delete for repository: %s", repositoryName)
		return nil
	}

	c.VerbosePrint("Deleting %d worktrees for repository: %s", len(worktrees), repositoryName)

	// Delete each worktree
	for _, worktree := range worktrees {
		c.VerbosePrint("Deleting worktree: %s/%s", worktree.Remote, worktree.Branch)

		if err := c.DeleteWorkTree(worktree.Branch, force, DeleteWorktreeOpts{
			RepositoryName: repositoryName,
		}); err != nil {
			return fmt.Errorf("failed to delete worktree %s/%s: %w", worktree.Remote, worktree.Branch, err)
		}
	}

	return nil
}

// deleteRepositoryDirectory deletes the repository directory if it's within the base path.
func (c *realCodeManager) deleteRepositoryDirectory(repositoryPath string) error {
	// Get config from ConfigManager
	cfg, err := c.configManager.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Check if the repository path is within the configured base path
	inBasePath, err := c.fs.IsPathWithinBase(cfg.RepositoriesDir, repositoryPath)
	if err != nil {
		c.VerbosePrint("Warning: failed to validate base path for repository %s: %v", repositoryPath, err)
		// Default to not deleting if validation fails
		return nil
	}

	if !inBasePath {
		c.VerbosePrint("Repository path %s is not within base path %s, skipping directory deletion",
			repositoryPath, cfg.RepositoriesDir)
		return nil
	}

	// Check if directory exists
	exists, err := c.fs.Exists(repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to check if repository directory exists: %w", err)
	}

	if !exists {
		c.VerbosePrint("Repository directory %s does not exist, skipping deletion", repositoryPath)
		return nil
	}

	// Delete the directory
	c.VerbosePrint("Deleting repository directory: %s", repositoryPath)
	if err := c.fs.RemoveAll(repositoryPath); err != nil {
		return fmt.Errorf("failed to delete repository directory: %w", err)
	}

	// Clean up empty parent directories up to the repositories directory
	if err := c.cleanupEmptyParentDirectories(repositoryPath); err != nil {
		c.VerbosePrint("Warning: failed to cleanup empty parent directories: %v", err)
		// Don't fail the entire operation if cleanup fails
	}

	return nil
}

// cleanupEmptyParentDirectories removes empty parent directories up to the repositories directory.
func (c *realCodeManager) cleanupEmptyParentDirectories(repositoryPath string) error {
	// Get config from ConfigManager
	cfg, err := c.configManager.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Get the parent directory of the repository
	parentDir := filepath.Dir(repositoryPath)

	// Keep cleaning up parent directories until we reach the repositories directory
	for parentDir != cfg.RepositoriesDir && parentDir != filepath.Dir(parentDir) {
		// Check if the directory exists
		exists, err := c.fs.Exists(parentDir)
		if err != nil {
			return fmt.Errorf("failed to check if parent directory exists: %w", err)
		}

		if !exists {
			// Directory doesn't exist, move up to next parent
			parentDir = filepath.Dir(parentDir)
			continue
		}

		// Check if the directory is empty
		isEmpty, err := c.isDirectoryEmpty(parentDir)
		if err != nil {
			return fmt.Errorf("failed to check if directory is empty: %w", err)
		}

		if !isEmpty {
			// Directory is not empty, stop cleanup
			c.VerbosePrint("Parent directory %s is not empty, stopping cleanup", parentDir)
			break
		}

		// Directory is empty, remove it
		c.VerbosePrint("Removing empty parent directory: %s", parentDir)
		if err := c.fs.Remove(parentDir); err != nil {
			return fmt.Errorf("failed to remove empty parent directory %s: %w", parentDir, err)
		}

		// Move up to next parent
		parentDir = filepath.Dir(parentDir)
	}

	return nil
}

// isDirectoryEmpty checks if a directory is empty.
func (c *realCodeManager) isDirectoryEmpty(dirPath string) (bool, error) {
	entries, err := c.fs.ReadDir(dirPath)
	if err != nil {
		return false, fmt.Errorf("failed to read directory: %w", err)
	}

	return len(entries) == 0, nil
}

// validateRepositoryName validates the repository name.
func (c *realCodeManager) validateRepositoryName(repositoryName string) error {
	if repositoryName == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	// Check for reserved names
	reservedNames := []string{".", "..", "status.yaml", "config.yaml"}
	for _, reserved := range reservedNames {
		if repositoryName == reserved {
			return fmt.Errorf("repository name '%s' is reserved", reserved)
		}
	}

	// Allow repository names that are URLs (contain slashes) or simple names
	// Only reject if it contains backslashes (Windows path separators)
	if strings.Contains(repositoryName, "\\") {
		return fmt.Errorf("repository name cannot contain backslashes")
	}

	return nil
}

// validateRepositoryNotInWorkspace checks if the repository is part of any workspace.
func (c *realCodeManager) validateRepositoryNotInWorkspace(repositoryName string) error {
	// Get all workspaces
	workspaces, err := c.statusManager.ListWorkspaces()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	// Check if repository is referenced in any workspace
	for workspaceName, workspace := range workspaces {
		if workspace.HasRepository(repositoryName) {
			return fmt.Errorf(
				"repository '%s' is part of workspace '%s'. Remove it from the workspace before deleting",
				repositoryName,
				workspaceName,
			)
		}
	}

	return nil
}
