package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetMainRepositoryPath gets the main repository path from a worktree path.
// If the path is already a main repository, it returns the same path.
// If the path is a worktree, it returns the main repository path.
func (g *realGit) GetMainRepositoryPath(worktreePath string) (string, error) {
	// Use git rev-parse --git-common-dir to get the main repository's .git directory
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"git rev-parse --git-common-dir failed: %w (command: git rev-parse --git-common-dir, output: %s)",
			err, string(output),
		)
	}

	gitCommonDir := strings.TrimSpace(string(output))
	if gitCommonDir == "" {
		return "", fmt.Errorf("git rev-parse --git-common-dir returned empty output")
	}

	// Convert to absolute path, resolving relative to worktreePath
	var absGitCommonDir string
	if filepath.IsAbs(gitCommonDir) {
		absGitCommonDir = gitCommonDir
	} else {
		// Resolve relative path from worktreePath
		absWorktreePath, err := filepath.Abs(worktreePath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for worktree: %w", err)
		}
		absGitCommonDir = filepath.Join(absWorktreePath, gitCommonDir)
		// Clean the path to resolve any ".." components
		absGitCommonDir = filepath.Clean(absGitCommonDir)
	}

	// The main repository path is the parent of .git directory
	mainRepoPath := filepath.Dir(absGitCommonDir)

	return mainRepoPath, nil
}
