// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/status"
)

// createWorktreesForWorkspace creates worktrees for all repositories in the workspace.
func (w *realWorkspace) createWorktreesForWorkspace(branch string, opts *CreateWorktreeOpts) error {
	w.verboseLogf("Creating worktrees for all repositories in workspace")

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace name for worktree-specific workspace file
	workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
	workspaceDir := filepath.Dir(w.OriginalFile)

	// Track created worktrees for cleanup on failure
	var createdWorktrees []struct {
		repoURL string
		branch  string
		path    string
	}

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
		createdWorktrees,
		worktreeWorkspacePath,
		opts,
	); err != nil {
		// Cleanup workspace file on failure
		if cleanupErr := w.fs.RemoveAll(worktreeWorkspacePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree workspace file: %v", cleanupErr)
		}
		return err
	}

	return nil
}

// createWorktreeDirectories creates worktree directories and executes Git worktree commands.
func (w *realWorkspace) createWorktreeDirectories(
	workspaceConfig *Config,
	workspaceDir string,
	branch string,
	createdWorktrees []struct {
		repoURL string
		branch  string
		path    string
	},
	worktreeWorkspacePath string,
	_ *CreateWorktreeOpts,
) error {
	for i, folder := range workspaceConfig.Folders {
		w.verboseLogf("Creating worktree %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		if err := w.createSingleWorktree(createSingleWorktreeParams{
			Folder:                folder,
			WorkspaceDir:          workspaceDir,
			Branch:                branch,
			CreatedWorktrees:      createdWorktrees,
			WorktreeWorkspacePath: worktreeWorkspacePath,
			Opts:                  nil,
		}); err != nil {
			return err
		}

		w.verboseLogf("✓ Worktree created successfully for %s", folder.Path)
	}

	return nil
}

// createSingleWorktreeParams contains parameters for creating a single worktree.
type createSingleWorktreeParams struct {
	Folder           Folder
	WorkspaceDir     string
	Branch           string
	CreatedWorktrees []struct {
		repoURL string
		branch  string
		path    string
	}
	WorktreeWorkspacePath string
	Opts                  *CreateWorktreeOpts
}

// createSingleWorktree creates a single worktree for a folder.
func (w *realWorkspace) createSingleWorktree(params createSingleWorktreeParams) error {
	resolvedPath := filepath.Join(params.WorkspaceDir, params.Folder.Path)
	repoURL, err := w.git.GetRepositoryName(resolvedPath)
	if err != nil {
		w.cleanupOnFailure(params.CreatedWorktrees, params.WorktreeWorkspacePath)
		return fmt.Errorf("failed to get repository URL for %s: %w", params.Folder.Path, err)
	}

	defaultRemote := w.getDefaultRemote(repoURL)
	worktreePath := w.buildWorktreePath(repoURL, defaultRemote, params.Branch)

	// Ensure branch exists
	if err := w.ensureBranchExists(ensureBranchExistsParams{
		ResolvedPath:          resolvedPath,
		Branch:                params.Branch,
		FolderPath:            params.Folder.Path,
		CreatedWorktrees:      params.CreatedWorktrees,
		WorktreeWorkspacePath: params.WorktreeWorkspacePath,
	}); err != nil {
		return err
	}

	// Create worktree
	if err := w.createWorktreeDirectory(
		resolvedPath, worktreePath, params.Branch,
		params.CreatedWorktrees, params.WorktreeWorkspacePath,
	); err != nil {
		return err
	}

	// Add worktree to status file
	if err := w.addWorktreeToStatusWithCleanup(
		repoURL, params.Branch, worktreePath, defaultRemote, resolvedPath,
	); err != nil {
		return err
	}

	return nil
}

// createWorktreeDirectory creates the worktree directory and Git worktree.
func (w *realWorkspace) createWorktreeDirectory(resolvedPath, worktreePath, branch string, createdWorktrees []struct {
	repoURL string
	branch  string
	path    string
}, worktreeWorkspacePath string) error {
	// Create worktree directory
	if err := w.fs.MkdirAll(worktreePath, 0755); err != nil {
		w.cleanupOnFailure(createdWorktrees, worktreeWorkspacePath)
		return fmt.Errorf("failed to create worktree directory %s: %w", worktreePath, err)
	}

	// Execute Git worktree creation command
	if err := w.git.CreateWorktree(resolvedPath, worktreePath, branch); err != nil {
		w.cleanupOnFailure(createdWorktrees, worktreeWorkspacePath)
		if cleanupErr := w.cleanupWorktreeDirectory(worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	return nil
}

// addWorktreeToStatusWithCleanup adds worktree to status with cleanup on failure.
func (w *realWorkspace) addWorktreeToStatusWithCleanup(
	repoURL, branch, worktreePath, defaultRemote, resolvedPath string,
) error {
	if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		Remote:       defaultRemote,
	}); err != nil {
		// Clean up worktree on failure
		if cleanupErr := w.git.RemoveWorktree(resolvedPath, worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up Git worktree: %v", cleanupErr)
		}
		if cleanupErr := w.fs.RemoveAll(worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}
	return nil
}

// ensureBranchExistsParams contains parameters for ensuring a branch exists.
type ensureBranchExistsParams struct {
	ResolvedPath     string
	Branch           string
	FolderPath       string
	CreatedWorktrees []struct {
		repoURL string
		branch  string
		path    string
	}
	WorktreeWorkspacePath string
}

// ensureBranchExists ensures that the specified branch exists in the repository.
func (w *realWorkspace) ensureBranchExists(params ensureBranchExistsParams) error {
	// Check if branch exists
	exists, err := w.git.BranchExists(params.ResolvedPath, params.Branch)
	if err != nil {
		// Cleanup on failure
		w.cleanupFailedWorktrees(params.CreatedWorktrees)
		w.cleanupWorktreeWorkspaceFile(params.WorktreeWorkspacePath)
		return fmt.Errorf("failed to check branch existence for %s: %w", params.FolderPath, err)
	}

	if !exists {
		w.verboseLogf("Branch %s does not exist in %s, creating from current branch", params.Branch, params.FolderPath)
		if err := w.git.CreateBranch(params.ResolvedPath, params.Branch); err != nil {
			// Cleanup on failure
			w.cleanupFailedWorktrees(params.CreatedWorktrees)
			w.cleanupWorktreeWorkspaceFile(params.WorktreeWorkspacePath)
			return fmt.Errorf("failed to create branch %s for %s: %w", params.Branch, params.FolderPath, err)
		}
	}

	return nil
}

// createDefaultBranchWorktree creates a worktree for the default branch of a repository.
func (w *realWorkspace) createDefaultBranchWorktree(repoURL, remoteName, branch, repoPath string) error {
	w.verboseLogf("Creating default branch worktree: %s:%s for repository %s", remoteName, branch, repoURL)

	// Generate worktree path using new structure
	worktreePath := w.buildWorktreePath(repoURL, remoteName, branch)

	// Check if worktree directory already exists
	if err := w.createWorktreeIfNotExists(repoPath, worktreePath, branch); err != nil {
		return err
	}

	// Add worktree to status file
	if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		Remote:       remoteName,
	}); err != nil {
		// Clean up worktree on failure
		if cleanupErr := w.git.RemoveWorktree(repoPath, worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up Git worktree: %v", cleanupErr)
		}
		if cleanupErr := w.fs.RemoveAll(worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	return nil
}

// createWorktreeIfNotExists creates a worktree if it doesn't already exist.
func (w *realWorkspace) createWorktreeIfNotExists(repoPath, worktreePath, branch string) error {
	exists, err := w.fs.Exists(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to check worktree directory existence: %w", err)
	}

	if exists {
		w.verboseLogf("Worktree directory already exists: %s", worktreePath)
		return nil
	}

	// Create worktree directory
	if err := w.fs.MkdirAll(worktreePath, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Create Git worktree
	if err := w.git.CreateWorktree(repoPath, worktreePath, branch); err != nil {
		// Clean up directory on failure
		if cleanupErr := w.fs.RemoveAll(worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	return nil
}
