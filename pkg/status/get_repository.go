package status

import "fmt"

// GetRepository retrieves a repository entry from the status file.
func (s *realManager) GetRepository(repoURL string) (*Repository, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Check if repository exists
	repo, exists := status.Repositories[repoURL]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrRepositoryNotFound, repoURL)
	}

	return &repo, nil
}
