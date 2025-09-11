package workspace

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/status"
)

// WorktreeInfo represents comprehensive worktree information for workspace operations.
type WorktreeInfo struct {
	Repository    string               // Repository URL
	Branch        string               // Branch name
	Remote        string               // Remote name
	WorktreePath  string               // Path to worktree directory
	WorkspaceFile string               // Path to worktree-specific workspace file
	Issue         *status.WorktreeInfo // Original worktree info including issue
}

// ListWorktrees lists worktrees for workspace mode.
func (w *realWorkspace) ListWorktrees() ([]status.WorktreeInfo, error) {
	w.logger.Logf("Listing worktrees for workspace mode")

	// Load workspace configuration (only if not already loaded)
	if w.file == "" {
		if err := w.Load(); err != nil {
			return nil, fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// Get workspace path
	workspacePath := w.getWorkspacePath()

	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspacePath)
	if err != nil {
		// If workspace not found, return empty list with no error
		if errors.Is(err, status.ErrWorkspaceNotFound) {
			return []status.WorktreeInfo{}, nil
		}
		return nil, err
	}

	// Get worktrees for each repository in the workspace
	workspaceWorktrees := make([]status.WorktreeInfo, 0)
	seenWorktrees := make(map[string]bool) // Track seen worktrees to avoid duplicates

	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := w.statusManager.GetRepository(repoURL)
		if err != nil {
			continue // Skip if repository not found
		}

		// Get worktrees for this repository
		for _, worktree := range repo.Worktrees {
			// Create a unique key for this worktree to avoid duplicates
			// Include repository URL to distinguish between worktrees from different repositories
			worktreeKey := fmt.Sprintf("%s:%s:%s", repoURL, worktree.Remote, worktree.Branch)
			if !seenWorktrees[worktreeKey] {
				workspaceWorktrees = append(workspaceWorktrees, worktree)
				seenWorktrees[worktreeKey] = true
			}
		}
	}

	return workspaceWorktrees, nil
}

// ListWorkspaceWorktrees lists all worktrees for a specific workspace by name.
// This method provides comprehensive worktree information including file paths.
func (w *realWorkspace) ListWorkspaceWorktrees(workspaceName string) ([]WorktreeInfo, error) {
	w.logger.Logf("Listing worktrees for workspace: %s", workspaceName)

	// Get workspace by name from status
	workspace, err := w.statusManager.GetWorkspaceByName(workspaceName)
	if err != nil {
		if errors.Is(err, status.ErrWorkspaceNotFound) {
			return []WorktreeInfo{}, nil
		}
		return nil, fmt.Errorf("failed to get workspace by name: %w", err)
	}

	// Get workspace path for file operations
	workspacePath := w.getWorkspacePathFromName(workspaceName)
	if workspacePath == "" {
		return nil, fmt.Errorf("failed to determine workspace path for: %s", workspaceName)
	}

	workspaceWorktrees := make([]WorktreeInfo, 0)
	seenWorktrees := make(map[string]bool) // Track seen worktrees to avoid duplicates

	// Process each repository in the workspace
	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := w.statusManager.GetRepository(repoURL)
		if err != nil {
			w.logger.Logf("Skipping repository %s: %v", repoURL, err)
			continue // Skip if repository not found
		}

		// Get worktrees for this repository
		for _, worktree := range repo.Worktrees {
			// Create a unique key for this worktree to avoid duplicates
			worktreeKey := fmt.Sprintf("%s:%s:%s", repoURL, worktree.Remote, worktree.Branch)
			if !seenWorktrees[worktreeKey] {
				// Determine worktree path
				worktreePath := w.getWorktreePath(repo.Path, worktree.Branch)

				// Determine workspace file path
				workspaceFile := w.getWorktreeWorkspaceFilePath(workspacePath, worktree.Branch)

				worktreeInfo := WorktreeInfo{
					Repository:    repoURL,
					Branch:        worktree.Branch,
					Remote:        worktree.Remote,
					WorktreePath:  worktreePath,
					WorkspaceFile: workspaceFile,
					Issue:         &worktree,
				}

				workspaceWorktrees = append(workspaceWorktrees, worktreeInfo)
				seenWorktrees[worktreeKey] = true
			}
		}
	}

	w.logger.Logf("Found %d worktrees for workspace %s", len(workspaceWorktrees), workspaceName)
	return workspaceWorktrees, nil
}

// getWorkspacePathFromName determines the workspace path from the workspace name.
// This is a helper method to support workspace operations by name.
func (w *realWorkspace) getWorkspacePathFromName(workspaceName string) string {
	// Get all workspaces to find the one with matching name
	workspaces, err := w.statusManager.ListWorkspaces()
	if err != nil {
		return ""
	}

	for path := range workspaces {
		// Extract workspace name from path (same logic as status manager)
		workspaceNameFromPath := w.extractWorkspaceNameFromPath(path)
		if workspaceNameFromPath == workspaceName {
			return path
		}
	}

	return ""
}

// getWorktreePath constructs the worktree path based on repository path and branch.
func (w *realWorkspace) getWorktreePath(repoPath, branch string) string {
	// Use the same logic as worktree creation
	return fmt.Sprintf("%s-%s", repoPath, branch)
}

// getWorktreeWorkspaceFilePath constructs the workspace file path for a specific worktree.
func (w *realWorkspace) getWorktreeWorkspaceFilePath(workspacePath, branch string) string {
	// Extract workspace name from path
	workspaceName := w.extractWorkspaceNameFromPath(workspacePath)
	if workspaceName == "" {
		return ""
	}

	// Construct worktree-specific workspace file path
	return fmt.Sprintf("%s-%s.code-workspace", workspaceName, branch)
}

// extractWorkspaceNameFromPath extracts the workspace name from the workspace file path.
// This replicates the logic from the status manager.
func (w *realWorkspace) extractWorkspaceNameFromPath(workspacePath string) string {
	// Get the base name of the file (without extension)
	baseName := filepath.Base(workspacePath)

	// Remove the .code-workspace extension if present
	baseName = strings.TrimSuffix(baseName, ".code-workspace")

	return baseName
}
