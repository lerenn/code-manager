//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Repository represents a repository entry in the status.yaml file
type Repository = status.Repository

// Remote represents a remote configuration for a repository
type Remote = status.Remote

// WorktreeInfo represents worktree information
type WorktreeInfo = status.WorktreeInfo

// StatusFile represents the structure of the status.yaml file
type StatusFile = status.Status

// TestSetup holds the test environment setup
type TestSetup struct {
	TempDir    string
	ConfigPath string
	RepoPath   string
	CmPath     string
	StatusPath string
}

// setupTestEnvironment creates a temporary test environment
func setupTestEnvironment(t *testing.T) *TestSetup {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cm-e2e-test-*")
	require.NoError(t, err)

	// Create subdirectories
	repoPath := filepath.Join(tempDir, "repo")
	cmPath := filepath.Join(tempDir, ".cm")
	statusPath := filepath.Join(cmPath, "status.yaml")

	// Create directories
	require.NoError(t, os.MkdirAll(repoPath, 0755))
	require.NoError(t, os.MkdirAll(cmPath, 0755))

	// Create test config using the config package
	testConfig := config.Config{
		RepositoriesDir: cmPath,
		StatusFile:      statusPath,
	}

	configPath := filepath.Join(tempDir, "config.yaml")
	configData, err := yaml.Marshal(testConfig)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	return &TestSetup{
		TempDir:    tempDir,
		ConfigPath: configPath,
		RepoPath:   repoPath,
		CmPath:     cmPath,
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

	// Get directory name for repository-specific configuration
	dirName := filepath.Base(repoPath)

	// Set the default branch to match the remote repository
	// octocat/Hello-World uses master, octocat/Spoon-Knife uses main
	defaultBranch := "master"
	if dirName == "Spoon-Knife" {
		defaultBranch = "main" // Spoon-Knife uses main
	}
	cmd = exec.Command("git", "branch", "-M", defaultBranch)
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Add a remote origin with a real public repository URL to avoid authentication issues
	// Use different repositories for Hello-World and Spoon-Knife to simulate real workspace scenario
	var remoteURL string
	switch dirName {
	case "Hello-World":
		remoteURL = "https://github.com/octocat/Hello-World.git"
	case "Spoon-Knife":
		remoteURL = "https://github.com/octocat/Spoon-Knife.git"
	default:
		// Default fallback for other test scenarios
		remoteURL = "https://github.com/octocat/Hello-World.git"
	}
	cmd = exec.Command("git", "remote", "add", "origin", remoteURL)
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test Repository\n\nThis is a test repository for CM e2e tests.\n"), 0644))

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
		return &StatusFile{
			Repositories: make(map[string]status.Repository),
			Workspaces:   make(map[string]status.Workspace),
		}
	}

	var status StatusFile
	require.NoError(t, yaml.Unmarshal(data, &status))
	return &status
}

// assertWorktreeExists checks that a worktree exists in the expected location
func assertWorktreeExists(t *testing.T, setup *TestSetup, branch string) {
	t.Helper()

	// The worktree should be created in the .cm directory with repo name and branch structure
	// Since we don't know the exact repo name, we'll check for any directory with the branch name
	worktreesDir := filepath.Join(setup.CmPath, "worktrees")

	// First check if worktrees directory exists
	if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
		// If worktrees directory doesn't exist, search in the .cm directory itself
		// This handles the case where worktrees are created directly in .cm/github.com/...
		var worktreePath string

		// List the contents of the .cm directory to see what's there
		cmEntries, err := os.ReadDir(setup.CmPath)
		if err != nil {
			t.Fatalf("Worktree directory should exist for branch %s", branch)
		}

		// Function to recursively search for worktree in directory structure
		var findWorktree func(dir string) string
		findWorktree = func(dir string) string {
			// Check if this directory has origin/branch structure
			originDir := filepath.Join(dir, "origin")
			if _, err := os.Stat(originDir); err == nil {
				branchDir := filepath.Join(originDir, branch)
				if _, err := os.Stat(branchDir); err == nil {
					return branchDir
				}
			}

			// If not, recursively check subdirectories
			entries, err := os.ReadDir(dir)
			if err != nil {
				return ""
			}

			for _, entry := range entries {
				if entry.IsDir() {
					subDir := filepath.Join(dir, entry.Name())
					if result := findWorktree(subDir); result != "" {
						return result
					}
				}
			}

			return ""
		}

		// Search for worktree in each top-level directory in .cm
		for _, entry := range cmEntries {
			if entry.IsDir() {
				repositoriesDir := filepath.Join(setup.CmPath, entry.Name())
				if result := findWorktree(repositoriesDir); result != "" {
					worktreePath = result
					break
				}
			}
		}

		if worktreePath != "" {
			// Found the worktree, continue with validation
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
			return
		}

		t.Fatalf("Worktree directory should exist for branch %s", branch)
		return
	}

	// If worktrees directory exists, use the original logic
	entries, err := os.ReadDir(worktreesDir)
	require.NoError(t, err)

	var worktreePath string

	// Function to recursively search for worktree in directory structure
	var findWorktree func(dir string) string
	findWorktree = func(dir string) string {
		// Check if this directory has origin/branch structure
		originDir := filepath.Join(dir, "origin")
		if _, err := os.Stat(originDir); err == nil {
			branchDir := filepath.Join(originDir, branch)
			if _, err := os.Stat(branchDir); err == nil {
				return branchDir
			}
		}

		// If not, recursively check subdirectories
		entries, err := os.ReadDir(dir)
		if err != nil {
			return ""
		}

		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(dir, entry.Name())
				if result := findWorktree(subDir); result != "" {
					return result
				}
			}
		}

		return ""
	}

	// Search for worktree in each top-level directory
	for _, entry := range entries {
		if entry.IsDir() {
			repositoriesDir := filepath.Join(worktreesDir, entry.Name())
			if result := findWorktree(repositoriesDir); result != "" {
				worktreePath = result
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

	// The worktree path should be in the .cm directory with repo name and branch structure
	// Since we don't know the exact repo name, we'll check for any path containing the branch name
	assert.Contains(t, string(output), branch, "Worktree should be listed in the original repository")
	assert.Contains(t, string(output), setup.CmPath, "Worktree should be in the .cm directory")
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

// getCurrentCommit gets the current commit hash of the repository
func getCurrentCommit(t *testing.T, repoPath string) (string, error) {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getPreviousCommit gets the previous commit hash on the current branch
func getPreviousCommit(t *testing.T, repoPath string) (string, error) {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "HEAD~1")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// createDummyCommit creates a dummy commit for testing purposes
func createDummyCommit(t *testing.T, repoPath string) error {
	t.Helper()

	// Set up Git environment variables for all commands
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Create a dummy file
	dummyFile := filepath.Join(repoPath, "dummy-test-file.txt")
	err := os.WriteFile(dummyFile, []byte("This is a dummy file for testing"), 0644)
	if err != nil {
		return err
	}

	// Add the file
	cmd := exec.Command("git", "add", "dummy-test-file.txt")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	if err := cmd.Run(); err != nil {
		return err
	}

	// Commit the file
	cmd = exec.Command("git", "commit", "-m", "Add dummy file for testing")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	return cmd.Run()
}
