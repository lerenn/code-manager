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

// TestOpenExistingWorktree tests opening an existing worktree with IDE
func TestOpenExistingWorktree(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory and create worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = wtmInstance.CreateWorkTree("feature/existing-ide", nil)
	require.NoError(t, err, "Worktree creation should succeed")

	// Open the worktree with IDE (dummy IDE will print the path to stdout)
	err = wtmInstance.OpenWorktree("feature/existing-ide", "dummy")
	require.NoError(t, err, "Opening worktree with IDE should succeed")

	// Verify that the worktree path in status.yaml is correct (not the original repo path)
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	worktreeEntry := status.Repositories[0]
	expectedPath := filepath.Join(setup.WtmPath, "repo", "feature", "existing-ide")
	assert.Equal(t, expectedPath, worktreeEntry.Path, "Worktree path should be the worktree directory, not the original repository")
}

// TestOpenNonExistentWorktree tests opening a non-existent worktree
func TestOpenNonExistentWorktree(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test opening a non-existent worktree
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

	err = wtmInstance.OpenWorktree("non-existent-branch", "dummy")
	assert.Error(t, err, "Opening non-existent worktree should fail")
	assert.Contains(t, err.Error(), "worktree not found", "Error should mention worktree not found")
}

// TestOpenWorktreeWithUnsupportedIDE tests opening a worktree with unsupported IDE
func TestOpenWorktreeWithUnsupportedIDE(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first
	wtmInstance := wtm.NewWTM(&config.Config{
		BasePath:   setup.WtmPath,
		StatusFile: setup.StatusPath,
	})

	// Change to repo directory and create worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = wtmInstance.CreateWorkTree("feature/unsupported-ide", nil)
	require.NoError(t, err, "Worktree creation should succeed")

	// Try to open with unsupported IDE
	err = wtmInstance.OpenWorktree("feature/unsupported-ide", "unsupported-ide")
	assert.Error(t, err, "Opening with unsupported IDE should fail")
	assert.Contains(t, err.Error(), "unsupported IDE", "Error should mention unsupported IDE")
}
