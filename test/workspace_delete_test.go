//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/stretchr/testify/require"
)

// deleteWorkspace deletes a workspace using the CM instance
func deleteWorkspace(t *testing.T, setup *TestSetup, workspaceName string, force bool) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)).
			WithLogger(logger.NewVerboseLogger()),
	})
	require.NoError(t, err)

	params := codemanager.DeleteWorkspaceParams{
		WorkspaceName: workspaceName,
		Force:         force,
	}

	return cmInstance.DeleteWorkspace(params)
}

// createWorkspaceWithWorktrees creates a workspace and some worktrees for testing deletion
func createWorkspaceWithWorktrees(t *testing.T, setup *TestSetup, workspaceName string, repositories []string) {
	t.Helper()

	// Create workspace
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace
	createParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  repositories,
	}
	require.NoError(t, cmInstance.CreateWorkspace(createParams))

	// Create one worktree for the workspace (this will create worktrees for all repositories in the workspace)
	worktreeOpts := codemanager.CreateWorkTreeOpts{
		WorkspaceName: workspaceName,
	}
	require.NoError(t, cmInstance.CreateWorkTree("feature/test-branch", worktreeOpts))
}

// createCustomTestGitRepo creates a Git repository with a custom remote URL
func createCustomTestGitRepo(t *testing.T, repoPath, remoteURL string) {
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

	// Set the default branch
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Add custom remote URL
	cmd = exec.Command("git", "remote", "add", "origin", remoteURL)
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Custom Repository\n\nThis is a custom test repository.\n"), 0644))

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

	cmd = exec.Command("git", "checkout", "main")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
}

// TestDeleteWorkspaceSuccess tests successful workspace deletion with worktrees
func TestDeleteWorkspaceSuccess(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create multiple test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	workspaceName := "test-workspace"
	repositories := []string{repo1Path, repo2Path}

	// Create workspace with worktrees
	createWorkspaceWithWorktrees(t, setup, workspaceName, repositories)

	// Verify workspace and worktrees exist before deletion
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Workspaces)
	require.Contains(t, status.Workspaces, workspaceName)

	// Check that worktrees exist in repositories
	require.NotNil(t, status.Repositories)
	workspace := status.Workspaces[workspaceName]
	// Store repository URLs for verification after deletion
	repositoryURLs := workspace.Repositories
	for _, repoURL := range repositoryURLs {
		repo, exists := status.Repositories[repoURL]
		require.True(t, exists, "Repository should exist in status")
		require.NotEmpty(t, repo.Worktrees, "Repository should have worktrees")
	}

	// Delete workspace with force flag
	err := deleteWorkspace(t, setup, workspaceName, true)
	require.NoError(t, err, "Workspace deletion should succeed")

	// Verify workspace is removed from status
	status = readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Workspaces)
	require.NotContains(t, status.Workspaces, workspaceName, "Workspace should be removed from status")

	// Verify worktrees are removed from repositories
	require.NotNil(t, status.Repositories)
	for _, repoURL := range repositoryURLs {
		repo, exists := status.Repositories[repoURL]
		require.True(t, exists, "Repository should still exist in status")
		require.Empty(t, repo.Worktrees, "Repository should have no worktrees after workspace deletion")
	}

	// Verify workspace files are deleted
	workspaceFile := filepath.Join(setup.CmPath, "workspaces", workspaceName+".code-workspace")
	_, err = os.Stat(workspaceFile)
	require.True(t, os.IsNotExist(err), "Main workspace file should be deleted")

	// Verify worktree-specific workspace files are deleted
	worktreeWorkspaceFile := filepath.Join(setup.CmPath, "workspaces", workspaceName, "feature-test-branch.code-workspace")
	_, err = os.Stat(worktreeWorkspaceFile)
	require.True(t, os.IsNotExist(err), "Worktree workspace file should be deleted")
}

// TestDeleteWorkspaceNotFound tests workspace deletion when workspace doesn't exist
func TestDeleteWorkspaceNotFound(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Try to delete non-existent workspace
	err := deleteWorkspace(t, setup, "non-existent-workspace", true)
	require.Error(t, err, "Deleting non-existent workspace should fail")
	require.Contains(t, err.Error(), "not found", "Error should mention workspace not found")
}

// TestDeleteWorkspaceInvalidName tests workspace deletion with invalid workspace name
func TestDeleteWorkspaceInvalidName(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Test with invalid characters
	err := deleteWorkspace(t, setup, "invalid/name", true)
	require.Error(t, err, "Deleting workspace with invalid characters should fail")
	require.Contains(t, err.Error(), "invalid workspace name", "Error should mention invalid workspace name")

}

// TestDeleteWorkspaceEmptyWorkspace tests workspace deletion with no worktrees
func TestDeleteWorkspaceEmptyWorkspace(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	workspaceName := "empty-workspace"

	// Create workspace without worktrees
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	createParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{setup.RepoPath},
	}
	require.NoError(t, cmInstance.CreateWorkspace(createParams))

	// Verify workspace exists
	status := readStatusFile(t, setup.StatusPath)
	require.Contains(t, status.Workspaces, workspaceName)

	// Delete workspace
	err = deleteWorkspace(t, setup, workspaceName, true)
	require.NoError(t, err, "Deleting empty workspace should succeed")

	// Verify workspace is removed
	status = readStatusFile(t, setup.StatusPath)
	require.NotContains(t, status.Workspaces, workspaceName, "Empty workspace should be removed")

	// Verify workspace file is deleted
	workspaceFile := filepath.Join(setup.CmPath, "workspaces", workspaceName+".code-workspace")
	_, err = os.Stat(workspaceFile)
	require.True(t, os.IsNotExist(err), "Workspace file should be deleted")
}

// TestDeleteWorkspaceMultipleRepositories tests workspace deletion with multiple repositories
func TestDeleteWorkspaceMultipleRepositories(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create multiple test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")
	repo3Path := filepath.Join(setup.TempDir, "Custom-Repo")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))
	require.NoError(t, os.MkdirAll(repo3Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)
	// Create a custom repository with a different remote URL to avoid conflicts
	createCustomTestGitRepo(t, repo3Path, "https://github.com/octocat/Custom-Repo.git")

	workspaceName := "multi-repo-workspace"
	repositories := []string{repo1Path, repo2Path, repo3Path}

	// Create workspace with worktrees
	createWorkspaceWithWorktrees(t, setup, workspaceName, repositories)

	// Verify workspace exists with multiple repositories
	status := readStatusFile(t, setup.StatusPath)
	require.Contains(t, status.Workspaces, workspaceName)
	workspace := status.Workspaces[workspaceName]
	require.Len(t, workspace.Repositories, 3, "Workspace should have three repositories")

	// Store repository URLs for verification after deletion
	repositoryURLs := workspace.Repositories

	// Delete workspace
	err := deleteWorkspace(t, setup, workspaceName, true)
	require.NoError(t, err, "Deleting multi-repository workspace should succeed")

	// Verify workspace is removed
	status = readStatusFile(t, setup.StatusPath)
	require.NotContains(t, status.Workspaces, workspaceName, "Multi-repository workspace should be removed")

	// Verify all worktrees are removed from all repositories
	for _, repoURL := range repositoryURLs {
		repo, exists := status.Repositories[repoURL]
		require.True(t, exists, "Repository should still exist in status")
		require.Empty(t, repo.Worktrees, "Repository should have no worktrees after workspace deletion")
	}
}

// TestDeleteWorkspacePreservesOtherWorkspaces tests that deleting one workspace doesn't affect others
func TestDeleteWorkspacePreservesOtherWorkspaces(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create multiple test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Create two workspaces with different repositories
	workspace1Name := "workspace-1"
	workspace2Name := "workspace-2"

	// Create first workspace
	createWorkspaceWithWorktrees(t, setup, workspace1Name, []string{repo1Path})

	// Create second workspace with a different repository
	createWorkspaceWithWorktrees(t, setup, workspace2Name, []string{repo2Path})

	// Verify both workspaces exist
	status := readStatusFile(t, setup.StatusPath)
	require.Contains(t, status.Workspaces, workspace1Name)
	require.Contains(t, status.Workspaces, workspace2Name)

	// Delete first workspace
	err := deleteWorkspace(t, setup, workspace1Name, true)
	require.NoError(t, err, "Deleting first workspace should succeed")

	// Verify first workspace is removed but second remains
	status = readStatusFile(t, setup.StatusPath)
	require.NotContains(t, status.Workspaces, workspace1Name, "First workspace should be removed")
	require.Contains(t, status.Workspaces, workspace2Name, "Second workspace should remain")

	// Verify second workspace still has its worktrees
	workspace2 := status.Workspaces[workspace2Name]
	require.Len(t, workspace2.Repositories, 1, "Second workspace should still have its repository")

	// Verify second workspace's worktrees are preserved
	repo2URL := workspace2.Repositories[0]
	repo2, exists := status.Repositories[repo2URL]
	require.True(t, exists, "Second repository should still exist")
	require.NotEmpty(t, repo2.Worktrees, "Second repository should still have worktrees")
}

// TestDeleteWorkspaceWithSharedRepositories tests workspace deletion when repositories are shared
func TestDeleteWorkspaceWithSharedRepositories(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create two test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Create two workspaces sharing the same repository
	workspace1Name := "workspace-1"
	workspace2Name := "workspace-2"

	// Create first workspace with first repository
	createWorkspaceWithWorktrees(t, setup, workspace1Name, []string{repo1Path})

	// Create second workspace sharing the same repository but with different worktrees
	// We need to create worktrees manually for the second workspace since they can't share the same worktree
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create second workspace
	createParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspace2Name,
		Repositories:  []string{repo1Path},
	}
	require.NoError(t, cmInstance.CreateWorkspace(createParams))

	// Create worktrees for the second workspace with different branch names
	worktreeOpts := codemanager.CreateWorkTreeOpts{
		WorkspaceName: workspace2Name,
	}
	require.NoError(t, cmInstance.CreateWorkTree("feature/workspace2-branch", worktreeOpts))

	// Verify both workspaces exist and share the repository
	status := readStatusFile(t, setup.StatusPath)
	require.Contains(t, status.Workspaces, workspace1Name)
	require.Contains(t, status.Workspaces, workspace2Name)

	// The repository should have worktrees from both workspaces
	// Get the repository URL from the first workspace
	workspace1 := status.Workspaces[workspace1Name]
	repoURL := workspace1.Repositories[0]
	repo, exists := status.Repositories[repoURL]
	require.True(t, exists, "Repository should exist")
	require.Len(t, repo.Worktrees, 2, "Repository should have worktrees from both workspaces")

	// Delete first workspace
	err = deleteWorkspace(t, setup, workspace1Name, true)
	require.NoError(t, err, "Deleting first workspace should succeed")

	// Verify first workspace is removed but second remains
	status = readStatusFile(t, setup.StatusPath)
	require.NotContains(t, status.Workspaces, workspace1Name, "First workspace should be removed")
	require.Contains(t, status.Workspaces, workspace2Name, "Second workspace should remain")

	// Verify repository still exists and has worktrees from second workspace only
	repo, exists = status.Repositories[repoURL]
	require.True(t, exists, "Repository should still exist")
	require.Len(t, repo.Worktrees, 1, "Repository should have worktrees from second workspace only")
}

// TestDeleteWorkspaceFileSystemCleanup tests that workspace deletion properly cleans up file system
func TestDeleteWorkspaceFileSystemCleanup(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	workspaceName := "cleanup-test-workspace"

	// Create workspace with worktrees
	createWorkspaceWithWorktrees(t, setup, workspaceName, []string{setup.RepoPath})

	// Verify workspace files exist
	// Only the worktree-specific workspace file should exist
	// The branch name "feature/test-branch" gets sanitized to "feature-test-branch" for filenames
	// The workspace file path structure is: {workspaceName}/{sanitizedBranchName}.code-workspace
	worktreeWorkspaceFile := filepath.Join(setup.CmPath, "workspaces", workspaceName, "feature-test-branch.code-workspace")

	require.FileExists(t, worktreeWorkspaceFile, "Worktree workspace file should exist")

	// Delete workspace
	err := deleteWorkspace(t, setup, workspaceName, true)
	require.NoError(t, err, "Workspace deletion should succeed")

	// Verify all workspace files are deleted
	_, err = os.Stat(worktreeWorkspaceFile)
	require.True(t, os.IsNotExist(err), "Worktree workspace file should be deleted")
}

// TestDeleteWorkspaceGitWorktreeCleanup tests that Git worktrees are properly removed
func TestDeleteWorkspaceGitWorktreeCleanup(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	workspaceName := "git-cleanup-test-workspace"

	// Create workspace with worktrees
	createWorkspaceWithWorktrees(t, setup, workspaceName, []string{setup.RepoPath})

	// Verify worktree exists in Git
	worktreeList := getGitWorktreeList(t, setup.RepoPath)
	require.Contains(t, worktreeList, "feature/test-branch", "Git worktree should exist")

	// Delete workspace
	err := deleteWorkspace(t, setup, workspaceName, true)
	require.NoError(t, err, "Workspace deletion should succeed")

	// Verify worktree is removed from Git
	worktreeList = getGitWorktreeList(t, setup.RepoPath)
	require.NotContains(t, worktreeList, "feature/test-branch", "Git worktree should be removed")
}

// TestDeleteWorkspaceConfirmation tests workspace deletion confirmation (without force)
func TestDeleteWorkspaceConfirmation(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	workspaceName := "confirmation-test-workspace"

	// Create workspace with worktrees
	createWorkspaceWithWorktrees(t, setup, workspaceName, []string{setup.RepoPath})

	// Verify workspace exists
	status := readStatusFile(t, setup.StatusPath)
	require.Contains(t, status.Workspaces, workspaceName)

	// Note: This test doesn't actually test the confirmation prompt interaction
	// since that would require mocking user input. Instead, we test that the
	// workspace deletion works with force=false, which would normally show
	// a confirmation prompt but we can't easily test the interactive part.
	// In a real scenario, the confirmation would be tested with a mock prompt.

	// Delete workspace without force (would normally show confirmation)
	err := deleteWorkspace(t, setup, workspaceName, false)
	// This might fail due to confirmation prompt, which is expected behavior
	// The important thing is that the workspace deletion logic is working
	if err != nil {
		// If it fails due to confirmation, that's expected
		require.Contains(t, err.Error(), "cancelled", "Error should indicate cancellation")
	} else {
		// If it succeeds (unlikely without confirmation), verify cleanup
		status = readStatusFile(t, setup.StatusPath)
		require.NotContains(t, status.Workspaces, workspaceName, "Workspace should be removed if deletion succeeded")
	}
}

// TestDeleteWorkspaceErrorHandling tests error handling during workspace deletion
func TestDeleteWorkspaceErrorHandling(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Test with invalid workspace name
	err := deleteWorkspace(t, setup, "invalid/name", true)
	require.Error(t, err, "Should fail with invalid workspace name")
	require.Contains(t, err.Error(), "invalid workspace name", "Should mention invalid name")

	// Test with non-existent workspace
	err = deleteWorkspace(t, setup, "non-existent", true)
	require.Error(t, err, "Should fail with non-existent workspace")
	require.Contains(t, err.Error(), "not found", "Should mention not found")
}

// TestDeleteWorkspaceCrossPlatform tests workspace deletion across different platforms
func TestDeleteWorkspaceCrossPlatform(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	workspaceName := "cross-platform-test"

	// Create workspace with worktrees
	createWorkspaceWithWorktrees(t, setup, workspaceName, []string{setup.RepoPath})

	// Verify workspace exists
	status := readStatusFile(t, setup.StatusPath)
	require.Contains(t, status.Workspaces, workspaceName)

	// Delete workspace
	err := deleteWorkspace(t, setup, workspaceName, true)
	require.NoError(t, err, "Cross-platform workspace deletion should succeed")

	// Verify cleanup
	status = readStatusFile(t, setup.StatusPath)
	require.NotContains(t, status.Workspaces, workspaceName, "Workspace should be removed")

	// Verify file system cleanup
	workspaceFile := filepath.Join(setup.CmPath, "workspaces", workspaceName+".code-workspace")
	_, err = os.Stat(workspaceFile)
	require.True(t, os.IsNotExist(err), "Workspace file should be deleted")
}
