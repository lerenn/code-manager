// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"
	"path/filepath"
)

// ValidateWorkspaceReferences validates that workspace references point to existing worktrees and repositories.
func (w *realWorkspace) ValidateWorkspaceReferences() error {
	w.verboseLogf("Validating workspace references")

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(w.OriginalFile)

	// Validate each repository in the workspace
	for _, folder := range workspaceConfig.Folders {
		if err := w.validateWorkspaceRepositoryReference(folder, workspaceDir); err != nil {
			return err
		}
	}

	return nil
}

// validateRepositoryPath validates that the repository path exists.
func (w *realWorkspace) validateRepositoryPath(folder Folder, resolvedPath string) error {
	// Check repository path exists
	exists, err := w.fs.Exists(resolvedPath)
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("repository not found in workspace: %s - %w", folder.Path, err)
	}

	if !exists {
		w.verboseLogf("Error: repository path does not exist")
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, folder.Path)
	}

	return nil
}

// validateRepositoryGit validates that the repository has a .git directory and git status works.
func (w *realWorkspace) validateRepositoryGit(folder Folder, resolvedPath string) error {
	// Verify path contains .git folder
	gitPath := filepath.Join(resolvedPath, ".git")
	exists, err := w.fs.Exists(gitPath)
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrRepositoryNotFound, folder.Path, err)
	}

	if !exists {
		w.verboseLogf("Error: .git directory not found in repository")
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, folder.Path)
	}

	// Execute git status to ensure repository is working
	w.verboseLogf("Executing git status in: %s", resolvedPath)
	_, err = w.git.Status(resolvedPath)
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrRepositoryNotFound, folder.Path, err)
	}

	return nil
}

// validateWorkspaceRepositoryReference validates a single repository reference in a workspace.
func (w *realWorkspace) validateWorkspaceRepositoryReference(folder Folder, workspaceDir string) error {
	w.verboseLogf("Validating workspace repository reference: %s", folder.Path)

	// Resolve relative path from workspace file location
	resolvedPath := filepath.Join(workspaceDir, folder.Path)

	// Validate repository exists and is a Git repository
	if err := w.validateRepositoryPath(folder, resolvedPath); err != nil {
		return err
	}

	if err := w.validateRepositoryGit(folder, resolvedPath); err != nil {
		return err
	}

	// Get repository URL
	repoURL, err := w.git.GetRepositoryName(resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
	}

	// Check if repository exists in status file
	repo, err := w.statusManager.GetRepository(repoURL)
	if err != nil {
		// Repository not in status, we need to add it
		w.verboseLogf("Repository %s not found in status, will add it", repoURL)
		if err := w.addRepositoryToStatus(repoURL, resolvedPath); err != nil {
			return fmt.Errorf("failed to add repository %s to status: %w", repoURL, err)
		}
		repo, err = w.statusManager.GetRepository(repoURL)
		if err != nil {
			return fmt.Errorf("failed to get repository %s after adding to status: %w", repoURL, err)
		}
	}

	// Check if repository has a default branch worktree
	if err := w.ensureRepositoryHasDefaultBranchWorktree(repoURL, repo); err != nil {
		return err
	}

	return nil
}

// validateWorkspaceForWorktreeCreation validates workspace state before worktree creation.
func (w *realWorkspace) validateWorkspaceForWorktreeCreation(branch string) error {
	w.verboseLogf("Validating workspace for worktree creation")

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(w.OriginalFile)

	for i, folder := range workspaceConfig.Folders {
		w.verboseLogf("Validating repository %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

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
			w.verboseLogf("Branch %s does not exist in %s, will create from current branch", branch, folder.Path)
		}

		// Validate directory creation permissions
		worktreePath := w.buildWorktreePath(repoURL, "origin", branch)

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
