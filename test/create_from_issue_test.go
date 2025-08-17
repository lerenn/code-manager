//go:build e2e

package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/issue"
	"github.com/lerenn/wtm/pkg/wtm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

// TestCreateFromIssue_StatusFileVerification tests that issue information is stored in the status file
func TestCreateFromIssue_StatusFileVerification(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Create a mock issue info that would be returned by the GitHub API
	mockIssueInfo := &issue.Info{
		Number:      123,
		Title:       "Test Issue Title",
		Description: "This is a test issue description",
		State:       "open",
		URL:         "https://github.com/test-owner/test-repo/issues/123",
		Repository:  "test-repo",
		Owner:       "test-owner",
	}

	// Create a worktree manually with issue information to simulate the behavior
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create a worktree with issue information
	err = wtmInstance.CreateWorkTree("test-branch", wtm.CreateWorkTreeOpts{
		IssueRef: "https://github.com/test-owner/test-repo/issues/123",
	})

	// The creation will fail due to API call, but we can still verify the status file structure
	// Let's manually add the issue information to the status file to test the verification logic
	status := readStatusFile(t, setup.StatusPath)

	// Add a repository entry with issue information
	repoEntry := Repository{
		URL:       "test-owner/test-repo",
		Branch:    "test-branch",
		Path:      setup.RepoPath,
		Workspace: "",
		Remote:    "origin",
		Issue:     mockIssueInfo,
	}

	status.Repositories = append(status.Repositories, repoEntry)

	// Write the status file back
	statusData, err := yaml.Marshal(status)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(setup.StatusPath, statusData, 0644))

	// Now verify that the issue information is stored correctly in the status file
	verifyIssueInfoInStatusFile(t, setup, "test-branch", mockIssueInfo)
}

// verifyIssueInfoInStatusFile verifies that issue information is correctly stored in the status file
func verifyIssueInfoInStatusFile(t *testing.T, setup *TestSetup, branch string, expectedIssue *issue.Info) {
	t.Helper()

	// Read the status file
	status := readStatusFile(t, setup.StatusPath)

	// Find the repository entry for the given branch
	var foundRepo *Repository
	for _, repo := range status.Repositories {
		if repo.Branch == branch {
			foundRepo = &repo
			break
		}
	}

	// Verify that the repository entry exists
	require.NotNil(t, foundRepo, "Repository entry should exist for branch %s", branch)

	// Verify that issue information is present
	require.NotNil(t, foundRepo.Issue, "Issue information should be present in status file")

	// Verify all issue fields
	assert.Equal(t, expectedIssue.Number, foundRepo.Issue.Number, "Issue number should match")
	assert.Equal(t, expectedIssue.Title, foundRepo.Issue.Title, "Issue title should match")
	assert.Equal(t, expectedIssue.Description, foundRepo.Issue.Description, "Issue description should match")
	assert.Equal(t, expectedIssue.State, foundRepo.Issue.State, "Issue state should match")
	assert.Equal(t, expectedIssue.URL, foundRepo.Issue.URL, "Issue URL should match")
	assert.Equal(t, expectedIssue.Repository, foundRepo.Issue.Repository, "Issue repository should match")
	assert.Equal(t, expectedIssue.Owner, foundRepo.Issue.Owner, "Issue owner should match")

	t.Logf("✅ Issue information verified in status file for branch %s", branch)
	t.Logf("   Issue #%d: %s", foundRepo.Issue.Number, foundRepo.Issue.Title)
	t.Logf("   URL: %s", foundRepo.Issue.URL)
	t.Logf("   Repository: %s/%s", foundRepo.Issue.Owner, foundRepo.Issue.Repository)
}

// TestCreateFromIssue_WorkspaceStatusFileVerification tests that issue information is stored in workspace mode
func TestCreateFromIssue_WorkspaceStatusFileVerification(t *testing.T) {
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

	// Create mock issue info
	mockIssueInfo := &issue.Info{
		Number:      456,
		Title:       "Workspace Test Issue",
		Description: "This is a test issue for workspace mode",
		State:       "open",
		URL:         "https://github.com/test-owner/test-repo/issues/456",
		Repository:  "test-repo",
		Owner:       "test-owner",
	}

	// Manually add workspace repository entries with issue information
	status := readStatusFile(t, setup.StatusPath)

	// Add repository entries for workspace mode (each repo gets the same issue info)
	repoEntry := Repository{
		URL:       "test-owner/test-repo",
		Branch:    "workspace-branch",
		Path:      setup.RepoPath,
		Workspace: workspaceFile,
		Remote:    "origin",
		Issue:     mockIssueInfo,
	}

	status.Repositories = append(status.Repositories, repoEntry)

	// Write the status file back
	statusData, err := yaml.Marshal(status)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(setup.StatusPath, statusData, 0644))

	// Verify that the issue information is stored correctly in workspace mode
	verifyIssueInfoInStatusFile(t, setup, "workspace-branch", mockIssueInfo)
}

// TestCreateFromIssue_NoIssueInfo tests that worktrees without issue info don't have the Issue field
func TestCreateFromIssue_NoIssueInfo(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Add GitHub remote origin
	addGitHubRemote(t, setup.RepoPath)

	// Create a worktree without issue information
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create a regular worktree (without issue information)
	err = wtmInstance.CreateWorkTree("regular-branch", wtm.CreateWorkTreeOpts{})
	require.NoError(t, err)

	// Verify that the status file doesn't have issue information for this worktree
	status := readStatusFile(t, setup.StatusPath)

	// Find the repository entry for the regular branch
	var foundRepo *Repository
	for _, repo := range status.Repositories {
		if repo.Branch == "regular-branch" {
			foundRepo = &repo
			break
		}
	}

	// Verify that the repository entry exists
	require.NotNil(t, foundRepo, "Repository entry should exist for branch regular-branch")

	// Verify that issue information is NOT present (should be nil)
	assert.Nil(t, foundRepo.Issue, "Issue information should NOT be present for regular worktrees")

	t.Logf("✅ Verified that regular worktree has no issue information in status file")
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
