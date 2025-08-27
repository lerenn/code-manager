package cm

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/status"
)

// ListWorktrees lists worktrees for the current project with mode detection.
func (c *realCM) ListWorktrees(force bool) ([]status.WorktreeInfo, ProjectType, error) {
	c.VerbosePrint("Listing worktrees with mode detection")

	// Detect project mode
	projectType, err := c.detectProjectMode()
	if err != nil {
		return nil, ProjectTypeNone, fmt.Errorf("failed to detect project mode: %w", err)
	}

	switch projectType {
	case ProjectTypeSingleRepo:
		worktrees, err := c.repository.ListWorktrees()
		return worktrees, ProjectTypeSingleRepo, c.translateListError(err)
	case ProjectTypeWorkspace:
		worktrees, err := c.workspace.ListWorktrees(force)
		return worktrees, ProjectTypeWorkspace, c.translateListError(err)
	case ProjectTypeNone:
		return nil, ProjectTypeNone, ErrNoGitRepositoryOrWorkspaceFound
	default:
		return nil, ProjectTypeNone, fmt.Errorf("unknown project type")
	}
}

// translateListError translates errors from list operations to CM package errors.
func (c *realCM) translateListError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific status errors and translate them
	if errors.Is(err, status.ErrConfigurationNotInitialized) {
		return ErrNotInitialized
	}

	// Return the original error if no translation is needed
	return err
}
