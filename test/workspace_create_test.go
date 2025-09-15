//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createWorkspace creates a workspace using the CM instance
func createWorkspace(t *testing.T, setup *TestSetup, workspaceName string, repositories []string) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	params := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  repositories,
	}

	return cmInstance.CreateWorkspace(params)
}

// TestCreateWorkspaceSuccess tests successful workspace creation with multiple repositories
func TestCreateWorkspaceSuccess(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create multiple test Git repositories with different remote URLs
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Test creating a workspace with multiple repositories
	err := createWorkspace(t, setup, "test-workspace", []string{repo1Path, repo2Path})
	require.NoError(t, err, "Workspace creation should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Workspaces, "Status file should have workspaces section")

	// Check that we have one workspace entry
	require.Len(t, status.Workspaces, 1, "Should have one workspace entry")

	// Check that the workspace exists
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")

	// Check that the workspace has the correct repositories
	require.Len(t, workspace.Repositories, 2, "Workspace should have two repositories")

	// Verify that repositories were added to the status file
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.Len(t, status.Repositories, 2, "Should have two repository entries")

	// Check that both repositories are in the workspace
	repoURLs := make(map[string]bool)
	for _, repoURL := range workspace.Repositories {
		repoURLs[repoURL] = true
	}

	// Verify both repositories are referenced in the workspace
	require.True(t, len(repoURLs) == 2, "Workspace should reference both repositories")
}

// TestCreateWorkspaceDuplicateName tests workspace creation with duplicate workspace name
func TestCreateWorkspaceDuplicateName(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create the workspace first time
	err := createWorkspace(t, setup, "test-workspace", []string{setup.RepoPath})
	require.NoError(t, err, "First workspace creation should succeed")

	// Try to create the same workspace again
	err = createWorkspace(t, setup, "test-workspace", []string{setup.RepoPath})
	assert.Error(t, err, "Second workspace creation should fail")
	assert.ErrorIs(t, err, codemanager.ErrWorkspaceAlreadyExists, "Error should mention workspace already exists")

	// Verify only one workspace entry exists in status file
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Workspaces, 1, "Should still have only one workspace entry")
}

// TestCreateWorkspaceInvalidName tests workspace creation with invalid workspace name
func TestCreateWorkspaceInvalidName(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a workspace with empty name
	err := createWorkspace(t, setup, "", []string{setup.RepoPath})
	assert.Error(t, err, "Workspace creation with empty name should fail")
	assert.ErrorIs(t, err, codemanager.ErrInvalidWorkspaceName, "Error should mention invalid workspace name")

	// Test creating a workspace with invalid characters
	err = createWorkspace(t, setup, "invalid/name", []string{setup.RepoPath})
	assert.Error(t, err, "Workspace creation with invalid characters should fail")
	assert.ErrorIs(t, err, codemanager.ErrInvalidWorkspaceName, "Error should mention invalid workspace name")

	// Verify status file exists but is empty (created during CM initialization)
	_, err = os.Stat(setup.StatusPath)
	assert.NoError(t, err, "Status file should exist (created during CM initialization)")

	// Verify status file is empty (no workspaces added)
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Workspaces, 0, "Status file should be empty for failed operation")
}

// TestCreateWorkspaceNoRepositories tests workspace creation with no repositories
func TestCreateWorkspaceNoRepositories(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Test creating a workspace with no repositories
	err := createWorkspace(t, setup, "test-workspace", []string{})
	assert.Error(t, err, "Workspace creation with no repositories should fail")
	assert.Contains(t, err.Error(), "at least one repository must be specified", "Error should mention repositories are required")

	// Verify status file exists but is empty (created during CM initialization)
	_, err = os.Stat(setup.StatusPath)
	assert.NoError(t, err, "Status file should exist (created during CM initialization)")

	// Verify status file is empty (no workspaces added)
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Workspaces, 0, "Status file should be empty for failed operation")
}

// TestCreateWorkspaceInvalidRepositories tests workspace creation with invalid repositories
func TestCreateWorkspaceInvalidRepositories(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Test creating a workspace with non-existent repository
	err := createWorkspace(t, setup, "test-workspace", []string{"/non/existent/path"})
	assert.Error(t, err, "Workspace creation with non-existent repository should fail")
	assert.ErrorIs(t, err, codemanager.ErrRepositoryNotFound, "Error should mention repository not found")

	// Test creating a workspace with invalid repository (not a git repo)
	invalidRepoPath := filepath.Join(setup.TempDir, "not-a-git-repo")
	require.NoError(t, os.MkdirAll(invalidRepoPath, 0755))
	// Create a file to make it not a git repository
	require.NoError(t, os.WriteFile(filepath.Join(invalidRepoPath, "file.txt"), []byte("not a git repo"), 0644))

	err = createWorkspace(t, setup, "test-workspace", []string{invalidRepoPath})
	assert.Error(t, err, "Workspace creation with invalid repository should fail")
	assert.ErrorIs(t, err, codemanager.ErrInvalidRepository, "Error should mention invalid repository")

	// Verify status file exists but is empty (created during CM initialization)
	_, err = os.Stat(setup.StatusPath)
	assert.NoError(t, err, "Status file should exist (created during CM initialization)")

	// Verify status file is empty (no workspaces added)
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Workspaces, 0, "Status file should be empty for failed operation")
}

// TestCreateWorkspaceWithRelativePaths tests workspace creation with relative paths
func TestCreateWorkspaceWithRelativePaths(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository in a subdirectory
	repoSubDir := filepath.Join(setup.TempDir, "subdir")
	require.NoError(t, os.MkdirAll(repoSubDir, 0755))
	createTestGitRepo(t, repoSubDir)

	// Change to the temp directory to test relative paths
	restore := safeChdir(t, setup.TempDir)
	defer restore()

	// Test creating a workspace with relative path
	err := createWorkspace(t, setup, "test-workspace", []string{"./subdir"})
	require.NoError(t, err, "Workspace creation with relative path should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Workspaces, "Status file should have workspaces section")

	// Check that we have one workspace entry
	require.Len(t, status.Workspaces, 1, "Should have one workspace entry")

	// Check that the workspace exists
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")

	// Check that the workspace has the correct repository
	require.Len(t, workspace.Repositories, 1, "Workspace should have one repository")

	// Verify that repository was added to the status file
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.Len(t, status.Repositories, 1, "Should have one repository entry")
}

// TestCreateWorkspaceWithMixedRepositoryTypes tests workspace creation with mixed repository types
func TestCreateWorkspaceWithMixedRepositoryTypes(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create multiple test Git repositories with different remote URLs
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Create a third repository with a custom remote URL to avoid conflicts
	repo3Path := filepath.Join(setup.TempDir, "subdir", "Custom-Repo")
	require.NoError(t, os.MkdirAll(repo3Path, 0755))

	// Create a custom Git repository with a different remote URL
	restore := safeChdir(t, repo3Path)
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

	// Set the default branch
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Add a custom remote URL
	cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/octocat/Custom-Repo.git")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create initial commit
	readmePath := filepath.Join(repo3Path, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Custom Repository\n\nThis is a custom test repository.\n"), 0644))

	cmd = exec.Command("git", "add", "README.md")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Change to the temp directory to test relative paths
	restore = safeChdir(t, setup.TempDir)
	defer restore()

	// Test creating a workspace with mixed repository types (absolute and relative paths)
	err := createWorkspace(t, setup, "mixed-workspace", []string{repo1Path, repo2Path, "./subdir/Custom-Repo"})
	require.NoError(t, err, "Workspace creation with mixed repository types should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Workspaces, "Status file should have workspaces section")

	// Check that we have one workspace entry
	require.Len(t, status.Workspaces, 1, "Should have one workspace entry")

	// Check that the workspace exists
	workspace, exists := status.Workspaces["mixed-workspace"]
	require.True(t, exists, "Workspace should exist in status file")

	// Check that the workspace has the correct repositories
	require.Len(t, workspace.Repositories, 3, "Workspace should have three repositories")

	// Verify that repositories were added to the status file
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.Len(t, status.Repositories, 3, "Should have three repository entries")
}

// TestCreateWorkspaceDuplicateRepositories tests workspace creation with duplicate repositories
func TestCreateWorkspaceDuplicateRepositories(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a workspace with duplicate repositories
	err := createWorkspace(t, setup, "test-workspace", []string{setup.RepoPath, setup.RepoPath})
	assert.Error(t, err, "Workspace creation with duplicate repositories should fail")
	assert.ErrorIs(t, err, codemanager.ErrDuplicateRepository, "Error should mention duplicate repository")

	// Verify status file exists but is empty (created during CM initialization)
	_, err = os.Stat(setup.StatusPath)
	assert.NoError(t, err, "Status file should exist (created during CM initialization)")

	// Verify status file is empty (no workspaces added)
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Workspaces, 0, "Status file should be empty for failed operation")
}
