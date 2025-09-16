// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/git"
)

// EnsureBranchExists ensures the specified branch exists, creating it if necessary.
func (w *realWorktree) EnsureBranchExists(repoPath, branch string) error {
	// First check for reference conflicts
	if err := w.git.CheckReferenceConflict(repoPath, branch); err != nil {
		return err
	}

	branchExists, err := w.git.BranchExists(repoPath, branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if !branchExists {
		return w.createBranchFromRemote(repoPath, branch)
	}

	return nil
}

// createBranchFromRemote creates a branch from remote or falls back to local default branch.
func (w *realWorktree) createBranchFromRemote(repoPath, branch string) error {
	w.logger.Logf("Branch %s does not exist locally, checking remote", branch)

	// Fetch from remote to ensure we have the latest changes and remote branch information
	w.logger.Logf("Fetching from origin to ensure repository is up to date")
	if err := w.git.FetchRemote(repoPath, "origin"); err != nil {
		return fmt.Errorf("failed to fetch from origin: %w", err)
	}

	// Check if the branch exists on the remote after fetching
	remoteBranchExists, err := w.git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   repoPath,
		RemoteName: "origin",
		Branch:     branch,
	})
	if err != nil {
		return fmt.Errorf("failed to check if branch exists on remote: %w", err)
	}

	if remoteBranchExists {
		return w.createLocalTrackingBranch(repoPath, branch)
	}

	return w.createBranchFromDefaultBranch(repoPath, branch)
}

// createLocalTrackingBranch creates a local tracking branch from the remote branch.
func (w *realWorktree) createLocalTrackingBranch(repoPath, branch string) error {
	w.logger.Logf("Branch %s exists on remote, creating local tracking branch", branch)
	if err := w.git.CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: "origin/" + branch,
	}); err != nil {
		return fmt.Errorf("failed to create local tracking branch %s from origin/%s: %w", branch, branch, err)
	}
	return nil
}

// createBranchFromDefaultBranch creates a new branch from the origin's default branch.
func (w *realWorktree) createBranchFromDefaultBranch(repoPath, branch string) error {
	w.logger.Logf("Branch %s does not exist on remote, creating from origin's default branch", branch)

	// Get the origin remote URL to determine the actual default branch
	originURL, err := w.git.GetRemoteURL(repoPath, "origin")
	if err != nil {
		w.logger.Logf("Warning: failed to get origin remote URL, falling back to local default branch: %v", err)
		return w.createBranchFromLocalDefaultBranch(repoPath, branch)
	}

	// Get the actual default branch from the remote repository
	remoteDefaultBranch, err := w.git.GetDefaultBranch(originURL)
	if err != nil {
		w.logger.Logf("Warning: failed to get default branch from remote, falling back to local default branch: %v", err)
		return w.createBranchFromLocalDefaultBranch(repoPath, branch)
	}

	w.logger.Logf("Remote default branch is: %s", remoteDefaultBranch)

	// Create the new branch from origin/default_branch (which should be up-to-date after fetch)
	originDefaultBranch := "origin/" + remoteDefaultBranch
	if err := w.git.CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: originDefaultBranch,
	}); err != nil {
		return fmt.Errorf("failed to create branch %s from %s: %w", branch, originDefaultBranch, err)
	}

	return nil
}

// createBranchFromLocalDefaultBranch creates a new branch from the local default branch.
func (w *realWorktree) createBranchFromLocalDefaultBranch(repoPath, branch string) error {
	w.logger.Logf("Creating branch %s from local default branch", branch)

	// Get the current branch (which should be the default branch)
	currentBranch, err := w.git.GetCurrentBranch(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	w.logger.Logf("Local default branch is: %s", currentBranch)

	// Create the new branch from the local default branch
	if err := w.git.CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: currentBranch,
	}); err != nil {
		return fmt.Errorf("failed to create branch %s from local %s: %w", branch, currentBranch, err)
	}

	return nil
}
