package repository

import (
	"errors"
	"fmt"
	"sort"

	"github.com/lerenn/code-manager/pkg/status"
)

// ListWorktrees lists all worktrees for the current repository.
func (r *realRepository) ListWorktrees() ([]status.WorktreeInfo, error) {
	r.deps.Logger.Logf("Listing worktrees for single repository mode")

	// Note: Repository validation is already done in mode detection, so we skip it here
	// to avoid duplicate validation calls

	// 1. Extract repository name from remote origin URL (fallback to local path if no remote)
	repoName, err := r.deps.Git.GetRepositoryName(r.repositoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository name: %w", err)
	}

	r.deps.Logger.Logf("Repository name: %s", repoName)

	// 2. Get repository from status file
	repo, err := r.deps.StatusManager.GetRepository(repoName)
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

	// 4. Sort worktrees by branch name for consistent ordering
	sort.Slice(worktrees, func(i, j int) bool {
		return worktrees[i].Branch < worktrees[j].Branch
	})

	r.deps.Logger.Logf("Found %d worktrees for current repository", len(worktrees))

	return worktrees, nil
}
