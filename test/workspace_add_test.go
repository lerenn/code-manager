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
	workspaceFile1 := filepath.Join(cfg.WorkspacesDir, "test-workspace", "master.code-workspace")
	workspaceFile2 := filepath.Join(cfg.WorkspacesDir, "test-workspace", "feature-test.code-workspace") // Branch name gets sanitized
	require.NotEmpty(t, repo3URL, "Should find repo3 URL")

	// Check that ALL workspace files exist and contain the new repository
	// This is critical - both files should be updated even if worktrees were only created for some branches
	for branchName, workspaceFile := range map[string]string{
		"master":       workspaceFile1,
		"feature/test": workspaceFile2,
	} {
		if _, err := os.Stat(workspaceFile); err == nil {
			content, err := os.ReadFile(workspaceFile)
			require.NoError(t, err, "Should be able to read workspace file for %s", branchName)

			var workspaceConfig struct {
				Folders []struct {
					Name string `json:"name"`
					Path string `json:"path"`
				} `json:"folders"`
			}
			err = json.Unmarshal(content, &workspaceConfig)
			require.NoError(t, err, "Should be able to parse workspace file JSON for %s", branchName)

			// Should have 3 folders (one for each repository)
			require.Len(t, workspaceConfig.Folders, 3,
				"Workspace file for %s should have 3 folders after adding repo3", branchName)

			// Verify repo3 folder entry exists with correct path
			expectedRepo3Path := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", branchName)
			foundRepo3 := false
			for _, folder := range workspaceConfig.Folders {
				if folder.Path == expectedRepo3Path {
					foundRepo3 = true
					break
				}
			}
			require.True(t, foundRepo3,
				"Workspace file for %s should contain repo3 folder entry with path %s", branchName, expectedRepo3Path)
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
	// Even though unique-branch doesn't exist in repo2, worktree will be created from default branch
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo2Path)
	require.NoError(t, err, "Adding repository should succeed")

	// Verify the repository was added to workspace
	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist")
	require.Len(t, workspace.Repositories, 2, "Workspace should have two repositories")

	// Verify that repo2 has worktree for unique-branch (created from default branch)
	// Find repo2 in status by checking all repositories
	var repo2Status Repository
	var repo2URL string
	found := false
	for url, repo := range status.Repositories {
		if strings.Contains(url, "Spoon-Knife") || strings.Contains(repo.Path, "Spoon-Knife") {
			repo2Status = repo
			repo2URL = url
			found = true
			break
		}
	}
	require.True(t, found, "Repository should be in status")
	// Should have worktree for unique-branch (created from default branch even though it didn't exist)
	require.Len(t, repo2Status.Worktrees, 1, "Repository should have worktree for unique-branch (created from default branch)")
	hasUniqueBranchWorktree := false
	for _, worktree := range repo2Status.Worktrees {
		if worktree.Branch == "unique-branch" {
			hasUniqueBranchWorktree = true
			break
		}
	}
	require.True(t, hasUniqueBranchWorktree, "Repo2 should have worktree for unique-branch")

	// CRITICAL: Verify workspace file is still updated even when no worktrees are created
	// This is the key test - workspace file should be updated even if the branch doesn't exist in the new repo
	cfg, err := config.NewManager(setup.ConfigPath).GetConfigWithFallback()
	require.NoError(t, err)

	workspaceFile := filepath.Join(cfg.WorkspacesDir, "test-workspace", "unique-branch.code-workspace")
	if _, err := os.Stat(workspaceFile); err == nil {
		content, err := os.ReadFile(workspaceFile)
		require.NoError(t, err, "Should be able to read workspace file for unique-branch")

		var workspaceConfig struct {
			Folders []struct {
				Name string `json:"name"`
				Path string `json:"path"`
			} `json:"folders"`
		}
		err = json.Unmarshal(content, &workspaceConfig)
		require.NoError(t, err, "Should be able to parse workspace file JSON for unique-branch")

		// Should have 2 folders (repo1 and repo2)
		require.Len(t, workspaceConfig.Folders, 2,
			"Workspace file for unique-branch should have 2 folders after adding repo2")

		// Verify repo2 folder entry exists with correct path (even though worktree doesn't exist)
		expectedRepo2Path := filepath.Join(cfg.RepositoriesDir, repo2URL, "origin", "unique-branch")
		foundRepo2 := false
		for _, folder := range workspaceConfig.Folders {
			if folder.Path == expectedRepo2Path {
				foundRepo2 = true
				break
			}
		}
		require.True(t, foundRepo2,
			"Workspace file for unique-branch should contain repo2 folder entry with path %s (even though worktree doesn't exist)",
			expectedRepo2Path)
	}
}

// TestAddRepositoryToWorkspaceWithSSHURL tests that workspace file paths use normalized URLs
// when adding repositories with SSH URL format remotes
func TestAddRepositoryToWorkspaceWithSSHURL(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create two test Git repositories
	repo1Path := filepath.Join(setup.TempDir, "extract-lgtm")
	repo2Path := filepath.Join(setup.TempDir, "another-repo")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Set up Git environment variables
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Set SSH URL format remotes for both repositories
	restore := safeChdir(t, repo1Path)
	sshURL1 := "ssh://git@forge.lab.home.lerenn.net/homelab/lgtm/origin/extract-lgtm.git"
	cmd := exec.Command("git", "remote", "set-url", "origin", sshURL1)
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	restore = safeChdir(t, repo2Path)
	sshURL2 := "ssh://git@forge.lab.home.lerenn.net/homelab/lgtm/origin/another-repo.git"
	cmd = exec.Command("git", "remote", "set-url", "origin", sshURL2)
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Switch both repositories to a temporary branch to avoid conflicts when creating worktrees
	restore = safeChdir(t, repo1Path)
	cmd = exec.Command("git", "checkout", "-b", "temp-branch")
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

	// Create worktree for master branch to generate workspace file
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	err = cmInstance.CreateWorkTree("master", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation for master should succeed")

	// Prepare repo2: create master branch and switch to temp-branch to avoid conflicts
	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "master")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	// Switch back to temp-branch to avoid conflicts when creating worktrees
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Add second repository to workspace
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo2Path)
	require.NoError(t, err, "Adding repository to workspace should succeed")

	// Verify the status.yaml file was updated
	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")

	// Check that the workspace now has two repositories
	require.Len(t, workspace.Repositories, 2, "Workspace should have two repositories after adding")

	// Verify that repositories in status use normalized URLs (not raw SSH URLs)
	normalizedURL1 := "forge.lab.home.lerenn.net/homelab/lgtm/origin/extract-lgtm"
	normalizedURL2 := "forge.lab.home.lerenn.net/homelab/lgtm/origin/another-repo"

	for _, repoURL := range workspace.Repositories {
		// Assert that no repository URL contains ssh:// or git@ protocol prefixes
		assert.NotContains(t, repoURL, "ssh://", "Repository URL should not contain ssh:// protocol prefix")
		assert.NotContains(t, repoURL, "git@", "Repository URL should not contain git@ protocol prefix")
		// Assert that URLs are normalized
		assert.True(t, repoURL == normalizedURL1 || repoURL == normalizedURL2,
			"Repository URL should be normalized: got %s, expected one of %s or %s",
			repoURL, normalizedURL1, normalizedURL2)
	}

	// Verify workspace file contains normalized paths
	workspaceFile := filepath.Join(setup.CmPath, "workspaces", "test-workspace", "master.code-workspace")
	require.FileExists(t, workspaceFile, "Workspace file should exist")

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

	// Should have 2 folders (one for each repository)
	require.Len(t, workspaceConfig.Folders, 2, "Workspace file should have 2 folders")

	// Verify all folder paths use normalized URLs
	cfg, err := config.NewManager(setup.ConfigPath).GetConfigWithFallback()
	require.NoError(t, err)

	expectedPath1 := filepath.Join(cfg.RepositoriesDir, normalizedURL1, "origin", "master")
	expectedPath2 := filepath.Join(cfg.RepositoriesDir, normalizedURL2, "origin", "master")

	for _, folder := range workspaceConfig.Folders {
		// Assert that no path contains ssh:// or git@ protocol prefixes
		assert.NotContains(t, folder.Path, "ssh://", "Folder path should not contain ssh:// protocol prefix: %s", folder.Path)
		assert.NotContains(t, folder.Path, "git@", "Folder path should not contain git@ protocol prefix: %s", folder.Path)
		// Assert that path uses normalized URL format
		assert.True(t, folder.Path == expectedPath1 || folder.Path == expectedPath2,
			"Folder path should use normalized URL: got %s, expected one of %s or %s",
			folder.Path, expectedPath1, expectedPath2)
	}
}

// TestAddRepositoryToWorkspaceUpdatesAllWorkspaceFiles tests that when adding a repository
// to a workspace with multiple branches, ALL existing workspace files are updated, not just
// the ones where worktrees are created. This reproduces the bug where workspace files were
// only updated for branches where worktrees were created.
func TestAddRepositoryToWorkspaceUpdatesAllWorkspaceFiles(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create three test Git repositories with different names to get different remote URLs
	// "Hello-World" maps to github.com/octocat/Hello-World
	// "Spoon-Knife" maps to github.com/octocat/Spoon-Knife
	// "New-Repo" maps to github.com/lerenn/lerenn.github.io
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")
	repo3Path := filepath.Join(setup.TempDir, "New-Repo")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))
	require.NoError(t, os.MkdirAll(repo3Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)
	createTestGitRepo(t, repo3Path)

	// Set up Git environment variables
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Switch all repositories to a temporary branch to avoid conflicts when creating worktrees
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

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create branch "extract-lgtm" in both repo1 and repo2
	restore = safeChdir(t, repo1Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile := filepath.Join(repo1Path, "lgtm.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("lgtm content"), 0644))
	cmd = exec.Command("git", "add", "lgtm.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo2Path, "lgtm.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("lgtm content"), 0644))
	cmd = exec.Command("git", "add", "lgtm.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create branch "extract-homeassistant" in both repo1 and repo2
	restore = safeChdir(t, repo1Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo1Path, "homeassistant.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("homeassistant content"), 0644))
	cmd = exec.Command("git", "add", "homeassistant.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo2Path, "homeassistant.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("homeassistant content"), 0644))
	cmd = exec.Command("git", "add", "homeassistant.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create worktrees for both branches in the workspace
	err = cmInstance.CreateWorkTree("extract-lgtm", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation for extract-lgtm should succeed")

	err = cmInstance.CreateWorkTree("extract-homeassistant", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation for extract-homeassistant should succeed")

	// Verify workspace files exist for both branches
	cfg, err := config.NewManager(setup.ConfigPath).GetConfigWithFallback()
	require.NoError(t, err)

	workspaceFile1 := filepath.Join(cfg.WorkspacesDir, "test-workspace", "extract-lgtm.code-workspace")
	workspaceFile2 := filepath.Join(cfg.WorkspacesDir, "test-workspace", "extract-homeassistant.code-workspace")

	require.FileExists(t, workspaceFile1, "Workspace file for extract-lgtm should exist")
	require.FileExists(t, workspaceFile2, "Workspace file for extract-homeassistant should exist")

	// Verify both workspace files have 2 folders (repo1 and repo2)
	for _, workspaceFile := range []string{workspaceFile1, workspaceFile2} {
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

		require.Len(t, workspaceConfig.Folders, 2, "Workspace file should have 2 folders before adding repo3")
	}

	// Prepare repo3: create ONLY extract-homeassistant branch (NOT extract-lgtm)
	// This ensures worktrees are only created for extract-homeassistant, not extract-lgtm
	restore = safeChdir(t, repo3Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo3Path, "homeassistant.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("homeassistant content"), 0644))
	cmd = exec.Command("git", "add", "homeassistant.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Add third repository to workspace
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo3Path)
	require.NoError(t, err, "Adding repository to workspace should succeed")

	// Verify status.yaml was updated
	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")
	require.Len(t, workspace.Repositories, 3, "Workspace should have three repositories after adding")

	// Find repo3 URL from status
	var repo3URL string
	for url := range status.Repositories {
		if strings.Contains(url, "repo3") || strings.Contains(url, "lerenn.github.io") {
			repo3URL = url
			break
		}
	}
	require.NotEmpty(t, repo3URL, "Should find repo3 URL")

	// Verify worktrees were created for BOTH branches (extract-lgtm and extract-homeassistant)
	// Even though extract-lgtm doesn't exist in repo3, it will be created from default branch
	repo3Status, exists := status.Repositories[repo3URL]
	require.True(t, exists, "Repo3 should exist in status")
	// Should have worktrees for both branches (they'll be created from default branch if they don't exist)
	require.Len(t, repo3Status.Worktrees, 2, "Repo3 should have worktrees for both branches")
	hasLgtmWorktree := false
	hasHomeassistantWorktree := false
	for _, worktree := range repo3Status.Worktrees {
		if worktree.Branch == "extract-lgtm" {
			hasLgtmWorktree = true
		}
		if worktree.Branch == "extract-homeassistant" {
			hasHomeassistantWorktree = true
		}
	}
	require.True(t, hasLgtmWorktree, "Repo3 should have worktree for extract-lgtm")
	require.True(t, hasHomeassistantWorktree, "Repo3 should have worktree for extract-homeassistant")

	// CRITICAL: Verify BOTH workspace files are updated with repo3
	// This is the main test - both files should be updated even though worktree was only created for one branch
	for branchName, workspaceFile := range map[string]string{
		"extract-lgtm":         workspaceFile1,
		"extract-homeassistant": workspaceFile2,
	} {
		content, err := os.ReadFile(workspaceFile)
		require.NoError(t, err, "Should be able to read workspace file for %s", branchName)

		var workspaceConfig struct {
			Folders []struct {
				Name string `json:"name"`
				Path string `json:"path"`
			} `json:"folders"`
		}
		err = json.Unmarshal(content, &workspaceConfig)
		require.NoError(t, err, "Should be able to parse workspace file JSON for %s", branchName)

		// Should have 3 folders (repo1, repo2, and repo3)
		require.Len(t, workspaceConfig.Folders, 3,
			"Workspace file for %s should have 3 folders after adding repo3", branchName)

		// Verify repo3 folder entry exists with correct path
		foundRepo3 := false
		expectedRepo3Path := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", branchName)
		for _, folder := range workspaceConfig.Folders {
			if folder.Path == expectedRepo3Path {
				foundRepo3 = true
				// Verify repository name extraction
				expectedRepoName := extractRepositoryNameFromURL(repo3URL)
				assert.Equal(t, expectedRepoName, folder.Name,
					"Repository name should be correctly extracted for %s", branchName)
				break
			}
		}
		require.True(t, foundRepo3,
			"Workspace file for %s should contain repo3 folder entry with path %s", branchName, expectedRepo3Path)
	}

	// Verify worktree exists for extract-homeassistant
	homeassistantWorktreePath := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", "extract-homeassistant")
	assert.DirExists(t, homeassistantWorktreePath, "Homeassistant worktree should exist")

	// Verify worktree DOES exist for extract-lgtm (created from default branch even though branch didn't exist)
	lgtmWorktreePath := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", "extract-lgtm")
	assert.DirExists(t, lgtmWorktreePath, "LGTM worktree should exist (created from default branch)")
}

// TestAddRepositoryToWorkspaceWithNoExistingWorktrees tests adding a repository
// to a workspace that has no existing worktrees (no workspace files exist yet).
// This should handle gracefully without errors.
func TestAddRepositoryToWorkspaceWithNoExistingWorktrees(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create two test Git repositories with different names to get different remote URLs
	// "Hello-World" maps to github.com/octocat/Hello-World
	// "Spoon-Knife" maps to github.com/octocat/Spoon-Knife
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Create workspace with first repository (no worktrees created)
	err := createWorkspace(t, setup, "test-workspace", []string{repo1Path})
	require.NoError(t, err, "Workspace creation should succeed")

	// Verify workspace has no worktrees in status
	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist")
	require.Len(t, workspace.Worktrees, 0, "Workspace should have no worktrees initially")

	// Verify no workspace files exist
	cfg, err := config.NewManager(setup.ConfigPath).GetConfigWithFallback()
	require.NoError(t, err)
	workspacesDir := cfg.WorkspacesDir
	workspaceDir := filepath.Join(workspacesDir, "test-workspace")
	_, err = os.Stat(workspaceDir)
	require.Error(t, err, "Workspace directory should not exist yet")

	// Add second repository to workspace
	// This should succeed gracefully even though no workspace files exist
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo2Path)
	require.NoError(t, err, "Adding repository should succeed even with no existing worktrees")

	// Verify status.yaml was updated
	status = readStatusFile(t, setup.StatusPath)
	workspace, exists = status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist")
	require.Len(t, workspace.Repositories, 2, "Workspace should have two repositories")
}

// TestAddRepositoryToWorkspaceCreatesWorktreesForAllBranches tests that when adding
// a repository to a workspace, worktrees are created for ALL branches in the workspace,
// not just branches that exist in all existing repositories.
func TestAddRepositoryToWorkspaceCreatesWorktreesForAllBranches(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create three test Git repositories with different names to get different remote URLs
	// "Hello-World" maps to github.com/octocat/Hello-World
	// "Spoon-Knife" maps to github.com/octocat/Spoon-Knife
	// "New-Repo" maps to github.com/lerenn/lerenn.github.io
	repo1Path := filepath.Join(setup.TempDir, "Hello-World")
	repo2Path := filepath.Join(setup.TempDir, "Spoon-Knife")
	repo3Path := filepath.Join(setup.TempDir, "New-Repo")

	require.NoError(t, os.MkdirAll(repo1Path, 0755))
	require.NoError(t, os.MkdirAll(repo2Path, 0755))
	require.NoError(t, os.MkdirAll(repo3Path, 0755))

	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)
	createTestGitRepo(t, repo3Path)

	// Set up Git environment variables
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Switch all repositories to a temporary branch to avoid conflicts when creating worktrees
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

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create branch "extract-lgtm" in both repo1 and repo2
	restore = safeChdir(t, repo1Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile := filepath.Join(repo1Path, "lgtm.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("lgtm content"), 0644))
	cmd = exec.Command("git", "add", "lgtm.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo2Path, "lgtm.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("lgtm content"), 0644))
	cmd = exec.Command("git", "add", "lgtm.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create branch "extract-homeassistant" in both repo1 and repo2
	restore = safeChdir(t, repo1Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo1Path, "homeassistant.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("homeassistant content"), 0644))
	cmd = exec.Command("git", "add", "homeassistant.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	restore = safeChdir(t, repo2Path)
	cmd = exec.Command("git", "checkout", "-b", "extract-homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo2Path, "homeassistant.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("homeassistant content"), 0644))
	cmd = exec.Command("git", "add", "homeassistant.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Create worktrees for both branches in the workspace
	err = cmInstance.CreateWorkTree("extract-lgtm", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation for extract-lgtm should succeed")

	err = cmInstance.CreateWorkTree("extract-homeassistant", codemanager.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err, "Worktree creation for extract-homeassistant should succeed")

	// Verify workspace has both branches
	cfg, err := config.NewManager(setup.ConfigPath).GetConfigWithFallback()
	require.NoError(t, err)

	status := readStatusFile(t, setup.StatusPath)
	workspace, exists := status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")
	require.Len(t, workspace.Worktrees, 2, "Workspace should have 2 branches")
	require.Contains(t, workspace.Worktrees, "extract-lgtm", "Workspace should have extract-lgtm branch")
	require.Contains(t, workspace.Worktrees, "extract-homeassistant", "Workspace should have extract-homeassistant branch")

	// CRITICAL: Prepare repo3 with BOTH branches (extract-lgtm AND extract-homeassistant)
	// This ensures worktrees are created for BOTH branches when adding repo3
	restore = safeChdir(t, repo3Path)
	// Create extract-lgtm branch
	cmd = exec.Command("git", "checkout", "-b", "extract-lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo3Path, "lgtm.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("lgtm content"), 0644))
	cmd = exec.Command("git", "add", "lgtm.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add lgtm")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	// Create extract-homeassistant branch
	cmd = exec.Command("git", "checkout", "-b", "extract-homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	testFile = filepath.Join(repo3Path, "homeassistant.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("homeassistant content"), 0644))
	cmd = exec.Command("git", "add", "homeassistant.txt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Add homeassistant")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "checkout", "temp-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())
	restore()

	// Add third repository to workspace
	err = addRepositoryToWorkspace(t, setup, "test-workspace", repo3Path)
	require.NoError(t, err, "Adding repository to workspace should succeed")

	// Verify status.yaml was updated
	status = readStatusFile(t, setup.StatusPath)
	workspace, exists = status.Workspaces["test-workspace"]
	require.True(t, exists, "Workspace should exist in status file")
	require.Len(t, workspace.Repositories, 3, "Workspace should have three repositories after adding")

	// Find repo3 URL from status
	var repo3URL string
	for url := range status.Repositories {
		if strings.Contains(url, "lerenn.github.io") {
			repo3URL = url
			break
		}
	}
	require.NotEmpty(t, repo3URL, "Should find repo3 URL")

	// CRITICAL: Verify worktrees were created for BOTH branches (extract-lgtm AND extract-homeassistant)
	repo3Status, exists := status.Repositories[repo3URL]
	require.True(t, exists, "Repo3 should exist in status")
	require.Len(t, repo3Status.Worktrees, 2, "Repo3 should have worktrees for BOTH branches")
	hasLgtmWorktree := false
	hasHomeassistantWorktree := false
	for _, worktree := range repo3Status.Worktrees {
		if worktree.Branch == "extract-lgtm" {
			hasLgtmWorktree = true
		}
		if worktree.Branch == "extract-homeassistant" {
			hasHomeassistantWorktree = true
		}
	}
	require.True(t, hasLgtmWorktree, "Repo3 should have worktree for extract-lgtm")
	require.True(t, hasHomeassistantWorktree, "Repo3 should have worktree for extract-homeassistant")

	// Verify worktree directories exist
	expectedLgtmPath := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", "extract-lgtm")
	expectedHomeassistantPath := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", "extract-homeassistant")
	require.DirExists(t, expectedLgtmPath, "Worktree directory for extract-lgtm should exist")
	require.DirExists(t, expectedHomeassistantPath, "Worktree directory for extract-homeassistant should exist")

	// Verify BOTH workspace files are updated with repo3
	workspaceFile1 := filepath.Join(cfg.WorkspacesDir, "test-workspace", "extract-lgtm.code-workspace")
	workspaceFile2 := filepath.Join(cfg.WorkspacesDir, "test-workspace", "extract-homeassistant.code-workspace")

	for branchName, workspaceFile := range map[string]string{
		"extract-lgtm":         workspaceFile1,
		"extract-homeassistant": workspaceFile2,
	} {
		content, err := os.ReadFile(workspaceFile)
		require.NoError(t, err, "Should be able to read workspace file for %s", branchName)

		var workspaceConfig struct {
			Folders []struct {
				Name string `json:"name"`
				Path string `json:"path"`
			} `json:"folders"`
		}
		err = json.Unmarshal(content, &workspaceConfig)
		require.NoError(t, err, "Should be able to parse workspace file JSON for %s", branchName)

		// Should have 3 folders (repo1, repo2, and repo3)
		require.Len(t, workspaceConfig.Folders, 3,
			"Workspace file for %s should have 3 folders after adding repo3", branchName)

		// Verify repo3 folder entry exists with correct path
		foundRepo3 := false
		expectedRepo3Path := filepath.Join(cfg.RepositoriesDir, repo3URL, "origin", branchName)
		for _, folder := range workspaceConfig.Folders {
			if folder.Path == expectedRepo3Path {
				foundRepo3 = true
				break
			}
		}
		require.True(t, foundRepo3, "Workspace file for %s should contain repo3 folder entry", branchName)
	}
}

// extractRepositoryNameFromURL extracts the repository name from a URL (helper for test)
func extractRepositoryNameFromURL(repoURL string) string {
	repoURL = strings.TrimSuffix(repoURL, "/")
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return repoURL
}
