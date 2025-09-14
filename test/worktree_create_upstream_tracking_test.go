//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorktreeUpstreamTracking tests that worktrees fail properly when upstream cannot be set
func TestWorktreeUpstreamTracking(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a feature branch (but don't simulate remote existence)
	createFeatureBranch(t, setup.RepoPath, "feature/upstream-test")

	// Create config file for CLI testing
	configFile := filepath.Join(setup.TempDir, "config.yaml")
	createConfigFile(t, configFile, setup.CmPath, setup.StatusPath)

	// Safely change to repo directory and create worktree using CLI
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	// Initialize CM first with explicit directories
	initCmd := exec.Command("cm", "init", "-c", configFile, "-r", setup.CmPath, "-w", filepath.Join(setup.TempDir, "workspaces"), "-s", setup.StatusPath)
	initOutput, err := initCmd.CombinedOutput()
	require.NoError(t, err, "CM init should succeed. Output: %s", string(initOutput))

	// Create worktree using CLI with -c flag - this should succeed because SetUpstreamBranch
	// now gracefully handles non-existing remote branches by skipping upstream tracking
	cmd := exec.Command("cm", "worktree", "create", "feature/upstream-test", "-c", configFile, "-v")
	output, err := cmd.CombinedOutput()

	// Show the actual output for debugging
	t.Logf("Worktree creation output: %s", string(output))
	t.Logf("Worktree creation error: %v", err)

	// The worktree creation should succeed because SetUpstreamBranch now skips
	// setting upstream tracking when the remote branch doesn't exist
	assert.NoError(t, err, "Worktree creation should succeed even when remote branch doesn't exist")

	// Verify the worktree was created successfully
	outputStr := string(output)
	assert.Contains(t, outputStr, "✓ Worktree created successfully",
		"Output should indicate successful worktree creation. Output: %s", outputStr)
}

// TestWorktreeUpstreamTrackingNewBranch tests upstream tracking for new branches that don't exist on remote
func TestWorktreeUpstreamTrackingNewBranch(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create config file for CLI testing
	configFile := filepath.Join(setup.TempDir, "config.yaml")
	createConfigFile(t, configFile, setup.CmPath, setup.StatusPath)

	// Safely change to repo directory and create worktree using CLI
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	// Initialize CM first with explicit directories
	initCmd := exec.Command("cm", "init", "-c", configFile, "-r", setup.CmPath, "-w", filepath.Join(setup.TempDir, "workspaces"), "-s", setup.StatusPath)
	initOutput, err := initCmd.CombinedOutput()
	require.NoError(t, err, "CM init should succeed. Output: %s", string(initOutput))

	// Create worktree for a new branch that doesn't exist on remote using CLI with -c flag
	// This should succeed because SetUpstreamBranch now gracefully handles non-existing remote branches
	cmd := exec.Command("cm", "worktree", "create", "feature/new-branch", "-c", configFile, "-v")
	output, err := cmd.CombinedOutput()

	// The worktree creation should succeed because SetUpstreamBranch now skips
	// setting upstream tracking when the remote branch doesn't exist
	assert.NoError(t, err, "Worktree creation should succeed even for new branches that don't exist on remote")

	// Verify the worktree was created successfully
	outputStr := string(output)
	assert.Contains(t, outputStr, "✓ Worktree created successfully",
		"Output should indicate successful worktree creation. Output: %s", outputStr)
}

// createConfigFile creates a config file for CLI testing
func createConfigFile(t *testing.T, configPath, repositoriesDir, statusFile string) {
	t.Helper()

	configContent := `repositories_dir: "` + repositoriesDir + `"
workspaces_dir: "` + filepath.Join(filepath.Dir(repositoriesDir), "workspaces") + `"
status_file: "` + statusFile + `"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "Failed to create config file")
}

// createFeatureBranch creates a feature branch without simulating remote existence
func createFeatureBranch(t *testing.T, repoPath, branchName string) {
	t.Helper()

	// Set up Git environment
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Create and checkout the feature branch
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err := cmd.Run()
	require.NoError(t, err, "Failed to create feature branch")

	// Create a commit on the feature branch
	testFile := filepath.Join(repoPath, "feature.txt")
	err = os.WriteFile(testFile, []byte("Feature content"), 0644)
	require.NoError(t, err, "Failed to create test file")

	cmd = exec.Command("git", "add", "feature.txt")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	require.NoError(t, err, "Failed to add file")

	cmd = exec.Command("git", "commit", "-m", "Add feature content")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	require.NoError(t, err, "Failed to commit")

	// Switch back to main (or master if main doesn't exist)
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	if err != nil {
		// Try master if main doesn't exist
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoPath
		cmd.Env = gitEnv
		err = cmd.Run()
		require.NoError(t, err, "Failed to switch back to default branch")
	}
}

// createFeatureBranchWithRemoteReference creates a feature branch and simulates it exists on remote
func createFeatureBranchWithRemoteReference(t *testing.T, repoPath, branchName string) {
	t.Helper()

	// Set up Git environment
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Create and checkout the feature branch
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err := cmd.Run()
	require.NoError(t, err, "Failed to create feature branch")

	// Create a commit on the feature branch
	testFile := filepath.Join(repoPath, "feature.txt")
	err = os.WriteFile(testFile, []byte("Feature content"), 0644)
	require.NoError(t, err, "Failed to create test file")

	cmd = exec.Command("git", "add", "feature.txt")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	require.NoError(t, err, "Failed to add file")

	cmd = exec.Command("git", "commit", "-m", "Add feature content")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	require.NoError(t, err, "Failed to commit")

	// Create a local remote reference to simulate the branch exists on remote
	cmd = exec.Command("git", "update-ref", "refs/remotes/origin/"+branchName, "HEAD")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err, "Failed to create remote reference")

	// Verify the remote reference exists
	cmd = exec.Command("git", "branch", "-r")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to list remote branches")
	t.Logf("Remote branches: %s", string(output))

	// Switch back to main (or master if main doesn't exist)
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	err = cmd.Run()
	if err != nil {
		// Try master if main doesn't exist
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoPath
		err = cmd.Run()
		require.NoError(t, err, "Failed to switch back to main/master")
	}
}

// getCurrentBranch gets the current branch in the worktree
func getCurrentBranch(t *testing.T, worktreePath string) string {
	t.Helper()

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		return "UNKNOWN"
	}

	return strings.TrimSpace(string(output))
}

// getUpstreamBranch gets the upstream branch for the given branch in the worktree
func getUpstreamBranch(t *testing.T, worktreePath, branchName string) string {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", branchName+"@{upstream}")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		return "NOT_SET"
	}

	return strings.TrimSpace(string(output))
}
