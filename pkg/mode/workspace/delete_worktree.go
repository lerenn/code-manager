package workspace

import (
	"fmt"
	"path/filepath"
)

// DeleteWorktree deletes worktrees for the workspace with the specified branch.
func (w *realWorkspace) DeleteWorktree(branch string, force bool) error {
	w.deps.Logger.Logf("Deleting worktrees for branch: %s", branch)

	// Load workspace configuration (only if not already loaded)
	if err := w.ensureWorkspaceLoaded(); err != nil {
		return err
	}

	// Get workspace name and worktree workspace path
	workspaceName, worktreeWorkspacePath, err := w.getWorkspaceInfo(branch)
	if err != nil {
		return err
	}

	// Get workspace and worktrees
	worktrees, err := w.getWorkspaceAndWorktrees(workspaceName)
	if err != nil {
		return err
	}

	// Delete worktrees for all repositories
	if err := w.deleteWorktreeRepositories(worktrees, force); err != nil {
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

	w.deps.Logger.Logf("Workspace worktree deletion completed successfully")
	return nil
}

// cleanupEmptyWorkspaceDirectory removes the workspace directory if it's empty.
func (w *realWorkspace) cleanupEmptyWorkspaceDirectory(workspaceDir string) error {
	// Check if directory exists
	exists, err := w.deps.FS.Exists(workspaceDir)
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
		w.deps.Logger.Logf("Removing empty workspace directory: %s", workspaceDir)
		if err := w.deps.FS.RemoveAll(workspaceDir); err != nil {
			return fmt.Errorf("failed to remove empty workspace directory: %w", err)
		}
	}

	return nil
}

// isDirectoryEmpty checks if a directory is empty (contains no files or subdirectories).
func (w *realWorkspace) isDirectoryEmpty(dirPath string) (bool, error) {
	// Use the filesystem interface to read directory contents
	entries, err := w.deps.FS.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	// Directory is empty if it has no entries
	return len(entries) == 0, nil
}

// removeWorktreeFromWorkspaceStatus removes a worktree entry from workspace status.
func (w *realWorkspace) removeWorktreeFromWorkspaceStatus(workspaceName, branch string) error {
	w.deps.Logger.Logf("Removing worktree '%s' from workspace '%s' status", branch, workspaceName)

	// Get current workspace
	workspace, err := w.deps.StatusManager.GetWorkspace(workspaceName)
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
	if err := w.deps.StatusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	w.deps.Logger.Logf("Successfully removed worktree '%s' from workspace '%s' status", branch, workspaceName)
	return nil
}

// cleanupWorkspaceFileAndDirectory removes the workspace file and cleans up the directory if empty.
func (w *realWorkspace) cleanupWorkspaceFileAndDirectory(worktreeWorkspacePath string, force bool) error {
	// Delete worktree-specific workspace file
	if err := w.deleteWorktreeWorkspaceFile(worktreeWorkspacePath, force); err != nil {
		return err
	}

	// Clean up workspace directory if it's empty
	workspaceDir := filepath.Dir(worktreeWorkspacePath)
	if err := w.cleanupEmptyWorkspaceDirectory(workspaceDir); err != nil {
		if !force {
			w.deps.Logger.Logf("Warning: failed to cleanup empty workspace directory: %v", err)
		}
	}

	return nil
}

// deleteWorktreeWorkspaceFile deletes the worktree-specific workspace file.
func (w *realWorkspace) deleteWorktreeWorkspaceFile(worktreeWorkspacePath string, force bool) error {
	if err := w.deps.FS.RemoveAll(worktreeWorkspacePath); err != nil {
		if !force {
			return fmt.Errorf("failed to remove worktree workspace file: %w", err)
		}
		w.deps.Logger.Logf("Warning: failed to remove worktree workspace file: %v", err)
	}
	return nil
}
