package codemanager

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	ws "github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/mode/workspace/interfaces"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
)

// RemoveRepositoryFromWorkspaceParams contains parameters for RemoveRepositoryFromWorkspace.
type RemoveRepositoryFromWorkspaceParams struct {
	WorkspaceName string // Name of the workspace
	Repository    string // Repository identifier (name, path, URL)
}

// RemoveRepositoryFromWorkspace removes a repository from an existing workspace.
func (c *realCodeManager) RemoveRepositoryFromWorkspace(params *RemoveRepositoryFromWorkspaceParams) error {
	return c.executeWithHooks(consts.RemoveRepositoryFromWorkspace, map[string]interface{}{
		"workspace_name": params.WorkspaceName,
		"repository":     params.Repository,
	}, func() error {
		return c.removeRepositoryFromWorkspace(params)
	})
}

// removeRepositoryFromWorkspace implements the business logic for removing a repository from a workspace.
func (c *realCodeManager) removeRepositoryFromWorkspace(params *RemoveRepositoryFromWorkspaceParams) error {
	c.VerbosePrint("Removing repository from workspace: %s", params.WorkspaceName)

	// Handle interactive selection and validate workspace
	workspaceName, workspace, err := c.handleRemoveRepositoryInteractiveSelection(params)
	if err != nil {
		return err
	}

	// Resolve and validate repository
	repoURL, err := c.resolveAndValidateRepositoryForRemove(workspace, workspaceName, params)
	if err != nil {
		return err
	}

	// Update params with final resolved values for success message
	params.WorkspaceName = workspaceName
	params.Repository = repoURL

	c.VerbosePrint("Removing repository '%s' from workspace '%s'", repoURL, workspaceName)

	// Update all .code-workspace files to remove the repository folder entries
	for _, branchName := range workspace.Worktrees {
		if err := c.removeRepositoryFromWorkspaceFile(workspaceName, branchName, repoURL); err != nil {
			return fmt.Errorf("failed to update workspace file for branch '%s': %w", branchName, err)
		}
	}

	// Remove repository from workspace.Repositories in status.yaml
	updatedRepos := make([]string, 0, len(workspace.Repositories))
	for _, repo := range workspace.Repositories {
		if repo != repoURL {
			updatedRepos = append(updatedRepos, repo)
		}
	}
	workspace.Repositories = updatedRepos

	if err := c.deps.StatusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("%w: failed to update workspace: %w", ErrStatusUpdate, err)
	}

	c.VerbosePrint("Repository '%s' removed from workspace '%s' successfully", repoURL, workspaceName)
	return nil
}

// removeRepositoryFromWorkspaceFile removes a repository folder from a workspace file.
func (c *realCodeManager) removeRepositoryFromWorkspaceFile(workspaceName, branchName, repoURL string) error {
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
		return nil // Not an error - workspace file might not exist
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

	// Build expected worktree path for this repository and branch
	expectedPath := filepath.Join(cfg.RepositoriesDir, repoURL, "origin", branchName)

	// Remove repository folder from Config.Folders
	updatedFolders, found := c.filterRepositoryFolder(workspaceConfig.Folders, expectedPath, workspaceFilePath)
	if !found {
		c.VerbosePrint("  Repository folder not found in workspace file: %s (skipping)", workspaceFilePath)
		return nil // Not an error - folder might not be in this workspace file
	}

	workspaceConfig.Folders = updatedFolders

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

// handleRemoveRepositoryInteractiveSelection handles interactive selection for workspace removal.
func (c *realCodeManager) handleRemoveRepositoryInteractiveSelection(
	params *RemoveRepositoryFromWorkspaceParams,
) (string, *status.Workspace, error) {
	// Handle interactive selection if workspace name not provided
	workspaceName := params.WorkspaceName
	if workspaceName == "" {
		result, err := c.promptSelectWorkspaceOnly()
		if err != nil {
			return "", nil, fmt.Errorf("failed to select workspace: %w", err)
		}
		if result.Type != prompt.TargetWorkspace {
			return "", nil, fmt.Errorf("selected target is not a workspace: %s", result.Type)
		}
		workspaceName = result.Name
		params.WorkspaceName = workspaceName
	}

	// Validate workspace exists
	workspace, err := c.deps.StatusManager.GetWorkspace(workspaceName)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrWorkspaceNotFound, err)
	}

	return workspaceName, workspace, nil
}

// resolveAndValidateRepositoryForRemove resolves and validates a repository for removal from workspace.
func (c *realCodeManager) resolveAndValidateRepositoryForRemove(
	workspace *status.Workspace,
	workspaceName string,
	params *RemoveRepositoryFromWorkspaceParams,
) (string, error) {
	// Resolve repository to get its URL
	repositoryName := params.Repository
	if repositoryName == "" {
		// If no repository provided, prompt for one from the workspace's repositories
		result, err := c.promptSelectRepositoryFromWorkspace(workspaceName)
		if err != nil {
			return "", fmt.Errorf("failed to select repository: %w", err)
		}
		if result.Type != prompt.TargetRepository {
			return "", fmt.Errorf("selected target is not a repository: %s", result.Type)
		}
		repositoryName = result.Name
		params.Repository = repositoryName
	}

	// Resolve repository path/name to get the repository URL
	resolvedRepo, err := c.resolveRepository(repositoryName)
	if err != nil {
		return "", fmt.Errorf("failed to resolve repository '%s': %w", repositoryName, err)
	}

	// Get repository URL from Git remote origin
	repoURL, err := c.deps.Git.GetRemoteURL(resolvedRepo, "origin")
	if err != nil {
		// If no origin remote, use the path as the identifier
		repoURL = resolvedRepo
	}

	// Check if repository is in workspace
	if !workspace.HasRepository(repoURL) {
		// Try with the resolved path as well
		if !workspace.HasRepository(resolvedRepo) {
			return "", fmt.Errorf("repository '%s' is not in workspace '%s'", repositoryName, workspaceName)
		}
		repoURL = resolvedRepo
	}

	return repoURL, nil
}

// filterRepositoryFolder filters out the repository folder from the folders list.
func (c *realCodeManager) filterRepositoryFolder(
	folders []interfaces.Folder,
	expectedPath, workspaceFilePath string,
) ([]interfaces.Folder, bool) {
	updatedFolders := make([]interfaces.Folder, 0, len(folders))
	found := false

	for _, folder := range folders {
		if folder.Path == expectedPath {
			found = true
			c.VerbosePrint("  Removing repository folder from workspace file: %s", workspaceFilePath)
			continue // Skip this folder
		}
		updatedFolders = append(updatedFolders, folder)
	}

	return updatedFolders, found
}

// promptSelectRepositoryFromWorkspace prompts the user to select a repository from a workspace.
func (c *realCodeManager) promptSelectRepositoryFromWorkspace(workspaceName string) (TargetSelectionResult, error) {
	// Get workspace to get its repositories
	workspace, err := c.deps.StatusManager.GetWorkspace(workspaceName)
	if err != nil {
		return TargetSelectionResult{}, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Build choices from workspace repositories
	var choices []prompt.TargetChoice
	for _, repoURL := range workspace.Repositories {
		choices = append(choices, prompt.TargetChoice{
			Type: prompt.TargetRepository,
			Name: repoURL,
		})
	}

	if len(choices) == 0 {
		return TargetSelectionResult{}, fmt.Errorf("workspace '%s' has no repositories", workspaceName)
	}

	// Use the prompt package to get repository selection
	selectedChoice, err := c.deps.Prompt.PromptSelectTarget(choices, false)
	if err != nil {
		return TargetSelectionResult{}, fmt.Errorf("failed to get repository selection: %w", err)
	}

	return TargetSelectionResult{
		Name: selectedChoice.Name,
		Type: selectedChoice.Type,
	}, nil
}
