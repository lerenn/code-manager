package workspace

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// DeleteWorkspace deletes an entire workspace and all its associated resources.
// This includes deleting all worktrees, workspace files, and removing the workspace from status.
func (w *realWorkspace) DeleteWorkspace(workspaceName string, force bool) error {
	w.logger.Logf("Starting workspace deletion for: %s (force: %v)", workspaceName, force)

	// Step 1: Validate workspace exists
	_, err := w.statusManager.GetWorkspaceByName(workspaceName)
	if err != nil {
		if errors.Is(err, status.ErrWorkspaceNotFound) {
			return fmt.Errorf("workspace '%s' not found", workspaceName)
		}
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Step 2: List all worktrees associated with the workspace
	worktrees, err := w.ListWorkspaceWorktrees(workspaceName)
	if err != nil {
		return fmt.Errorf("failed to list workspace worktrees: %w", err)
	}

	// Step 3: Show confirmation prompt (unless force flag is used)
	if !force {
		if err := w.showDeletionConfirmation(workspaceName, worktrees); err != nil {
			return fmt.Errorf("deletion cancelled: %w", err)
		}
	}

	// Step 4: Delete all worktrees associated with the workspace
	w.logger.Logf("Deleting %d worktrees for workspace %s", len(worktrees), workspaceName)
	for _, worktree := range worktrees {
		if err := w.deleteWorktreeFromWorkspace(worktree); err != nil {
			return fmt.Errorf("failed to delete worktree %s/%s: %w", worktree.Repository, worktree.Branch, err)
		}
		w.logger.Logf("✓ Deleted worktree: %s/%s", worktree.Repository, worktree.Branch)
	}

	// Step 5: Delete workspace files
	if err := w.deleteWorkspaceFiles(workspaceName, worktrees); err != nil {
		return fmt.Errorf("failed to delete workspace files: %w", err)
	}

	// Step 6: Remove workspace from status file
	if err := w.statusManager.RemoveWorkspace(workspaceName); err != nil {
		return fmt.Errorf("failed to remove workspace from status: %w", err)
	}

	w.logger.Logf("✓ Workspace '%s' deleted successfully", workspaceName)
	return nil
}

// showDeletionConfirmation shows a detailed confirmation prompt for workspace deletion.
func (w *realWorkspace) showDeletionConfirmation(workspaceName string, worktrees []WorktreeInfo) error {
	// Build confirmation message
	var message strings.Builder
	message.WriteString(fmt.Sprintf("Are you sure you want to delete workspace '%s'?\n\n", workspaceName))

	message.WriteString("This will delete:\n")

	// Group worktrees by repository for better display
	repoWorktrees := make(map[string][]string)
	for _, worktree := range worktrees {
		repoWorktrees[worktree.Repository] = append(repoWorktrees[worktree.Repository], worktree.Branch)
	}

	if len(worktrees) > 0 {
		message.WriteString(fmt.Sprintf("- %d worktrees across %d repositories:\n", len(worktrees), len(repoWorktrees)))
		for repo, branches := range repoWorktrees {
			message.WriteString(fmt.Sprintf("  * %s: %s\n", repo, strings.Join(branches, ", ")))
		}
	} else {
		message.WriteString("- No worktrees to delete\n")
	}

	// Add workspace file information
	workspacePath := w.getWorkspacePathFromName(workspaceName)
	if workspacePath != "" {
		message.WriteString(fmt.Sprintf("- Workspace file: %s\n", workspacePath))
	}

	if len(worktrees) > 0 {
		message.WriteString(fmt.Sprintf("- %d worktree-specific workspace files\n", len(worktrees)))
	}

	message.WriteString("\nType 'yes' to confirm deletion: ")

	// Show confirmation prompt
	confirmed, err := w.prompt.PromptForConfirmation(message.String(), false)
	if err != nil {
		return fmt.Errorf("failed to get user confirmation: %w", err)
	}

	if !confirmed {
		return errors.New("user cancelled deletion")
	}

	return nil
}

// deleteWorktreeFromWorkspace deletes a single worktree from the workspace.
func (w *realWorkspace) deleteWorktreeFromWorkspace(worktreeInfo WorktreeInfo) error {
	// Create a worktree instance to handle the deletion
	worktreeParams := worktree.NewWorktreeParams{
		FS:              w.fs,
		Git:             w.git,
		StatusManager:   w.statusManager,
		Logger:          w.logger,
		Prompt:          w.prompt,
		RepositoriesDir: w.config.RepositoriesDir,
	}

	wt := w.worktreeProvider(worktreeParams)

	// Get repository information for deletion
	repo, err := w.statusManager.GetRepository(worktreeInfo.Repository)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	// Delete the worktree using the existing worktree deletion logic
	deleteParams := worktree.DeleteParams{
		RepoURL:      worktreeInfo.Repository,
		Branch:       worktreeInfo.Branch,
		WorktreePath: worktreeInfo.WorktreePath,
		RepoPath:     repo.Path,
		Force:        true, // Use force=true since we're in workspace deletion mode
	}

	return wt.Delete(deleteParams)
}

// deleteWorkspaceFiles deletes all workspace files associated with the workspace.
func (w *realWorkspace) deleteWorkspaceFiles(workspaceName string, worktrees []WorktreeInfo) error {
	w.logger.Logf("Deleting workspace files for: %s", workspaceName)

	// Get workspace path
	workspacePath := w.getWorkspacePathFromName(workspaceName)
	if workspacePath == "" {
		return fmt.Errorf("failed to determine workspace path for: %s", workspaceName)
	}

	// Delete main workspace file
	exists, err := w.fs.Exists(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to check if workspace file exists: %w", err)
	}
	if exists {
		if err := w.fs.RemoveAll(workspacePath); err != nil {
			return fmt.Errorf("failed to delete main workspace file %s: %w", workspacePath, err)
		}
		w.logger.Logf("✓ Deleted: %s", workspacePath)
	}

	// Delete worktree-specific workspace files
	for _, worktree := range worktrees {
		if worktree.WorkspaceFile != "" {
			exists, err := w.fs.Exists(worktree.WorkspaceFile)
			if err != nil {
				return fmt.Errorf("failed to check if worktree workspace file exists: %w", err)
			}
			if exists {
				if err := w.fs.RemoveAll(worktree.WorkspaceFile); err != nil {
					return fmt.Errorf("failed to delete worktree workspace file %s: %w", worktree.WorkspaceFile, err)
				}
				w.logger.Logf("✓ Deleted: %s", worktree.WorkspaceFile)
			}
		}
	}

	return nil
}
