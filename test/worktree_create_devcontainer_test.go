//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeCreate_WithDevcontainer(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create devcontainer configuration
	devcontainerDir := filepath.Join(setup.RepoPath, ".devcontainer")
	err := os.MkdirAll(devcontainerDir, 0755)
	require.NoError(t, err)

	devcontainerConfig := `{
		"name": "Test Devcontainer",
		"image": "mcr.microsoft.com/vscode/devcontainers/base:ubuntu"
	}`
	devcontainerFile := filepath.Join(devcontainerDir, "devcontainer.json")
	err = os.WriteFile(devcontainerFile, []byte(devcontainerConfig), 0644)
	require.NoError(t, err)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	// Create a test branch first (required for detached worktrees)
	branchName := "feature-branch"
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	require.NoError(t, cmd.Run())

	// Create a worktree - should be detached due to devcontainer
	err = cmInstance.CreateWorkTree(branchName, codemanager.CreateWorkTreeOpts{
		RepositoryName: ".",
	})
	require.NoError(t, err)

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Find the repository entry
	var repoEntry *Repository
	for _, repo := range status.Repositories {
		if strings.Contains(repo.Path, setup.RepoPath) {
			repoEntry = &repo
			break
		}
	}
	require.NotNil(t, repoEntry, "Repository should be found in status")

	// Find the worktree entry
	var worktreeEntry *WorktreeInfo
	for _, worktree := range repoEntry.Worktrees {
		if worktree.Branch == branchName {
			worktreeEntry = &worktree
			break
		}
	}
	require.NotNil(t, worktreeEntry, "Worktree should be found in status")
	assert.True(t, worktreeEntry.Detached, "Expected worktree to be marked as detached in status")

	// Verify the worktree directory exists and is a detached clone
	// Use the new worktree structure: $base_path/<repo_url>/<remote_name>/<branch>
	// The repository URL is normalized from the remote origin URL
	repoURL := "github.com/octocat/Hello-World" // This matches the remote URL set in createTestGitRepo
	worktreePath := filepath.Join(setup.CmPath, repoURL, "origin", branchName)
	_, err = os.Stat(worktreePath)
	require.NoError(t, err, "Worktree directory should exist")

	// Verify it's a detached worktree (standalone clone)
	gitDir := filepath.Join(worktreePath, ".git")
	gitDirInfo, err := os.Stat(gitDir)
	require.NoError(t, err)
	assert.True(t, gitDirInfo.IsDir(), "Expected .git to be a directory (standalone clone), not a file (worktree reference)")

	// Verify the worktree works (can checkout, has files)
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, branchName, strings.TrimSpace(string(output)))

	// Verify test file exists (README.md is created by createTestGitRepo)
	testFilePath := filepath.Join(worktreePath, "README.md")
	_, err = os.Stat(testFilePath)
	assert.NoError(t, err, "Expected README.md file to exist in worktree")
}

func TestWorktreeCreate_WithRootDevcontainer(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create root devcontainer configuration
	devcontainerConfig := `{
		"name": "Test Devcontainer",
		"image": "mcr.microsoft.com/vscode/devcontainers/base:ubuntu"
	}`
	devcontainerFile := filepath.Join(setup.RepoPath, ".devcontainer.json")
	err := os.WriteFile(devcontainerFile, []byte(devcontainerConfig), 0644)
	require.NoError(t, err)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	// Create a test branch first (required for detached worktrees)
	branchName := "feature-branch"
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	require.NoError(t, cmd.Run())

	// Create a worktree - should be detached due to devcontainer
	err = cmInstance.CreateWorkTree(branchName, codemanager.CreateWorkTreeOpts{
		RepositoryName: ".",
	})
	require.NoError(t, err)

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Find the repository entry
	var repoEntry *Repository
	for _, repo := range status.Repositories {
		if strings.Contains(repo.Path, setup.RepoPath) {
			repoEntry = &repo
			break
		}
	}
	require.NotNil(t, repoEntry, "Repository should be found in status")

	// Find the worktree entry
	var worktreeEntry *WorktreeInfo
	for _, worktree := range repoEntry.Worktrees {
		if worktree.Branch == branchName {
			worktreeEntry = &worktree
			break
		}
	}
	require.NotNil(t, worktreeEntry, "Worktree should be found in status")
	assert.True(t, worktreeEntry.Detached, "Expected worktree to be marked as detached in status")

	// Verify the worktree directory exists and is a detached clone
	// Use the new worktree structure: $base_path/<repo_url>/<remote_name>/<branch>
	// The repository URL is normalized from the remote origin URL
	repoURL := "github.com/octocat/Hello-World" // This matches the remote URL set in createTestGitRepo
	worktreePath := filepath.Join(setup.CmPath, repoURL, "origin", branchName)
	_, err = os.Stat(worktreePath)
	require.NoError(t, err, "Worktree directory should exist")

	// Verify it's a detached worktree (standalone clone)
	gitDir := filepath.Join(worktreePath, ".git")
	gitDirInfo, err := os.Stat(gitDir)
	require.NoError(t, err)
	assert.True(t, gitDirInfo.IsDir(), "Expected .git to be a directory (standalone clone), not a file (worktree reference)")
}

func TestWorktreeCreate_WithoutDevcontainer(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository (no devcontainer config)
	createTestGitRepo(t, setup.RepoPath)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	// Create a worktree - should be regular worktree (not detached)
	branchName := "feature-branch"
	err = cmInstance.CreateWorkTree(branchName, codemanager.CreateWorkTreeOpts{
		RepositoryName: ".",
	})
	require.NoError(t, err)

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Find the repository entry
	var repoEntry *Repository
	for _, repo := range status.Repositories {
		if strings.Contains(repo.Path, setup.RepoPath) {
			repoEntry = &repo
			break
		}
	}
	require.NotNil(t, repoEntry, "Repository should be found in status")

	// Find the worktree entry
	var worktreeEntry *WorktreeInfo
	for _, worktree := range repoEntry.Worktrees {
		if worktree.Branch == branchName {
			worktreeEntry = &worktree
			break
		}
	}
	require.NotNil(t, worktreeEntry, "Worktree should be found in status")
	assert.False(t, worktreeEntry.Detached, "Expected worktree to NOT be marked as detached in status")

	// Verify the worktree directory exists and is a regular worktree
	// Use the new worktree structure: $base_path/<repo_url>/<remote_name>/<branch>
	// The repository URL is normalized from the remote origin URL
	repoURL := "github.com/octocat/Hello-World" // This matches the remote URL set in createTestGitRepo
	worktreePath := filepath.Join(setup.CmPath, repoURL, "origin", branchName)
	_, err = os.Stat(worktreePath)
	require.NoError(t, err, "Worktree directory should exist")

	// Verify it's a regular worktree (not detached)
	gitFile := filepath.Join(worktreePath, ".git")
	gitFileInfo, err := os.Stat(gitFile)
	require.NoError(t, err)
	assert.False(t, gitFileInfo.IsDir(), "Expected .git to be a file (worktree reference), not a directory (standalone clone)")
}
