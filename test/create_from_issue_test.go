//go:build e2e

package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/wtm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createWorktreeFromIssueWithBranchParams contains parameters for createWorktreeFromIssueWithBranch.
type createWorktreeFromIssueWithBranchParams struct {
	Setup      *TestSetup
	BranchName string
	IssueRef   string
}

// createWorktreeFromIssueWithIDEParams contains parameters for createWorktreeFromIssueWithIDE.
type createWorktreeFromIssueWithIDEParams struct {
	Setup    *TestSetup
	IDEName  string
	IssueRef string
}

func TestCreateFromIssue_InvalidIssueReference(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with invalid issue reference
	err := createWorktreeFromIssue(t, setup, "invalid-issue-ref")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issue reference format")
}

func TestCreateFromIssue_InvalidIssueNumber(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with invalid issue number format
	err := createWorktreeFromIssue(t, setup, "owner/repo#abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issue number")
}

func TestCreateFromIssue_InvalidOwnerRepoFormat(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with invalid owner/repo format
	err := createWorktreeFromIssue(t, setup, "owner#123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid owner/repo format")
}

func TestCreateFromIssue_IssueNumberRequiresContext(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with issue number only (now supported with repository context)
	err := createWorktreeFromIssue(t, setup, "123")
	assert.Error(t, err)
	// Should fail due to API call (issue not found), not parsing
	assert.NotContains(t, err.Error(), "invalid issue reference format")
}

func TestCreateFromIssue_ValidGitHubURL(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with valid GitHub URL format
	// Note: This will fail because we don't have a real GitHub API connection
	// but it should parse the URL correctly
	err := createWorktreeFromIssue(t, setup, "https://github.com/owner/repo/issues/123")
	assert.Error(t, err)
	// Should fail due to API call, not parsing
	assert.NotContains(t, err.Error(), "invalid issue reference format")
}

func TestCreateFromIssue_ValidOwnerRepoFormat(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with valid owner/repo#issue format
	// Note: This will fail because we don't have a real GitHub API connection
	// but it should parse the format correctly
	err := createWorktreeFromIssue(t, setup, "owner/repo#456")
	assert.Error(t, err)
	// Should fail due to API call, not parsing
	assert.NotContains(t, err.Error(), "invalid issue reference format")
}

func TestCreateFromIssue_WithCustomBranchName(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with custom branch name
	// Note: This will fail because we don't have a real GitHub API connection
	// but it should parse the issue reference correctly
	err := createWorktreeFromIssueWithBranch(t, createWorktreeFromIssueWithBranchParams{
		Setup:      setup,
		BranchName: "custom-branch-name",
		IssueRef:   "https://github.com/owner/repo/issues/123",
	})
	assert.Error(t, err)
	// Should fail due to API call, not parsing
	assert.NotContains(t, err.Error(), "invalid issue reference format")
}

func TestCreateFromIssue_WithIDE(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Test with IDE flag
	// Note: This will fail because we don't have a real GitHub API connection
	// but it should parse the issue reference correctly
	err := createWorktreeFromIssueWithIDE(t, createWorktreeFromIssueWithIDEParams{
		Setup:    setup,
		IDEName:  "cursor",
		IssueRef: "https://github.com/owner/repo/issues/123",
	})
	assert.Error(t, err)
	// Should fail due to API call, not parsing
	assert.NotContains(t, err.Error(), "invalid issue reference format")
}

// Helper functions

func createWorktreeFromIssue(t *testing.T, setup *TestSetup, issueRef string) error {
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory and create worktree from issue
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	return wtmInstance.CreateWorkTree("", wtm.CreateWorkTreeOpts{IssueRef: issueRef})
}

func createWorktreeFromIssueWithBranch(t *testing.T, params createWorktreeFromIssueWithBranchParams) error {
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   params.Setup.WtmPath,
		StatusFile: params.Setup.StatusPath,
	})

	// Change to repo directory and create worktree from issue
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(params.Setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	return wtmInstance.CreateWorkTree(params.BranchName, wtm.CreateWorkTreeOpts{IssueRef: params.IssueRef})
}

func createWorktreeFromIssueWithIDE(t *testing.T, params createWorktreeFromIssueWithIDEParams) error {
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   params.Setup.WtmPath,
		StatusFile: params.Setup.StatusPath,
	})

	// Change to repo directory and create worktree from issue
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(params.Setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	return wtmInstance.CreateWorkTree("", wtm.CreateWorkTreeOpts{IssueRef: params.IssueRef, IDEName: params.IDEName})
}

// addGitHubRemote adds a GitHub remote origin to the repository
func addGitHubRemote(t *testing.T, repoPath string) {
	t.Helper()

	// Change to the repository directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(repoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Add GitHub remote origin
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test-owner/test-repo.git")
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	require.NoError(t, cmd.Run())
}

// TestCreateFromIssue_WorkspaceMode tests the create from issue functionality in workspace mode
func TestCreateFromIssue_WorkspaceMode(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Create a workspace file
	workspaceFile := filepath.Join(setup.TempDir, "test.code-workspace")
	workspaceContent := fmt.Sprintf(`{
		"folders": [
			{
				"name": "repo1",
				"path": "%s"
			}
		]
	}`, setup.RepoPath)

	err := os.WriteFile(workspaceFile, []byte(workspaceContent), 0644)
	require.NoError(t, err)

	// Change to workspace directory
	err = os.Chdir(filepath.Dir(workspaceFile))
	require.NoError(t, err)

	// Test with valid GitHub URL format in workspace mode
	// Note: This will fail because we don't have a real GitHub API connection
	// but it should parse the URL correctly
	err = createWorktreeFromIssue(t, setup, "https://github.com/owner/repo/issues/123")
	assert.Error(t, err)
	// Should fail due to API call, not parsing
	assert.NotContains(t, err.Error(), "invalid issue reference format")
}
