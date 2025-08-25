// Package worktree provides worktree management functionality for CM.
package worktree

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
)

//go:generate mockgen -source=worktree.go -destination=mockworktree.gen.go -package=worktree

// Worktree interface provides worktree management capabilities.
type Worktree interface {
	// BuildPath constructs a worktree path from repository URL, remote name, and branch.
	BuildPath(repoURL, remoteName, branch string) string

	// Create creates a new worktree with proper validation and cleanup.
	Create(params CreateParams) error

	// Delete deletes a worktree with proper cleanup and confirmation.
	Delete(params DeleteParams) error

	// ValidateCreation validates that worktree creation is possible.
	ValidateCreation(params ValidateCreationParams) error

	// ValidateDeletion validates that worktree deletion is possible.
	ValidateDeletion(params ValidateDeletionParams) error

	// EnsureBranchExists ensures the specified branch exists, creating it if necessary.
	EnsureBranchExists(repoPath, branch string) error

	// AddToStatus adds the worktree to the status file.
	AddToStatus(params AddToStatusParams) error

	// RemoveFromStatus removes the worktree from the status file.
	RemoveFromStatus(repoURL, branch string) error

	// CleanupDirectory removes the worktree directory.
	CleanupDirectory(worktreePath string) error

	// Exists checks if a worktree exists for the specified branch.
	Exists(repoPath, branch string) (bool, error)

	// GetPath gets the path of a worktree for a branch.
	GetPath(repoPath, branch string) (string, error)
}

// CreateParams contains parameters for worktree creation.
type CreateParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	RepoPath     string
	Remote       string
	IssueInfo    *issue.Info
	Force        bool
}

// DeleteParams contains parameters for worktree deletion.
type DeleteParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	RepoPath     string
	Force        bool
}

// ValidateCreationParams contains parameters for worktree creation validation.
type ValidateCreationParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	RepoPath     string
}

// ValidateDeletionParams contains parameters for worktree deletion validation.
type ValidateDeletionParams struct {
	RepoURL string
	Branch  string
}

// AddToStatusParams contains parameters for adding worktree to status.
type AddToStatusParams struct {
	RepoURL       string
	Branch        string
	WorktreePath  string
	WorkspacePath string
	Remote        string
	IssueInfo     *issue.Info
}

// realWorktree provides the real implementation of the Worktree interface.
type realWorktree struct {
	fs            fs.FS
	git           git.Git
	statusManager status.Manager
	logger        logger.Logger
	prompt        prompt.Prompt
	basePath      string
	verbose       bool
}

// NewWorktreeParams contains parameters for creating a new Worktree instance.
type NewWorktreeParams struct {
	FS            fs.FS
	Git           git.Git
	StatusManager status.Manager
	Logger        logger.Logger
	Prompt        prompt.Prompt
	BasePath      string
	Verbose       bool
}

// NewWorktree creates a new Worktree instance.
func NewWorktree(params NewWorktreeParams) Worktree {
	return &realWorktree{
		fs:            params.FS,
		git:           params.Git,
		statusManager: params.StatusManager,
		logger:        params.Logger,
		prompt:        params.Prompt,
		basePath:      params.BasePath,
		verbose:       params.Verbose,
	}
}

// BuildPath constructs a worktree path from repository URL, remote name, and branch.
func (w *realWorktree) BuildPath(repoURL, remoteName, branch string) string {
	// Use structure: $base_path/<repo_url>/<remote_name>/<branch>
	return filepath.Join(w.basePath, repoURL, remoteName, branch)
}

// Create creates a new worktree with proper validation and cleanup.
func (w *realWorktree) Create(params CreateParams) error {
	w.verboseLogf("Creating worktree for %s:%s at %s", params.Remote, params.Branch, params.WorktreePath)

	// Validate creation
	if err := w.ValidateCreation(ValidateCreationParams{
		RepoURL:      params.RepoURL,
		Branch:       params.Branch,
		WorktreePath: params.WorktreePath,
		RepoPath:     params.RepoPath,
	}); err != nil {
		return err
	}

	// Ensure branch exists
	if err := w.EnsureBranchExists(params.RepoPath, params.Branch); err != nil {
		return err
	}

	// Create worktree directory
	if err := w.createWorktreeDirectory(params.WorktreePath); err != nil {
		return err
	}

	// Create Git worktree
	if err := w.git.CreateWorktree(params.RepoPath, params.WorktreePath, params.Branch); err != nil {
		// Clean up directory on failure
		if cleanupErr := w.cleanupWorktreeDirectory(params.WorktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	w.verboseLogf("✓ Worktree created successfully for %s:%s", params.Remote, params.Branch)
	return nil
}

// Delete deletes a worktree with proper cleanup and confirmation.
func (w *realWorktree) Delete(params DeleteParams) error {
	w.verboseLogf("Deleting worktree for %s at %s", params.Branch, params.WorktreePath)

	// Validate deletion
	if err := w.ValidateDeletion(ValidateDeletionParams{
		RepoURL: params.RepoURL,
		Branch:  params.Branch,
	}); err != nil {
		return err
	}

	// Prompt for confirmation unless force flag is used
	if !params.Force {
		if err := w.promptForConfirmation(params.Branch, params.WorktreePath); err != nil {
			return err
		}
	}

	// Remove worktree from Git tracking first
	if err := w.git.RemoveWorktree(params.RepoPath, params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree from Git: %w", err)
	}

	// Remove worktree directory
	if err := w.fs.RemoveAll(params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Remove entry from status file
	if err := w.RemoveFromStatus(params.RepoURL, params.Branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status: %w", err)
	}

	w.verboseLogf("✓ Worktree deleted successfully for %s", params.Branch)
	return nil
}

// ValidateCreation validates that worktree creation is possible.
func (w *realWorktree) ValidateCreation(params ValidateCreationParams) error {
	// Check if worktree directory already exists
	exists, err := w.fs.Exists(params.WorktreePath)
	if err != nil {
		return fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, params.WorktreePath)
	}

	// Check if worktree already exists in status file
	existingWorktree, err := w.statusManager.GetWorktree(params.RepoURL, params.Branch)
	if err == nil && existingWorktree != nil {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, params.RepoURL, params.Branch)
	}

	// Create worktree directory structure
	if err := w.fs.MkdirAll(filepath.Dir(params.WorktreePath), 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return nil
}

// ValidateDeletion validates that worktree deletion is possible.
func (w *realWorktree) ValidateDeletion(params ValidateDeletionParams) error {
	// Check if worktree exists in status file
	existingWorktree, err := w.statusManager.GetWorktree(params.RepoURL, params.Branch)
	if err != nil || existingWorktree == nil {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotInStatus, params.RepoURL, params.Branch)
	}

	return nil
}

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

	if branchExists {
		return nil
	}

	return w.createBranchFromDefault(repoPath, branch)
}

// createBranchFromDefault creates a new branch from the default branch.
func (w *realWorktree) createBranchFromDefault(repoPath, branch string) error {
	w.verboseLogf("Branch %s does not exist, creating from upstream default branch", branch)

	// Fetch from remote to ensure we have the latest changes
	if err := w.fetchFromOrigin(repoPath); err != nil {
		return err
	}

	// Get remote URL and default branch
	remoteURL, err := w.git.GetRemoteURL(repoPath, "origin")
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteDefaultBranch, err := w.git.GetDefaultBranch(remoteURL)
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}

	// Try to create from local branches first
	if err := w.createBranchFromLocal(repoPath, branch, remoteDefaultBranch); err == nil {
		return nil
	}

	// Fall back to remote default branch
	return w.createBranchFromRemote(repoPath, branch, remoteDefaultBranch)
}

// fetchFromOrigin fetches from the origin remote.
func (w *realWorktree) fetchFromOrigin(repoPath string) error {
	w.verboseLogf("Fetching from origin to ensure repository is up to date")
	if err := w.git.FetchRemote(repoPath, "origin"); err != nil {
		return fmt.Errorf("failed to fetch from origin: %w", err)
	}
	return nil
}

// createBranchFromRemote attempts to create a branch from the remote default branch.
func (w *realWorktree) createBranchFromRemote(repoPath, branch, remoteDefaultBranch string) error {
	w.verboseLogf("Attempting to create branch %s from remote default branch %s", branch, remoteDefaultBranch)

	err := w.git.CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: "origin/" + remoteDefaultBranch,
	})

	if err == nil {
		w.verboseLogf("Successfully created branch %s from remote default branch %s", branch, remoteDefaultBranch)
		return nil
	}

	w.verboseLogf("Failed to create from remote default branch, falling back to local branches: %v", err)
	return err
}

// createBranchFromLocal creates a branch from a local branch.
func (w *realWorktree) createBranchFromLocal(repoPath, branch, remoteDefaultBranch string) error {
	fromBranch, err := w.findLocalDefaultBranch(repoPath, remoteDefaultBranch)
	if err != nil {
		return err
	}

	// Create the new branch from the determined branch
	if err := w.git.CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: fromBranch,
	}); err != nil {
		return fmt.Errorf("failed to create branch %s from %s: %w", branch, fromBranch, err)
	}

	return nil
}

// findLocalDefaultBranch finds a suitable local branch to create from.
func (w *realWorktree) findLocalDefaultBranch(repoPath, remoteDefaultBranch string) (string, error) {
	// First, try to find a local branch with the same name as the remote's default branch
	exists, err := w.git.BranchExists(repoPath, remoteDefaultBranch)
	if err == nil && exists {
		return remoteDefaultBranch, nil
	}

	// If no local branch with the same name, try common default branch names
	possibleDefaultBranches := []string{"main", "master"}
	for _, possibleBranch := range possibleDefaultBranches {
		exists, err := w.git.BranchExists(repoPath, possibleBranch)
		if err != nil {
			w.verboseLogf("Warning: failed to check if branch %s exists: %v", possibleBranch, err)
			continue
		}
		if exists {
			return possibleBranch, nil
		}
	}

	// If no local branch found, return an error to trigger fallback to remote
	return "", fmt.Errorf("no suitable local branch found")
}

// AddToStatus adds the worktree to the status file.
func (w *realWorktree) AddToStatus(params AddToStatusParams) error {
	if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:       params.RepoURL,
		Branch:        params.Branch,
		WorktreePath:  params.WorktreePath,
		WorkspacePath: params.WorkspacePath,
		Remote:        params.Remote,
		IssueInfo:     params.IssueInfo,
	}); err != nil {
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}
	return nil
}

// RemoveFromStatus removes the worktree from the status file.
func (w *realWorktree) RemoveFromStatus(repoURL, branch string) error {
	return w.statusManager.RemoveWorktree(repoURL, branch)
}

// CleanupDirectory removes the worktree directory.
func (w *realWorktree) CleanupDirectory(worktreePath string) error {
	if err := w.fs.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}
	return nil
}

// Exists checks if a worktree exists for the specified branch.
func (w *realWorktree) Exists(repoPath, branch string) (bool, error) {
	return w.git.WorktreeExists(repoPath, branch)
}

// GetPath gets the path of a worktree for a branch.
func (w *realWorktree) GetPath(repoPath, branch string) (string, error) {
	return w.git.GetWorktreePath(repoPath, branch)
}

// createWorktreeDirectory creates the worktree directory.
func (w *realWorktree) createWorktreeDirectory(worktreePath string) error {
	if err := w.fs.MkdirAll(worktreePath, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}
	return nil
}

// cleanupWorktreeDirectory removes the worktree directory.
func (w *realWorktree) cleanupWorktreeDirectory(worktreePath string) error {
	if err := w.fs.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to cleanup worktree directory: %w", err)
	}
	return nil
}

// promptForConfirmation prompts the user for confirmation before deletion.
func (w *realWorktree) promptForConfirmation(branch, worktreePath string) error {
	message := fmt.Sprintf(
		"You are about to delete the worktree for branch '%s'\nWorktree path: %s\nAre you sure you want to continue?",
		branch, worktreePath,
	)

	result, err := w.prompt.PromptForConfirmation(message, false)
	if err != nil {
		return err
	}

	if !result {
		return ErrDeletionCancelled
	}

	return nil
}

// verboseLogf logs a formatted message only in verbose mode.
func (w *realWorktree) verboseLogf(msg string, args ...interface{}) {
	if w.verbose {
		w.logger.Logf(fmt.Sprintf(msg, args...))
	}
}
