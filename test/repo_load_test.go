//go:build e2e

package test

import (
	"os"
	"testing"

	"github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadWorktree loads a worktree using the CM instance
func loadWorktree(t *testing.T, setup *TestSetup, branchArg string) error {
	t.Helper()

	cmInstance, err := cm.NewCM(&config.Config{
		BasePath:   setup.CmPath,
		StatusFile: setup.StatusPath,
	})

	require.NoError(t, err)
	// Safely change to repo directory and load worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	return cmInstance.LoadWorktree(branchArg)
}

func TestLoadWorktreeWithOptionalRemote(t *testing.T) {
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
		assert.Contains(t, err.Error(), "branch name contains invalid character")
	})

	// Test 4: Error case - empty argument
	t.Run("LoadEmptyArgument", func(t *testing.T) {
		err := loadWorktree(t, setup, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "argument cannot be empty")
	})
}

func TestLoadWorktreeWithNewRemote(t *testing.T) {
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

	// Test loading from a new remote (this will fail in real scenario but tests the parsing)
	t.Run("LoadFromNewRemote", func(t *testing.T) {
		err := loadWorktree(t, setup, "george-wicked:patch-1")
		assert.NoError(t, err)
	})
}
