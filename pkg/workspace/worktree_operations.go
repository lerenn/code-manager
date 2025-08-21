// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"errors"

	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
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
	worktreeWorkspacePath string,
	_ *CreateWorktreeOpts,
) error {
	for i, folder := range workspaceConfig.Folders {
		w.verboseLogf("Creating worktree %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		if err := w.createSingleWorktree(createSingleWorktreeParams{
			Folder:                folder,
			WorkspaceDir:          workspaceDir,
			Branch:                branch,
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
	Folder                Folder
	WorkspaceDir          string
	Branch                string
	WorktreeWorkspacePath string
	Opts                  *CreateWorktreeOpts
}

// createSingleWorktree creates a single worktree for a folder.
func (w *realWorkspace) createSingleWorktree(params createSingleWorktreeParams) error {
	resolvedPath := filepath.Join(params.WorkspaceDir, params.Folder.Path)
	repoURL, err := w.git.GetRepositoryName(resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to get repository URL for %s: %w", params.Folder.Path, err)
	}

	defaultRemote := w.getDefaultRemote(repoURL)
	worktreePath := w.worktree.BuildPath(repoURL, defaultRemote, params.Branch)

	// Create worktree using the worktree package
	err = w.worktree.Create(worktree.CreateParams{
		RepoURL:      repoURL,
		Branch:       params.Branch,
		WorktreePath: worktreePath,
		RepoPath:     resolvedPath,
		Remote:       defaultRemote,
		IssueInfo:    nil,
		Force:        false,
	})

	if err != nil {
		// Cleanup workspace file on failure
		if cleanupErr := w.fs.RemoveAll(params.WorktreeWorkspacePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree workspace file: %v", cleanupErr)
		}
		return err
	}

	// Add to status file with auto-repository handling
	if err := w.addWorktreeToStatus(
		repoURL, params.Branch, worktreePath, defaultRemote,
		params.WorktreeWorkspacePath, params.Folder.Path,
	); err != nil {
		// Clean up worktree on status failure
		if cleanupErr := w.worktree.CleanupDirectory(worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory after status failure: %v", cleanupErr)
		}
		return err
	}

	return nil
}

// addWorktreeToStatus adds a worktree to the status file with auto-repository handling.
func (w *realWorkspace) addWorktreeToStatus(
	repoURL, branch, worktreePath, remote, workspacePath, folderPath string,
) error {
	// Try to add worktree to status
	addParams := worktree.AddToStatusParams{
		RepoURL:       repoURL,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: workspacePath,
		Remote:        remote,
		IssueInfo:     nil,
	}

	err := w.worktree.AddToStatus(addParams)
	if err == nil {
		return nil
	}

	// If repository not found, try to auto-add it
	if !errors.Is(err, status.ErrRepositoryNotFound) {
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	repoPath := filepath.Join(filepath.Dir(w.OriginalFile), folderPath)
	if addErr := w.autoAddRepositoryToStatus(repoURL, repoPath); addErr != nil {
		return fmt.Errorf("failed to auto-add repository to status: %w", addErr)
	}

	// Try adding the worktree again
	if err := w.worktree.AddToStatus(addParams); err != nil {
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	return nil
}

// autoAddRepositoryToStatus automatically adds a repository to the status file.
func (w *realWorkspace) autoAddRepositoryToStatus(repoURL, repoPath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if it's a Git repository
	exists, err := w.fs.Exists(filepath.Join(absPath, ".git"))
	if err != nil {
		return fmt.Errorf("failed to check .git existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("not a Git repository: .git directory not found")
	}

	// Get remotes information
	remotes := make(map[string]status.Remote)

	// Check for origin remote
	originURL, err := w.git.GetRemoteURL(absPath, "origin")
	if err == nil && originURL != "" {
		remotes["origin"] = status.Remote{
			DefaultBranch: "main", // Default to main, could be enhanced to detect actual default branch
		}
	}

	// Add the repository to status
	if err := w.statusManager.AddRepository(repoURL, status.AddRepositoryParams{
		Path:    absPath,
		Remotes: remotes,
	}); err != nil {
		return fmt.Errorf("failed to add repository to status: %w", err)
	}

	return nil
}

// createDefaultBranchWorktree creates a worktree for the default branch of a repository.
func (w *realWorkspace) createDefaultBranchWorktree(repoURL, remoteName, branch, repoPath string) error {
	w.verboseLogf("Creating default branch worktree: %s:%s for repository %s", remoteName, branch, repoURL)

	// Generate worktree path using worktree package
	worktreePath := w.worktree.BuildPath(repoURL, remoteName, branch)

	// Create worktree using the worktree package
	err := w.worktree.Create(worktree.CreateParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		RepoPath:     repoPath,
		Remote:       remoteName,
		IssueInfo:    nil,
		Force:        false,
	})

	if err != nil {
		return err
	}

	// Add to status file with auto-repository handling
	if err := w.addWorktreeToStatus(repoURL, branch, worktreePath, remoteName, "", repoPath); err != nil {
		// Clean up worktree on status failure
		if cleanupErr := w.worktree.CleanupDirectory(worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory after status failure: %v", cleanupErr)
		}
		return err
	}

	return nil
}
