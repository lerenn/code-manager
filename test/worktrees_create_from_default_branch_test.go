//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorktreeFromDefaultBranch(t *testing.T) {
	// Create temporary test environment
	tempDir, err := os.MkdirTemp("", "cm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create CM instance with temporary config
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			RepositoriesDir: tempDir,
			StatusFile:      filepath.Join(tempDir, "status.yaml"),
		},
	})

	require.NoError(t, err)
	// Create a test repository
	repoPath := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	// Initialize git repository
	runGitCommand(t, repoPath, "init")
	runGitCommand(t, repoPath, "config", "user.name", "Test User")
	runGitCommand(t, repoPath, "config", "user.email", "test@example.com")

	// Create initial commit on main branch
	runGitCommand(t, repoPath, "checkout", "-b", "main")
	runGitCommand(t, repoPath, "commit", "--allow-empty", "-m", "Initial commit")

	// Create a feature branch and make some changes
	runGitCommand(t, repoPath, "checkout", "-b", "feature-branch")
	runGitCommand(t, repoPath, "commit", "--allow-empty", "-m", "Feature commit")

	// Switch back to main and make changes there too
	runGitCommand(t, repoPath, "checkout", "main")
	runGitCommand(t, repoPath, "commit", "--allow-empty", "-m", "Main branch update")

	// Add remote (use a real public repository)
	runGitCommand(t, repoPath, "remote", "add", "origin", "https://github.com/octocat/Hello-World.git")

	// Fetch from remote to get the remote branches
	runGitCommand(t, repoPath, "fetch", "origin")

	// Create a local master branch that tracks origin/master
	runGitCommand(t, repoPath, "checkout", "-b", "master", "origin/master")

	// Change to the repository directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	require.NoError(t, os.Chdir(repoPath))

	// Create a worktree for a new branch
	// This should now create the branch from origin/main (default branch) instead of current branch
	newBranchName := "new-feature-branch"
	err = cmInstance.CreateWorkTree(newBranchName)
	require.NoError(t, err)

	// Verify that the new branch was created from the remote default branch (master)
	// Check the commit history of the new branch
	output := runGitCommand(t, repoPath, "log", "--oneline", newBranchName)
	masterOutput := runGitCommand(t, repoPath, "log", "--oneline", "master")
	mainOutput := runGitCommand(t, repoPath, "log", "--oneline", "main")
	featureOutput := runGitCommand(t, repoPath, "log", "--oneline", "feature-branch")

	// The new branch should have the same commits as the remote master branch
	assert.Equal(t, strings.TrimSpace(masterOutput), strings.TrimSpace(output),
		"New branch should be based on remote default branch (master), not current branch")

	// Verify that the new branch does NOT have the same commits as the local main branch
	assert.NotEqual(t, strings.TrimSpace(mainOutput), strings.TrimSpace(output),
		"New branch should not be based on local main branch")

	// Verify that the new branch does NOT have the feature-branch commit
	assert.NotEqual(t, strings.TrimSpace(featureOutput), strings.TrimSpace(output),
		"New branch should not be based on feature-branch")

	// Verify that the worktree was created and added to status
	worktrees, _, err := cmInstance.ListWorktrees(false)
	require.NoError(t, err)

	found := false
	for _, wt := range worktrees {
		if wt.Branch == newBranchName {
			found = true
			break
		}
	}
	assert.True(t, found, "New worktree should be listed in status")
}

func runGitCommand(t *testing.T, repoPath string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	require.NoError(t, err, "Git command failed: %s", strings.Join(args, " "))
	return string(output)
}
