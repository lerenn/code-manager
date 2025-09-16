package status

import "fmt"

// AddRepository adds a repository entry to the status file.
func (s *realManager) AddRepository(repoURL string, params AddRepositoryParams) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check for duplicate repository
	if _, exists := status.Repositories[repoURL]; exists {
		return fmt.Errorf("%w: %s", ErrRepositoryAlreadyExists, repoURL)
	}

	// Create new repository entry
	repo := Repository{
		Path:      params.Path,
		Remotes:   params.Remotes,
		Worktrees: make(map[string]WorktreeInfo),
	}

	// Add to repositories map
	status.Repositories[repoURL] = repo

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}
