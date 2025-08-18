//go:build e2e

package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/cm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createWorktree creates a worktree using the CM instance
func createWorktree(t *testing.T, setup *TestSetup, branch string) error {
	t.Helper()

	cmInstance := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	return cmInstance.CreateWorkTree(branch)
}

// TestCreateWorktreeSingleRepo tests creating a worktree in single repository mode
func TestCreateWorktreeSingleRepo(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree for the feature branch
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Command should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Check that we have one repository entry
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Check that the worktree entry exists
	entry := status.Repositories[0]
	assert.Equal(t, "feature/test-branch", entry.Branch, "Branch should match")
	assert.NotEmpty(t, entry.URL, "Repository URL should be set")
	assert.NotEmpty(t, entry.Path, "Repository path should be set")

	// Verify the worktree exists in the .cm directory
	assertWorktreeExists(t, setup, "feature/test-branch")

	// Verify the worktree is properly linked in the original repository
	assertWorktreeInRepo(t, setup, "feature/test-branch")
}

// TestCreateWorktreeNonExistentBranch tests creating a worktree for a non-existent branch
// Note: The CLI actually creates the branch if it doesn't exist, so this test verifies that behavior
func TestCreateWorktreeNonExistentBranch(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree for a non-existent branch
	err := createWorktree(t, setup, "non-existent-branch")
	require.NoError(t, err, "Command should succeed and create the branch")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Check that we have one repository entry
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Check that the worktree entry exists
	entry := status.Repositories[0]
	// The branch name might include "heads/" prefix, so we check if it contains our branch name
	assert.Contains(t, entry.Branch, "non-existent-branch", "Branch should contain the branch name")
	assert.NotEmpty(t, entry.URL, "Repository URL should be set")
	assert.NotEmpty(t, entry.Path, "Repository path should be set")

	// Verify the worktree exists in the .cm directory
	// Use the actual branch name from the status file
	assertWorktreeExists(t, setup, entry.Branch)

	// Verify the worktree is properly linked in the original repository
	assertWorktreeInRepo(t, setup, entry.Branch)
}

// TestCreateWorktreeAlreadyExists tests creating a worktree that already exists
func TestCreateWorktreeAlreadyExists(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create the worktree first time
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "First creation should succeed")

	// Try to create the same worktree again
	err = createWorktree(t, setup, "feature/test-branch")
	assert.Error(t, err, "Second creation should fail")
	assert.Contains(t, err.Error(), "already exists", "Error should mention worktree already exists")

	// Verify only one worktree entry exists in status file
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have only one worktree entry")
}

// TestCreateWorktreeOutsideGitRepo tests creating a worktree outside a Git repository
func TestCreateWorktreeOutsideGitRepo(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Don't create a Git repository, just use an empty directory

	// Test creating a worktree outside a Git repository
	err := createWorktree(t, setup, "feature/test-branch")
	assert.Error(t, err, "Command should fail outside Git repository")
	assert.Contains(t, err.Error(), "no Git repository or workspace found", "Error should mention no Git repository found")

	// Verify status file exists but is empty (created during CM initialization)
	_, err = os.Stat(setup.StatusPath)
	assert.NoError(t, err, "Status file should exist (created during CM initialization)")

	// Verify status file is empty (no worktrees added)
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 0, "Status file should be empty for failed operation")
}

// TestCreateWorktreeWithVerboseFlag tests creating a worktree with verbose output
func TestCreateWorktreeWithVerboseFlag(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree with verbose flag
	cmInstance := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})
	cmInstance.SetVerbose(true)

	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	err := cmInstance.CreateWorkTree("feature/test-branch")
	require.NoError(t, err, "Command should succeed")

	// Verify the worktree was created successfully
	assertWorktreeExists(t, setup, "feature/test-branch")
	assertWorktreeInRepo(t, setup, "feature/test-branch")
}

// TestCreateWorktreeWithQuietFlag tests creating a worktree with quiet output
func TestCreateWorktreeWithQuietFlag(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree with quiet flag (quiet mode is handled by the logger, not the CM interface)
	err := createWorktree(t, setup, "feature/test-branch")
	require.NoError(t, err, "Command should succeed")

	// Verify the worktree was created successfully
	assertWorktreeExists(t, setup, "feature/test-branch")
	assertWorktreeInRepo(t, setup, "feature/test-branch")
}

// TestCreateWorktreeWithIDE tests creating a worktree with IDE opening
func TestCreateWorktreeWithIDE(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree with IDE opening
	cmInstance := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	ideName := "dummy"

	// Create worktree with IDE (dummy IDE will print the path to stdout)
	err := cmInstance.CreateWorkTree("feature/test-ide", cm.CreateWorkTreeOpts{IDEName: ideName})
	require.NoError(t, err, "Command should succeed")

	// Verify the worktree was created
	assertWorktreeExists(t, setup, "feature/test-ide")
	assertWorktreeInRepo(t, setup, "feature/test-ide")

	// Verify status.yaml was updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Verify that the original repository path in status.yaml is correct (not the worktree path)
	worktreeEntry := status.Repositories[0]
	expectedPath, err := filepath.EvalSymlinks(setup.RepoPath)
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(worktreeEntry.Path)
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath, "Path should be the original repository directory, not the worktree directory")
}

// TestCreateWorktreeWithUnsupportedIDE tests creating a worktree with unsupported IDE
func TestCreateWorktreeWithUnsupportedIDE(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree with unsupported IDE
	cmInstance := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	ideName := "unsupported-ide"
	err := cmInstance.CreateWorkTree("feature/unsupported-ide", cm.CreateWorkTreeOpts{IDEName: ideName})
	assert.Error(t, err, "Command should fail with unsupported IDE")
	assert.Contains(t, err.Error(), "unsupported IDE", "Error should mention unsupported IDE")
}
