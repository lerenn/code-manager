// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/status"
)

// addRepositoryToStatus adds a repository to the status file with its remotes.
func (w *realWorkspace) addRepositoryToStatus(repoURL, repoPath string) error {
	w.verboseLogf("Adding repository %s to status file", repoURL)

	// Get repository remotes
	remotes, err := w.getRepositoryRemotes(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get remotes for repository %s: %w", repoURL, err)
	}

	// Build remotes map with default branches
	remotesMap := make(map[string]status.Remote)
	for remoteName, remoteURL := range remotes {
		// Get default branch for this remote
		defaultBranch, err := w.git.GetDefaultBranch(remoteURL)
		if err != nil {
			w.verboseLogf("Warning: failed to get default branch for remote %s: %v", remoteName, err)
			// Use a fallback default branch - try to detect from local repository
			localDefaultBranch, localErr := w.git.GetCurrentBranch(repoPath)
			if localErr != nil {
				w.verboseLogf("Warning: failed to get local default branch: %v", localErr)
				// Use hardcoded fallback
				defaultBranch = "main"
			} else {
				defaultBranch = localDefaultBranch
			}
		}

		remotesMap[remoteName] = status.Remote{
			DefaultBranch: defaultBranch,
		}
	}

	// Add repository to status
	if err := w.statusManager.AddRepository(repoURL, status.AddRepositoryParams{
		Path:    repoPath,
		Remotes: remotesMap,
	}); err != nil {
		return fmt.Errorf("failed to add repository to status: %w", err)
	}

	return nil
}

// getRepositoryRemotes gets all remotes for a repository.
func (w *realWorkspace) getRepositoryRemotes(repoPath string) (map[string]string, error) {
	w.verboseLogf("Getting remotes for repository: %s", repoPath)

	// For now, we'll use a simplified approach that just gets the origin remote
	// In a real implementation, we would need to add a method to the Git interface
	// to get all remotes, but for now we'll assume origin exists
	remotes := make(map[string]string)

	// Check if origin remote exists
	exists, err := w.git.RemoteExists(repoPath, "origin")
	if err != nil {
		return nil, fmt.Errorf("failed to check if origin remote exists: %w", err)
	}

	if exists {
		// Get origin remote URL
		remoteURL, err := w.git.GetRemoteURL(repoPath, "origin")
		if err != nil {
			return nil, fmt.Errorf("failed to get origin remote URL: %w", err)
		}
		remotes["origin"] = remoteURL
	}

	return remotes, nil
}

// getDefaultRemote gets the default remote for a repository.
func (w *realWorkspace) getDefaultRemote(repoURL string) string {
	defaultRemote := "origin"
	repo, err := w.statusManager.GetRepository(repoURL)
	if err == nil {
		// Check if origin exists, otherwise use the first available remote
		if _, exists := repo.Remotes[defaultRemote]; !exists {
			for remoteName := range repo.Remotes {
				defaultRemote = remoteName
				break
			}
		}
	}
	return defaultRemote
}

// ensureRepositoryHasDefaultBranchWorktree ensures a repository has a worktree for its default branch.
func (w *realWorkspace) ensureRepositoryHasDefaultBranchWorktree(repoURL string, repo *status.Repository) error {
	w.verboseLogf("Ensuring repository %s has default branch worktree", repoURL)

	// Find the default remote (usually "origin")
	defaultRemote := "origin"
	if _, exists := repo.Remotes[defaultRemote]; !exists {
		// If origin doesn't exist, use the first available remote
		for remoteName := range repo.Remotes {
			defaultRemote = remoteName
			break
		}
	}

	if defaultRemote == "" {
		return fmt.Errorf("no remotes found for repository %s", repoURL)
	}

	// Get default branch for the default remote
	defaultBranch := repo.Remotes[defaultRemote].DefaultBranch
	if defaultBranch == "" {
		return fmt.Errorf("no default branch found for remote %s in repository %s", defaultRemote, repoURL)
	}

	// Check if worktree already exists for default branch
	worktreeKey := fmt.Sprintf("%s:%s", defaultRemote, defaultBranch)
	if _, exists := repo.Worktrees[worktreeKey]; exists {
		w.verboseLogf("Worktree %s already exists for repository %s", worktreeKey, repoURL)
		return nil
	}

	// Check if there's already a worktree for this branch in the repository
	// This is important for test environments where worktrees might already exist
	worktreeExists, err := w.worktree.Exists(repo.Path, defaultBranch)
	if err != nil {
		w.verboseLogf("Warning: failed to check if worktree exists for branch %s: %v", defaultBranch, err)
		// Continue with creation attempt
	} else if worktreeExists {
		w.verboseLogf("Worktree already exists for branch %s in repository %s, skipping creation", defaultBranch, repoURL)
		return nil
	}

	// Create worktree for default branch
	w.verboseLogf("Creating worktree %s for repository %s", worktreeKey, repoURL)
	if err := w.createDefaultBranchWorktree(repoURL, defaultRemote, defaultBranch, repo.Path); err != nil {
		return fmt.Errorf("failed to create default branch worktree for repository %s: %w", repoURL, err)
	}

	return nil
}
