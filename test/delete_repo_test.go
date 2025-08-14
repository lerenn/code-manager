//go:build e2e

package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/wtm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deleteWorktree deletes a worktree using the WTM instance
func deleteWorktree(t *testing.T, setup *TestSetup, branch string, force bool) error {
	t.Helper()

	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory and delete worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	return wtmInstance.DeleteWorkTree(branch, force)
}

// TestDeleteWorktreeSingleRepo tests deleting a worktree in single repository mode
func TestDeleteWorktreeSingleRepo(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first
	err := createWorktree(t, setup, "feature/test-delete-branch")
	require.NoError(t, err, "Worktree creation should succeed")

	// Verify the worktree was created
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Verify the worktree exists in the .wtm directory
	assertWorktreeExists(t, setup, "feature/test-delete-branch")

	// Delete the worktree with force flag
	err = deleteWorktree(t, setup, "feature/test-delete-branch", true)
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted from status.yaml
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 0, "Should have no repository entries after deletion")

	// Verify the worktree directory was removed
	worktreePath := filepath.Join(setup.WtmPath, "test-repo", "feature/test-delete-branch")
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
	assert.NotContains(t, worktrees, "feature/test-delete-branch", "Worktree should not be in Git's tracking")
}

// TestDeleteWorktreeNonExistentBranch tests deleting a non-existent worktree
func TestDeleteWorktreeNonExistentBranch(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Try to delete a non-existent worktree
	err := deleteWorktree(t, setup, "non-existent-branch", true)
	assert.Error(t, err, "Should fail when deleting non-existent worktree")
	assert.Contains(t, err.Error(), "worktree not found in status file")
}

// TestDeleteWorktreeVerboseMode tests deleting a worktree with verbose output
func TestDeleteWorktreeVerboseMode(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first
	err := createWorktree(t, setup, "feature/verbose-test")
	require.NoError(t, err, "Worktree creation should succeed")

	// Delete the worktree with verbose mode
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})
	wtmInstance.SetVerbose(true)

	// Change to repo directory and delete worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = wtmInstance.DeleteWorkTree("feature/verbose-test", true)
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 0, "Should have no repository entries after deletion")
}

// TestDeleteWorktreeCLI tests deleting a worktree using the WTM instance directly
func TestDeleteWorktreeCLI(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first using WTM instance
	err := createWorktree(t, setup, "feature/cli-test")
	require.NoError(t, err, "Worktree creation should succeed")

	// Verify the worktree was created
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Delete worktree using WTM instance with force flag
	err = deleteWorktree(t, setup, "feature/cli-test", true)
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 0, "Should have no repository entries after deletion")

	// Verify the worktree directory was removed
	worktreePath := filepath.Join(setup.WtmPath, "test-repo", "feature/cli-test")
	_, err = os.Stat(worktreePath)
	assert.True(t, os.IsNotExist(err), "Worktree directory should be removed")
}

// TestDeleteWorktreeCLIWithVerbose tests deleting a worktree using the WTM instance with verbose output
func TestDeleteWorktreeCLIWithVerbose(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first using WTM instance
	err := createWorktree(t, setup, "feature/verbose-cli-test")
	require.NoError(t, err, "Worktree creation should succeed")

	// Delete worktree using WTM instance with force flag and verbose mode
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})
	wtmInstance.SetVerbose(true)

	// Change to repo directory and delete worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = wtmInstance.DeleteWorkTree("feature/verbose-cli-test", true)
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 0, "Should have no repository entries after deletion")
}

// TestDeleteWorktreeCLINonExistentBranch tests delete command with non-existent branch
func TestDeleteWorktreeCLINonExistentBranch(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Try to delete non-existent worktree using WTM instance
	err := deleteWorktree(t, setup, "non-existent-branch", true)
	assert.Error(t, err, "Should fail when deleting non-existent worktree")
	assert.Contains(t, err.Error(), "worktree not found in status file")
}
