//go:build e2e

package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deleteWorktreeParams contains parameters for deleteWorktree.
type deleteWorktreeParams struct {
	Setup  *TestSetup
	Branch string
	Force  bool
}

// deleteWorktree deletes a worktree using the CM instance
func deleteWorktree(t *testing.T, params deleteWorktreeParams) error {
	t.Helper()

	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			BasePath:   params.Setup.CmPath,
			StatusFile: params.Setup.StatusPath,
		},
	})

	require.NoError(t, err)
	// Change to repo directory and delete worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(params.Setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	return cmInstance.DeleteWorkTree(params.Branch, params.Force)
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

	// Verify the worktree exists in the .cm directory
	assertWorktreeExists(t, setup, "feature/test-delete-branch")

	// Delete the worktree with force flag
	err = deleteWorktree(t, deleteWorktreeParams{
		Setup:  setup,
		Branch: "feature/test-delete-branch",
		Force:  true,
	})
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted from status.yaml
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")

	// Check that the repository has no worktrees
	repo := status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should have no worktrees after deletion")

	// Verify the worktree directory was removed (using the correct path structure)
	// The worktree should be in worktrees/github.com/octocat/Hello-World/origin/feature/test-delete-branch
	worktreePath := filepath.Join(setup.CmPath, "worktrees", "github.com", "test", "repo", "origin", "feature/test-delete-branch")
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
	err := deleteWorktree(t, deleteWorktreeParams{
		Setup:  setup,
		Branch: "non-existent-branch",
		Force:  true,
	})
	assert.Error(t, err, "Should fail when deleting non-existent worktree")
	assert.ErrorIs(t, err, cm.ErrWorktreeNotInStatus)
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
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			BasePath:   setup.CmPath,
			StatusFile: setup.StatusPath,
		},
	})

	require.NoError(t, err)
	// Change to repo directory and delete worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = cmInstance.DeleteWorkTree("feature/verbose-test", true)
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")

	// Check that the repository has no worktrees
	repo := status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should have no worktrees after deletion")
}

// TestDeleteWorktreeCLI tests deleting a worktree using the CM instance directly
func TestDeleteWorktreeCLI(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first using CM instance
	err := createWorktree(t, setup, "feature/cli-test")
	require.NoError(t, err, "Worktree creation should succeed")

	// Verify the worktree was created
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Delete worktree using CM instance with force flag
	err = deleteWorktree(t, deleteWorktreeParams{
		Setup:  setup,
		Branch: "feature/cli-test",
		Force:  true,
	})
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted
	status = readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")

	// Check that the repository has no worktrees
	repo := status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should have no worktrees after deletion")

	// Verify the worktree directory was removed (using the correct path structure)
	worktreePath := filepath.Join(setup.CmPath, "worktrees", "github.com", "test", "repo", "origin", "feature/cli-test")
	_, err = os.Stat(worktreePath)
	assert.True(t, os.IsNotExist(err), "Worktree directory should be removed")
}

// TestDeleteWorktreeCLIWithVerbose tests deleting a worktree using the CM instance with verbose output
func TestDeleteWorktreeCLIWithVerbose(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first using CM instance
	err := createWorktree(t, setup, "feature/verbose-cli-test")
	require.NoError(t, err, "Worktree creation should succeed")

	// Delete worktree using CM instance with force flag and verbose mode
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			BasePath:   setup.CmPath,
			StatusFile: setup.StatusPath,
		},
	})

	require.NoError(t, err)
	// Change to repo directory and delete worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = cmInstance.DeleteWorkTree("feature/verbose-cli-test", true)
	require.NoError(t, err, "Worktree deletion should succeed")

	// Verify the worktree was deleted
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have repository entry after deletion")

	// Check that the repository has no worktrees
	repo := status.Repositories["github.com/octocat/Hello-World"]
	assert.Len(t, repo.Worktrees, 0, "Repository should have no worktrees after deletion")
}

// TestDeleteWorktreeCLINonExistentBranch tests delete command with non-existent branch
func TestDeleteWorktreeCLINonExistentBranch(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Try to delete non-existent worktree using CM instance
	err := deleteWorktree(t, deleteWorktreeParams{
		Setup:  setup,
		Branch: "non-existent-branch",
		Force:  true,
	})
	assert.Error(t, err, "Should fail when deleting non-existent worktree")
	assert.ErrorIs(t, err, cm.ErrWorktreeNotInStatus)
}
