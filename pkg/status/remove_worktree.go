package status

import (
	"fmt"
)

// RemoveWorktree removes a worktree entry from the status file.
func (s *realManager) RemoveWorktree(repoURL, branch string) error {
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	repo, err := s.validateRepository(status, repoURL)
	if err != nil {
		return err
	}

	if err := s.deleteWorktreeFromRepo(&repo, branch); err != nil {
		return fmt.Errorf("%w for repository %s branch %s", err, repoURL, branch)
	}

	status.Repositories[repoURL] = repo

	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// validateRepository checks if the repository exists in the status.
func (s *realManager) validateRepository(status *Status, repoURL string) (Repository, error) {
	repo, exists := status.Repositories[repoURL]
	if !exists {
		return Repository{}, fmt.Errorf("%w: %s", ErrRepositoryNotFound, repoURL)
	}
	return repo, nil
}

// deleteWorktreeFromRepo finds and deletes the worktree entry from the repository.
func (s *realManager) deleteWorktreeFromRepo(repo *Repository, branch string) error {
	for worktreeKey, worktree := range repo.Worktrees {
		if worktree.Branch == branch {
			delete(repo.Worktrees, worktreeKey)
			return nil
		}
	}
	return ErrWorktreeNotFound
}
