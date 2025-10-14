package status

import (
	"fmt"
	"log"
)

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

	log.Printf("    [RemoveWorktree] Before deletion: repo.Worktrees = %v", repo.Worktrees)

	// Find and remove the worktree entry
	found := false
	for worktreeKey, worktree := range repo.Worktrees {
		if worktree.Branch == branch {
			log.Printf("    [RemoveWorktree] Deleting worktree with key: %s, branch: %s", worktreeKey, branch)
			delete(repo.Worktrees, worktreeKey)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotFound, repoURL, branch)
	}

	log.Printf("    [RemoveWorktree] After deletion: repo.Worktrees = %v", repo.Worktrees)

	// Update repository
	status.Repositories[repoURL] = repo

	log.Printf("    [RemoveWorktree] After update: status.Repositories[%s].Worktrees = %v",
		repoURL, status.Repositories[repoURL].Worktrees)

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	log.Printf("    [RemoveWorktree] After save: status saved successfully")

	return nil
}
