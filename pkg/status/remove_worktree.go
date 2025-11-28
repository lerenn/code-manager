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

	s.logRemoveWorktreeBefore(repo.Worktrees)

	if err := s.deleteWorktreeFromRepo(&repo, branch); err != nil {
		return fmt.Errorf("%w for repository %s branch %s", err, repoURL, branch)
	}

	s.logRemoveWorktreeAfter(repo.Worktrees, repoURL, status)

	status.Repositories[repoURL] = repo

	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	s.logRemoveWorktreeSave()

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
			s.logRemoveWorktreeDelete(worktreeKey, branch)
			delete(repo.Worktrees, worktreeKey)
			return nil
		}
	}
	return ErrWorktreeNotFound
}

// logRemoveWorktreeBefore logs the worktrees before deletion.
func (s *realManager) logRemoveWorktreeBefore(worktrees map[string]WorktreeInfo) {
	if s.logger != nil {
		s.logger.Logf("    [RemoveWorktree] Before deletion: repo.Worktrees = %v", worktrees)
	}
}

// logRemoveWorktreeDelete logs the deletion of a worktree.
func (s *realManager) logRemoveWorktreeDelete(worktreeKey, branch string) {
	if s.logger != nil {
		s.logger.Logf("    [RemoveWorktree] Deleting worktree with key: %s, branch: %s", worktreeKey, branch)
	}
}

// logRemoveWorktreeAfter logs the worktrees after deletion.
func (s *realManager) logRemoveWorktreeAfter(worktrees map[string]WorktreeInfo, repoURL string, status *Status) {
	if s.logger != nil {
		s.logger.Logf("    [RemoveWorktree] After deletion: repo.Worktrees = %v", worktrees)
		s.logger.Logf("    [RemoveWorktree] After update: status.Repositories[%s].Worktrees = %v",
			repoURL, status.Repositories[repoURL].Worktrees)
	}
}

// logRemoveWorktreeSave logs after saving the status.
func (s *realManager) logRemoveWorktreeSave() {
	if s.logger != nil {
		s.logger.Logf("    [RemoveWorktree] After save: status saved successfully")
	}
}
