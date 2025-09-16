package status

import "fmt"

// RemoveWorktree removes a worktree entry from the status file.
func (s *realManager) RemoveWorktree(repoURL, branch string) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check if repository exists
	repo, exists := status.Repositories[repoURL]
	if !exists {
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, repoURL)
	}

	// Find and remove the worktree entry
	found := false
	for worktreeKey, worktree := range repo.Worktrees {
		if worktree.Branch == branch {
			delete(repo.Worktrees, worktreeKey)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotFound, repoURL, branch)
	}

	// Update repository
	status.Repositories[repoURL] = repo

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}
