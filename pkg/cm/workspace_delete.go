package cm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/status"
)

// DeleteWorkspace deletes a workspace and all associated resources.
func (c *realCM) DeleteWorkspace(params DeleteWorkspaceParams) error {
	return c.executeWithHooks("delete_workspace", map[string]interface{}{
		"workspace_name": params.WorkspaceName,
		"force":          params.Force,
	}, func() error {
		return c.deleteWorkspace(params)
	})
}

// deleteWorkspace implements the workspace deletion business logic.
func (c *realCM) deleteWorkspace(params DeleteWorkspaceParams) error {
	c.VerbosePrint("Deleting workspace: %s", params.WorkspaceName)

	// Validate workspace name
	if err := c.validateWorkspaceName(params.WorkspaceName); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidWorkspaceName, err)
	}

	// Check if workspace exists and get it
	workspace, err := c.statusManager.GetWorkspace(params.WorkspaceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("%w: workspace '%s' not found", ErrWorkspaceNotFound, params.WorkspaceName)
		}
		return fmt.Errorf("failed to check workspace existence: %w", err)
	}

	// List all worktrees associated with the workspace
	worktrees, err := c.listWorkspaceWorktreesFromWorkspace(workspace)
	if err != nil {
		return fmt.Errorf("failed to list workspace worktrees: %w", err)
	}

	// Show confirmation prompt with deletion summary (unless --force)
	if !params.Force {
		if err := c.showDeletionConfirmation(params.WorkspaceName, workspace, worktrees); err != nil {
			return fmt.Errorf("deletion cancelled: %w", err)
		}
	}

	// Delete all worktrees associated with the workspace
	if err := c.deleteWorkspaceWorktrees(workspace, worktrees, params.Force); err != nil {
		return fmt.Errorf("failed to delete workspace worktrees: %w", err)
	}

	// Delete workspace files
	if err := c.deleteWorkspaceFiles(params.WorkspaceName, worktrees); err != nil {
		return fmt.Errorf("failed to delete workspace files: %w", err)
	}

	// Remove workspace entry from status file
	if err := c.removeWorkspaceFromStatus(params.WorkspaceName); err != nil {
		return fmt.Errorf("failed to remove workspace from status: %w", err)
	}

	c.VerbosePrint("Workspace '%s' deleted successfully", params.WorkspaceName)
	return nil
}

// showDeletionConfirmation shows a confirmation prompt with detailed deletion summary.
func (c *realCM) showDeletionConfirmation(workspaceName string, workspace *status.Workspace, worktrees []status.WorktreeInfo) error {
	// Group worktrees by repository for better display
	repoWorktrees := make(map[string][]string)

	// Process each repository in the workspace
	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := c.statusManager.GetRepository(repoURL)
		if err != nil {
			continue
		}

		// Find worktrees for this repository
		for _, worktree := range worktrees {
			// Check if this worktree belongs to this repository
			if _, exists := repo.Worktrees[worktree.Branch]; exists {
				repoWorktrees[repoURL] = append(repoWorktrees[repoURL], worktree.Branch)
			}
		}
	}

	// Build confirmation message
	var message strings.Builder
	message.WriteString(fmt.Sprintf("Are you sure you want to delete workspace '%s'?\n\n", workspaceName))
	message.WriteString("This will delete:\n")

	if len(worktrees) > 0 {
		message.WriteString(fmt.Sprintf("- %d worktrees across %d repositories:\n", len(worktrees), len(repoWorktrees)))
		for repoURL, branches := range repoWorktrees {
			message.WriteString(fmt.Sprintf("  * %s: %s\n", repoURL, strings.Join(branches, ", ")))
		}
	} else {
		message.WriteString("- No worktrees to delete\n")
	}

	// Add workspace file information
	workspaceFile := c.getWorkspaceFilePath(workspaceName)
	message.WriteString(fmt.Sprintf("- Workspace file: %s\n", workspaceFile))

	if len(worktrees) > 0 {
		message.WriteString(fmt.Sprintf("- %d worktree-specific workspace files\n", len(worktrees)))
	}

	message.WriteString("\nType 'yes' to confirm deletion: ")

	// Get user confirmation
	confirmed, err := c.prompt.PromptForConfirmation(message.String(), false)
	if err != nil {
		return fmt.Errorf("failed to get user confirmation: %w", err)
	}

	if !confirmed {
		return fmt.Errorf("deletion cancelled by user")
	}

	return nil
}

// deleteWorkspaceWorktrees deletes all worktrees associated with a workspace.
func (c *realCM) deleteWorkspaceWorktrees(workspace *status.Workspace, worktrees []status.WorktreeInfo, force bool) error {
	c.VerbosePrint("Deleting %d worktrees for workspace", len(worktrees))

	// Process each repository in the workspace
	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := c.statusManager.GetRepository(repoURL)
		if err != nil {
			c.VerbosePrint("  ⚠ Skipping repository %s: %v", repoURL, err)
			continue
		}

		// Find worktrees for this repository
		for _, worktree := range worktrees {
			// Check if this worktree belongs to this repository
			if repoWorktree, exists := repo.Worktrees[worktree.Branch]; exists && repoWorktree.Remote == worktree.Remote {
				c.VerbosePrint("  Deleting worktree: %s/%s", worktree.Remote, worktree.Branch)

				// Get worktree path
				worktreePath := c.BuildWorktreePath(repoURL, worktree.Remote, worktree.Branch)

				// Remove worktree from Git
				if err := c.git.RemoveWorktree(repo.Path, worktreePath, force); err != nil {
					return fmt.Errorf("failed to remove worktree %s/%s: %w", worktree.Remote, worktree.Branch, err)
				}

				// Remove worktree from status
				if err := c.statusManager.RemoveWorktree(repoURL, worktree.Branch); err != nil {
					return fmt.Errorf("failed to remove worktree from status %s/%s: %w", worktree.Remote, worktree.Branch, err)
				}

				c.VerbosePrint("    ✓ Deleted worktree: %s/%s", worktree.Remote, worktree.Branch)
			}
		}
	}

	return nil
}

// deleteWorkspaceFiles deletes workspace files.
func (c *realCM) deleteWorkspaceFiles(workspaceName string, worktrees []status.WorktreeInfo) error {
	c.VerbosePrint("Deleting workspace files")

	// Delete main workspace file
	workspaceFile := c.getWorkspaceFilePath(workspaceName)
	if err := c.deleteWorkspaceFile(workspaceFile); err != nil {
		return fmt.Errorf("failed to delete main workspace file: %w", err)
	}

	// Delete worktree-specific workspace files
	for _, worktree := range worktrees {
		worktreeWorkspaceFile := c.getWorktreeWorkspaceFilePath(workspaceName, worktree.Branch)
		if err := c.deleteWorkspaceFile(worktreeWorkspaceFile); err != nil {
			c.VerbosePrint("    ⚠ Could not delete worktree workspace file %s: %v", worktreeWorkspaceFile, err)
			// Continue with other files even if one fails
		}
	}

	return nil
}

// deleteWorkspaceFile deletes a single workspace file.
func (c *realCM) deleteWorkspaceFile(filePath string) error {
	// Check if file exists
	exists, err := c.fs.Exists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if !exists {
		c.VerbosePrint("    File does not exist, skipping: %s", filePath)
		return nil
	}

	// Delete the file
	if err := c.fs.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	c.VerbosePrint("    ✓ Deleted: %s", filePath)
	return nil
}

// removeWorkspaceFromStatus removes the workspace entry from the status file.
func (c *realCM) removeWorkspaceFromStatus(workspaceName string) error {
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
func (c *realCM) getWorkspaceFilePath(workspaceName string) string {
	return filepath.Join(c.config.WorkspacesDir, fmt.Sprintf("%s.code-workspace", workspaceName))
}

// getWorktreeWorkspaceFilePath returns the path to a worktree-specific workspace file.
func (c *realCM) getWorktreeWorkspaceFilePath(workspaceName, branch string) string {
	return filepath.Join(c.config.WorkspacesDir, fmt.Sprintf("%s-%s.code-workspace", workspaceName, branch))
}
