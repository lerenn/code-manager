package repository

import (
	"fmt"
	"path/filepath"
	"strings"
)

// IsGitRepository checks if the current directory is a Git repository (including worktrees).
func (r *realRepository) IsGitRepository() (bool, error) {
	r.deps.Logger.Logf("Checking if directory %s is a Git repository...", r.repositoryPath)

	// Check if .git exists
	gitPath := filepath.Join(r.repositoryPath, ".git")
	exists, err := r.deps.FS.Exists(gitPath)
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		r.deps.Logger.Logf("No .git found")
		return false, nil
	}

	// Check if .git is a directory (regular repository)
	isDir, err := r.deps.FS.IsDir(gitPath)
	if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	if isDir {
		r.deps.Logger.Logf("Git repository detected (.git directory)")
		return true, nil
	}

	// If .git is not a directory, it must be a file (worktree)
	// Validate that it's actually a Git worktree file by checking for 'gitdir:' prefix
	r.deps.Logger.Logf("Checking if .git file is a valid worktree file...")

	content, err := r.deps.FS.ReadFile(gitPath)
	if err != nil {
		r.deps.Logger.Logf("Failed to read .git file: %v", err)
		return false, nil
	}

	contentStr := strings.TrimSpace(string(content))
	if !strings.HasPrefix(contentStr, "gitdir:") {
		r.deps.Logger.Logf(".git file exists but is not a valid worktree file (missing 'gitdir:' prefix)")
		return false, nil
	}

	r.deps.Logger.Logf("Git worktree detected (.git file)")
	return true, nil
}
