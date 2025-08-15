//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/wtm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// listWorktrees lists worktrees using the WTM instance
func listWorktrees(t *testing.T, setup *TestSetup) ([]Repository, error) {
	t.Helper()

	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory and list worktrees
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	worktrees, _, err := wtmInstance.ListWorktrees()
	return worktrees, err
}

// runListCommand runs the wtm list command and captures output
func runListCommand(t *testing.T, setup *TestSetup, args ...string) (string, error) {
	t.Helper()

	// Build the wtm binary path
	wtmPath := filepath.Join(setup.TempDir, "wtm")

	// Get the current working directory and go up one level to project root
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Dir(currentDir) // Go up one level from test directory

	// Build the binary from the project root
	buildCmd := exec.Command("go", "build", "-o", wtmPath, "cmd/wtm/main.go")
	buildCmd.Dir = projectRoot
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("Build failed with output: %s", string(buildOutput))
		t.Logf("Current directory: %s", currentDir)
		t.Logf("Project root: %s", projectRoot)
		require.NoError(t, err, "Failed to build wtm binary")
	}

	// Prepare command arguments
	cmdArgs := append([]string{"list"}, args...)
	cmdArgs = append(cmdArgs, "--config", setup.ConfigPath)

	cmd := exec.Command(wtmPath, cmdArgs...)
	cmd.Dir = setup.RepoPath

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// TestListWorktreesEmpty tests listing worktrees when none exist
func TestListWorktreesEmpty(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test listing worktrees when none exist
	worktrees, err := listWorktrees(t, setup)
	require.NoError(t, err, "Command should succeed")
	assert.Len(t, worktrees, 0, "Should return empty list when no worktrees exist")

	// Test CLI command output
	output, err := runListCommand(t, setup)
	require.NoError(t, err, "CLI command should succeed")
	assert.Contains(t, output, "No worktrees found for current repository", "Should show appropriate message for empty list")
}

// TestListWorktreesWithWorktrees tests listing worktrees when some exist
func TestListWorktreesWithWorktrees(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create some worktrees first
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Should create first worktree")

	err = createWorktree(t, setup, "bugfix/issue-123")
	require.NoError(t, err, "Should create second worktree")

	// Test listing worktrees
	worktrees, err := listWorktrees(t, setup)
	require.NoError(t, err, "Command should succeed")
	assert.Len(t, worktrees, 2, "Should return 2 worktrees")

	// Verify worktree details
	branchNames := make([]string, len(worktrees))
	for i, wt := range worktrees {
		branchNames[i] = wt.Branch
		assert.NotEmpty(t, wt.URL, "Repository URL should be set")
		assert.NotEmpty(t, wt.Path, "Repository path should be set")
		expectedPath, err := filepath.EvalSymlinks(setup.RepoPath)
		require.NoError(t, err)
		actualPath, err := filepath.EvalSymlinks(wt.Path)
		require.NoError(t, err)
		assert.Equal(t, expectedPath, actualPath, "Path should be the original repository directory")
	}

	// Check that both expected branches are present
	assert.Contains(t, branchNames, "feature/test-branch", "Should include feature branch")
	assert.Contains(t, branchNames, "bugfix/issue-123", "Should include bugfix branch")

	// Test CLI command output
	output, err := runListCommand(t, setup)
	require.NoError(t, err, "CLI command should succeed")

	// Should show repository name and worktrees
	assert.Contains(t, output, "Worktrees for", "Should show repository header")
	assert.Contains(t, output, "feature/test-branch", "Should show feature branch")
	assert.Contains(t, output, "bugfix/issue-123", "Should show bugfix branch")
}

// TestListWorktreesVerboseMode tests listing worktrees with verbose output
func TestListWorktreesVerboseMode(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Should create worktree")

	// Test CLI command with verbose output
	output, err := runListCommand(t, setup, "--verbose")
	require.NoError(t, err, "CLI command should succeed")

	// Should show verbose output
	assert.Contains(t, output, "Starting worktree listing", "Should show verbose start message")
	assert.Contains(t, output, "Checking for .git directory", "Should show Git detection")
	assert.Contains(t, output, "Git repository detected", "Should show repository detection")
	assert.Contains(t, output, "Listing worktrees for single repository mode", "Should show mode detection")
	assert.Contains(t, output, "Repository name:", "Should show repository name extraction")
	assert.Contains(t, output, "Found", "Should show worktree count")
	assert.Contains(t, output, "Worktrees for", "Should show final output")
}

// TestListWorktreesQuietMode tests listing worktrees with quiet output
func TestListWorktreesQuietMode(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Should create worktree")

	// Test CLI command with quiet output
	output, err := runListCommand(t, setup, "--quiet")
	require.NoError(t, err, "CLI command should succeed")

	// Should show only the worktree list, no verbose messages
	assert.Contains(t, output, "Worktrees for", "Should show repository header")
	assert.Contains(t, output, "feature/test-branch", "Should show worktree")
	assert.NotContains(t, output, "Starting worktree listing", "Should not show verbose messages")
	assert.NotContains(t, output, "Checking for .git directory", "Should not show verbose messages")
}

// TestListWorktreesNoRepository tests listing worktrees when not in a Git repository
func TestListWorktreesNoRepository(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Don't create a Git repository, just create the temp directory

	// Test listing worktrees when not in a Git repository
	worktrees, err := listWorktrees(t, setup)
	require.Error(t, err, "Should return error when not in Git repository")
	assert.Contains(t, err.Error(), "no Git repository or workspace found", "Should show appropriate error message")
	assert.Nil(t, worktrees, "Should return nil worktrees")

	// Test CLI command output
	output, err := runListCommand(t, setup)
	require.Error(t, err, "CLI command should fail")
	assert.Contains(t, output, "no Git repository or workspace found", "Should show appropriate error message")
}

// TestListWorktreesWorkspaceMode tests listing worktrees in workspace mode (placeholder)
func TestListWorktreesWorkspaceMode(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a workspace file instead of a Git repository
	workspaceFile := filepath.Join(setup.RepoPath, "test.code-workspace")
	workspaceContent := `{
		"folders": [
			{
				"name": "Test Project",
				"path": "."
			}
		]
	}`
	require.NoError(t, os.WriteFile(workspaceFile, []byte(workspaceContent), 0644))

	// Test listing worktrees in workspace mode
	worktrees, err := listWorktrees(t, setup)
	require.NoError(t, err, "Should succeed in workspace mode")
	assert.Len(t, worktrees, 0, "Should return empty list in workspace mode (placeholder)")

	// Test CLI command output
	output, err := runListCommand(t, setup)
	require.NoError(t, err, "CLI command should succeed")
	assert.Contains(t, output, "No worktrees found for current workspace", "Should show appropriate message for workspace mode")
}

// TestListWorktreesStatusFileCorruption tests listing worktrees with corrupted status file
func TestListWorktreesStatusFileCorruption(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a corrupted status file
	corruptedContent := `repositories:
  - name: test-repo
    branch: test-branch
    path: /invalid/path
    invalid_field: this should cause an error
    [invalid yaml: this is not valid yaml`
	require.NoError(t, os.WriteFile(setup.StatusPath, []byte(corruptedContent), 0644))

	// Test listing worktrees with corrupted status file
	worktrees, err := listWorktrees(t, setup)
	require.Error(t, err, "Should return error with corrupted status file")
	assert.Contains(t, err.Error(), "failed to load worktrees from status file", "Should show appropriate error message")
	assert.Nil(t, worktrees, "Should return nil worktrees")
}

// TestListWorktreesMultipleRepositories tests that only worktrees for current repository are shown
func TestListWorktreesMultipleRepositories(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create worktrees for current repository
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Should create worktree for current repo")

	// Manually add a worktree for a different repository to the status file
	status := readStatusFile(t, setup.StatusPath)
	status.Repositories = append(status.Repositories, Repository{
		URL:    "github.com/other/repo",
		Branch: "feature/other-branch",
		Path:   filepath.Join(setup.WtmPath, "github.com/other/repo/feature/other-branch"),
	})

	// Write the updated status file
	statusData, err := yaml.Marshal(status)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(setup.StatusPath, statusData, 0644))

	// Test listing worktrees - should only show current repository worktrees
	worktrees, err := listWorktrees(t, setup)
	require.NoError(t, err, "Command should succeed")
	assert.Len(t, worktrees, 1, "Should only return worktrees for current repository")
	assert.Equal(t, "feature/test-branch", worktrees[0].Branch, "Should only show current repository worktree")

	// Test CLI command output
	output, err := runListCommand(t, setup)
	require.NoError(t, err, "CLI command should succeed")
	assert.Contains(t, output, "feature/test-branch", "Should show current repository worktree")
	assert.NotContains(t, output, "feature/other-branch", "Should not show other repository worktree")
}

// TestListWorktreesRepositoryNameExtraction tests repository name extraction from different Git configurations
func TestListWorktreesRepositoryNameExtraction(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Set up remote origin URL
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Add remote origin
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/testuser/testrepo.git")
	require.NoError(t, cmd.Run())

	// Create a worktree
	err = createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Should create worktree")

	// Test listing worktrees
	worktrees, err := listWorktrees(t, setup)
	require.NoError(t, err, "Command should succeed")
	assert.Len(t, worktrees, 1, "Should return one worktree")

	// Verify repository name was extracted correctly
	worktree := worktrees[0]
	assert.Equal(t, "github.com/testuser/testrepo", worktree.URL, "Should extract repository name from remote origin")
	assert.Equal(t, "feature/test-branch", worktree.Branch, "Should have correct branch name")
}

// TestListWorktreesNoRemoteOrigin tests repository name extraction when no remote origin is configured
func TestListWorktreesNoRemoteOrigin(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree (this will use local path as repository name)
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Should create worktree")

	// Test listing worktrees
	worktrees, err := listWorktrees(t, setup)
	require.NoError(t, err, "Command should succeed")
	assert.Len(t, worktrees, 1, "Should return one worktree")

	// Verify repository name was extracted correctly (should use local path)
	worktree := worktrees[0]
	assert.NotEmpty(t, worktree.URL, "Should have repository name")
	assert.Equal(t, "feature/test-branch", worktree.Branch, "Should have correct branch name")

	// The URL should be the directory name (since no remote origin)
	// This will depend on the temp directory name, so we just check it's not empty
	assert.True(t, len(worktree.URL) > 0, "Should have non-empty repository name")
}
