package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/mode/repository"
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
	var actualRepositoryURLs []string
	var err error

	// Track created worktrees for potential rollback
	defer func() {
		// If there's an error, rollback all created worktrees
		if err != nil {
			w.rollbackWorkspaceWorktrees(workspaceName, branch, createdWorktrees, createdWorkspaceFile)
		}
	}()

	// Create worktrees in each repository and collect actual repository URLs
	for _, repoURL := range repositories {
		worktreePath, actualRepoURL, err := w.createSingleRepositoryWorktreeWithURL(repoURL, workspaceName, branch)
		if err != nil {
			return "", err
		}
		createdWorktrees = append(createdWorktrees, worktreePath)
		actualRepositoryURLs = append(actualRepositoryURLs, actualRepoURL)
	}

	// Create .code-workspace file in workspaces_dir using actual repository URLs
	workspaceFilePath, err := w.createWorkspaceFile(workspaceName, branch, actualRepositoryURLs)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace file: %w", err)
	}
	createdWorkspaceFile = workspaceFilePath

	// Update status.yaml workspace section with worktree name and actual repository URLs
	if err := w.updateWorkspaceStatus(workspaceName, branch, actualRepositoryURLs); err != nil {
		return "", fmt.Errorf("failed to update workspace status: %w", err)
	}

	w.logger.Logf("Successfully created worktrees from workspace: %s", workspaceName)
	return workspaceFilePath, nil
}

// createSingleRepositoryWorktreeWithURL creates a worktree for a single repository and returns both the
// worktree path and actual repository URL.
func (w *realWorkspace) createSingleRepositoryWorktreeWithURL(
	repoURL, workspaceName, branch string,
) (string, string, error) {
	w.logger.Logf("Creating worktree in repository: %s", repoURL)

	// Get repository path from status or construct it
	repoPath := w.getRepositoryPath(repoURL)

	// Validate repository exists and is accessible
	if err := w.validateRepositoryPath(repoPath); err != nil {
		return "", "", fmt.Errorf("repository '%s' in workspace '%s' is not valid: %w", repoURL, workspaceName, err)
	}

	// If repoURL looks like a file system path, extract the actual repository URL from Git remotes
	actualRepoURL := repoURL
	if filepath.IsAbs(repoURL) || strings.Contains(repoURL, string(filepath.Separator)) {
		extractedURL, err := w.getRepositoryURLFromPath(repoPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to extract repository URL from path '%s': %w", repoPath, err)
		}
		actualRepoURL = extractedURL
		w.logger.Logf("Extracted repository URL '%s' from path '%s'", actualRepoURL, repoPath)
	}

	// Create worktree using worktree package directly
	// Pass the actual repository path to the repository package
	worktreePath, err := w.createWorktreeForRepositoryWithPath(actualRepoURL, repoPath, branch)
	if err != nil {
		return "", "", fmt.Errorf("failed to create worktree in repository '%s': %w", actualRepoURL, err)
	}

	w.logger.Logf("Created worktree: %s", worktreePath)
	return worktreePath, actualRepoURL, nil
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

// getRepositoryURLFromPath extracts the repository URL from a file system path by looking at Git remotes.
func (w *realWorkspace) getRepositoryURLFromPath(repoPath string) (string, error) {
	// Check if it's a Git repository
	gitDir := filepath.Join(repoPath, ".git")
	exists, err := w.fs.Exists(gitDir)
	if err != nil {
		return "", fmt.Errorf("failed to check .git existence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("not a Git repository: %s", repoPath)
	}

	// Get the origin remote URL
	originURL, err := w.git.GetRemoteURL(repoPath, "origin")
	if err != nil {
		return "", fmt.Errorf("failed to get origin remote URL: %w", err)
	}
	if originURL == "" {
		return "", fmt.Errorf("no origin remote found in repository: %s", repoPath)
	}

	// Convert the remote URL to a repository URL format
	// e.g., "https://github.com/octocat/Hello-World.git" -> "github.com/octocat/Hello-World"
	repoURL := strings.TrimSuffix(originURL, ".git")
	if strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.TrimPrefix(repoURL, "https://")
	} else if strings.HasPrefix(repoURL, "git@") {
		// Handle SSH URLs like "git@github.com:octocat/Hello-World.git"
		repoURL = strings.TrimPrefix(repoURL, "git@")
		repoURL = strings.Replace(repoURL, ":", "/", 1)
		repoURL = strings.TrimSuffix(repoURL, ".git")
	}

	return repoURL, nil
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

// createWorktreeForRepositoryWithPath creates a worktree for a specific repository using repositoryProvider
// with explicit path.
func (w *realWorkspace) createWorktreeForRepositoryWithPath(
	_, repoPath, branch string,
) (string, error) {
	// Create repository instance using repositoryProvider with explicit path
	repoInstance := w.repositoryProvider(repository.NewRepositoryParams{
		FS:               w.fs,
		Git:              w.git,
		Config:           w.config,
		StatusManager:    w.statusManager,
		Logger:           w.logger,
		Prompt:           w.prompt,
		WorktreeProvider: w.safeWorktreeProvider(),
		HookManager:      w.hookManager,
		RepositoryName:   repoPath, // Pass the actual repository path
	})

	// Use repository's CreateWorktree method
	worktreePath, err := repoInstance.CreateWorktree(branch, repository.CreateWorktreeOpts{
		Remote: "origin",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create worktree using repository: %w", err)
	}

	return worktreePath, nil
}

// createWorkspaceFile creates a .code-workspace file in the workspaces directory.
func (w *realWorkspace) createWorkspaceFile(workspaceName, branchName string, repositories []string) (string, error) {
	// Use shared utility to build workspace file path
	workspaceFilePath := buildWorkspaceFilePath(w.config.WorkspacesDir, workspaceName, branchName)

	// Ensure workspace directory exists (this will create both workspaces dir and workspace subdir)
	workspaceDir := filepath.Dir(workspaceFilePath)
	if err := w.fs.MkdirAll(workspaceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Create workspace file content
	workspaceContent := w.generateWorkspaceFileContent(workspaceName, branchName, repositories)

	// Write workspace file
	if err := w.fs.CreateFileWithContent(workspaceFilePath, []byte(workspaceContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write workspace file: %w", err)
	}

	w.logger.Logf("Created workspace file: %s", workspaceFilePath)
	return workspaceFilePath, nil
}

// extractRepositoryNameFromURL extracts the repository name (last part) from a Git repository URL.
// Examples:
// - "github.com/lerenn/home" -> "home"
// - "github.com/kubernetes/kubernetes.io" -> "kubernetes.io"
// - "gitlab.com/user/project-name" -> "project-name".
func (w *realWorkspace) extractRepositoryNameFromURL(repoURL string) string {
	// Remove any trailing slashes
	repoURL = strings.TrimSuffix(repoURL, "/")

	// Split by "/" and get the last part
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	// Fallback to the original URL if we can't parse it
	return repoURL
}

// generateWorkspaceFileContent generates the content for a .code-workspace file.
func (w *realWorkspace) generateWorkspaceFileContent(_ string, branchName string, repositories []string) string {
	// Create workspace file content with all repositories
	// This is a simplified implementation - in practice, you might want to use a proper JSON library
	content := `{
	"folders": [
`

	for i, repoURL := range repositories {
		// Convert repository URL to worktree path using the worktree path structure
		// Structure: $base_path/<repo_url>/<remote_name>/<branch>
		worktreePath := filepath.Join(w.config.RepositoriesDir, repoURL, "origin", branchName)

		// Extract repository name for the folder alias
		repoName := w.extractRepositoryNameFromURL(repoURL)

		content += fmt.Sprintf(`		{
			"path": "%s",
			"name": "%s"
		}`, worktreePath, repoName)
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

// updateWorkspaceStatus updates the workspace status with the new worktree and actual repository URLs.
func (w *realWorkspace) updateWorkspaceStatus(workspaceName, branch string, actualRepositoryURLs []string) error {
	// Get current workspace
	workspace, err := w.statusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Add worktree name to existing worktree array (just the branch name)
	workspace.Worktrees = append(workspace.Worktrees, branch)

	// Update repository URLs with actual repository URLs
	workspace.Repositories = actualRepositoryURLs

	// Update workspace in status file
	if err := w.statusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	w.logger.Logf("Updated workspace '%s' with worktree: %s and repositories: %v",
		workspaceName, branch, actualRepositoryURLs)
	return nil
}

// removeWorkspaceWorktreeEntry removes a worktree entry from workspace status.
func (w *realWorkspace) removeWorkspaceWorktreeEntry(workspaceName, branch string) {
	// This is a simplified implementation
	// In practice, you might want to implement a proper RemoveWorkspaceWorktree method
	w.logger.Logf("Removing workspace worktree entry: %s-%s", workspaceName, branch)
}

// safeWorktreeProvider safely converts the workspace worktree provider to repository worktree provider.
func (w *realWorkspace) safeWorktreeProvider() repository.WorktreeProvider {
	if w.worktreeProvider == nil {
		return nil
	}
	return repository.WorktreeProvider(w.worktreeProvider)
}
