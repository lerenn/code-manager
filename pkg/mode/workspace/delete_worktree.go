package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/status"
)

// DeleteWorktree deletes worktrees for the workspace with the specified branch.
// This method works with workspace names from the status file, not workspace files.
func (w *realWorkspace) DeleteWorktree(workspaceName, branch string, force bool) error {
	w.logger.Logf("Deleting worktrees for workspace: %s, branch: %s", workspaceName, branch)

	// Validate workspace and worktree exist
	workspace, err := w.validateWorkspaceAndWorktree(workspaceName, branch)
	if err != nil {
		return err
	}

	// Use shared utility to build workspace file path
	worktreeWorkspacePath := buildWorkspaceFilePath(w.config.WorkspacesDir, workspaceName, branch)

	// Delete worktrees for all repositories in the workspace
	if err := w.deleteWorktreeRepositoriesFromWorkspace(workspace, branch, force); err != nil {
		return err
	}

	// Clean up workspace file and directory
	if err := w.cleanupWorkspaceFileAndDirectory(worktreeWorkspacePath, force); err != nil {
		return err
	}

	// Remove worktree entry from workspace status
	if err := w.removeWorktreeFromWorkspaceStatus(workspaceName, branch); err != nil {
		return err
	}

	w.logger.Logf("Workspace worktree deletion completed successfully")
	return nil
}

// cleanupEmptyWorkspaceDirectory removes the workspace directory if it's empty.
func (w *realWorkspace) cleanupEmptyWorkspaceDirectory(workspaceDir string) error {
	// Check if directory exists
	exists, err := w.fs.Exists(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to check if workspace directory exists: %w", err)
	}
	if !exists {
		return nil // Directory doesn't exist, nothing to clean up
	}

	// Check if directory is empty
	isEmpty, err := w.isDirectoryEmpty(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to check if workspace directory is empty: %w", err)
	}

	if isEmpty {
		w.logger.Logf("Removing empty workspace directory: %s", workspaceDir)
		if err := w.fs.RemoveAll(workspaceDir); err != nil {
			return fmt.Errorf("failed to remove empty workspace directory: %w", err)
		}
	}

	return nil
}

// isDirectoryEmpty checks if a directory is empty (contains no files or subdirectories).
func (w *realWorkspace) isDirectoryEmpty(dirPath string) (bool, error) {
	// Use the filesystem interface to read directory contents
	entries, err := w.fs.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	// Directory is empty if it has no entries
	return len(entries) == 0, nil
}

// deleteWorktreeRepositoriesFromWorkspace deletes worktrees for all repositories in the workspace.
func (w *realWorkspace) deleteWorktreeRepositoriesFromWorkspace(
	workspace *status.Workspace, branch string, force bool,
) error {
	w.logger.Logf("Deleting worktrees for workspace repositories: %d repositories, branch: %s, force: %v",
		len(workspace.Repositories), branch, force)

	var errors []error
	for _, repoURL := range workspace.Repositories {
		// Get repository from status
		repo, err := w.statusManager.GetRepository(repoURL)
		if err != nil {
			w.logger.Logf("Warning: failed to get repository %s from status: %v", repoURL, err)
			continue
		}

		// Check if this repository has the worktree
		worktreeKey := fmt.Sprintf("origin:%s", branch) // Assuming origin remote
		if worktreeInfo, exists := repo.Worktrees[worktreeKey]; exists {
			// Delete the worktree using repository package
			if err := w.deleteSingleRepositoryWorktree(repoURL, worktreeInfo, force); err != nil {
				errors = append(errors, fmt.Errorf("failed to delete worktree in repository %s: %w", repoURL, err))
			}
		}
	}

	if len(errors) > 0 {
		if len(errors) == len(workspace.Repositories) {
			// All deletions failed
			return fmt.Errorf("failed to delete worktrees in all repositories: %v", errors)
		}
		// Some deletions failed
		w.logger.Logf("Some worktrees failed to delete: %v", errors)
		return fmt.Errorf("some worktrees failed to delete: %v", errors)
	}

	w.logger.Logf("Successfully deleted worktrees in all %d repositories", len(workspace.Repositories))
	return nil
}

// deleteSingleRepositoryWorktree deletes a worktree in a specific repository.
func (w *realWorkspace) deleteSingleRepositoryWorktree(
	repoURL string, worktreeInfo status.WorktreeInfo, force bool,
) error {
	w.logger.Logf("Deleting worktree for repository: %s, branch: %s", repoURL, worktreeInfo.Branch)

	// Use repository URL directly as it's already the full path
	repoPath := repoURL

	// Create repository instance using repositoryProvider
	repoInstance := w.repositoryProvider(repository.NewRepositoryParams{
		FS:               w.fs,
		Git:              w.git,
		Config:           w.config,
		StatusManager:    w.statusManager,
		Logger:           w.logger,
		Prompt:           w.prompt,
		WorktreeProvider: w.safeWorktreeProvider(),
		HookManager:      w.hookManager,
		RepositoryName:   repoPath,
	})

	// Use repository's DeleteWorktree method
	if err := repoInstance.DeleteWorktree(worktreeInfo.Branch, force); err != nil {
		return fmt.Errorf("failed to delete worktree for branch %s: %w", worktreeInfo.Branch, err)
	}

	w.logger.Logf("Successfully deleted worktree for branch %s in repository %s", worktreeInfo.Branch, repoURL)
	return nil
}

// removeWorktreeFromWorkspaceStatus removes a worktree entry from workspace status.
func (w *realWorkspace) removeWorktreeFromWorkspaceStatus(workspaceName, branch string) error {
	w.logger.Logf("Removing worktree '%s' from workspace '%s' status", branch, workspaceName)

	// Get current workspace
	workspace, err := w.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Remove worktree from the list
	var updatedWorktrees []string
	for _, worktree := range workspace.Worktrees {
		if worktree != branch {
			updatedWorktrees = append(updatedWorktrees, worktree)
		}
	}
	workspace.Worktrees = updatedWorktrees

	// Update workspace in status file
	if err := w.statusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	w.logger.Logf("Successfully removed worktree '%s' from workspace '%s' status", branch, workspaceName)
	return nil
}

// validateWorkspaceAndWorktree validates that the workspace exists and contains the specified worktree.
func (w *realWorkspace) validateWorkspaceAndWorktree(workspaceName, branch string) (*status.Workspace, error) {
	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return nil, fmt.Errorf("workspace '%s' not found in status.yaml: %w", workspaceName, err)
	}

	// Check if the worktree exists in the workspace worktrees list
	for _, worktree := range workspace.Worktrees {
		if worktree == branch {
			return workspace, nil
		}
	}

	return nil, fmt.Errorf("worktree '%s' not found in workspace '%s'", branch, workspaceName)
}

// cleanupWorkspaceFileAndDirectory removes the workspace file and cleans up the directory if empty.
func (w *realWorkspace) cleanupWorkspaceFileAndDirectory(worktreeWorkspacePath string, force bool) error {
	// Delete worktree-specific workspace file
	if err := w.fs.RemoveAll(worktreeWorkspacePath); err != nil {
		if !force {
			return fmt.Errorf("failed to remove worktree workspace file: %w", err)
		}
		w.logger.Logf("Warning: failed to remove worktree workspace file: %v", err)
	}

	// Clean up workspace directory if it's empty
	workspaceDir := filepath.Dir(worktreeWorkspacePath)
	if err := w.cleanupEmptyWorkspaceDirectory(workspaceDir); err != nil {
		if !force {
			w.logger.Logf("Warning: failed to cleanup empty workspace directory: %v", err)
		}
	}

	return nil
}
