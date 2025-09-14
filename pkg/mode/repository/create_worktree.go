package repository

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// CreateWorktree creates a worktree for the repository with the specified branch.
func (r *realRepository) CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error) {
	r.logger.Logf("Creating worktree for single repository with branch: %s", branch)

	// Validate repository
	validationResult, err := r.ValidateRepository(ValidationParams{Branch: branch})
	if err != nil {
		return "", err
	}

	// Get remote from options
	remote := r.extractRemote(opts)

	// Create and validate worktree instance
	worktreeInstance, worktreePath, err := r.createAndValidateWorktreeInstance(validationResult.RepoURL, branch, remote)
	if err != nil {
		return "", err
	}

	// Get current directory
	currentDir, err := filepath.Abs(r.repositoryPath)
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get issue info from options
	issueInfo := r.extractIssueInfo(opts)

	// Create the worktree (with --no-checkout)
	if err := r.createWorktreeWithNoCheckout(
		worktreeInstance, validationResult.RepoURL, branch, worktreePath, currentDir, issueInfo, remote,
	); err != nil {
		return "", err
	}

	// Execute worktree checkout hooks (for git-crypt setup, etc.)
	if err := r.executeWorktreeCheckoutHooks(
		worktreeInstance, worktreePath, branch, currentDir, validationResult.RepoURL,
	); err != nil {
		return "", err
	}

	// Now checkout the branch
	if err := r.checkoutBranchInWorktree(worktreeInstance, worktreePath, branch, remote); err != nil {
		// Clean up worktree on checkout failure
		r.cleanupWorktreeOnError(worktreeInstance, worktreePath, "checkout failure")
		return "", err
	}

	// Add to status file with auto-repository handling
	if err := r.addWorktreeToStatusAndHandleCleanup(
		worktreeInstance, validationResult.RepoURL, branch, worktreePath, issueInfo, remote,
	); err != nil {
		return "", err
	}

	r.logger.Logf("Successfully created worktree for branch %s at %s", branch, worktreePath)

	return worktreePath, nil
}

// extractIssueInfo extracts issue info from options if provided.
func (r *realRepository) extractIssueInfo(opts []CreateWorktreeOpts) *issue.Info {
	if len(opts) > 0 && opts[0].IssueInfo != nil {
		return opts[0].IssueInfo
	}
	return nil
}

// extractRemote extracts remote name from options if provided, otherwise returns DefaultRemote.
func (r *realRepository) extractRemote(opts []CreateWorktreeOpts) string {
	if len(opts) > 0 && opts[0].Remote != "" {
		return opts[0].Remote
	}
	return DefaultRemote
}

// cleanupWorktreeOnError cleans up worktree directory on error with logging.
func (r *realRepository) cleanupWorktreeOnError(worktreeInstance worktree.Worktree, worktreePath, context string) {
	if cleanupErr := worktreeInstance.CleanupDirectory(worktreePath); cleanupErr != nil {
		r.logger.Logf("Warning: failed to clean up worktree directory after %s: %v", context, cleanupErr)
	}
}

// executeWorktreeCheckoutHooks executes worktree checkout hooks with proper error handling.
func (r *realRepository) executeWorktreeCheckoutHooks(
	worktreeInstance worktree.Worktree,
	worktreePath, branch, currentDir, repoURL string,
) error {
	if r.hookManager == nil {
		return nil
	}

	ctx := &hooks.HookContext{
		OperationName: "CreateWorkTree",
		Parameters: map[string]interface{}{
			"worktreePath": worktreePath,
			"branch":       branch,
			"repoPath":     currentDir,
			"repoURL":      repoURL,
		},
		Results:  make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	if err := r.hookManager.ExecuteWorktreeCheckoutHooks("CreateWorkTree", ctx); err != nil {
		// Cleanup failed worktree
		r.cleanupWorktreeOnError(worktreeInstance, worktreePath, "hook failure")
		return fmt.Errorf("worktree checkout hooks failed: %w", err)
	}

	return nil
}

// createAndValidateWorktreeInstance creates and validates a worktree instance.
func (r *realRepository) createAndValidateWorktreeInstance(
	repoURL, branch, remote string,
) (worktree.Worktree, string, error) {
	// Create worktree instance using provider
	worktreeInstance := r.worktreeProvider(worktree.NewWorktreeParams{
		FS:              r.fs,
		Git:             r.git,
		StatusManager:   r.statusManager,
		Logger:          r.logger,
		Prompt:          r.prompt,
		RepositoriesDir: r.config.RepositoriesDir,
	})

	// Build worktree path
	worktreePath := worktreeInstance.BuildPath(repoURL, remote, branch)
	r.logger.Logf("Worktree path: %s", worktreePath)

	// Validate creation
	if err := worktreeInstance.ValidateCreation(worktree.ValidateCreationParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     r.repositoryPath,
	}); err != nil {
		return nil, "", err
	}

	return worktreeInstance, worktreePath, nil
}

// createWorktreeWithNoCheckout creates the worktree with --no-checkout flag.
func (r *realRepository) createWorktreeWithNoCheckout(
	worktreeInstance worktree.Worktree,
	repoURL, branch, worktreePath, currentDir string,
	issueInfo *issue.Info,
	remote string,
) error {
	return worktreeInstance.Create(worktree.CreateParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     currentDir,
		Remote:       remote,
		IssueInfo:    issueInfo,
		Force:        false,
	})
}

// checkoutBranchInWorktree checks out the branch in the worktree with proper error handling.
func (r *realRepository) checkoutBranchInWorktree(
	worktreeInstance worktree.Worktree,
	worktreePath, branch, remote string,
) error {
	if err := worktreeInstance.CheckoutBranch(worktreePath, branch); err != nil {
		return fmt.Errorf("failed to checkout branch in worktree: %w", err)
	}

	// Set upstream branch tracking to enable push without specifying remote/branch
	if err := r.git.SetUpstreamBranch(worktreePath, remote, branch); err != nil {
		return fmt.Errorf("failed to set upstream branch tracking: %w", err)
	}

	return nil
}

// addWorktreeToStatusAndHandleCleanup adds the worktree to status and handles cleanup on failure.
func (r *realRepository) addWorktreeToStatusAndHandleCleanup(
	worktreeInstance worktree.Worktree,
	repoURL string,
	branch string,
	worktreePath string,
	issueInfo *issue.Info,
	remote string,
) error {
	if err := r.AddWorktreeToStatus(StatusParams{
		RepoURL:       repoURL,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: "",
		Remote:        remote,
		IssueInfo:     issueInfo,
	}); err != nil {
		// Clean up worktree on status failure
		r.cleanupWorktreeOnError(worktreeInstance, worktreePath, "status failure")
		return err
	}
	return nil
}
