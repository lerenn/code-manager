package status

import (
	"fmt"
	"log"
)

// RemoveWorkspace removes a workspace entry from the status file.
func (s *realManager) RemoveWorkspace(workspaceName string) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	log.Printf("    [RemoveWorkspace] After load: status.Repositories[github.com/octocat/Hello-World].Worktrees = %v",
		status.Repositories["github.com/octocat/Hello-World"].Worktrees)

	// Check if workspace exists
	if _, exists := status.Workspaces[workspaceName]; !exists {
		return fmt.Errorf("%w: %s", ErrWorkspaceNotFound, workspaceName)
	}

	// Remove workspace from status
	delete(status.Workspaces, workspaceName)

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	log.Printf("    [RemoveWorkspace] After save: status saved successfully")

	// Update internal workspaces map
	s.computeWorkspacesMap(status.Workspaces)

	return nil
}
