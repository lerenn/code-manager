package codemanager

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/branch"
	"github.com/lerenn/code-manager/pkg/status"
)

// DeleteWorkspace deletes a workspace and all associated resources.
func (c *realCodeManager) DeleteWorkspace(params DeleteWorkspaceParams) error {
	return c.executeWithHooks("delete_workspace", map[string]interface{}{
		"workspace_name": params.WorkspaceName,
		"force":          params.Force,
	}, func() error {
		return c.deleteWorkspace(params)
	})
}

// deleteWorkspace implements the workspace deletion business logic.
func (c *realCodeManager) deleteWorkspace(params DeleteWorkspaceParams) error {
	c.VerbosePrint("Deleting workspace: %s", params.WorkspaceName)

	// Validate workspace name
	if err := c.validateWorkspaceName(params.WorkspaceName); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidWorkspaceName, err)
	}

	// Get workspace and worktrees
	workspace, worktrees, err := c.getWorkspaceAndWorktrees(params.WorkspaceName)
	if err != nil {
		return err
	}

	// Show confirmation prompt with deletion summary (unless --force)
	if !params.Force {
		if err := c.showDeletionConfirmation(params.WorkspaceName, workspace, worktrees); err != nil {
			return fmt.Errorf("deletion cancelled: %w", err)
		}
	}

	// Perform workspace deletion steps
	if err := c.performWorkspaceDeletion(params.WorkspaceName, workspace, worktrees, params.Force); err != nil {
		return err
	}

	c.VerbosePrint("Workspace '%s' deleted successfully", params.WorkspaceName)
	return nil
}

// getWorkspaceAndWorktrees retrieves workspace and associated worktrees.
func (c *realCodeManager) getWorkspaceAndWorktrees(workspaceName string) (
	*status.Workspace, []status.WorktreeInfo, error) {
	// Check if workspace exists and get it
	workspace, err := c.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, nil, fmt.Errorf("%w: workspace '%s' not found", ErrWorkspaceNotFound, workspaceName)
		}
		return nil, nil, fmt.Errorf("failed to check workspace existence: %w", err)
	}

	// List all worktrees associated with the workspace
	worktrees, err := c.listWorkspaceWorktreesFromWorkspace(workspace)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list workspace worktrees: %w", err)
	}

	// Debug: Print worktrees found
	c.VerbosePrint("Found %d worktrees for workspace %s:", len(worktrees), workspaceName)
	for i, worktree := range worktrees {
		c.VerbosePrint("  [%d] %s/%s", i, worktree.Remote, worktree.Branch)
	}

	return workspace, worktrees, nil
}

// performWorkspaceDeletion performs all the deletion steps for a workspace.
func (c *realCodeManager) performWorkspaceDeletion(
	workspaceName string,
	workspace *status.Workspace,
	worktrees []status.WorktreeInfo,
	force bool,
) error {
	// Delete all worktrees associated with the workspace
	if err := c.deleteWorkspaceWorktrees(workspace, worktrees, force); err != nil {
		return fmt.Errorf("failed to delete workspace worktrees: %w", err)
	}

	// Remove worktree entries from workspace status
	if err := c.removeWorktreesFromWorkspaceStatus(workspaceName, worktrees); err != nil {
		return fmt.Errorf("failed to remove worktrees from workspace status: %w", err)
	}

	// Delete workspace files
	if err := c.deleteWorkspaceFiles(workspaceName, worktrees); err != nil {
		return fmt.Errorf("failed to delete workspace files: %w", err)
	}

	// Remove workspace entry from status file
	if err := c.removeWorkspaceFromStatus(workspaceName); err != nil {
		return fmt.Errorf("failed to remove workspace from status: %w", err)
	}

	return nil
}

// showDeletionConfirmation shows a confirmation prompt with detailed deletion summary.
func (c *realCodeManager) showDeletionConfirmation(
	workspaceName string,
	workspace *status.Workspace,
	_ []status.WorktreeInfo, // Unused parameter - worktree info is derived from workspace
) error {
	// Group worktrees by repository for better display
	repoWorktrees := c.groupWorktreesByRepository(workspace)

	// Count unique worktree references (not total worktrees across all repos)
	uniqueWorktrees := len(workspace.Worktrees)

	// Build confirmation message
	message := c.buildDeletionConfirmationMessage(workspaceName, uniqueWorktrees, repoWorktrees)

	// Get user confirmation
	confirmed, err := c.prompt.PromptForConfirmation(message, false)
	if err != nil {
		return fmt.Errorf("failed to get user confirmation: %w", err)
	}

	if !confirmed {
		return fmt.Errorf("deletion cancelled by user")
	}

	return nil
}

// groupWorktreesByRepository groups worktrees by repository for display purposes.
func (c *realCodeManager) groupWorktreesByRepository(workspace *status.Workspace) map[string][]string {
	repoWorktrees := make(map[string][]string)

	// Process each repository in the workspace to get accurate counts
	for _, repoURL := range workspace.Repositories {
		repoBranches := c.findWorktreesForRepository(workspace, repoURL)
		if len(repoBranches) > 0 {
			repoWorktrees[repoURL] = repoBranches
		}
	}

	return repoWorktrees
}

// findWorktreesForRepository finds worktrees for a specific repository.
func (c *realCodeManager) findWorktreesForRepository(workspace *status.Workspace, repoURL string) []string {
	// Get repository to check its worktrees
	repo, err := c.statusManager.GetRepository(repoURL)
	if err != nil {
		return nil
	}

	// Find worktrees for this repository that match workspace worktree references
	var repoBranches []string
	for _, worktreeRef := range workspace.Worktrees {
		// Look for worktrees in this repository that match the workspace worktree reference
		for worktreeKey, worktree := range repo.Worktrees {
			if worktree.Branch == worktreeRef {
				repoBranches = append(repoBranches, worktree.Branch)
				c.VerbosePrint("    Found worktree %s in repository %s (key: %s)", worktree.Branch, repoURL, worktreeKey)
				break // Found this worktree reference in this repository, move to next reference
			}
		}
	}

	return repoBranches
}

// buildDeletionConfirmationMessage builds the confirmation message for workspace deletion.
func (c *realCodeManager) buildDeletionConfirmationMessage(
	workspaceName string,
	uniqueWorktrees int,
	repoWorktrees map[string][]string,
) string {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("Are you sure you want to delete workspace '%s'?\n\n", workspaceName))
	message.WriteString("This will delete:\n")

	if uniqueWorktrees > 0 {
		message.WriteString(fmt.Sprintf("- %d worktree(s) across %d repositories:\n", uniqueWorktrees, len(repoWorktrees)))
		for repoURL, branches := range repoWorktrees {
			message.WriteString(fmt.Sprintf("  * %s: %s\n", repoURL, strings.Join(branches, ", ")))
		}
	} else {
		message.WriteString("- No worktrees to delete\n")
	}

	// Add workspace file information
	workspaceFile := c.getWorkspaceFilePath(workspaceName)
	message.WriteString(fmt.Sprintf("- Workspace file: %s\n", workspaceFile))
	message.WriteString("\nType 'yes' to confirm deletion: ")

	return message.String()
}

// deleteWorkspaceWorktrees deletes all worktrees associated with a workspace.
func (c *realCodeManager) deleteWorkspaceWorktrees(
	workspace *status.Workspace,
	worktrees []status.WorktreeInfo,
	force bool,
) error {
	c.VerbosePrint("Deleting %d worktrees for workspace", len(worktrees))

	// Process each worktree directly
	for _, worktree := range worktrees {
		if err := c.deleteSingleWorkspaceWorktree(workspace, worktree, force); err != nil {
			return err
		}
	}

	return nil
}

// deleteSingleWorkspaceWorktree deletes a single worktree from a workspace.
func (c *realCodeManager) deleteSingleWorkspaceWorktree(
	workspace *status.Workspace,
	worktree status.WorktreeInfo,
	force bool,
) error {
	c.VerbosePrint("  Deleting worktree: %s/%s", worktree.Remote, worktree.Branch)

	// Find which repository this worktree belongs to
	repoURL, repoPath, err := c.findWorktreeRepository(workspace, worktree)
	if err != nil {
		return err
	}

	// Get config from ConfigManager
	cfg, err := c.configManager.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Get worktree path and remove from Git
	worktreePath := filepath.Join(cfg.RepositoriesDir, repoURL, worktree.Remote, worktree.Branch)
	if err := c.removeWorktreeFromGit(repoPath, worktreePath, worktree, force); err != nil {
		return err
	}

	// Remove worktree from status
	if err := c.removeWorktreeFromStatus(repoURL, worktree); err != nil {
		return err
	}

	c.VerbosePrint("    ✓ Deleted worktree: %s/%s", worktree.Remote, worktree.Branch)
	return nil
}

// findWorktreeRepository finds the repository that contains a specific worktree.
func (c *realCodeManager) findWorktreeRepository(
	workspace *status.Workspace,
	worktree status.WorktreeInfo,
) (string, string, error) {
	for _, currentRepoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := c.statusManager.GetRepository(currentRepoURL)
		if err != nil {
			c.VerbosePrint("    ⚠ Skipping repository %s: %v", currentRepoURL, err)
			continue
		}

		// Check if this worktree belongs to this repository
		// The worktrees are stored with "remote:branch" as the key
		worktreeKey := fmt.Sprintf("%s:%s", worktree.Remote, worktree.Branch)
		if _, exists := repo.Worktrees[worktreeKey]; exists {
			return currentRepoURL, repo.Path, nil
		}
	}

	c.VerbosePrint("    ⚠ Worktree %s/%s not found in any workspace repository", worktree.Remote, worktree.Branch)
	return "", "", fmt.Errorf("worktree %s/%s not found in any workspace repository", worktree.Remote, worktree.Branch)
}

// removeWorktreeFromGit removes a worktree from Git.
func (c *realCodeManager) removeWorktreeFromGit(
	repoPath, worktreePath string, worktree status.WorktreeInfo, force bool) error {
	// Debug: Print the paths
	c.VerbosePrint("    Repository Path: %s", repoPath)
	c.VerbosePrint("    Worktree Path: %s", worktreePath)

	// Check if worktree path exists
	if exists, err := c.fs.Exists(worktreePath); err == nil {
		c.VerbosePrint("    Worktree path exists: %v", exists)
	} else {
		c.VerbosePrint("    Error checking worktree path existence: %v", err)
	}

	// Remove worktree from Git
	if err := c.git.RemoveWorktree(repoPath, worktreePath, force); err != nil {
		return fmt.Errorf("failed to remove worktree %s/%s: %w", worktree.Remote, worktree.Branch, err)
	}

	return nil
}

// removeWorktreeFromStatus removes a worktree from the status file.
func (c *realCodeManager) removeWorktreeFromStatus(repoURL string, worktree status.WorktreeInfo) error {
	worktreeKey := fmt.Sprintf("%s:%s", worktree.Remote, worktree.Branch)
	c.VerbosePrint("    Removing worktree from status with key: %s", worktreeKey)

	// Check if worktree still exists in status before removing
	repo, err := c.statusManager.GetRepository(repoURL)
	if err == nil {
		c.VerbosePrint("    Repository worktrees before removal: %v", repo.Worktrees)
	}

	if err := c.statusManager.RemoveWorktree(repoURL, worktree.Branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status %s/%s: %w", worktree.Remote, worktree.Branch, err)
	}

	return nil
}

// deleteWorkspaceFiles deletes workspace files.
func (c *realCodeManager) deleteWorkspaceFiles(workspaceName string, worktrees []status.WorktreeInfo) error {
	c.VerbosePrint("Deleting workspace files")

	// Delete main workspace file if it exists
	workspaceFile := c.getWorkspaceFilePath(workspaceName)
	if exists, err := c.fs.Exists(workspaceFile); err == nil && exists {
		if err := c.deleteWorkspaceFile(workspaceFile); err != nil {
			return fmt.Errorf("failed to delete main workspace file: %w", err)
		}
		c.VerbosePrint("    ✓ Deleted main workspace file: %s", workspaceFile)
	} else {
		c.VerbosePrint("    Main workspace file does not exist: %s", workspaceFile)
	}

	// Delete worktree-specific workspace files
	for _, worktree := range worktrees {
		worktreeWorkspaceFile := c.getWorktreeWorkspaceFilePath(workspaceName, worktree.Branch)
		if exists, err := c.fs.Exists(worktreeWorkspaceFile); err == nil && exists {
			if err := c.deleteWorkspaceFile(worktreeWorkspaceFile); err != nil {
				c.VerbosePrint("    ⚠ Could not delete worktree workspace file %s: %v", worktreeWorkspaceFile, err)
				// Continue with other files even if one fails
			} else {
				c.VerbosePrint("    ✓ Deleted worktree workspace file: %s", worktreeWorkspaceFile)
			}
		} else {
			c.VerbosePrint("    Worktree workspace file does not exist: %s", worktreeWorkspaceFile)
		}
	}

	return nil
}

// deleteWorkspaceFile deletes a single workspace file.
func (c *realCodeManager) deleteWorkspaceFile(filePath string) error {
	// Delete the file (existence was already checked by the caller)
	if err := c.fs.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	c.VerbosePrint("    ✓ Deleted: %s", filePath)
	return nil
}

// removeWorktreesFromWorkspaceStatus removes worktree entries from workspace status.
func (c *realCodeManager) removeWorktreesFromWorkspaceStatus(
	workspaceName string, worktrees []status.WorktreeInfo) error {
	c.VerbosePrint("Removing worktree entries from workspace status")

	// Get current workspace
	workspace, err := c.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Create a map of worktree branches to remove for efficient lookup
	worktreesToRemove := make(map[string]bool)
	for _, worktree := range worktrees {
		worktreesToRemove[worktree.Branch] = true
	}

	// Remove worktree entries from workspace
	var remainingWorktrees []string
	for _, worktreeRef := range workspace.Worktrees {
		if !worktreesToRemove[worktreeRef] {
			remainingWorktrees = append(remainingWorktrees, worktreeRef)
		}
	}

	// Update workspace with remaining worktrees
	workspace.Worktrees = remainingWorktrees
	if err := c.statusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	c.VerbosePrint("  ✓ Removed %d worktree entries from workspace status", len(worktrees))
	return nil
}

// removeWorkspaceFromStatus removes the workspace entry from the status file.
func (c *realCodeManager) removeWorkspaceFromStatus(workspaceName string) error {
	c.VerbosePrint("Removing workspace from status file")

	// Use the status manager to remove the workspace
	// Note: This assumes the status manager will have a RemoveWorkspace method
	// For now, we'll use a placeholder that will be implemented in Phase 1
	if err := c.statusManager.RemoveWorkspace(workspaceName); err != nil {
		return fmt.Errorf("failed to remove workspace from status: %w", err)
	}

	c.VerbosePrint("  ✓ Removed workspace from status file")
	return nil
}

// getWorkspaceFilePath returns the path to the main workspace file.
func (c *realCodeManager) getWorkspaceFilePath(workspaceName string) string {
	cfg, err := c.configManager.GetConfigWithFallback()
	if err != nil {
		// Fallback to a default path if config cannot be loaded
		return filepath.Join("~/Code/workspaces", fmt.Sprintf("%s.code-workspace", workspaceName))
	}
	return filepath.Join(cfg.WorkspacesDir, fmt.Sprintf("%s.code-workspace", workspaceName))
}

// getWorktreeWorkspaceFilePath returns the path to a worktree-specific workspace file.
func (c *realCodeManager) getWorktreeWorkspaceFilePath(workspaceName, branchName string) string {
	cfg, err := c.configManager.GetConfigWithFallback()
	if err != nil {
		// Fallback to a default path if config cannot be loaded
		sanitizedBranchForFilename := branch.SanitizeBranchNameForFilename(branchName)
		return filepath.Join("~/Code/workspaces",
			fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename))
	}

	// Sanitize branch name for filename (replace / with -)
	sanitizedBranchForFilename := branch.SanitizeBranchNameForFilename(branchName)
	return filepath.Join(cfg.WorkspacesDir,
		fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename))
}
