package repository

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// StatusParams contains parameters for status operations.
type StatusParams struct {
	RepoURL       string
	Branch        string
	WorktreePath  string
	WorkspacePath string
	Remote        string
	IssueInfo     *issue.Info
}

// AddWorktreeToStatus adds the worktree to the status file with proper error handling.
func (r *realRepository) AddWorktreeToStatus(params StatusParams) error {
	// Create worktree instance using provider
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: r.config.RepositoriesDir,
	})

	if err := worktreeInstance.AddToStatus(worktree.AddToStatusParams{
		RepoURL:       params.RepoURL,
		Branch:        params.Branch,
		WorktreePath:  params.WorktreePath,
		WorkspacePath: params.WorkspacePath,
		Remote:        params.Remote,
		IssueInfo:     params.IssueInfo,
	}); err != nil {
		return r.handleStatusAddError(err, params)
	}
	return nil
}

// handleStatusAddError handles errors when adding worktree to status.
func (r *realRepository) handleStatusAddError(err error, params StatusParams) error {
	// Check if the error is due to repository not found, and auto-add it
	if errors.Is(err, status.ErrRepositoryNotFound) {
		return r.handleRepositoryNotFoundError(params)
	}

	// Clean up created directory on status update failure
	r.cleanupWorktreeDirectory(params.WorktreePath)
	return fmt.Errorf("failed to add worktree to status: %w", err)
}

// handleRepositoryNotFoundError handles the case when repository is not found in status.
func (r *realRepository) handleRepositoryNotFoundError(params StatusParams) error {
	currentDir, err := filepath.Abs(r.repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if addErr := r.AutoAddRepositoryToStatus(params.RepoURL, currentDir); addErr != nil {
		// Clean up created directory on status update failure
		r.cleanupWorktreeDirectory(params.WorktreePath)
		return fmt.Errorf("failed to auto-add repository to status: %w", addErr)
	}

	// Try adding the worktree again
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: r.config.RepositoriesDir,
	})

	if err := worktreeInstance.AddToStatus(worktree.AddToStatusParams{
		RepoURL:       params.RepoURL,
		Branch:        params.Branch,
		WorktreePath:  params.WorktreePath,
		WorkspacePath: params.WorkspacePath,
		Remote:        params.Remote,
		IssueInfo:     params.IssueInfo,
	}); err != nil {
		// Clean up created directory on status update failure
		r.cleanupWorktreeDirectory(params.WorktreePath)
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	return nil
}

// AutoAddRepositoryToStatus automatically adds a repository to the status file.
func (r *realRepository) AutoAddRepositoryToStatus(repoURL, repoPath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if it's a Git repository
	exists, err := r.fs.Exists(filepath.Join(absPath, ".git"))
	if err != nil {
		return fmt.Errorf("failed to check .git existence: %w", err)
	}
	if !exists {
		return ErrNotAGitRepository
	}

	// Get remotes information
	remotes := make(map[string]status.Remote)

	// Check for origin remote (default remote)
	originURL, err := r.git.GetRemoteURL(absPath, DefaultRemote)
	if err == nil && originURL != "" {
		remotes[DefaultRemote] = status.Remote{
			DefaultBranch: "main", // Default to main, could be enhanced to detect actual default branch
		}
	}

	// Add the repository to status
	if err := r.statusManager.AddRepository(repoURL, status.AddRepositoryParams{
		Path:    absPath,
		Remotes: remotes,
	}); err != nil {
		return fmt.Errorf("failed to add repository to status: %w", err)
	}

	return nil
}

// cleanupWorktreeDirectory cleans up the worktree directory.
func (r *realRepository) cleanupWorktreeDirectory(worktreePath string) {
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: r.config.RepositoriesDir,
	})
	if cleanupErr := worktreeInstance.CleanupDirectory(worktreePath); cleanupErr != nil {
		r.logger.Logf("Warning: failed to clean up directory after status update failure: %v", cleanupErr)
	}
}
