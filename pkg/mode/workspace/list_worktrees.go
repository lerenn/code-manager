package workspace

import (
	"errors"
	"fmt"

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
