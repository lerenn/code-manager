//go:build e2e

package test

import (
	"os"
	"testing"

	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadWorktree loads a worktree using the CM instance
func loadWorktree(t *testing.T, setup *TestSetup, branchArg string) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})

	require.NoError(t, err)
	// Safely change to repo directory and load worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	return cmInstance.LoadWorktree(branchArg, codemanager.LoadWorktreeOpts{
		RepositoryName: ".",
	})
}

func TestLoadWorktreeRepoModeWithOptionalRemote(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Safely change to repo directory for setup
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	// Add origin remote
	require.NoError(t, os.WriteFile(".git/config", []byte(`[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = https://github.com/octocat/Hello-World.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`), 0644))

	// Test 1: Load branch without specifying remote (should default to origin)
	t.Run("LoadBranchWithoutRemote", func(t *testing.T) {
		err := loadWorktree(t, setup, "test")
		assert.NoError(t, err)
	})

	// Test 2: Load branch with explicit remote
	t.Run("LoadBranchWithExplicitRemote", func(t *testing.T) {
		err := loadWorktree(t, setup, "origin:octocat-patch-1")
		assert.NoError(t, err)
	})

	// Test 3: Error case - invalid format
	t.Run("LoadInvalidFormat", func(t *testing.T) {
		err := loadWorktree(t, setup, "invalid:branch:format")
		assert.Error(t, err)
		assert.ErrorIs(t, err, codemanager.ErrBranchNameContainsColon)
	})

	// Test 4: Error case - empty argument
	t.Run("LoadEmptyArgument", func(t *testing.T) {
		err := loadWorktree(t, setup, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get branch name")
	})
}

func TestLoadWorktreeRepoModeWithNewRemote(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Safely change to repo directory for setup
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	// Add origin remote
	require.NoError(t, os.WriteFile(".git/config", []byte(`[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = https://github.com/octocat/Hello-World.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`), 0644))

	// Test loading from a new remote (this will fail because remote doesn't exist)
	t.Run("LoadFromNewRemote", func(t *testing.T) {
		err := loadWorktree(t, setup, "george-wicked:patch-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "remote 'george-wicked' not found")
	})
}

// TestWorktreeLoadWithRepository tests loading a worktree with RepositoryName option
func TestWorktreeLoadWithRepository(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Initialize CM in the repository
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Initialize CM from within the repository
	restore := safeChdir(t, setup.RepoPath)
	err = cmInstance.Init(codemanager.InitOpts{
		NonInteractive:  true,
		RepositoriesDir: setup.CmPath,
		StatusFile:      setup.StatusPath,
	})
	restore()
	require.NoError(t, err)

	// Load worktree using RepositoryName option (use a branch that exists)
	err = cmInstance.LoadWorktree("test", codemanager.LoadWorktreeOpts{
		RepositoryName: setup.RepoPath,
	})
	require.NoError(t, err)

	// Verify worktree was created
	worktrees, err := cmInstance.ListWorktrees(codemanager.ListWorktreesOpts{
		RepositoryName: setup.RepoPath,
	})
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)
	assert.Equal(t, "test", worktrees[0].Branch)
}
