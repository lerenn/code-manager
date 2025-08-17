//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Repository represents a repository entry in the status.yaml file
type Repository = status.Repository

// StatusFile represents the structure of the status.yaml file
type StatusFile = status.Status

// TestSetup holds the test environment setup
type TestSetup struct {
	TempDir    string
	ConfigPath string
	RepoPath   string
	WtmPath    string
	StatusPath string
}

// setupTestEnvironment creates a temporary test environment
func setupTestEnvironment(t *testing.T) *TestSetup {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "wtm-e2e-test-*")
	require.NoError(t, err)

	// Create subdirectories
	repoPath := filepath.Join(tempDir, "repo")
	wtmPath := filepath.Join(tempDir, ".wtm")
	statusPath := filepath.Join(wtmPath, "status.yaml")

	// Create directories
	require.NoError(t, os.MkdirAll(repoPath, 0755))
	require.NoError(t, os.MkdirAll(wtmPath, 0755))

	// Create test config using the config package
	testConfig := &config.Config{
		BasePath:   wtmPath,
		StatusFile: statusPath,
	}

	configPath := filepath.Join(tempDir, "config.yaml")
	configData, err := yaml.Marshal(testConfig)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	return &TestSetup{
		TempDir:    tempDir,
		ConfigPath: configPath,
		RepoPath:   repoPath,
		WtmPath:    wtmPath,
		StatusPath: statusPath,
	}
}

// cleanupTestEnvironment removes the temporary test environment
func cleanupTestEnvironment(t *testing.T, setup *TestSetup) {
	t.Helper()
	if setup != nil && setup.TempDir != "" {
		require.NoError(t, os.RemoveAll(setup.TempDir))
	}
}

// safeChdir safely changes to a directory and restores the original directory
func safeChdir(t *testing.T, targetDir string) func() {
	t.Helper()

	// Ensure the target directory exists
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Get current directory, but don't fail if it doesn't exist
	originalDir, err := os.Getwd()
	if err != nil {
		// If we can't get the current directory, use a fallback
		originalDir = "/tmp"
		t.Logf("Warning: could not get current directory, using fallback: %s", originalDir)
	}

	// Verify the target directory exists before changing to it
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Fatalf("Target directory does not exist: %s", targetDir)
	}

	// Change to target directory
	err = os.Chdir(targetDir)
	require.NoError(t, err)

	// Return a function to restore the original directory
	return func() {
		if restoreErr := os.Chdir(originalDir); restoreErr != nil {
			t.Logf("Warning: failed to restore original directory: %v", restoreErr)
		}
	}
}

// createTestGitRepo creates a Git repository with some test content
func createTestGitRepo(t *testing.T, repoPath string) {
	t.Helper()

	// Safely change to the repository directory for setup
	restore := safeChdir(t, repoPath)
	defer restore()

	// Set up Git environment variables for all commands
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Initialize Git repository
	cmd := exec.Command("git", "init")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test Repository\n\nThis is a test repository for WTM e2e tests.\n"), 0644))

	cmd = exec.Command("git", "add", "README.md")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create a test branch
	cmd = exec.Command("git", "checkout", "-b", "feature/test-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create a file in the feature branch
	featureFile := filepath.Join(repoPath, "feature.txt")
	require.NoError(t, os.WriteFile(featureFile, []byte("This is a feature file\n"), 0644))

	cmd = exec.Command("git", "add", "feature.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Add feature file")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Switch back to the default branch (could be main, master, or something else)
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Env = gitEnv
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	currentBranch := strings.TrimSpace(string(output))

	// We need to be on the default branch, not on feature/test-branch

	// If we're on feature/test-branch, switch to the default branch
	if currentBranch == "feature/test-branch" {
		// Find the default branch (usually the first branch created)
		cmd = exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
		cmd.Env = gitEnv
		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		branches := strings.Split(strings.TrimSpace(string(output)), "\n")

		// Find the default branch (not feature/test-branch)
		var defaultBranchName string
		for _, branch := range branches {
			if branch != "feature/test-branch" && branch != "" {
				defaultBranchName = branch
				break
			}
		}

		if defaultBranchName != "" {
			cmd = exec.Command("git", "checkout", defaultBranchName)
			cmd.Env = gitEnv
			require.NoError(t, cmd.Run())
		}
	}
}

// readStatusFile reads and parses the status.yaml file
func readStatusFile(t *testing.T, statusPath string) *StatusFile {
	t.Helper()

	data, err := os.ReadFile(statusPath)
	if err != nil {
		return &StatusFile{Repositories: []Repository{}}
	}

	var status StatusFile
	require.NoError(t, yaml.Unmarshal(data, &status))
	return &status
}

// assertWorktreeExists checks that a worktree exists in the expected location
func assertWorktreeExists(t *testing.T, setup *TestSetup, branch string) {
	t.Helper()

	// The worktree should be created in the .wtm directory with repo name and branch structure
	// Since we don't know the exact repo name, we'll check for any directory with the branch name
	entries, err := os.ReadDir(setup.WtmPath)
	require.NoError(t, err)

	var worktreePath string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "status.yaml" {
			// This is likely the repository directory
			repoDir := filepath.Join(setup.WtmPath, entry.Name())
			branchDir := filepath.Join(repoDir, branch)
			if _, err := os.Stat(branchDir); err == nil {
				worktreePath = branchDir
				break
			}
		}
	}

	require.NotEmpty(t, worktreePath, "Worktree directory should exist for branch %s", branch)
	assert.DirExists(t, worktreePath, "Worktree directory should exist")

	// Check that it's a valid Git worktree
	gitDir := filepath.Join(worktreePath, ".git")
	assert.FileExists(t, gitDir, "Worktree should have .git file")

	// Check that the branch is correct
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, branch, strings.TrimSpace(string(output)), "Worktree should be on the correct branch")
}

// assertWorktreeInRepo checks that the worktree is properly linked in the original repository
func assertWorktreeInRepo(t *testing.T, setup *TestSetup, branch string) {
	t.Helper()

	// Check that the worktree is listed in the original repository
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = setup.RepoPath
	output, err := cmd.Output()
	require.NoError(t, err)

	// The worktree path should be in the .wtm directory with repo name and branch structure
	// Since we don't know the exact repo name, we'll check for any path containing the branch name
	assert.Contains(t, string(output), branch, "Worktree should be listed in the original repository")
	assert.Contains(t, string(output), setup.WtmPath, "Worktree should be in the .wtm directory")
}

// getGitWorktreeList gets the list of worktrees from Git
func getGitWorktreeList(t *testing.T, repoPath string) string {
	t.Helper()

	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	require.NoError(t, err)
	return string(output)
}
