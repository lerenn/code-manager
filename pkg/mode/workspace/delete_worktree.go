package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DeleteWorktree deletes worktrees for the workspace with the specified branch.
func (w *realWorkspace) DeleteWorktree(branch string, force bool) error {
	w.logger.Logf("Deleting worktrees for branch: %s", branch)

	// Load workspace configuration (only if not already loaded)
	if err := w.ensureWorkspaceLoaded(); err != nil {
		return err
	}

	// Get worktrees for this workspace and branch
	workspaceWorktrees, err := w.getWorkspaceWorktrees(branch)
	if err != nil {
		return err
	}

	if len(workspaceWorktrees) == 0 {
		return fmt.Errorf("no worktrees found for workspace branch %s", branch)
	}

	// Get workspace name and worktree workspace path
	_, worktreeWorkspacePath, err := w.getWorkspaceInfo(branch)
	if err != nil {
		return err
	}

	// Delete worktrees for all repositories
	if err := w.deleteWorktreeRepositories(workspaceWorktrees, force); err != nil {
		return err
	}

	// Delete worktree-specific workspace file
	if err := w.deleteWorktreeWorkspaceFile(worktreeWorkspacePath, force); err != nil {
		return err
	}

	// Remove worktree entries from status file
	if err := w.removeWorktreeStatusEntries(workspaceWorktrees, force); err != nil {
		return err
	}

	w.logger.Logf("Workspace worktree deletion completed successfully")
	return nil
}

// ensureWorkspaceLoaded ensures the workspace is loaded.
func (w *realWorkspace) ensureWorkspaceLoaded() error {
	if w.file == "" {
		if err := w.Load(); err != nil {
			return fmt.Errorf("failed to load workspace: %w", err)
		}
	}
	return nil
}

// getWorkspaceInfo gets workspace name and worktree workspace path.
func (w *realWorkspace) getWorkspaceInfo(branch string) (string, string, error) {
	// Get workspace name for worktree-specific workspace file
	workspaceConfig, err := w.ParseFile(w.file)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse workspace file: %w", err)
	}
	workspaceName := w.GetName(workspaceConfig, w.file)

	// Sanitize branch name for filename (replace slashes with hyphens)
	sanitizedBranchForFilename := strings.ReplaceAll(branch, "/", "-")

	cfg, err := w.configManager.GetConfigWithFallback()
	if err != nil {
		return "", "", fmt.Errorf("failed to get config: %w", err)
	}
	worktreeWorkspacePath := filepath.Join(
		cfg.WorkspacesDir,
		fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename),
	)

	return workspaceName, worktreeWorkspacePath, nil
}

// deleteWorktreeWorkspaceFile deletes the worktree-specific workspace file.
func (w *realWorkspace) deleteWorktreeWorkspaceFile(worktreeWorkspacePath string, force bool) error {
	if err := w.fs.RemoveAll(worktreeWorkspacePath); err != nil {
		if !force {
			return fmt.Errorf("failed to remove worktree workspace file: %w", err)
		}
		w.logger.Logf("Warning: failed to remove worktree workspace file: %v", err)
	}
	return nil
}
