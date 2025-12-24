package status

import (
	"fmt"
)

// UpdateWorkspace updates an existing workspace entry in the status file.
func (s *realManager) UpdateWorkspace(workspaceName string, workspace Workspace) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	if s.logger != nil {
		s.logger.Logf("    [UpdateWorkspace] After load: status.Repositories[github.com/octocat/Hello-World].Worktrees = %v",
			status.Repositories["github.com/octocat/Hello-World"].Worktrees)
	}

	// Check if workspace exists
	if _, exists := status.Workspaces[workspaceName]; !exists {
		return fmt.Errorf("%w: %s", ErrWorkspaceNotFound, workspaceName)
	}

	// Update workspace entry
	status.Workspaces[workspaceName] = workspace

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	if s.logger != nil {
		s.logger.Logf("    [UpdateWorkspace] After save: status saved successfully")
	}

	// Update internal workspaces map
	s.computeWorkspacesMap(status.Workspaces)

	return nil
}
