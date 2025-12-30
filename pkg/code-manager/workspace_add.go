package codemanager

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/git"
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
func (c *realCodeManager) AddRepositoryToWorkspace(params *AddRepositoryToWorkspaceParams) error {
	return c.executeWithHooks(consts.AddRepositoryToWorkspace, map[string]interface{}{
		"workspace_name": params.WorkspaceName,
		"repository":     params.Repository,
	}, func() error {
		return c.addRepositoryToWorkspace(params)
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
	originalRepoPath, finalRepoURL, err := c.resolveAndValidateRepositoryForAdd(workspace, workspaceName, repositoryName)
	if err != nil {
		return err
	}

	// Update params with final resolved values for success message
	params.WorkspaceName = workspaceName
	params.Repository = finalRepoURL

	// Try to create worktrees for ALL branches in the workspace
	// If a branch doesn't exist in the new repository, it will be skipped gracefully
	branchesToCreate := workspace.Worktrees

	c.VerbosePrint(
		"Will attempt to create worktrees for %d branches in workspace: %v",
		len(branchesToCreate), branchesToCreate,
	)

	// Update workspace in status.yaml to include new repository
	workspace.Repositories = append(workspace.Repositories, finalRepoURL)
	if err := c.deps.StatusManager.UpdateWorkspace(workspaceName, *workspace); err != nil {
		return fmt.Errorf("%w: failed to update workspace: %w", ErrStatusUpdate, err)
	}

	// Update all existing workspace files to include the new repository
	// This must happen before worktree creation so that all workspace files are updated,
	// regardless of whether worktrees are created for all branches
	if err := c.updateAllWorkspaceFilesForNewRepository(workspaceName, finalRepoURL, workspace.Worktrees); err != nil {
		return fmt.Errorf("failed to update workspace files: %w", err)
	}

	// Create worktrees for all branches in the workspace
	// If a branch doesn't exist in the new repository, it will be skipped gracefully
	// Pass originalRepoPath to check for local branches that might not exist in cloned repository
	if err := c.createWorktreesForBranches(branchesToCreate, finalRepoURL, workspaceName, originalRepoPath); err != nil {
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
	rawRepoURL, err := c.deps.Git.GetRemoteURL(resolvedRepo, "origin")
	if err != nil {
		// If no origin remote, use the path as the identifier
		// This is intentional behavior - we fall back to using the path when no remote exists
		//nolint: nilerr // Intentionally returning nil when no origin remote exists
		return resolvedRepo, resolvedRepo, nil
	}

	// Normalize the repository URL before checking status
	// This ensures consistent format (host/path) regardless of URL protocol (ssh://, git@, https://)
	normalizedRepoURL, err := c.normalizeRepositoryURL(rawRepoURL)
	if err != nil {
		// If normalization fails, fall back to using the path as the identifier
		c.VerbosePrint("  ⚠ Failed to normalize repository URL '%s': %v, using path as identifier", rawRepoURL, err)
		return resolvedRepo, resolvedRepo, nil
	}

	// Check if repository already exists in status using the normalized URL
	var finalRepoURL string
	if existingRepo, err := c.deps.StatusManager.GetRepository(normalizedRepoURL); err == nil && existingRepo != nil {
		finalRepoURL = normalizedRepoURL
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

// addDefaultBranchWorktreeIfNeeded adds the default branch worktree to status if needed.
func (c *realCodeManager) addDefaultBranchWorktreeIfNeeded(
	repoStatus *status.Repository,
	finalRepoURL string,
	branchesToCreate []string,
) {
	if repoStatus == nil || repoStatus.Remotes == nil {
		return
	}

	originRemote, ok := repoStatus.Remotes["origin"]
	if !ok {
		return
	}

	defaultBranch := originRemote.DefaultBranch
	if !contains(branchesToCreate, defaultBranch) {
		return
	}

	// Default branch is in the list - check if worktree already exists (cloned repo location)
	expectedWorktreePath := c.BuildWorktreePath(finalRepoURL, "origin", defaultBranch)
	if repoStatus.Path != expectedWorktreePath {
		return
	}

	// The repository is at the default branch worktree location
	// Add it to status if it doesn't already exist
	existingWorktree, getErr := c.deps.StatusManager.GetWorktree(finalRepoURL, defaultBranch)
	if getErr != nil || existingWorktree == nil {
		// Add default branch worktree to status
		if addErr := c.deps.StatusManager.AddWorktree(status.AddWorktreeParams{
			RepoURL:      finalRepoURL,
			Branch:       defaultBranch,
			WorktreePath: repoStatus.Path,
			Remote:       "origin",
			Detached:     false,
		}); addErr != nil {
			c.VerbosePrint("  Note: Error adding default branch worktree to status: %v", addErr)
		} else {
			c.VerbosePrint("  Added default branch worktree to status: %s", defaultBranch)
		}
	}
}

// contains checks if a string slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// createWorktreesForBranches creates worktrees for the given branches.
func (c *realCodeManager) createWorktreesForBranches(
	branchesToCreate []string,
	finalRepoURL, workspaceName, originalRepoPath string,
) error {
	// Get repository from status to check if we need to add default branch worktree
	repoStatus, err := c.deps.StatusManager.GetRepository(finalRepoURL)
	if err == nil && repoStatus != nil {
		c.addDefaultBranchWorktreeIfNeeded(repoStatus, finalRepoURL, branchesToCreate)
	}

	for _, branchName := range branchesToCreate {
		c.VerbosePrint("Creating worktree for branch '%s' in repository '%s'", branchName, finalRepoURL)

		// Create worktree in new repository
		worktreePath, actualRepoURL, err := c.createWorktreeForBranchInRepository(finalRepoURL, branchName, originalRepoPath)
		if err != nil {
			return fmt.Errorf("failed to create worktree for branch '%s' in repository '%s': %w", branchName, finalRepoURL, err)
		}

		// Skip workspace file update if worktree was not created (branch doesn't exist)
		if worktreePath == "" {
			c.VerbosePrint("  Skipping workspace file update for branch '%s' (branch doesn't exist in repository)", branchName)
			continue
		}

		// Update the .code-workspace file for that branch
		if err := c.updateWorkspaceFileForNewRepository(workspaceName, branchName, actualRepoURL); err != nil {
			return fmt.Errorf("failed to update workspace file for branch '%s': %w", branchName, err)
		}

		// Final verification: ensure the branch actually exists after worktree creation
		c.verifyAndCleanupWorktree(finalRepoURL, branchName)
	}
	return nil
}

// verifyAndCleanupWorktree verifies that a branch exists after worktree creation and removes it from status if not.
func (c *realCodeManager) verifyAndCleanupWorktree(repoURL, branchName string) {
	repoStatus, repoErr := c.deps.StatusManager.GetRepository(repoURL)
	if repoErr != nil || repoStatus == nil {
		return
	}

	mainRepoPath, getMainErr := c.deps.Git.GetMainRepositoryPath(repoStatus.Path)
	if getMainErr != nil {
		return
	}

	branchExists, checkErr := c.deps.Git.BranchExists(mainRepoPath, branchName)
	if checkErr != nil || branchExists {
		return
	}

	// Branch doesn't exist - check remote
	remoteExists, remoteErr := c.deps.Git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   mainRepoPath,
		RemoteName: "origin",
		Branch:     branchName,
	})
	if remoteErr != nil || remoteExists {
		return
	}

	// Branch doesn't exist on remote either - remove worktree from status
	existingWorktree, statusErr := c.deps.StatusManager.GetWorktree(repoURL, branchName)
	if statusErr == nil && existingWorktree != nil {
		c.VerbosePrint("  Final check: Branch '%s' does not exist, removing worktree from status", branchName)
		_ = c.deps.StatusManager.RemoveWorktree(repoURL, branchName)
	}
}

// updateAllWorkspaceFilesForNewRepository updates all existing workspace files for a workspace
// to include a new repository. It iterates through all branches in workspace.Worktrees and
// updates each corresponding workspace file.
func (c *realCodeManager) updateAllWorkspaceFilesForNewRepository(
	workspaceName, repoURL string, branches []string,
) error {
	c.VerbosePrint("Updating all workspace files for workspace '%s' with new repository '%s'", workspaceName, repoURL)

	for _, branchName := range branches {
		c.VerbosePrint("  Updating workspace file for branch '%s'", branchName)
		if err := c.updateWorkspaceFileForNewRepository(workspaceName, branchName, repoURL); err != nil {
			// Log error but continue with other branches
			// Some workspace files might not exist yet, which is handled gracefully in updateWorkspaceFileForNewRepository
			c.VerbosePrint("  ⚠ Failed to update workspace file for branch '%s': %v", branchName, err)
			// Only return error if it's not a "file doesn't exist" case
			// updateWorkspaceFileForNewRepository returns nil if file doesn't exist, so any error here is real
			return fmt.Errorf("failed to update workspace file for branch '%s': %w", branchName, err)
		}
	}

	c.VerbosePrint("  ✓ Updated all workspace files for workspace '%s'", workspaceName)
	return nil
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

// checkExistingWorktree checks if an existing worktree in status is valid and returns its path if so.
func (c *realCodeManager) checkExistingWorktree(
	repoURL, branchName, managedRepoPath string,
) (string, bool) {
	existingWorktree, err := c.deps.StatusManager.GetWorktree(repoURL, branchName)
	if err != nil || existingWorktree == nil {
		return "", false
	}

	// Worktree exists in status, check if directory actually exists
	worktreePath := c.BuildWorktreePath(repoURL, existingWorktree.Remote, branchName)
	exists, err := c.deps.FS.Exists(worktreePath)
	if err != nil || !exists {
		// Worktree exists in status but directory is missing - continue to create it
		c.VerbosePrint(
			"  Worktree exists in status but directory is missing, recreating worktree for branch '%s'",
			branchName,
		)
		return "", false
	}

	// Both status entry and directory exist
	// Verify the branch actually exists in the repository before using the worktree
	mainRepoPath, getMainErr := c.deps.Git.GetMainRepositoryPath(managedRepoPath)
	if getMainErr != nil {
		return "", false
	}

	branchExists, checkErr := c.deps.Git.BranchExists(mainRepoPath, branchName)
	if checkErr != nil || !branchExists {
		// Branch doesn't exist, don't use the worktree - it was incorrectly added
		c.VerbosePrint(
			"  Worktree exists in status for branch '%s' but branch doesn't exist in repository, removing from status",
			branchName,
		)
		_ = c.deps.StatusManager.RemoveWorktree(repoURL, branchName)
		return "", false
	}

	// Branch exists, use existing worktree
	c.VerbosePrint(
		"  Worktree already exists for branch '%s' in repository '%s', skipping creation",
		branchName, repoURL,
	)
	c.VerbosePrint("  Using existing worktree path: %s", worktreePath)
	return worktreePath, true
}

// logBranchStatus logs the status of a branch for worktree creation.
// Worktree creation will always proceed because the worktree package's EnsureBranchExists
// will handle creating the branch from remote (if it exists) or from default branch (if it doesn't).
func (c *realCodeManager) logBranchStatus(
	repoURL, branchName, mainRepoPath, originalRepoPath string,
) {
	// Check if branch exists in managed repository
	branchExists, checkErr := c.deps.Git.BranchExists(mainRepoPath, branchName)
	if checkErr == nil && branchExists {
		c.VerbosePrint("  Branch '%s' exists in repository '%s', proceeding with worktree creation", branchName, repoURL)
		return
	}

	// Check original repository if available (for local branches)
	if originalRepoPath != "" {
		originalBranchExists, originalCheckErr := c.deps.Git.BranchExists(originalRepoPath, branchName)
		if originalCheckErr == nil && originalBranchExists {
			c.VerbosePrint(
				"  Branch '%s' exists in original repository '%s' (local branch), proceeding with worktree creation",
				branchName, originalRepoPath,
			)
			return
		}
	}

	// Check remote if we didn't find it locally
	if checkErr == nil {
		remoteExists, remoteErr := c.deps.Git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
			RepoPath:   mainRepoPath,
			RemoteName: "origin",
			Branch:     branchName,
		})
		if remoteErr == nil && remoteExists {
			// Branch exists on remote but not fetched yet - let repository mode fetch and create it
			c.VerbosePrint("  Branch '%s' exists on remote but not fetched yet, repository mode will fetch it", branchName)
		} else {
			// Branch doesn't exist on remote - worktree creation will create it from default branch
			c.VerbosePrint(
				"  Branch '%s' does not exist on remote in repository '%s', worktree creation will create it from default branch",
				branchName, repoURL,
			)
		}
	} else {
		c.VerbosePrint(
			"  Warning: Failed to check if branch '%s' exists: %v, will attempt creation anyway",
			branchName, checkErr,
		)
	}
}

// shouldSkipWorktreeCreation checks if worktree creation should be skipped.
// Currently always returns false because worktree creation will handle creating
// branches that don't exist (from remote or default branch).
func (c *realCodeManager) shouldSkipWorktreeCreation(
	repoURL, branchName, mainRepoPath, originalRepoPath string,
) bool {
	c.logBranchStatus(repoURL, branchName, mainRepoPath, originalRepoPath)
	// Always attempt worktree creation - the worktree package's EnsureBranchExists
	// will handle creating the branch from remote (if it exists) or from default branch (if it doesn't)
	return false
}

// handleWorktreeCreationError handles errors from worktree creation.
func (c *realCodeManager) handleWorktreeCreationError(err error, repoURL, branchName string) (string, string, error) {
	// Check if error is because worktree already exists and handle it gracefully
	if existingPath := c.handleWorktreeExistsError(err, repoURL, branchName); existingPath != "" {
		return existingPath, repoURL, nil
	}

	// Check if error is because branch doesn't exist
	if c.isBranchNotFoundError(err) {
		// Branch doesn't exist - make sure no worktree was incorrectly added to status
		existingWorktree, statusErr := c.deps.StatusManager.GetWorktree(repoURL, branchName)
		if statusErr == nil && existingWorktree != nil {
			c.VerbosePrint("  Removing incorrectly added worktree for non-existent branch '%s' from status", branchName)
			_ = c.deps.StatusManager.RemoveWorktree(repoURL, branchName)
		}
		c.VerbosePrint("  Branch '%s' does not exist in repository '%s', skipping worktree creation", branchName, repoURL)
		return "", repoURL, nil
	}

	return "", "", fmt.Errorf("failed to create worktree in repository '%s': %w", repoURL, err)
}

// isBranchNotFoundError checks if an error indicates that a branch was not found.
func (c *realCodeManager) isBranchNotFoundError(err error) bool {
	errStr := strings.ToLower(err.Error())
	patterns := []string{
		"not found",
		"does not exist",
		"not found on remote",
		"could not resolve",
		"invalid reference",
		"no such ref",
	}
	for _, pattern := range patterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	// Check for branch-specific patterns
	if strings.Contains(errStr, "branch") {
		if strings.Contains(errStr, "not exist") || strings.Contains(errStr, "not found") {
			return true
		}
	}
	// Check for fatal branch errors
	if strings.Contains(errStr, "fatal:") && strings.Contains(errStr, "branch") {
		return true
	}
	return false
}

// createWorktreeForBranchInRepository creates a worktree for a specific branch in a repository.
func (c *realCodeManager) createWorktreeForBranchInRepository(
	repoURL, branchName, originalRepoPath string,
) (string, string, error) {
	// Get repository from status to get the actual managed path
	repoStatus, err := c.deps.StatusManager.GetRepository(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to get repository from status: %w", err)
	}

	managedRepoPath := repoStatus.Path

	// Check if worktree already exists in status and is valid
	if worktreePath, exists := c.checkExistingWorktree(repoURL, branchName, managedRepoPath); exists {
		return worktreePath, repoURL, nil
	}

	// Get the main repository path (in case managedRepoPath is a worktree)
	mainRepoPath, err := c.deps.Git.GetMainRepositoryPath(managedRepoPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to get main repository path from '%s': %w", managedRepoPath, err)
	}

	// Check if branch exists before trying to create worktree
	if c.shouldSkipWorktreeCreation(repoURL, branchName, mainRepoPath, originalRepoPath) {
		return "", repoURL, nil
	}

	// Get repository instance using the main repository path
	repoProvider := c.deps.RepositoryProvider
	repoInstance := repoProvider(repo.NewRepositoryParams{
		Dependencies:   c.deps,
		RepositoryName: mainRepoPath,
	})

	// Create worktree using repository mode
	worktreePath, err := repoInstance.CreateWorktree(branchName, repo.CreateWorktreeOpts{
		Remote: "origin",
	})
	if err != nil {
		return c.handleWorktreeCreationError(err, repoURL, branchName)
	}

	// Verify the branch exists after worktree creation
	c.verifyBranchExistsAfterCreation(repoURL, branchName, mainRepoPath)

	// Return the repoURL we already have (it's the canonical identifier)
	return worktreePath, repoURL, nil
}

// verifyBranchExistsAfterCreation verifies that a branch exists after worktree creation.
func (c *realCodeManager) verifyBranchExistsAfterCreation(repoURL, branchName, mainRepoPath string) {
	branchExistsAfter, checkErrAfter := c.deps.Git.BranchExists(mainRepoPath, branchName)
	if checkErrAfter != nil || branchExistsAfter {
		return
	}

	// Branch doesn't exist locally - check remote
	remoteExistsAfter, remoteErrAfter := c.deps.Git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   mainRepoPath,
		RemoteName: "origin",
		Branch:     branchName,
	})
	if remoteErrAfter != nil || remoteExistsAfter {
		c.VerbosePrint("  Branch '%s' exists on remote but not fetched, worktree is valid", branchName)
		return
	}

	// Branch doesn't exist on remote either - remove worktree from status
	existingWorktree, statusErr := c.deps.StatusManager.GetWorktree(repoURL, branchName)
	if statusErr == nil && existingWorktree != nil {
		c.VerbosePrint("  Branch '%s' does not exist after worktree creation, removing worktree from status", branchName)
		_ = c.deps.StatusManager.RemoveWorktree(repoURL, branchName)
	}
}

// handleWorktreeExistsError handles the case where a worktree creation error indicates
// the worktree already exists. Returns the existing worktree path if valid, empty string otherwise.
func (c *realCodeManager) handleWorktreeExistsError(err error, repoURL, branchName string) string {
	// Check if error is because worktree already exists
	// This could be from validation (wrong repo URL) or from Git (correct repo)
	worktreeExists := strings.Contains(err.Error(), "worktree already exists") ||
		strings.Contains(err.Error(), "already used by worktree")
	if !worktreeExists {
		return ""
	}

	// Check if worktree actually exists for the correct repository URL
	existingWorktree, checkErr := c.deps.StatusManager.GetWorktree(repoURL, branchName)
	if checkErr != nil || existingWorktree == nil {
		return ""
	}

	// Worktree exists in status for the correct repository
	worktreePath := c.BuildWorktreePath(repoURL, existingWorktree.Remote, branchName)
	exists, dirErr := c.deps.FS.Exists(worktreePath)
	if dirErr != nil || !exists {
		return ""
	}

	c.VerbosePrint(
		"  Worktree already exists for branch '%s' in repository '%s', skipping creation",
		branchName, repoURL,
	)
	c.VerbosePrint("  Using existing worktree path: %s", worktreePath)
	return worktreePath
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
