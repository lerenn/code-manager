package status

import "fmt"

// ListWorkspaces lists all workspaces in the status file.
func (s *realManager) ListWorkspaces() (map[string]Workspace, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	return status.Workspaces, nil
}
