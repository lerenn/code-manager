package status

import "fmt"

// AddWorkspace adds a workspace entry to the status file.
func (s *realManager) AddWorkspace(workspacePath string, params AddWorkspaceParams) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check for duplicate workspace
	if _, exists := status.Workspaces[workspacePath]; exists {
		return fmt.Errorf("%w: %s", ErrWorkspaceAlreadyExists, workspacePath)
	}

	// Create new workspace entry
	workspace := Workspace{
		Worktrees:    []string{}, // Empty initially, populated when worktrees are created
		Repositories: params.Repositories,
	}

	// Add to workspaces map
	status.Workspaces[workspacePath] = workspace

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}
