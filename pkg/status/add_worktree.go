package status

import "fmt"

// AddWorktree adds a worktree entry to the status file.
func (s *realManager) AddWorktree(params AddWorktreeParams) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Ensure repository exists
	if _, exists := status.Repositories[params.RepoURL]; !exists {
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, params.RepoURL)
	}

	// Check for duplicate worktree entry
	worktreeKey := fmt.Sprintf("%s:%s", params.Remote, params.Branch)
	if _, exists := status.Repositories[params.RepoURL].Worktrees[worktreeKey]; exists {
		return fmt.Errorf("%w for repository %s worktree %s", ErrWorktreeAlreadyExists, params.RepoURL, worktreeKey)
	}

	// Create new worktree entry
	worktreeInfo := WorktreeInfo{
		Remote: params.Remote,
		Branch: params.Branch,
		Issue:  params.IssueInfo,
	}

	// Add to repository's worktrees
	repo := status.Repositories[params.RepoURL]
	if repo.Worktrees == nil {
		repo.Worktrees = make(map[string]WorktreeInfo)
	}
	repo.Worktrees[worktreeKey] = worktreeInfo
	status.Repositories[params.RepoURL] = repo

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}
