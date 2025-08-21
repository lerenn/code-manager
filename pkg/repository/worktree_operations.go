// Package repository provides Git repository management functionality for CM.
package repository

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/issue"
)

// WorktreeCreationParams contains parameters for worktree creation.
type WorktreeCreationParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	IssueInfo    *issue.Info
}

// WorktreeDeletionParams contains parameters for worktree deletion.
type WorktreeDeletionParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	Force        bool
}

// PrepareWorktreeCreation validates the repository and prepares the worktree path.
func (r *realRepository) PrepareWorktreeCreation(branch string) (string, string, error) {
	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{Branch: branch})
	if err != nil {
		return "", "", err
	}

	// Prepare worktree path
	worktreePath, err := r.PrepareWorktreePath(validationResult.RepoURL, branch)
	if err != nil {
		return "", "", err
	}

	return validationResult.RepoURL, worktreePath, nil
}

// PrepareWorktreePath prepares the worktree directory path.
func (r *realRepository) PrepareWorktreePath(repoURL, branch string) (string, error) {
	// Create worktree directory path
	worktreePath := r.BuildWorktreePath(repoURL, "origin", branch)

	r.VerbosePrint("Worktree path: %s", worktreePath)

	// Check if worktree directory already exists
	exists, err := r.FS.Exists(worktreePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return "", fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, worktreePath)
	}

	// Create worktree directory structure
	if err := r.FS.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return worktreePath, nil
}

// ExecuteWorktreeCreation creates the branch and worktree.
func (r *realRepository) ExecuteWorktreeCreation(params WorktreeCreationParams) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Ensure branch exists
	if err := r.EnsureBranchExists(currentDir, params.Branch); err != nil {
		return err
	}

	// Create worktree with cleanup
	if err := r.CreateWorktreeWithCleanup(CreateWorktreeWithCleanupParams{
		RepoURL:      params.RepoURL,
		Branch:       params.Branch,
		WorktreePath: params.WorktreePath,
		CurrentDir:   currentDir,
		IssueInfo:    params.IssueInfo,
	}); err != nil {
		return err
	}

	return nil
}

// EnsureBranchExists ensures the branch exists, creating it if necessary.
func (r *realRepository) EnsureBranchExists(currentDir, branch string) error {
	branchExists, err := r.Git.BranchExists(currentDir, branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if !branchExists {
		r.VerbosePrint("Branch %s does not exist, creating from current branch", branch)
		if err := r.Git.CreateBranch(currentDir, branch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}
	}

	return nil
}

// CreateWorktreeWithCleanupParams contains parameters for CreateWorktreeWithCleanup.
type CreateWorktreeWithCleanupParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	CurrentDir   string
	IssueInfo    *issue.Info
}

// CreateWorktreeWithCleanup creates the worktree with proper cleanup on failure.
func (r *realRepository) CreateWorktreeWithCleanup(params CreateWorktreeWithCleanupParams) error {
	// Update status file with worktree entry (before creating the worktree for proper cleanup)
	if err := r.AddWorktreeToStatus(StatusParams{
		RepoURL:       params.RepoURL,
		Branch:        params.Branch,
		WorktreePath:  params.WorktreePath,
		WorkspacePath: "",
		Remote:        "origin", // Set the remote to "origin" for repository worktrees
		IssueInfo:     params.IssueInfo,
	}); err != nil {
		return err
	}

	// Create the Git worktree
	if err := r.Git.CreateWorktree(params.CurrentDir, params.WorktreePath, params.Branch); err != nil {
		// Clean up on worktree creation failure
		r.CleanupOnWorktreeCreationFailure(params.RepoURL, params.Branch, params.WorktreePath)
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	return nil
}

// PrepareWorktreeDeletion validates the repository and prepares the worktree deletion.
func (r *realRepository) PrepareWorktreeDeletion(branch string) (string, string, error) {
	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{})
	if err != nil {
		return "", "", err
	}

	// Check if worktree exists in status file
	if err := r.ValidateWorktreeExists(validationResult.RepoURL, branch); err != nil {
		return "", "", err
	}

	// Get worktree path from Git
	worktreePath, err := r.Git.GetWorktreePath(validationResult.RepoPath, branch)
	if err != nil {
		return "", "", fmt.Errorf("failed to get worktree path: %w", err)
	}

	r.VerbosePrint("Worktree path: %s", worktreePath)

	return validationResult.RepoURL, worktreePath, nil
}

// ExecuteWorktreeDeletion deletes the worktree with proper cleanup.
func (r *realRepository) ExecuteWorktreeDeletion(params WorktreeDeletionParams) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Prompt for confirmation unless force flag is used
	if !params.Force {
		if err := r.PromptForConfirmation(params.Branch, params.WorktreePath); err != nil {
			return err
		}
	}

	// Remove worktree from Git tracking first
	if err := r.Git.RemoveWorktree(currentDir, params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree from Git: %w", err)
	}

	// Remove worktree directory
	if err := r.FS.RemoveAll(params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Remove entry from status file
	if err := r.RemoveWorktreeFromStatus(params.RepoURL, params.Branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status: %w", err)
	}

	return nil
}

// PromptForConfirmation prompts the user for confirmation before deletion.
func (r *realRepository) PromptForConfirmation(branch, worktreePath string) error {
	fmt.Printf("You are about to delete the worktree for branch '%s'\n", branch)
	fmt.Printf("Worktree path: %s\n", worktreePath)
	fmt.Print("Are you sure you want to continue? (y/N): ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	result, err := r.ParseConfirmationInput(input)
	if err != nil {
		return err
	}

	if !result {
		return ErrDeletionCancelled
	}

	return nil
}
