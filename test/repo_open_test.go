//go:build e2e

package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/hooks/ide_opening"
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
	cmInstance, err := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	require.NoError(t, err)
	// Change to repo directory and create worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = cmInstance.CreateWorkTree("feature/existing-ide")
	require.NoError(t, err, "Worktree creation should succeed")

	// Open the worktree with IDE (dummy IDE will print the path to stdout)
	err = cmInstance.OpenWorktree("feature/existing-ide", "dummy")
	require.NoError(t, err, "Opening worktree with IDE should succeed")

	// Verify that the original repository path in status.yaml is correct (not the worktree path)
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Get the first repository from the map
	var repo Repository
	for _, r := range status.Repositories {
		repo = r
		break
	}

	expectedPath, err := filepath.EvalSymlinks(setup.RepoPath)
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(repo.Path)
	require.NoError(t, err)
	assert.Equal(t, expectedPath, actualPath, "Path should be the original repository directory, not the worktree directory")
}

// TestOpenNonExistentWorktree tests opening a non-existent worktree
func TestOpenNonExistentWorktree(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test opening a non-existent worktree
	cmInstance, err := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	require.NoError(t, err)
	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = cmInstance.OpenWorktree("non-existent-branch", "dummy")
	assert.Error(t, err, "Opening non-existent worktree should fail")
	assert.ErrorIs(t, err, cm.ErrWorktreeNotInStatus, "Error should mention worktree not found")
}

// TestOpenWorktreeWithUnsupportedIDE tests opening a worktree with unsupported IDE
func TestOpenWorktreeWithUnsupportedIDE(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a worktree first
	cmInstance, err := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	require.NoError(t, err)
	// Change to repo directory and create worktree
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(setup.RepoPath)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = cmInstance.CreateWorkTree("feature/unsupported-ide")
	require.NoError(t, err, "Worktree creation should succeed")

	// Try to open with unsupported IDE
	err = cmInstance.OpenWorktree("feature/unsupported-ide", "unsupported-ide")
	assert.Error(t, err, "Opening with unsupported IDE should fail")
	assert.ErrorIs(t, err, ide_opening.ErrUnsupportedIDE, "Error should mention unsupported IDE")
}
