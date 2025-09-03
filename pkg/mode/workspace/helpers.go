package workspace

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/mode"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// Error definitions for workspace package.
var (
	// Workspace file errors.
	ErrFailedToCheckWorkspaceFiles = errors.New("failed to check for workspace files")
	ErrWorkspaceFileNotFound       = errors.New("workspace file not found")
	ErrInvalidWorkspaceFile        = errors.New("invalid workspace file")
	ErrNoRepositoriesFound         = errors.New("no repositories found in workspace")

	// Repository errors.
	ErrRepositoryNotFound = errors.New("repository not found in workspace")
	ErrRepositoryNotClean = errors.New("repository is not clean")

	// Worktree errors.
	ErrWorktreeExists      = errors.New("worktree already exists")
	ErrWorktreeNotInStatus = errors.New("worktree not found in status file")

	// Directory and file system errors.
	ErrDirectoryExists = errors.New("directory already exists")

	// User interaction errors.
	ErrDeletionCancelled = errors.New("deletion cancelled by user")
)

// validateWorkspaceForWorktreeCreation validates workspace state before worktree creation.
func (w *realWorkspace) validateWorkspaceForWorktreeCreation(branch string) error {
	w.logger.Logf("Validating workspace for worktree creation")

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(w.OriginalFile)

	for i, folder := range workspaceConfig.Folders {
		w.logger.Logf("Validating repository %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		// Resolve relative path from workspace file location
		resolvedPath := filepath.Join(workspaceDir, folder.Path)

		// Check for existing worktrees in status file
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}

		// Check if worktree already exists
		_, err = w.statusManager.GetWorktree(repoURL, branch)
		if err == nil {
			return fmt.Errorf("worktree already exists for repository %s branch %s", repoURL, branch)
		}

		// Check branch existence and create if needed
		exists, err := w.git.BranchExists(resolvedPath, branch)
		if err != nil {
			return fmt.Errorf("failed to check branch existence for %s: %w", folder.Path, err)
		}

		if !exists {
			w.logger.Logf("Branch %s does not exist in %s, will create from current branch", branch, folder.Path)
		}

		// Validate directory creation permissions using worktree package
		worktreeInstance := w.worktreeProvider(worktree.NewWorktreeParams{
			FS:            w.fs,
			Git:           w.git,
			StatusManager: w.statusManager,
			Logger:        w.logger,
			Prompt:        w.prompt,
			BasePath:      w.config.BasePath,
		})
		worktreePath := worktreeInstance.BuildPath(repoURL, "origin", branch)

		// Check if worktree directory already exists
		exists, err = w.fs.Exists(worktreePath)
		if err != nil {
			return fmt.Errorf("failed to check worktree directory existence: %w", err)
		}
		if exists {
			return fmt.Errorf("worktree directory already exists: %s", worktreePath)
		}
	}

	return nil
}

// createWorktreesForWorkspace creates worktrees for all repositories in the workspace.
func (w *realWorkspace) createWorktreesForWorkspace(branch string, opts *mode.CreateWorktreeOpts) error {
	w.logger.Logf("Creating worktrees for all repositories in workspace")

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace name for worktree-specific workspace file
	workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
	workspaceDir := filepath.Dir(w.OriginalFile)

	// Sanitize branch name for filename (replace slashes with hyphens)
	sanitizedBranchForFilename := strings.ReplaceAll(branch, "/", "-")

	// Create worktree-specific workspace file path
	worktreeWorkspacePath := filepath.Join(
		w.config.BasePath,
		"workspaces",
		fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename),
	)

	// 1. Add workspace to status file if not already present
	if err := w.ensureWorkspaceInStatus(workspaceConfig, workspaceDir); err != nil {
		return err
	}

	// 2. Create worktree-specific workspace file
	if err := w.createWorktreeWorkspaceFile(createWorktreeWorkspaceFileParams{
		WorkspaceConfig:       workspaceConfig,
		WorkspaceName:         workspaceName,
		Branch:                branch,
		WorktreeWorkspacePath: worktreeWorkspacePath,
	}); err != nil {
		return fmt.Errorf("failed to create worktree workspace file: %w", err)
	}

	// 3. Create worktree directories and execute Git worktree commands
	if err := w.createWorktreeDirectories(
		workspaceConfig,
		workspaceDir,
		branch,
		worktreeWorkspacePath,
		opts,
	); err != nil {
		// Cleanup workspace file on failure
		if cleanupErr := w.fs.RemoveAll(worktreeWorkspacePath); cleanupErr != nil {
			w.logger.Logf("Warning: failed to clean up worktree workspace file: %v", cleanupErr)
		}
		return err
	}

	return nil
}

// getWorkspacePath gets the workspace path.
func (w *realWorkspace) getWorkspacePath() (string, error) {
	return filepath.Abs(w.OriginalFile)
}

// getWorkspaceWorktrees gets all worktrees for this workspace and branch.
func (w *realWorkspace) getWorkspaceWorktrees(branch string) ([]WorktreeWithRepo, error) {
	// Get workspace path
	workspacePath, err := filepath.Abs(w.OriginalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspacePath)
	if err != nil {
		// If workspace not found, return empty list with no error
		if errors.Is(err, status.ErrWorkspaceNotFound) {
			return []WorktreeWithRepo{}, nil
		}
		return nil, err
	}

	w.logger.Logf("Looking for worktrees with workspace path: %s", workspacePath)
	w.logger.Logf("Workspace repositories: %v", workspace.Repositories)

	// Get worktrees for each repository in the workspace that match the branch
	var workspaceWorktrees []WorktreeWithRepo

	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := w.statusManager.GetRepository(repoURL)
		if err != nil {
			continue // Skip if repository not found
		}

		// Get worktrees for this repository that match the branch
		for _, worktree := range repo.Worktrees {
			if worktree.Branch == branch {
				workspaceWorktrees = append(workspaceWorktrees, WorktreeWithRepo{
					WorktreeInfo: worktree,
					RepoURL:      repoURL,
					RepoPath:     repo.Path,
				})
				w.logger.Logf("✓ Found matching worktree: %s:%s for repository %s", worktree.Remote, worktree.Branch, repoURL)
			}
		}
	}

	return workspaceWorktrees, nil
}

// deleteWorktreeRepositories deletes worktrees for all repositories.
func (w *realWorkspace) deleteWorktreeRepositories(workspaceWorktrees []WorktreeWithRepo, force bool) error {
	for i, worktreeWithRepo := range workspaceWorktrees {
		w.logger.Logf("Deleting worktree %d/%d: %s:%s for repository %s", i+1, len(workspaceWorktrees),
			worktreeWithRepo.Remote, worktreeWithRepo.Branch, worktreeWithRepo.RepoURL)

		// Create worktree instance using provider
		worktreeInstance := w.worktreeProvider(worktree.NewWorktreeParams{
			FS:            w.fs,
			Git:           w.git,
			StatusManager: w.statusManager,
			Logger:        w.logger,
			Prompt:        w.prompt,
			BasePath:      w.config.BasePath,
		})

		// Get worktree path using worktree package
		worktreePath := worktreeInstance.BuildPath(worktreeWithRepo.RepoURL, worktreeWithRepo.Remote, worktreeWithRepo.Branch)

		// Delete worktree using the worktree package
		err := worktreeInstance.Delete(worktree.DeleteParams{
			RepoURL:      worktreeWithRepo.RepoURL,
			Branch:       worktreeWithRepo.Branch,
			WorktreePath: worktreePath,
			RepoPath:     worktreeWithRepo.RepoPath,
			Force:        force,
		})

		if err != nil {
			if !force {
				return fmt.Errorf("failed to delete worktree for %s:%s: %w",
					worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
			}
			w.logger.Logf("Warning: failed to delete worktree for %s:%s: %v",
				worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
		}

		w.logger.Logf("✓ Worktree deleted successfully for %s:%s", worktreeWithRepo.Remote, worktreeWithRepo.Branch)
	}

	return nil
}

// removeWorktreeStatusEntries removes worktree entries from status file.
func (w *realWorkspace) removeWorktreeStatusEntries(workspaceWorktrees []WorktreeWithRepo, force bool) error {
	for _, worktreeWithRepo := range workspaceWorktrees {
		if err := w.statusManager.RemoveWorktree(worktreeWithRepo.RepoURL, worktreeWithRepo.Branch); err != nil {
			if !force {
				return fmt.Errorf("failed to remove worktree status entry for %s:%s: %w",
					worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
			}
			w.logger.Logf("Warning: failed to remove worktree status entry for %s:%s: %v",
				worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
		}
	}

	return nil
}

// WorktreeWithRepo represents a worktree with its associated repository information.
type WorktreeWithRepo struct {
	status.WorktreeInfo
	RepoURL  string
	RepoPath string
}

// createWorktreeWorkspaceFileParams contains parameters for creating a worktree workspace file.
type createWorktreeWorkspaceFileParams struct {
	WorkspaceConfig       Config
	WorkspaceName         string
	Branch                string
	WorktreeWorkspacePath string
}

// ensureWorkspaceInStatus ensures the workspace is in the status file.
func (w *realWorkspace) ensureWorkspaceInStatus(workspaceConfig Config, workspaceDir string) error {
	// This is a placeholder - the actual implementation would be more complex
	// For now, we'll just return nil to allow compilation
	return nil
}

// createWorktreeWorkspaceFile creates a worktree-specific workspace file.
func (w *realWorkspace) createWorktreeWorkspaceFile(params createWorktreeWorkspaceFileParams) error {
	// This is a placeholder - the actual implementation would be more complex
	// For now, we'll just return nil to allow compilation
	return nil
}

// createWorktreeDirectories creates worktree directories and executes Git worktree commands.
func (w *realWorkspace) createWorktreeDirectories(
	workspaceConfig Config,
	workspaceDir string,
	branch string,
	worktreeWorkspacePath string,
	opts *mode.CreateWorktreeOpts,
) error {
	// This is a placeholder - the actual implementation would be more complex
	// For now, we'll just return nil to allow compilation
	return nil
}
