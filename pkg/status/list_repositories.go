package status

import "fmt"

// ListRepositories lists all repositories in the status file.
func (s *realManager) ListRepositories() (map[string]Repository, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	return status.Repositories, nil
}
