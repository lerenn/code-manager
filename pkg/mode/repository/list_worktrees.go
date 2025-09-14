package repository

import (
	"errors"
	"fmt"

	"github.com/lerenn/code-manager/pkg/status"
)

// ListWorktrees lists all worktrees for the current repository.
func (r *realRepository) ListWorktrees() ([]status.WorktreeInfo, error) {
	r.logger.Logf("Listing worktrees for single repository mode")

	// Note: Repository validation is already done in mode detection, so we skip it here
	// to avoid duplicate validation calls

	// 1. Extract repository name from remote origin URL (fallback to local path if no remote)
	repoName, err := r.git.GetRepositoryName(r.repositoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository name: %w", err)
	}

	r.logger.Logf("Repository name: %s", repoName)

	// 2. Get repository from status file
	repo, err := r.statusManager.GetRepository(repoName)
	if err != nil {
		// If repository not found, return empty list
		// But propagate other errors (like status file corruption)
		if errors.Is(err, status.ErrRepositoryNotFound) {
			return []status.WorktreeInfo{}, nil
		}
		return nil, err
	}

	// 3. Convert repository worktrees to WorktreeInfo slice
	var worktrees []status.WorktreeInfo
	for _, worktree := range repo.Worktrees {
		worktrees = append(worktrees, worktree)
	}

	r.logger.Logf("Found %d worktrees for current repository", len(worktrees))

	return worktrees, nil
}
