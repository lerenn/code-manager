// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/branch"
	"github.com/lerenn/code-manager/pkg/dependencies"
	"github.com/lerenn/code-manager/pkg/mode/workspace/interfaces"
	"github.com/lerenn/code-manager/pkg/status"
)

// Workspace interface provides workspace management capabilities.
// This interface is now defined in pkg/mode/workspace/interfaces to avoid circular imports.
type Workspace = interfaces.Workspace

// CreateWorktreeOpts contains optional parameters for worktree creation in workspace mode.
type CreateWorktreeOpts = interfaces.CreateWorktreeOpts

// Config represents the configuration of a workspace.
type Config = interfaces.Config

// Folder represents a folder in a workspace.
type Folder = interfaces.Folder

// NewWorkspaceParams contains parameters for creating a new Workspace instance.
type NewWorkspaceParams = interfaces.NewWorkspaceParams

// realWorkspace represents a workspace and provides methods for workspace operations.
type realWorkspace struct {
	deps *dependencies.Dependencies
	file string
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace(params NewWorkspaceParams) Workspace {
	// Cast interface{} to concrete type
	deps := params.Dependencies.(*dependencies.Dependencies)
	if deps == nil {
		deps = dependencies.New()
	}

	return &realWorkspace{
		deps: deps,
		file: params.File,
	}
}

// buildWorkspaceFilePath constructs the workspace file path for a given workspace name and branch.
// This is a shared utility function used by create, delete, and open operations.
// The workspace file is named: {workspaceName}/{sanitizedBranchName}.code-workspace.
func buildWorkspaceFilePath(workspacesDir, workspaceName, branchName string) string {
	// Sanitize branch name for filename (replace / with -)
	sanitizedBranchForFilename := branch.SanitizeBranchNameForFilename(branchName)

	// Create workspace file path
	workspaceFileName := fmt.Sprintf("%s/%s.code-workspace", workspaceName, sanitizedBranchForFilename)
	return filepath.Join(workspacesDir, workspaceFileName)
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
func (w *realWorkspace) getWorkspaceInfo(branchName string) (string, string, error) {
	// Get workspace name for worktree-specific workspace file
	workspaceConfig, err := w.ParseFile(w.file)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse workspace file: %w", err)
	}
	workspaceName := w.GetName(workspaceConfig, w.file)

	// Sanitize branch name for filename (replace slashes with hyphens)
	sanitizedBranchForFilename := branch.SanitizeBranchNameForFilename(branchName)

	cfg, err := w.deps.Config.GetConfigWithFallback()
	if err != nil {
		return "", "", fmt.Errorf("failed to get config: %w", err)
	}
	worktreeWorkspacePath := filepath.Join(
		cfg.WorkspacesDir,
		fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename),
	)

	return workspaceName, worktreeWorkspacePath, nil
}

// getWorkspaceAndWorktrees retrieves workspace and associated worktrees.
func (w *realWorkspace) getWorkspaceAndWorktrees(workspaceName string) (
	worktrees []status.WorktreeInfo, err error) {
	// Check if workspace exists and get it
	workspace, err := w.deps.StatusManager.GetWorkspace(workspaceName)
	if err != nil {
		return nil, fmt.Errorf("workspace '%s' not found in status.yaml: %w", workspaceName, err)
	}

	// List all worktrees associated with the workspace
	worktrees = w.listWorkspaceWorktreesFromWorkspace(workspace)

	return worktrees, nil
}

// listWorkspaceWorktreesFromWorkspace lists worktrees from workspace definition.
func (w *realWorkspace) listWorkspaceWorktreesFromWorkspace(
	workspace *status.Workspace,
) []status.WorktreeInfo {
	var allWorktrees []status.WorktreeInfo

	// Process each repository in the workspace
	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := w.deps.StatusManager.GetRepository(repoURL)
		if err != nil {
			w.deps.Logger.Logf("Warning: failed to get repository %s from status: %v", repoURL, err)
			continue
		}

		// Find worktrees for this repository that match workspace worktree references
		for _, worktreeRef := range workspace.Worktrees {
			// Look for worktrees in this repository that match the workspace worktree reference
			for worktreeKey, worktree := range repo.Worktrees {
				if worktree.Branch == worktreeRef {
					allWorktrees = append(allWorktrees, worktree)
					w.deps.Logger.Logf("Found worktree %s in repository %s (key: %s)", worktree.Branch, repoURL, worktreeKey)
					break // Found this worktree reference in this repository, move to next reference
				}
			}
		}
	}

	return allWorktrees
}

// getWorkspaceWorktrees gets all worktrees for the current workspace.
func (w *realWorkspace) getWorkspaceWorktrees() ([]status.WorktreeInfo, error) {
	// Get workspace name
	workspaceConfig, err := w.ParseFile(w.file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workspace file: %w", err)
	}
	workspaceName := w.GetName(workspaceConfig, w.file)

	// Get workspace and worktrees
	worktrees, err := w.getWorkspaceAndWorktrees(workspaceName)
	if err != nil {
		return nil, err
	}

	return worktrees, nil
}

// deleteWorktreeRepositories deletes worktrees for all repositories.
func (w *realWorkspace) deleteWorktreeRepositories(worktrees []status.WorktreeInfo, force bool) error {
	w.deps.Logger.Logf("Deleting %d worktrees for workspace", len(worktrees))

	// Process each worktree directly
	for _, worktree := range worktrees {
		if err := w.deleteSingleWorkspaceWorktree(worktree, force); err != nil {
			return err
		}
	}

	return nil
}

// deleteSingleWorkspaceWorktree deletes a single worktree from a workspace.
func (w *realWorkspace) deleteSingleWorkspaceWorktree(worktree status.WorktreeInfo, force bool) error {
	w.deps.Logger.Logf("  Deleting worktree: %s/%s", worktree.Remote, worktree.Branch)

	// Get config from ConfigManager
	cfg, err := w.deps.Config.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Get worktree path and remove from Git
	worktreePath := filepath.Join(cfg.RepositoriesDir, worktree.Remote, worktree.Branch)
	if err := w.removeWorktreeFromGit(worktreePath, worktree, force); err != nil {
		return err
	}

	// Remove worktree from status
	if err := w.removeWorktreeFromStatus(worktree.Remote, worktree); err != nil {
		return err
	}

	w.deps.Logger.Logf("    ✓ Deleted worktree: %s/%s", worktree.Remote, worktree.Branch)
	return nil
}

// removeWorktreeFromGit removes a worktree from Git.
func (w *realWorkspace) removeWorktreeFromGit(worktreePath string, worktree status.WorktreeInfo, force bool) error {
	// Debug: Print the paths
	w.deps.Logger.Logf("    Worktree Path: %s", worktreePath)

	// Check if worktree path exists
	if exists, err := w.deps.FS.Exists(worktreePath); err == nil {
		w.deps.Logger.Logf("    Worktree path exists: %v", exists)
	} else {
		w.deps.Logger.Logf("    Error checking worktree path existence: %v", err)
	}

	// Remove worktree from Git
	if err := w.deps.Git.RemoveWorktree(".", worktreePath, force); err != nil {
		return fmt.Errorf("failed to remove worktree %s/%s: %w", worktree.Remote, worktree.Branch, err)
	}

	return nil
}

// removeWorktreeFromStatus removes a worktree from the status file.
func (w *realWorkspace) removeWorktreeFromStatus(repoURL string, worktree status.WorktreeInfo) error {
	worktreeKey := fmt.Sprintf("%s:%s", worktree.Remote, worktree.Branch)
	w.deps.Logger.Logf("    Removing worktree from status with key: %s", worktreeKey)

	if err := w.deps.StatusManager.RemoveWorktree(repoURL, worktree.Branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status %s/%s: %w", worktree.Remote, worktree.Branch, err)
	}

	return nil
}

// ListWorktrees lists all worktrees for the workspace.
func (w *realWorkspace) ListWorktrees() ([]status.WorktreeInfo, error) {
	// Load workspace configuration (only if not already loaded)
	if err := w.ensureWorkspaceLoaded(); err != nil {
		return nil, err
	}

	// Get all worktrees for this workspace
	worktrees, err := w.getWorkspaceWorktrees()
	if err != nil {
		return nil, err
	}

	return worktrees, nil
}
