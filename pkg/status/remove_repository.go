package status

import "fmt"

// RemoveRepository removes a repository entry from the status file.
func (s *realManager) RemoveRepository(repoURL string) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check if repository exists
	if _, exists := status.Repositories[repoURL]; !exists {
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, repoURL)
	}

	// Remove repository from status
	delete(status.Repositories, repoURL)

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}
