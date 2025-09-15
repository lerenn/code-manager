//go:build e2e

package test

import (
	"os"
	"path/filepath"
	"testing"

	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deleteAllWorktreesParams contains parameters for deleteAllWorktrees.
type deleteAllWorktreesParams struct {
	Setup *TestSetup
	Force bool
}

// deleteAllWorktrees deletes all worktrees using the CM instance
func deleteAllWorktrees(t *testing.T, params deleteAllWorktreesParams) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		ConfigManager: config.NewManager(params.Setup.ConfigPath),
	})

	require.NoError(t, err)
	// Change to repo directory and delete all worktrees
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(params.Setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	return cmInstance.DeleteAllWorktrees(params.Force)
}

// TestDeleteAllWorktreesRepoMode tests deleting all worktrees in single repository mode
func TestDeleteAllWorktreesRepoMode(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create multiple worktrees first
	branches := []string{"feature/test-delete-all-branch1", "feature/test-delete-all-branch2", "feature/test-delete-all-branch3"}
	for _, branch := range branches {
		err := createWorktree(t, setup, branch)
		require.NoError(t, err, "Worktree creation should succeed for branch %s", branch)
	}

	// Verify the worktrees were created
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should have one repository entry")
	repo := status.Repositories["github.com/octocat/Hello-World"]
	require.Len(t, repo.Worktrees, len(branches), "Should have %d worktrees", len(branches))

	// Verify the worktrees exist in the .cm directory
	for _, branch := range branches {
		assertWorktreeExists(t, setup, branch)
	}

	// Delete all worktrees with force flag
	err := deleteAllWorktrees(t, deleteAllWorktreesParams{
		Setup: setup,
		Force: true,
	})
	require.NoError(t, err, "Delete all worktrees should succeed")

	// Verify all worktrees were deleted from status.yaml
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")

	// Check that the repository has no worktrees
	repo = status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should have no worktrees after deletion")

	// Verify all worktree directories were removed
	for _, branch := range branches {
		worktreePath := filepath.Join(setup.CmPath, "worktrees", "github.com", "test", "repo", "origin", branch)
		_, err = os.Stat(worktreePath)
		assert.True(t, os.IsNotExist(err), "Worktree directory should be removed for branch %s", branch)
	}

	// Verify the worktrees are no longer in Git's tracking
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Check git worktree list
	worktrees := getGitWorktreeList(t, setup.RepoPath)
	for _, branch := range branches {
		assert.NotContains(t, worktrees, branch, "Worktree should not be in Git's tracking for branch %s", branch)
	}
}

// TestDeleteAllWorktreesRepoModeNoWorktrees tests deleting all worktrees when none exist
func TestDeleteAllWorktreesRepoModeNoWorktrees(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first to ensure the repository is in the status file
	branch := "feature/temp-worktree"
	err := createWorktree(t, setup, branch)
	require.NoError(t, err, "Worktree creation should succeed")

	// Delete the worktree individually first
	err = deleteWorktree(t, deleteWorktreeParams{
		Setup:  setup,
		Branch: branch,
		Force:  true,
	})
	require.NoError(t, err, "Individual worktree deletion should succeed")

	// Verify no worktrees exist now
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should have one repository entry")
	repo := status.Repositories["github.com/octocat/Hello-World"]
	require.Len(t, repo.Worktrees, 0, "Should have no worktrees after individual deletion")

	// Delete all worktrees (should succeed even when none exist)
	err = deleteAllWorktrees(t, deleteAllWorktreesParams{
		Setup: setup,
		Force: true,
	})
	require.NoError(t, err, "Delete all worktrees should succeed even when none exist")

	// Verify status remains unchanged
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")
	repo = status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should still have no worktrees after deletion")
}

// TestDeleteAllWorktreesRepoModeSingleWorktree tests deleting all worktrees when only one exists
func TestDeleteAllWorktreesRepoModeSingleWorktree(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a single worktree
	branch := "feature/single-worktree-test"
	err := createWorktree(t, setup, branch)
	require.NoError(t, err, "Worktree creation should succeed")

	// Verify the worktree was created
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should have one repository entry")
	repo := status.Repositories["github.com/octocat/Hello-World"]
	require.Len(t, repo.Worktrees, 1, "Should have one worktree")

	// Verify the worktree exists in the .cm directory
	assertWorktreeExists(t, setup, branch)

	// Delete all worktrees with force flag
	err = deleteAllWorktrees(t, deleteAllWorktreesParams{
		Setup: setup,
		Force: true,
	})
	require.NoError(t, err, "Delete all worktrees should succeed")

	// Verify the worktree was deleted from status.yaml
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")

	// Check that the repository has no worktrees
	repo = status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should have no worktrees after deletion")

	// Verify the worktree directory was removed
	worktreePath := filepath.Join(setup.CmPath, "worktrees", "github.com", "test", "repo", "origin", branch)
	_, err = os.Stat(worktreePath)
	assert.True(t, os.IsNotExist(err), "Worktree directory should be removed")

	// Verify the worktree is no longer in Git's tracking
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Check git worktree list
	worktrees := getGitWorktreeList(t, setup.RepoPath)
	assert.NotContains(t, worktrees, branch, "Worktree should not be in Git's tracking")
}

// TestDeleteAllWorktreesWithModifiedFiles tests deleting all worktrees with modified files using force flag
func TestDeleteAllWorktreesWithModifiedFiles(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create multiple worktrees
	branches := []string{
		"feature/test-modified-1",
		"feature/test-modified-2",
		"feature/test-modified-3",
	}

	for _, branch := range branches {
		err := createWorktree(t, setup, branch)
		require.NoError(t, err, "Worktree creation should succeed for branch %s", branch)
	}

	// Verify the worktrees were created
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should have one repository entry")
	repo := status.Repositories["github.com/octocat/Hello-World"]
	require.Len(t, repo.Worktrees, len(branches), "Should have %d worktrees", len(branches))

	// Create modified files in each worktree
	for _, branch := range branches {
		worktreePath := findWorktreePath(t, setup, branch)
		require.NotEmpty(t, worktreePath, "Should be able to find worktree path for %s", branch)

		// Create a modified file
		modifiedFilePath := filepath.Join(worktreePath, "modified-file.txt")
		err := os.WriteFile(modifiedFilePath, []byte("This file has been modified in "+branch), 0644)
		require.NoError(t, err, "Should be able to create modified file in worktree %s", branch)

		// Create an untracked file
		untrackedFilePath := filepath.Join(worktreePath, "untracked-file.txt")
		err = os.WriteFile(untrackedFilePath, []byte("This is an untracked file in "+branch), 0644)
		require.NoError(t, err, "Should be able to create untracked file in worktree %s", branch)
	}

	// Note: We skip testing the confirmation prompt in E2E tests since it requires user input.
	// The important part is testing that the force flag is properly propagated to git.
	// The confirmation prompt is tested in unit tests with mocked prompts.

	// Now delete all worktrees with force flag - this should succeed
	err := deleteAllWorktrees(t, deleteAllWorktreesParams{
		Setup: setup,
		Force: true,
	})
	require.NoError(t, err, "Delete all worktrees with force flag should succeed")

	// Verify all worktrees were deleted from status.yaml
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")

	// Check that the repository has no worktrees
	repo = status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should have no worktrees after deletion")

	// Verify all worktree directories were removed
	for _, branch := range branches {
		worktreePath := findWorktreePath(t, setup, branch)
		_, statErr := os.Stat(worktreePath)
		assert.True(t, os.IsNotExist(statErr), "Worktree directory for %s should be removed", branch)
	}

	// Verify all worktrees are no longer in Git's tracking
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Check git worktree list
	worktrees := getGitWorktreeList(t, setup.RepoPath)
	for _, branch := range branches {
		assert.NotContains(t, worktrees, branch, "Worktree %s should not be in Git's tracking", branch)
	}
}
