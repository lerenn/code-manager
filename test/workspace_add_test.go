//go:build e2e

package test

import (
	"encoding/json"
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

// addRepositoryToWorkspace adds a repository to an existing workspace using the CM instance
func addRepositoryToWorkspace(t *testing.T, setup *TestSetup, workspaceName, repository string) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	params := codemanager.AddRepositoryToWorkspaceParams{
		WorkspaceName: workspaceName,
		Repository:    repository,
	}

	return cmInstance.AddRepositoryToWorkspace(&params)
}

// TestAddRepositoryToWorkspaceSuccess tests successful addition of repository to workspace
func TestAddRepositoryToWorkspaceSuccess(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create two test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Create workspace with first repository
	err := createWorkspace(t, setup, "test-workspace", []string{repo1Path})
	require.NoError(t, err, "Workspace creation should succeed")

	// Note: We don't create worktrees here because when adding a repository,
	// worktrees are only created for branches that have worktrees in ALL existing repositories.
	// Since we only have one repository initially, no worktrees will be created when adding the second one.

	// Add second repository to workspace
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo2Path)
	require.NoError(t, err, "Adding repository to workspace should succeed")

	// Verify the status.yaml file was updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Workspaces, "Status file should have workspaces section")

	// Check that the workspace exists
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")

	// Check that the workspace now has two repositories
	require.Len(t, workspace.Repositories, 2, "Workspace should have two repositories after adding")

	// Verify that both repositories are in the workspace
	repoURLs := make(map[string]bool)
	for _, repoURL := range workspace.Repositories {
		repoURLs[repoURL] = true
	}
	require.Len(t, repoURLs, 2, "Workspace should reference both repositories")
}

// TestAddRepositoryToWorkspaceWithWorktrees tests adding repository when workspace has worktrees
func TestAddRepositoryToWorkspaceWithWorktrees(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create three test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")
	repo3Path := filepath.Join(setup.TempDir, "New-Repo")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))
	require.NoError(t, os.MkdirAll(repo3Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)
	createTestGitRepo(t, repo3Path)

	// Switch all repositories to a temporary branch to avoid conflicts when creating worktrees
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	restore := safeChdir(t, repo1Path)
	cmd := exec.Command("git", "checkout", "-b", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()
	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "-b", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()
	restore = safeChdir(t, repo3Path)
	cmd = exec.Command("git", "checkout", "-b", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create workspace with first two repositories
	err := createWorkspace(t, setup, "test-workspace", []string{repo1Path, repo2Path})
	require.NoError(t, err, "Workspace creation should succeed")

	// Create worktrees for multiple branches
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// First, we need to ensure the repositories are cloned to the CM-managed location
	// by creating worktrees. But we need to do this carefully to avoid conflicts.
	// Create worktrees for master branch - this will create worktrees for all repos in the workspace
	err = cmInstance.CreateWorkTree("master", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation for master should succeed")

	// Create worktrees for feature branch (need to create the branch first)
	// Switch to repo1 to create feature branch
	restore = safeChdir(t, repo1Path)

	cmd = exec.Command("git", "checkout", "-b", "feature/test")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create a commit on feature branch
	testFile := filepath.Join(repo1Path, "feature.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("feature content"), 0644))
	cmd = exec.Command("git", "add", "feature.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add feature")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	// Switch back to temp-branch to avoid conflicts when creating worktrees
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Do the same for repo2
	restore()
	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "-b", "feature/test")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo2Path, "feature.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("feature content"), 0644))
	cmd = exec.Command("git", "add", "feature.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add feature")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	// Switch back to temp-branch to avoid conflicts when creating worktrees
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create worktrees for feature branch
	err = cmInstance.CreateWorkTree("feature/test", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation for feature/test should succeed")

	// Prepare repo3: create feature/test branch and switch to a different branch to avoid conflicts
	restore = safeChdir(t, repo3Path)
	cmd = exec.Command("git", "checkout", "-b", "feature/test")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo3Path, "feature.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("feature content"), 0644))
	cmd = exec.Command("git", "add", "feature.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add feature")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	// Switch back to temp-branch to avoid conflicts when creating worktrees
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Add third repository to workspace
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo3Path)
	require.NoError(t, err, "Adding repository to workspace should succeed")

	// Verify the status.yaml file was updated
	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")

	// Check that the workspace now has three repositories
	require.Len(t, workspace.Repositories, 3, "Workspace should have three repositories after adding")

	// Verify that worktrees were created for the new repository for both branches
	// (since both branches exist in all existing repositories)
	// Get repository URL from status by checking all repositories
	// The new repository should be github.com/lerenn/lerenn.github.io
	var repo3Status Repository
	found := false
	expectedRepoURL := "github.com/lerenn/lerenn.github.io"
	for url, repo := range status.Repositories {
		if url == expectedRepoURL || strings.Contains(url, "lerenn.github.io") {
			repo3Status = repo
			found = true
			break
		}
	}
	require.True(t, found, "New repository should be in status (looking for %s)", expectedRepoURL)
	require.Len(t, repo3Status.Worktrees, 2, "New repository should have worktrees for both branches")

	// Verify that worktrees actually exist in the file system
	cfg, err := config.NewManager(setup.ConfigPath).GetConfigWithFallback()
	require.NoError(t, err)

	// Find the repo3 URL from status (reuse expectedRepoURL from above)
	var repo3URL string
	for url := range status.Repositories {
		if url == expectedRepoURL || strings.Contains(url, "lerenn.github.io") {
			repo3URL = url
			break
		}
	}
	require.NotEmpty(t, repo3URL, "Should find repo3 URL (looking for %s)", expectedRepoURL)

	// Verify worktrees exist for both branches
	// Note: branch name is "feature/test" but it gets sanitized in paths, so we check for "feature/test"
	masterWorktreePath := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", "master")
	featureWorktreePath := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", "feature/test")
	assert.DirExists(t, masterWorktreePath, "Master worktree should exist")
	assert.DirExists(t, featureWorktreePath, "Feature worktree should exist")

	// Verify workspace files were updated
	workspaceFile1 := filepath.Join(setup.CmPath, "workspaces", "test-workspace", "master.code-workspace")
	workspaceFile2 := filepath.Join(setup.CmPath, "workspaces", "test-workspace", "feature-test.code-workspace") // Branch name gets sanitized

	// Check that workspace files exist and contain the new repository
	for _, workspaceFile := range []string{workspaceFile1, workspaceFile2} {
		if _, err := os.Stat(workspaceFile); err == nil {
			content, err := os.ReadFile(workspaceFile)
			require.NoError(t, err)

			var workspaceConfig struct {
				Folders []struct {
					Name string `json:"name"`
					Path string `json:"path"`
				} `json:"folders"`
			}
			err = json.Unmarshal(content, &workspaceConfig)
			require.NoError(t, err)

			// Should have 3 folders (one for each repository)
			require.Len(t, workspaceConfig.Folders, 3, "Workspace file should have 3 folders")
		}
	}
}

// TestAddRepositoryToWorkspaceDuplicate tests adding duplicate repository
func TestAddRepositoryToWorkspaceDuplicate(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Create workspace with repository
	err := createWorkspace(t, setup, "test-workspace", []string{setup.RepoPath})
	require.NoError(t, err, "Workspace creation should succeed")

	// Try to add the same repository again
	err = addRepositoryToWorkspace(t, setup, "test-workspace", setup.RepoPath)
	assert.Error(t, err, "Adding duplicate repository should fail")
	assert.ErrorIs(t, err, codemanager.ErrDuplicateRepository, "Error should mention duplicate repository")

	// Verify status file still has only one repository
	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist")
	require.Len(t, workspace.Repositories, 1, "Workspace should still have only one repository")
}

// TestAddRepositoryToWorkspaceNotFound tests adding repository when workspace doesn't exist
func TestAddRepositoryToWorkspaceNotFound(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Try to add repository to non-existent workspace
	err := addRepositoryToWorkspace(t, setup, "non-existent-workspace", setup.RepoPath)
	assert.Error(t, err, "Adding repository to non-existent workspace should fail")
	assert.ErrorIs(t, err, codemanager.ErrWorkspaceNotFound, "Error should mention workspace not found")
}

// TestAddRepositoryToWorkspaceInvalidRepository tests adding invalid repository
func TestAddRepositoryToWorkspaceInvalidRepository(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create workspace first
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	createTestGitRepo(t, repo1Path)

	err := createWorkspace(t, setup, "test-workspace", []string{repo1Path})
	require.NoError(t, err, "Workspace creation should succeed")

	// Try to add non-existent repository
	err = addRepositoryToWorkspace(t, setup, "test-workspace", "/non/existent/path")
	assert.Error(t, err, "Adding non-existent repository should fail")
	assert.ErrorIs(t, err, codemanager.ErrRepositoryNotFound, "Error should mention repository not found")
}

// TestAddRepositoryToWorkspaceNoMatchingBranches tests when no branches match
func TestAddRepositoryToWorkspaceNoMatchingBranches(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create two test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Switch both repositories to a temporary branch to avoid conflicts when creating worktrees
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	restore := safeChdir(t, repo1Path)
	cmd := exec.Command("git", "checkout", "-b", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()
	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "-b", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create workspace with first repository
	err := createWorkspace(t, setup, "test-workspace", []string{repo1Path})
	require.NoError(t, err, "Workspace creation should succeed")

	// Create worktree for a branch that doesn't exist in repo2
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create a unique branch in repo1 only
	restore = safeChdir(t, repo1Path)

	cmd = exec.Command("git", "checkout", "-b", "unique-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	testFile := filepath.Join(repo1Path, "unique.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("unique content"), 0644))
	cmd = exec.Command("git", "add", "unique.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add unique")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	// Switch back to temp-branch to avoid conflicts when creating worktrees
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create worktree for unique branch in the workspace (cmInstance already created above)
	err = cmInstance.CreateWorkTree("unique-branch", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation should succeed")

	// Add second repository to workspace
	// Since unique-branch doesn't exist in repo2, no worktrees should be created for it
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo2Path)
	require.NoError(t, err, "Adding repository should succeed even if no matching branches")

	// Verify the repository was added to workspace
	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist")
	require.Len(t, workspace.Repositories, 2, "Workspace should have two repositories")

	// Verify that repo2 has no worktrees (since unique-branch doesn't exist in it)
	// Find repo2 in status by checking all repositories
	var repo2Status Repository
	found := false
	for url, repo := range status.Repositories {
		if strings.Contains(url, "Spoon-Knife") || strings.Contains(repo.Path, "Spoon-Knife") {
			repo2Status = repo
			found = true
			break
		}
	}
	require.True(t, found, "Repository should be in status")
	// Should have no worktrees since unique-branch doesn't exist in repo2
	require.Len(t, repo2Status.Worktrees, 0, "Repository should have no worktrees for non-matching branch")
}
