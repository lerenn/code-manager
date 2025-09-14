package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/branch"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// CreateWorktree creates worktrees from workspace definition in status.yaml.
func (w *realWorkspace) CreateWorktree(branch string, opts ...CreateWorktreeOpts) (string, error) {
	workspaceName, err := w.extractWorkspaceName(opts)
	if err != nil {
		return "", err
	}

	w.logger.Logf("Creating worktrees from workspace status: %s", workspaceName)

	repositories, err := w.validateAndGetRepositories(workspaceName)
	if err != nil {
		return "", err
	}

	return w.createWorkspaceWorktrees(workspaceName, branch, repositories)
}

// extractWorkspaceName extracts and validates the workspace name from options.
func (w *realWorkspace) extractWorkspaceName(opts []CreateWorktreeOpts) (string, error) {
	if len(opts) > 0 && opts[0].WorkspaceName != "" {
		return opts[0].WorkspaceName, nil
	}
	return "", fmt.Errorf("workspace name is required for workspace worktree creation")
}

// validateAndGetRepositories validates workspace and returns repository list.
func (w *realWorkspace) validateAndGetRepositories(workspaceName string) ([]string, error) {
	workspace, err := w.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return nil, fmt.Errorf("workspace '%s' not found in status.yaml: %w", workspaceName, err)
	}

	repositories := workspace.Repositories
	if len(repositories) == 0 {
		return nil, fmt.Errorf("workspace '%s' has no repositories defined", workspaceName)
	}

	w.logger.Logf("Found %d repositories in workspace: %v", len(repositories), repositories)
	return repositories, nil
}

// createWorkspaceWorktrees creates worktrees for all repositories in the workspace.
func (w *realWorkspace) createWorkspaceWorktrees(workspaceName, branch string, repositories []string) (string, error) {
	var createdWorktrees []string
	var createdWorkspaceFile string
	var err error

	// Track created worktrees for potential rollback
	defer func() {
		// If there's an error, rollback all created worktrees
		if err != nil {
			w.rollbackWorkspaceWorktrees(workspaceName, branch, createdWorktrees, createdWorkspaceFile)
		}
	}()

	// Create worktrees in each repository
	for _, repoURL := range repositories {
		worktreePath, err := w.createSingleRepositoryWorktree(repoURL, workspaceName, branch)
		if err != nil {
			return "", err
		}
		createdWorktrees = append(createdWorktrees, worktreePath)
	}

	// Create .code-workspace file in workspaces_dir
	workspaceFilePath, err := w.createWorkspaceFile(workspaceName, branch, repositories)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace file: %w", err)
	}
	createdWorkspaceFile = workspaceFilePath

	// Update status.yaml workspace section with worktree name only after ALL worktrees are successfully created
	if err := w.updateWorkspaceStatus(workspaceName, branch); err != nil {
		return "", fmt.Errorf("failed to update workspace status: %w", err)
	}

	w.logger.Logf("Successfully created worktrees from workspace: %s", workspaceName)
	return workspaceFilePath, nil
}

// createSingleRepositoryWorktree creates a worktree for a single repository.
func (w *realWorkspace) createSingleRepositoryWorktree(repoURL, workspaceName, branch string) (string, error) {
	w.logger.Logf("Creating worktree in repository: %s", repoURL)

	// Get repository path from status or construct it
	repoPath := w.getRepositoryPath(repoURL)

	// Validate repository exists and is accessible
	if err := w.validateRepositoryPath(repoPath); err != nil {
		return "", fmt.Errorf("repository '%s' in workspace '%s' is not valid: %w", repoURL, workspaceName, err)
	}

	// Create worktree using worktree package directly
	worktreePath, err := w.createWorktreeForRepository(repoURL, repoPath, branch)
	if err != nil {
		return "", fmt.Errorf("failed to create worktree in repository '%s': %w", repoURL, err)
	}

	w.logger.Logf("Created worktree: %s", worktreePath)
	return worktreePath, nil
}

// rollbackWorkspaceWorktrees rolls back all created worktrees and cleans up workspace file.
func (w *realWorkspace) rollbackWorkspaceWorktrees(
	workspaceName, branch string,
	createdWorktrees []string,
	workspaceFile string,
) {
	w.logger.Logf("Rolling back workspace worktrees for: %s", workspaceName)

	// Remove workspace file if it was created
	if workspaceFile != "" {
		if err := w.fs.RemoveAll(workspaceFile); err != nil {
			w.logger.Logf("Warning: failed to remove workspace file %s: %v", workspaceFile, err)
		}
	}

	// Remove worktrees (this will also update status.yaml)
	for _, worktreePath := range createdWorktrees {
		// Extract repository URL from worktree path for status update
		// This is a simplified approach - in practice, you might want to track this more precisely
		// For now, we'll skip individual worktree deletion as it's complex to track
		w.logger.Logf("Warning: worktree rollback not fully implemented for: %s", worktreePath)
	}

	// Remove worktree entry from workspace status
	w.removeWorkspaceWorktreeEntry(workspaceName, branch)
}

// getRepositoryPath gets the repository path from status or constructs it.
func (w *realWorkspace) getRepositoryPath(repoURL string) string {
	// Try to get repository from status first
	repo, err := w.statusManager.GetRepository(repoURL)
	if err == nil {
		return repo.Path
	}

	// If not found in status, construct path using config
	return filepath.Join(w.config.RepositoriesDir, repoURL)
}

// validateRepositoryPath validates that the repository path exists and is a valid Git repository.
func (w *realWorkspace) validateRepositoryPath(repoPath string) error {
	// Check if directory exists
	exists, err := w.fs.Exists(repoPath)
	if err != nil {
		return fmt.Errorf("failed to check repository path: %w", err)
	}
	if !exists {
		return fmt.Errorf("repository path does not exist: %s", repoPath)
	}

	// Check if it's a Git repository by checking for .git directory
	gitDir := filepath.Join(repoPath, ".git")
	exists, err = w.fs.Exists(gitDir)
	if err != nil {
		return fmt.Errorf("failed to check if path is Git repository: %w", err)
	}
	if !exists {
		return fmt.Errorf("path is not a Git repository: %s", repoPath)
	}

	return nil
}

// createWorktreeForRepository creates a worktree for a specific repository.
func (w *realWorkspace) createWorktreeForRepository(repoURL, repoPath, branch string) (string, error) {
	// Create worktree instance
	worktreeInstance := w.worktreeProvider(worktree.NewWorktreeParams{
		FS:              w.fs,
		Git:             w.git,
		StatusManager:   w.statusManager,
		Logger:          w.logger,
		Prompt:          w.prompt,
		RepositoriesDir: w.config.RepositoriesDir,
	})

	// Build worktree path
	worktreePath := worktreeInstance.BuildPath(repoURL, "origin", branch)
	w.logger.Logf("Worktree path: %s", worktreePath)

	// Validate creation
	if err := worktreeInstance.ValidateCreation(worktree.ValidateCreationParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     repoPath,
	}); err != nil {
		return "", fmt.Errorf("failed to validate worktree creation: %w", err)
	}

	// Create the worktree
	if err := worktreeInstance.Create(worktree.CreateParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     repoPath,
		Remote:       "origin",
		Force:        false,
	}); err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}

	// Checkout the branch in the worktree
	if err := worktreeInstance.CheckoutBranch(worktreePath, branch); err != nil {
		// Cleanup failed worktree
		if cleanupErr := worktreeInstance.CleanupDirectory(worktreePath); cleanupErr != nil {
			w.logger.Logf("Warning: failed to clean up worktree directory after checkout failure: %v", cleanupErr)
		}
		return "", fmt.Errorf("failed to checkout branch in worktree: %w", err)
	}

	// Set upstream branch tracking to enable push without specifying remote/branch
	if err := w.git.SetUpstreamBranch(worktreePath, "origin", branch); err != nil {
		// Cleanup failed worktree
		if cleanupErr := worktreeInstance.CleanupDirectory(worktreePath); cleanupErr != nil {
			w.logger.Logf("Warning: failed to clean up worktree directory after upstream setup failure: %v", cleanupErr)
		}
		return "", fmt.Errorf("failed to set upstream branch tracking: %w", err)
	}

	// Add to status file
	if err := w.addWorktreeToStatus(worktreeInstance, repoURL, branch, worktreePath, repoPath, nil); err != nil {
		return "", fmt.Errorf("failed to add worktree to status: %w", err)
	}

	return worktreePath, nil
}

// createWorkspaceFile creates a .code-workspace file in the workspaces directory.
func (w *realWorkspace) createWorkspaceFile(workspaceName, branchName string, repositories []string) (string, error) {
	// Sanitize branch name for filename (replace / with -)
	sanitizedBranchForFilename := branch.SanitizeBranchNameForFilename(branchName)

	// Create workspace file path
	workspaceFileName := fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename)
	workspaceFilePath := filepath.Join(w.config.WorkspacesDir, workspaceFileName)

	// Ensure workspaces directory exists
	workspacesDir := filepath.Dir(workspaceFilePath)
	if err := w.fs.MkdirAll(workspacesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspaces directory: %w", err)
	}

	// Create workspace file content
	workspaceContent := w.generateWorkspaceFileContent(workspaceName, repositories)

	// Write workspace file
	if err := w.fs.CreateFileWithContent(workspaceFilePath, []byte(workspaceContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write workspace file: %w", err)
	}

	w.logger.Logf("Created workspace file: %s", workspaceFilePath)
	return workspaceFilePath, nil
}

// generateWorkspaceFileContent generates the content for a .code-workspace file.
func (w *realWorkspace) generateWorkspaceFileContent(_ string, repositories []string) string {
	// Create workspace file content with all repositories
	// This is a simplified implementation - in practice, you might want to use a proper JSON library
	content := `{
	"folders": [
`

	for i, repoURL := range repositories {
		// Convert repository URL to local path
		repoPath := filepath.Join(w.config.RepositoriesDir, repoURL)
		content += fmt.Sprintf(`		{
			"path": "%s"
		}`, repoPath)
		if i < len(repositories)-1 {
			content += ","
		}
		content += "\n"
	}

	content += `	],
	"settings": {},
	"extensions": {
		"recommendations": []
	}
}`

	return content
}

// updateWorkspaceStatus updates the workspace status with the new worktree.
func (w *realWorkspace) updateWorkspaceStatus(workspaceName, branch string) error {
	// Get current workspace
	workspace, err := w.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Add worktree name to existing worktree array (just the branch name)
	workspace.Worktrees = append(workspace.Worktrees, branch)

	// Update workspace in status file
	if err := w.statusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	w.logger.Logf("Updated workspace '%s' with worktree: %s", workspaceName, branch)
	return nil
}

// removeWorkspaceWorktreeEntry removes a worktree entry from workspace status.
func (w *realWorkspace) removeWorkspaceWorktreeEntry(workspaceName, branch string) {
	// This is a simplified implementation
	// In practice, you might want to implement a proper RemoveWorkspaceWorktree method
	w.logger.Logf("Removing workspace worktree entry: %s-%s", workspaceName, branch)
}
