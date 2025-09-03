package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/mode"
)

// CreateWorktree creates worktrees for all repositories in the workspace.
func (w *realWorkspace) CreateWorktree(branch string, opts ...mode.CreateWorktreeOpts) (string, error) {
	w.logger.Logf("Creating worktrees for branch: %s", branch)

	// 1. Load and validate workspace configuration (only if not already loaded)
	if w.OriginalFile == "" {
		if err := w.Load(false); err != nil {
			return "", fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// 2. Validate all repositories in workspace
	if err := w.Validate(); err != nil {
		return "", fmt.Errorf("failed to validate workspace: %w", err)
	}

	// 3. Pre-validate worktree creation for all repositories
	if err := w.validateWorkspaceForWorktreeCreation(branch); err != nil {
		return "", fmt.Errorf("failed to validate workspace for worktree creation: %w", err)
	}

	// 4. Create worktrees for all repositories
	var workspaceOpts *mode.CreateWorktreeOpts
	if len(opts) > 0 {
		workspaceOpts = &opts[0]
	}
	if err := w.createWorktreesForWorkspace(branch, workspaceOpts); err != nil {
		return "", fmt.Errorf("failed to create worktrees: %w", err)
	}

	// 5. Calculate and return the worktree path
	worktreePath := filepath.Join(
		w.config.BasePath,
		"workspaces",
		fmt.Sprintf("workspace-%s", branch),
	)

	w.logger.Logf("Workspace worktree creation completed successfully")
	return worktreePath, nil
}
