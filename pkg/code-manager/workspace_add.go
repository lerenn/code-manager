package codemanager

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/mode/workspace/interfaces"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
)

// AddRepositoryToWorkspaceParams contains parameters for AddRepositoryToWorkspace.
type AddRepositoryToWorkspaceParams struct {
	WorkspaceName string // Name of the workspace
	Repository    string // Repository identifier (name, path, URL)
}

// AddRepositoryToWorkspace adds a repository to an existing workspace.
func (c *realCodeManager) AddRepositoryToWorkspace(params AddRepositoryToWorkspaceParams) error {
	return c.executeWithHooks(consts.AddRepositoryToWorkspace, map[string]interface{}{
		"workspace_name": params.WorkspaceName,
		"repository":     params.Repository,
	}, func() error {
		return c.addRepositoryToWorkspace(&params)
	})
}

// addRepositoryToWorkspace implements the business logic for adding a repository to a workspace.
func (c *realCodeManager) addRepositoryToWorkspace(params *AddRepositoryToWorkspaceParams) error {
	c.VerbosePrint("Adding repository to workspace: %s", params.WorkspaceName)

	// Handle interactive selection
	workspaceName, repositoryName, err := c.handleAddRepositoryInteractiveSelection(params)
	if err != nil {
		return err
	}

	// Validate workspace exists
	workspace, err := c.deps.StatusManager.GetWorkspace(workspaceName)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWorkspaceNotFound, err)
	}

	// Check if repository already in workspace (prevent duplicates)
	if workspace.HasRepository(repositoryName) {
		return fmt.Errorf(
			"%w: repository '%s' already exists in workspace '%s'",
			ErrDuplicateRepository, repositoryName, workspaceName,
		)
	}

	// Resolve and validate repository
	_, finalRepoURL, err := c.resolveAndValidateRepositoryForAdd(workspace, workspaceName, repositoryName)
	if err != nil {
		return err
	}

	// Update params with final resolved values for success message
	params.WorkspaceName = workspaceName
	params.Repository = finalRepoURL

	// Determine branches that have worktrees in ALL existing repositories (before adding new one)
	branchesToCreate := c.getBranchesWithWorktreesInAllRepos(workspace, workspace.Repositories)

	c.VerbosePrint("Found %d branches with worktrees in all repositories: %v", len(branchesToCreate), branchesToCreate)

	// Update workspace in status.yaml to include new repository
	workspace.Repositories = append(workspace.Repositories, finalRepoURL)
	if err := c.deps.StatusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("%w: failed to update workspace: %w", ErrStatusUpdate, err)
	}

	// Create worktrees for each branch that has worktrees in all repositories
	if err := c.createWorktreesForBranches(branchesToCreate, finalRepoURL, workspaceName); err != nil {
		return err
	}

	c.VerbosePrint("Repository '%s' added to workspace '%s' successfully", finalRepoURL, workspaceName)
	return nil
}

// handleAddRepositoryInteractiveSelection handles interactive selection for workspace and repository.
func (c *realCodeManager) handleAddRepositoryInteractiveSelection(
	params *AddRepositoryToWorkspaceParams,
) (string, string, error) {
	// Handle interactive selection if workspace name not provided
	workspaceName := params.WorkspaceName
	if workspaceName == "" {
		result, err := c.promptSelectWorkspaceOnly()
		if err != nil {
			return "", "", fmt.Errorf("failed to select workspace: %w", err)
		}
		if result.Type != prompt.TargetWorkspace {
			return "", "", fmt.Errorf("selected target is not a workspace: %s", result.Type)
		}
		workspaceName = result.Name
		params.WorkspaceName = workspaceName
	}

	// Handle interactive selection if repository name not provided
	repositoryName := params.Repository
	if repositoryName == "" {
		result, err := c.promptSelectRepositoryOnly()
		if err != nil {
			return "", "", fmt.Errorf("failed to select repository: %w", err)
		}
		if result.Type != prompt.TargetRepository {
			return "", "", fmt.Errorf("selected target is not a repository: %s", result.Type)
		}
		repositoryName = result.Name
		params.Repository = repositoryName
	}

	return workspaceName, repositoryName, nil
}

// resolveAndValidateRepositoryForAdd resolves and validates a repository for adding to workspace.
func (c *realCodeManager) resolveAndValidateRepositoryForAdd(
	workspace *status.Workspace,
	workspaceName, repositoryName string,
) (string, string, error) {
	// Resolve repository path/name
	resolvedRepo, err := c.resolveRepository(repositoryName)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve repository '%s': %w", repositoryName, err)
	}

	// Get repository URL from Git remote origin
	repoURL, err := c.deps.Git.GetRemoteURL(resolvedRepo, "origin")
	if err != nil {
		// If no origin remote, use the path as the identifier
		repoURL = resolvedRepo
	}

	// Check if repository already exists in status using the remote URL
	var finalRepoURL string
	if existingRepo, err := c.deps.StatusManager.GetRepository(repoURL); err == nil && existingRepo != nil {
		finalRepoURL = repoURL
		c.VerbosePrint("  ✓ %s (already exists in status)", repositoryName)
	} else {
		// Add new repository to status file
		finalRepoURL, err = c.addRepositoryToStatus(resolvedRepo)
		if err != nil {
			return "", "", fmt.Errorf("%w: failed to add repository '%s': %w", ErrRepositoryAddition, repositoryName, err)
		}
		c.VerbosePrint("  ✓ %s (added to status)", repositoryName)
	}

	// Check if the resolved URL is already in the workspace (in case of URL mismatch)
	if workspace.HasRepository(finalRepoURL) {
		return "", "", fmt.Errorf(
			"%w: repository with URL '%s' already exists in workspace '%s'",
			ErrDuplicateRepository, finalRepoURL, workspaceName,
		)
	}

	return resolvedRepo, finalRepoURL, nil
}

// createWorktreesForBranches creates worktrees for the given branches.
func (c *realCodeManager) createWorktreesForBranches(
	branchesToCreate []string,
	finalRepoURL, workspaceName string,
) error {
	for _, branchName := range branchesToCreate {
		c.VerbosePrint("Creating worktree for branch '%s' in repository '%s'", branchName, finalRepoURL)

		// Create worktree in new repository
		_, actualRepoURL, err := c.createWorktreeForBranchInRepository(finalRepoURL, branchName)
		if err != nil {
			return fmt.Errorf("failed to create worktree for branch '%s' in repository '%s': %w", branchName, finalRepoURL, err)
		}

		// Update the .code-workspace file for that branch
		if err := c.updateWorkspaceFileForNewRepository(workspaceName, branchName, actualRepoURL); err != nil {
			return fmt.Errorf("failed to update workspace file for branch '%s': %w", branchName, err)
		}
	}
	return nil
}

// getBranchesWithWorktreesInAllRepos returns branches that have worktrees in ALL existing repositories.
func (c *realCodeManager) getBranchesWithWorktreesInAllRepos(
	workspace *status.Workspace, existingRepos []string,
) []string {
	var branchesWithAllRepos []string

	// Safety check
	if workspace == nil {
		return branchesWithAllRepos
	}

	// For each branch in workspace.Worktrees
	for _, branchName := range workspace.Worktrees {
		if c.branchHasWorktreesInAllRepos(branchName, existingRepos) {
			branchesWithAllRepos = append(branchesWithAllRepos, branchName)
			c.VerbosePrint("  ✓ Branch '%s' has worktrees in all repositories", branchName)
		}
	}

	return branchesWithAllRepos
}

// branchHasWorktreesInAllRepos checks if a branch has worktrees in all given repositories.
func (c *realCodeManager) branchHasWorktreesInAllRepos(branchName string, existingRepos []string) bool {
	// Check each repository in the list
	for _, repoURL := range existingRepos {
		// Get repository from status
		repo, err := c.deps.StatusManager.GetRepository(repoURL)
		if err != nil || repo == nil {
			c.VerbosePrint("  ⚠ Skipping repository %s: %v", repoURL, err)
			return false
		}

		// Check if repository has a worktree for that branch
		if !c.repositoryHasWorktreeForBranch(repo, branchName) {
			c.VerbosePrint("  Branch '%s' missing in repository '%s'", branchName, repoURL)
			return false
		}
	}

	return true
}

// repositoryHasWorktreeForBranch checks if a repository has a worktree for the given branch.
func (c *realCodeManager) repositoryHasWorktreeForBranch(repo *status.Repository, branchName string) bool {
	if repo.Worktrees == nil {
		return false
	}

	for _, worktree := range repo.Worktrees {
		if worktree.Branch == branchName {
			return true
		}
	}

	return false
}

// updateWorkspaceFileForNewRepository updates a workspace file to include a new repository.
func (c *realCodeManager) updateWorkspaceFileForNewRepository(workspaceName, branchName, repoURL string) error {
	// Get config to access WorkspacesDir
	cfg, err := c.deps.Config.GetConfigWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Construct workspace file path using the same utility as workspace mode
	workspaceFilePath := ws.BuildWorkspaceFilePath(cfg.WorkspacesDir, workspaceName, branchName)

	// Check if workspace file exists
	exists, err := c.deps.FS.Exists(workspaceFilePath)
	if err != nil {
		return fmt.Errorf("failed to check if workspace file exists: %w", err)
	}
	if !exists {
		c.VerbosePrint("  ⚠ Workspace file does not exist: %s (skipping update)", workspaceFilePath)
		return nil // Not an error - workspace file might not exist yet
	}

	// Read existing workspace file
	content, err := c.deps.FS.ReadFile(workspaceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read workspace file: %w", err)
	}

	// Parse JSON
	var workspaceConfig interfaces.Config
	if err := json.Unmarshal(content, &workspaceConfig); err != nil {
		return fmt.Errorf("failed to parse workspace file JSON: %w", err)
	}

	// Check if repository already in folders (prevent duplicates)
	for _, folder := range workspaceConfig.Folders {
		// Extract expected worktree path
		expectedPath := filepath.Join(cfg.RepositoriesDir, repoURL, "origin", branchName)
		if folder.Path == expectedPath {
			c.VerbosePrint("  Repository already in workspace file: %s", workspaceFilePath)
			return nil // Already added, skip
		}
	}

	// Extract repository name from URL
	repoName := extractRepositoryNameFromURL(repoURL)

	// Build worktree path
	worktreePath := filepath.Join(cfg.RepositoriesDir, repoURL, "origin", branchName)

	// Add new folder to Config.Folders
	newFolder := interfaces.Folder{
		Name: repoName,
		Path: worktreePath,
	}
	workspaceConfig.Folders = append(workspaceConfig.Folders, newFolder)

	// Marshal back to JSON
	updatedContent, err := json.MarshalIndent(workspaceConfig, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace file JSON: %w", err)
	}

	// Write file atomically
	if err := c.deps.FS.WriteFileAtomic(workspaceFilePath, updatedContent, 0644); err != nil {
		return fmt.Errorf("failed to write workspace file: %w", err)
	}

	c.VerbosePrint("  Updated workspace file: %s", workspaceFilePath)
	return nil
}

// createWorktreeForBranchInRepository creates a worktree for a specific branch in a repository.
func (c *realCodeManager) createWorktreeForBranchInRepository(repoURL, branchName string) (string, string, error) {
	// Get repository from status to get the actual managed path
	// This ensures we use the CM-managed repository location, not the original path
	repoStatus, err := c.deps.StatusManager.GetRepository(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to get repository from status: %w", err)
	}

	// Check if worktree already exists in status
	existingWorktree, err := c.deps.StatusManager.GetWorktree(repoURL, branchName)
	if err == nil && existingWorktree != nil {
		// Worktree already exists, build the path and return it
		c.VerbosePrint("  Worktree already exists for branch '%s' in repository '%s', skipping creation", branchName, repoURL)
		worktreePath := c.BuildWorktreePath(repoURL, existingWorktree.Remote, branchName)
		c.VerbosePrint("  Using existing worktree path: %s", worktreePath)
		return worktreePath, repoURL, nil
	}

	// Use the repository path from status (CM-managed location)
	managedRepoPath := repoStatus.Path

	// Get repository instance using the managed path
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: managedRepoPath,
	})

	// Create worktree using repository mode
	worktreePath, err := repoInstance.CreateWorktree(branchName, repo.CreateWorktreeOpts{
		Remote: "origin",
	})
	if err != nil {
		// Check if error is because worktree already exists (from Git side)
		worktreeExists := strings.Contains(err.Error(), "worktree already exists") ||
			strings.Contains(err.Error(), "already used by worktree")
		if worktreeExists {
			c.VerbosePrint(
				"  Worktree already exists for branch '%s' in repository '%s', skipping creation",
				branchName, repoURL,
			)
			// Build the expected worktree path
			worktreePath := c.BuildWorktreePath(repoURL, "origin", branchName)
			c.VerbosePrint("  Using existing worktree path: %s", worktreePath)
			return worktreePath, repoURL, nil
		}
		return "", "", fmt.Errorf("failed to create worktree in repository '%s': %w", repoURL, err)
	}

	// Return the repoURL we already have (it's the canonical identifier)
	return worktreePath, repoURL, nil
}

// extractRepositoryNameFromURL extracts the repository name (last part) from a Git repository URL.
func extractRepositoryNameFromURL(repoURL string) string {
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
