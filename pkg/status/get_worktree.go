package status

import "fmt"

// GetWorktree retrieves the status of a specific worktree.
func (s *realManager) GetWorktree(repoURL, branch string) (*WorktreeInfo, error) {
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

	// Find the worktree entry
	for _, worktree := range repo.Worktrees {
		if worktree.Branch == branch {
			return &worktree, nil
		}
	}

	return nil, fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotFound, repoURL, branch)
}
