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

// createWorktreeForUpstreamTest creates a worktree using the CM instance
func createWorktreeForUpstreamTest(t *testing.T, setup *TestSetup, branch string) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})

	require.NoError(t, err)
	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	return cmInstance.CreateWorkTree(branch, codemanager.CreateWorkTreeOpts{
		RepositoryName: ".",
	})
}

// TestWorktreeUpstreamTracking tests that worktrees fail properly when upstream cannot be set
func TestWorktreeUpstreamTracking(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create a feature branch (but don't simulate remote existence)
	createFeatureBranch(t, setup.RepoPath, "feature/upstream-test")

	// Create worktree using the same pattern as other working tests
	err := createWorktreeForUpstreamTest(t, setup, "feature/upstream-test")

	// The worktree creation should succeed because SetUpstreamBranch now skips
	// setting upstream tracking when the remote branch doesn't exist
	assert.NoError(t, err, "Worktree creation should succeed even when remote branch doesn't exist")

	// Verify the worktree was created successfully by listing worktrees
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	restore := safeChdir(t, setup.RepoPath)
	worktrees, err := cmInstance.ListWorktrees(codemanager.ListWorktreesOpts{
		RepositoryName: ".",
	})
	restore()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1, "Should have one worktree")
	assert.Equal(t, "feature/upstream-test", worktrees[0].Branch, "Worktree should be on the correct branch")
}

// TestWorktreeUpstreamTrackingNewBranch tests upstream tracking for new branches that don't exist on remote
func TestWorktreeUpstreamTrackingNewBranch(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create worktree using the same pattern as other working tests
	err := createWorktreeForUpstreamTest(t, setup, "feature/new-branch")

	// The worktree creation should succeed because SetUpstreamBranch now skips
	// setting upstream tracking when the remote branch doesn't exist
	assert.NoError(t, err, "Worktree creation should succeed even for new branches that don't exist on remote")

	// Verify the worktree was created successfully by listing worktrees
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	restore := safeChdir(t, setup.RepoPath)
	worktrees, err := cmInstance.ListWorktrees(codemanager.ListWorktreesOpts{
		RepositoryName: ".",
	})
	restore()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1, "Should have one worktree")
	assert.Equal(t, "feature/new-branch", worktrees[0].Branch, "Worktree should be on the correct branch")
}

// createFeatureBranch creates a feature branch without simulating remote existence
func createFeatureBranch(t *testing.T, repoPath, branchName string) {
	t.Helper()

	// Set up Git environment
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Create and checkout the feature branch
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err := cmd.Run()
	require.NoError(t, err, "Failed to create feature branch")

	// Create a commit on the feature branch
	testFile := filepath.Join(repoPath, "feature.txt")
	err = os.WriteFile(testFile, []byte("Feature content"), 0644)
	require.NoError(t, err, "Failed to create test file")

	cmd = exec.Command("git", "add", "feature.txt")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	require.NoError(t, err, "Failed to add file")

	cmd = exec.Command("git", "commit", "-m", "Add feature content")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	require.NoError(t, err, "Failed to commit")

	// Switch back to main (or master if main doesn't exist)
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	err = cmd.Run()
	if err != nil {
		// Try master if main doesn't exist
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoPath
		cmd.Env = gitEnv
		err = cmd.Run()
		require.NoError(t, err, "Failed to switch back to default branch")
	}
}
