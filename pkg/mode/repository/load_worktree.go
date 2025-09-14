package repository

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/git"
)

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (r *realRepository) LoadWorktree(remoteSource, branchName string) (string, error) {
	r.logger.Logf("Loading branch: remote=%s, branch=%s", remoteSource, branchName)

	// 1. Validate current directory is a Git repository
	gitExists, err := r.IsGitRepository()
	if err != nil {
		return "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !gitExists {
		return "", ErrGitRepositoryNotFound
	}

	// 2. Parse remote source (default to "origin" if not specified)
	if remoteSource == "" {
		remoteSource = DefaultRemote
	}

	// 3. Validate remote exists and is a valid Git hosting service URL
	if err := r.validateRemote(remoteSource); err != nil {
		return "", err
	}

	// 4. Handle remote management
	if err := r.HandleRemoteManagement(remoteSource); err != nil {
		return "", err
	}

	// 5. Fetch from the remote
	r.logger.Logf("Fetching from remote '%s'", remoteSource)
	if err := r.git.FetchRemote(r.repositoryPath, remoteSource); err != nil {
		return "", fmt.Errorf("%w: %w", git.ErrFetchFailed, err)
	}

	// 6. Validate branch exists on remote
	r.logger.Logf("Checking if branch '%s' exists on remote '%s'", branchName, remoteSource)
	exists, err := r.git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   r.repositoryPath,
		RemoteName: remoteSource,
		Branch:     branchName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf(
			"%w: branch '%s' not found on remote '%s'",
			git.ErrBranchNotFoundOnRemote,
			branchName,
			remoteSource,
		)
	}

	// 7. Create worktree for the branch (using existing worktree creation logic directly)
	r.logger.Logf("Creating worktree for branch '%s'", branchName)
	worktreePath, err := r.CreateWorktree(branchName, CreateWorktreeOpts{Remote: remoteSource})
	return worktreePath, err
}
