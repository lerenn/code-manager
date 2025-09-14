// Package repository provides Git repository management functionality for CM.
package repository

import (
	"fmt"
	"path/filepath"
)

// ValidationParams contains parameters for repository validation.
type ValidationParams struct {
	CurrentDir string
	Branch     string
}

// ValidationResult contains the result of repository validation.
type ValidationResult struct {
	RepoURL  string
	RepoPath string
}

// ValidateRepository validates the repository and returns repository information.
func (r *realRepository) ValidateRepository(params ValidationParams) (*ValidationResult, error) {
	// Get current working directory if not provided
	if params.CurrentDir == "" {
		currentDir, err := filepath.Abs(r.repositoryPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		params.CurrentDir = currentDir
	}

	// Validate Git repository
	if err := r.validateGitRepository(); err != nil {
		return nil, err
	}

	// Get repository URL
	repoURL, err := r.getRepositoryURL(params.CurrentDir)
	if err != nil {
		return nil, err
	}

	// Only perform additional validation if a branch is provided
	if params.Branch != "" {
		// Check if worktree already exists
		if err := r.validateWorktreeNotExists(repoURL, params.Branch); err != nil {
			return nil, err
		}

		// Validate repository state (only if worktree doesn't exist)
		if err := r.validateRepositoryState(params.CurrentDir); err != nil {
			return nil, err
		}
	}

	return &ValidationResult{
		RepoURL:  repoURL,
		RepoPath: params.CurrentDir,
	}, nil
}

// validateGitRepository validates that we're in a Git repository.
func (r *realRepository) validateGitRepository() error {
	isSingleRepo, err := r.IsGitRepository()
	if err != nil {
		return fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return fmt.Errorf("current directory is not a Git repository")
	}
	return nil
}

// getRepositoryURL gets the repository URL from remote origin URL with fallback to local path.
func (r *realRepository) getRepositoryURL(currentDir string) (string, error) {
	repoURL, err := r.git.GetRepositoryName(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}
	r.logger.Logf("Repository URL: %s", repoURL)
	return repoURL, nil
}

// validateWorktreeNotExists checks if worktree already exists in status file.
func (r *realRepository) validateWorktreeNotExists(repoURL, branch string) error {
	existingWorktree, err := r.statusManager.GetWorktree(repoURL, branch)
	if err == nil && existingWorktree != nil {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, repoURL, branch)
	}
	return nil
}

// validateRepositoryState validates that the repository is in a clean state.
func (r *realRepository) validateRepositoryState(currentDir string) error {
	isClean, err := r.git.IsClean(currentDir)
	if err != nil {
		return fmt.Errorf("failed to check repository state: %w", err)
	}
	if !isClean {
		return fmt.Errorf("%w: repository is not in a clean state", ErrRepositoryNotClean)
	}
	return nil
}

// ValidateWorktreeExists validates that a worktree exists in the status file.
func (r *realRepository) ValidateWorktreeExists(repoURL, branch string) error {
	existingWorktree, err := r.statusManager.GetWorktree(repoURL, branch)
	if err != nil || existingWorktree == nil {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotInStatus, repoURL, branch)
	}
	return nil
}

// ValidateOriginRemote validates that the origin remote exists and is a valid Git hosting service URL.
func (r *realRepository) ValidateOriginRemote() error {
	r.logger.Logf("Validating origin remote")

	// Check if origin remote exists
	exists, err := r.git.RemoteExists(r.repositoryPath, "origin")
	if err != nil {
		return fmt.Errorf("failed to check origin remote: %w", err)
	}
	if !exists {
		return ErrOriginRemoteNotFound
	}

	// Get origin remote URL
	originURL, err := r.git.GetRemoteURL(r.repositoryPath, "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	// Validate that it's a valid Git hosting service URL
	if r.ExtractHostFromURL(originURL) == "" {
		return ErrOriginRemoteInvalidURL
	}

	return nil
}
