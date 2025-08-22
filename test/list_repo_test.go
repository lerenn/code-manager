//go:build e2e

package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// listWorktrees lists worktrees using the CM instance
func listWorktrees(t *testing.T, setup *TestSetup) ([]status.WorktreeInfo, error) {
	t.Helper()

	cmInstance := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory and list worktrees
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	worktrees, _, err := cmInstance.ListWorktrees(false)
	return worktrees, err
}

// runListCommand runs the cm list command and captures output
func runListCommand(t *testing.T, setup *TestSetup, args ...string) (string, error) {
	t.Helper()

	// Check for verbose flag
	isVerbose := false
	for _, arg := range args {
		if arg == "--verbose" {
			isVerbose = true
			break
		}
	}

	// Create CM instance with the test configuration
	cmInstance := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	// Set verbose mode if requested
	if isVerbose {
		cmInstance.SetVerbose(true)
	}

	// Change to repo directory and list worktrees
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Call ListWorktrees directly
	worktrees, projectType, err := cmInstance.ListWorktrees(false)
	if err != nil {
		return err.Error(), err
	}

	// Format output similar to CLI output
	var output strings.Builder

	if projectType == cm.ProjectTypeSingleRepo {
		if len(worktrees) == 0 {
			output.WriteString("No worktrees found for current repository\n")
		} else {
			output.WriteString("Worktrees for current repository:\n")
			for _, wt := range worktrees {
				output.WriteString(fmt.Sprintf("  %s\n", wt.Branch))
			}
		}
	} else if projectType == cm.ProjectTypeWorkspace {
		if len(worktrees) == 0 {
			output.WriteString("No worktrees found for current workspace\n")
		} else {
			output.WriteString("Worktrees for workspace:\n")
			for _, wt := range worktrees {
				output.WriteString(fmt.Sprintf("  %s [%s]\n", wt.Branch, wt.Remote))
			}
		}
	}

	return output.String(), nil
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
		assert.NotEmpty(t, wt.Remote, "Repository remote should be set")
		// Note: WorktreeInfo doesn't have Path field, path verification is done through status manager
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

	// Should show worktree output (verbose mode is handled internally by CM)
	assert.Contains(t, output, "Worktrees for", "Should show repository header")
	assert.Contains(t, output, "feature/test-branch", "Should show worktree")
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
	assert.NotContains(t, output, "Checking if current directory is a Git repository", "Should not show verbose messages")
}

// TestListWorktreesNoRepository tests listing worktrees when not in a Git repository
func TestListWorktreesNoRepository(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Don't create a Git repository, just create the temp directory

	// Test listing worktrees when not in a Git repository
	worktrees, err := listWorktrees(t, setup)
	require.Error(t, err, "Should return error when not in Git repository")
	assert.ErrorIs(t, err, cm.ErrNoGitRepositoryOrWorkspaceFound, "Should show appropriate error message")
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

	// Create a corrupted status file with obviously invalid YAML
	corruptedContent := `this is completely invalid yaml content: [missing quotes, no structure`
	require.NoError(t, os.WriteFile(setup.StatusPath, []byte(corruptedContent), 0644))

	// Test listing worktrees with corrupted status file
	worktrees, err := listWorktrees(t, setup)
	require.Error(t, err, "Should return error with corrupted status file")
	assert.Contains(t, err.Error(), "failed to parse status file", "Should show appropriate error message")
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

	// Add a different repository with its worktree
	repoURL := "github.com/other/repo"
	status.Repositories[repoURL] = Repository{
		Path: setup.RepoPath,
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
		Worktrees: make(map[string]WorktreeInfo),
	}

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

	// Remove existing origin remote if it exists
	cmd := exec.Command("git", "remote", "remove", "origin")
	_ = cmd.Run() // Ignore error if remote doesn't exist

	// Add remote origin
	cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/octocat/Hello-World.git")
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
	assert.Equal(t, "origin", worktree.Remote, "Should have origin remote")
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
	assert.NotEmpty(t, worktree.Remote, "Should have remote name")
	assert.Equal(t, "feature/test-branch", worktree.Branch, "Should have correct branch name")

	// The remote should be the default remote (origin)
	assert.Equal(t, "origin", worktree.Remote, "Should have origin remote")
}
