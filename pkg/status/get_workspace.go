package status

import "fmt"

// GetWorkspace retrieves a workspace entry from the status file.
func (s *realManager) GetWorkspace(workspacePath string) (*Workspace, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Check if workspace exists
	workspace, exists := status.Workspaces[workspacePath]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, workspacePath)
	}

	return &workspace, nil
}
