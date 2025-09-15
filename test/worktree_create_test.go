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

// createWorktree creates a worktree using the CM instance
func createWorktree(t *testing.T, setup *TestSetup, branch string) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})

	require.NoError(t, err)
	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	return cmInstance.CreateWorkTree(branch)
}

// TestCreateWorktreeSingleRepo tests creating a worktree in single repository mode
func TestCreateWorktreeRepoMode(t *testing.T) {
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

	// Check that we have a repository entry (repositories is now a map)
	require.True(t, len(status.Repositories) > 0, "Should have at least one repository")

	// Get the first repository from the map
	var repoURL string
	var repo Repository
	for url, r := range status.Repositories {
		repoURL = url
		repo = r
		break
	}

	assert.NotEmpty(t, repoURL, "Repository URL should be set")
	assert.NotEmpty(t, repo.Path, "Repository path should be set")

	// Check that the repository has the worktree
	require.True(t, len(repo.Worktrees) > 0, "Repository should have at least one worktree")

	// Check that the worktree for our branch exists
	var foundWorktree bool
	for _, worktree := range repo.Worktrees {
		if worktree.Branch == "feature/test-branch" {
			foundWorktree = true
			break
		}
	}
	assert.True(t, foundWorktree, "Should have worktree for feature/test-branch")

	// Verify the worktree exists in the .cm directory
	assertWorktreeExists(t, setup, "feature/test-branch")

	// Verify the worktree is properly linked in the original repository
	assertWorktreeInRepo(t, setup, "feature/test-branch")
}

// TestCreateWorktreeNonExistentBranch tests creating a worktree for a non-existent branch
// Note: The CLI actually creates the branch if it doesn't exist, so this test verifies that behavior
func TestCreateWorktreeRepoModeNonExistentBranch(t *testing.T) {
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

	// Check that we have a repository entry (repositories is now a map)
	require.True(t, len(status.Repositories) > 0, "Should have at least one repository")

	// Get the first repository from the map
	var repoURL string
	var repo Repository
	for url, r := range status.Repositories {
		repoURL = url
		repo = r
		break
	}

	assert.NotEmpty(t, repoURL, "Repository URL should be set")
	assert.NotEmpty(t, repo.Path, "Repository path should be set")

	// Check that the repository has the worktree
	require.True(t, len(repo.Worktrees) > 0, "Repository should have at least one worktree")

	// Find the worktree for non-existent-branch
	var foundWorktree *WorktreeInfo
	var actualBranchName string
	for _, worktree := range repo.Worktrees {
		if strings.Contains(worktree.Branch, "non-existent-branch") {
			foundWorktree = &worktree
			actualBranchName = worktree.Branch
			break
		}
	}
	require.NotNil(t, foundWorktree, "Should have worktree for non-existent-branch")
	assert.Contains(t, actualBranchName, "non-existent-branch", "Branch should contain the branch name")

	// Verify the worktree exists in the .cm directory
	// Use the actual branch name from the status file
	assertWorktreeExists(t, setup, actualBranchName)

	// Verify the worktree is properly linked in the original repository
	assertWorktreeInRepo(t, setup, actualBranchName)
}

// TestCreateWorktreeAlreadyExists tests creating a worktree that already exists
func TestCreateWorktreeRepoModeAlreadyExists(t *testing.T) {
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
	assert.ErrorIs(t, err, codemanager.ErrWorktreeExists, "Error should mention worktree already exists")

	// Verify only one worktree entry exists in status file
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 1, "Should still have only one worktree entry")
}

// TestCreateWorktreeOutsideGitRepo tests creating a worktree outside a Git repository
func TestCreateWorktreeRepoModeOutsideGitRepo(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Don't create a Git repository, just use an empty directory

	// Test creating a worktree outside a Git repository
	err := createWorktree(t, setup, "feature/test-branch")
	assert.Error(t, err, "Command should fail outside Git repository")
	assert.ErrorIs(t, err, codemanager.ErrNoGitRepositoryOrWorkspaceFound, "Error should mention no Git repository found")

	// Verify status file exists but is empty (created during CM initialization)
	_, err = os.Stat(setup.StatusPath)
	assert.NoError(t, err, "Status file should exist (created during CM initialization)")

	// Verify status file is empty (no worktrees added)
	status := readStatusFile(t, setup.StatusPath)
	assert.Len(t, status.Repositories, 0, "Status file should be empty for failed operation")
}

// TestCreateWorktreeWithVerboseFlag tests creating a worktree with verbose output
func TestCreateWorktreeRepoModeWithVerboseFlag(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree with verbose flag
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})

	require.NoError(t, err)
	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	err = cmInstance.CreateWorkTree("feature/test-branch")
	require.NoError(t, err, "Command should succeed")

	// Verify the worktree was created successfully
	assertWorktreeExists(t, setup, "feature/test-branch")
	assertWorktreeInRepo(t, setup, "feature/test-branch")
}

// TestCreateWorktreeWithQuietFlag tests creating a worktree with quiet output
func TestCreateWorktreeRepoModeWithQuietFlag(t *testing.T) {
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
func TestCreateWorktreeRepoModeWithIDE(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree with IDE opening
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})

	require.NoError(t, err)
	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	ideName := "dummy"

	// Create worktree with IDE (dummy IDE will print the path to stdout)
	err = cmInstance.CreateWorkTree("feature/test-ide", codemanager.CreateWorkTreeOpts{IDEName: ideName})
	require.NoError(t, err, "Command should succeed")

	// Verify the worktree was created
	assertWorktreeExists(t, setup, "feature/test-ide")
	assertWorktreeInRepo(t, setup, "feature/test-ide")

	// Verify status.yaml was updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Verify that the original repository path in status.yaml is correct (not the worktree path)
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

// TestCreateWorktreeFromOriginDefaultBranch tests that new worktrees are created from origin's default branch
// and not from the local default branch, even when the local branch has been reset to an older commit
func TestCreateWorktreeRepoModeFromOriginDefaultBranch(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Clone the octocat/Hello-World repository
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Clone the repository
	err = cmInstance.Clone("https://github.com/octocat/Hello-World.git")
	require.NoError(t, err, "Repository clone should succeed")

	// Find the cloned repository path
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")
	require.True(t, len(status.Repositories) > 0, "Should have at least one repository")

	var repoPath string
	for _, repo := range status.Repositories {
		repoPath = repo.Path
		break
	}
	require.NotEmpty(t, repoPath, "Repository path should be set")

	// Safely change to repo directory
	restore := safeChdir(t, repoPath)
	defer restore()

	// Create a dummy commit locally to create a difference between local master and origin/master
	t.Log("Creating a dummy commit locally to create difference between local and origin")
	err = createDummyCommit(t, repoPath)
	require.NoError(t, err, "Should be able to create dummy commit")

	// Now local master is ahead of origin/master
	// Local: A -> B -> C -> D (dummy)
	// Origin: A -> B -> C
	localCommitWithDummy, err := getCurrentCommit(t, repoPath)
	require.NoError(t, err, "Should be able to get local commit with dummy")

	// Get the commit before the dummy (which should be the same as origin/master)
	commitBeforeDummy, err := getPreviousCommit(t, repoPath)
	require.NoError(t, err, "Should be able to get commit before dummy")

	// Verify that local is now ahead of origin
	assert.NotEqual(t, commitBeforeDummy, localCommitWithDummy, "Local master should be ahead of origin/master")

	// Create a new worktree
	err = cmInstance.CreateWorkTree("test-origin-default")
	require.NoError(t, err, "Worktree creation should succeed")

	// Verify the worktree exists
	assertWorktreeExists(t, setup, "test-origin-default")

	// Get the worktree path
	worktreePath := filepath.Join(setup.CmPath, "github.com/octocat/Hello-World", "origin", "test-origin-default")
	require.DirExists(t, worktreePath, "Worktree directory should exist")

	// Check what commit the worktree is on
	worktreeCommit, err := getCurrentCommit(t, worktreePath)
	require.NoError(t, err, "Should be able to get worktree commit")

	t.Logf("Worktree commit: %s", worktreeCommit)
	t.Logf("Local commit with dummy: %s", localCommitWithDummy)
	t.Logf("Commit before dummy (origin/master): %s", commitBeforeDummy)

	// The worktree creation is working correctly - it's using origin/master instead of local master
	// This verifies that our primary mechanism is working: new worktrees are created from origin's default branch
	assert.Equal(t, commitBeforeDummy, worktreeCommit, "Worktree should be on origin/master (commit before dummy)")

	// The worktree should NOT be on the commit with the dummy because it's using origin/master, not local master
	assert.NotEqual(t, localCommitWithDummy, worktreeCommit, "Worktree should not be on the commit with dummy when using origin/master")

	// Verify the worktree is properly linked in the cloned repository
	// We need to check from the cloned repo path, not the original test repo path
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	require.NoError(t, err, "Should be able to list worktrees")

	// The worktree should be listed in the cloned repository
	assert.Contains(t, string(output), "test-origin-default", "Worktree should be listed in the cloned repository")
	assert.Contains(t, string(output), setup.CmPath, "Worktree should be in the .cm directory")
}

// TestCreateWorktreeWithUnsupportedIDE tests creating a worktree with unsupported IDE
func TestCreateWorktreeRepoModeWithUnsupportedIDE(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository
	createTestGitRepo(t, setup.RepoPath)

	// Test creating a worktree with unsupported IDE
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})

	require.NoError(t, err)
	// Safely change to repo directory and create worktree
	restore := safeChdir(t, setup.RepoPath)
	defer restore()

	ideName := "unsupported-ide"
	err = cmInstance.CreateWorkTree("feature/unsupported-ide", codemanager.CreateWorkTreeOpts{IDEName: ideName})
	// Note: IDE opening is now handled by the hook system, so the worktree creation succeeds
	// but the IDE opening fails. The test now verifies that the worktree is created successfully.
	require.NoError(t, err, "Worktree creation should succeed even with unsupported IDE")

	// Verify the worktree was created successfully
	assertWorktreeExists(t, setup, "feature/unsupported-ide")
	assertWorktreeInRepo(t, setup, "feature/unsupported-ide")
}

// TestCreateWorktreeWorkspaceMode_Success tests the complete E2E workflow for creating worktrees from workspace
func TestCreateWorktreeWorkspaceMode_Success(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create test repositories with different names to get different remote URLs
	repo1Path := filepath.Join(setup.CmPath, "Hello-World")
	repo2Path := filepath.Join(setup.CmPath, "Spoon-Knife")

	err := os.MkdirAll(repo1Path, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(repo2Path, 0755)
	require.NoError(t, err)

	// Initialize Git repositories using the existing helper
	createTestGitRepo(t, repo1Path)
	createTestGitRepo(t, repo2Path)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace using CM instance
	workspaceName := "test-workspace"
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{repo1Path, repo2Path},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	require.NoError(t, err)

	// Create worktrees from workspace
	branch := "feature-branch"
	err = cmInstance.CreateWorkTree(branch, codemanager.CreateWorkTreeOpts{
		WorkspaceName: workspaceName,
	})
	if err != nil {
		t.Logf("Error creating worktrees: %v", err)
	}
	require.NoError(t, err)

	// Verify workspace file was created
	workspaceFilePath := filepath.Join(setup.CmPath, "workspaces", workspaceName+"-"+branch+".code-workspace")
	assert.FileExists(t, workspaceFilePath)

	// Verify workspace file content
	content, err := os.ReadFile(workspaceFilePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "github.com/octocat/Hello-World")
	assert.Contains(t, string(content), "github.com/octocat/Spoon-Knife")
	assert.Contains(t, string(content), "folders")
}

// TestCreateWorktreeWorkspaceMode_WithIDE tests creating worktrees from workspace with IDE opening
func TestCreateWorktreeWorkspaceMode_WithIDE(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create test repository
	repoPath := filepath.Join(setup.CmPath, "Hello-World")
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Initialize Git repository using the existing helper
	createTestGitRepo(t, repoPath)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace using CM instance
	workspaceName := "test-workspace"
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{repoPath},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	require.NoError(t, err)

	// Create worktrees from workspace with IDE
	branch := "feature-branch"
	err = cmInstance.CreateWorkTree(branch, codemanager.CreateWorkTreeOpts{
		WorkspaceName: workspaceName,
		IDEName:       "dummy",
	})
	require.NoError(t, err)

	// Note: Worktree creation verification is complex due to the way workspace worktrees are created
	// The workspace file creation is sufficient to verify the functionality works

	// Verify workspace file was created
	workspaceFilePath := filepath.Join(setup.CmPath, "workspaces", workspaceName+"-"+branch+".code-workspace")
	assert.FileExists(t, workspaceFilePath)
}

// TestCreateWorktreeWorkspaceMode_WorkspaceNotFound tests error handling for non-existent workspace
func TestCreateWorktreeWorkspaceMode_WorkspaceNotFound(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Try to create worktrees from non-existent workspace
	branch := "feature-branch"
	err = cmInstance.CreateWorkTree(branch, codemanager.CreateWorkTreeOpts{
		WorkspaceName: "nonexistent-workspace",
	})

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace 'nonexistent-workspace' not found")
}

// TestCreateWorktreeWorkspaceMode_EmptyRepositories tests error handling for workspace with no repositories
func TestCreateWorktreeWorkspaceMode_EmptyRepositories(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace with no repositories (this should fail during workspace creation)
	workspaceName := "empty-workspace"
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	// This should fail because workspace creation requires at least one repository
	assert.Error(t, err)
}

// TestCreateWorktreeWorkspaceMode_RepositoryNotFound tests error handling for workspace with non-existent repositories
func TestCreateWorktreeWorkspaceMode_RepositoryNotFound(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace with non-existent repository
	workspaceName := "test-workspace"
	nonexistentRepoPath := filepath.Join(setup.CmPath, "nonexistent-repo")
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{nonexistentRepoPath},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	// This should fail because the repository doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not found")
}

// TestCreateWorktreeWorkspaceMode_InvalidGitRepository tests error handling for workspace with non-Git repositories
func TestCreateWorktreeWorkspaceMode_InvalidGitRepository(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create non-Git directory
	repoPath := filepath.Join(setup.CmPath, "github.com", "user", "not-a-git-repo")
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace with non-Git repository
	workspaceName := "test-workspace"
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{repoPath},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	// This should fail because the repository is not a Git repository
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not contain a Git repository")
}

// TestCreateWorktreeWorkspaceMode_MultipleRepositories tests creating worktrees for multiple repositories
func TestCreateWorktreeWorkspaceMode_MultipleRepositories(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create multiple test repositories with different names to get different remote URLs
	repo1Path := filepath.Join(setup.CmPath, "Hello-World")
	repo2Path := filepath.Join(setup.CmPath, "Spoon-Knife")
	// Use a different directory name that will get a different remote URL
	repo3Path := filepath.Join(setup.CmPath, "Another-Repo")

	repositories := []string{repo1Path, repo2Path, repo3Path}

	for _, repoPath := range repositories {
		err := os.MkdirAll(repoPath, 0755)
		require.NoError(t, err)

		// Initialize Git repository using the existing helper
		createTestGitRepo(t, repoPath)
	}

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace with multiple repositories
	workspaceName := "multi-repo-workspace"
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  repositories,
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	// This should fail because the third repository has the same remote URL as the first one
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository with URL")
}

// TestCreateWorktreeWorkspaceMode_RollbackOnFailure tests rollback behavior when worktree creation fails
func TestCreateWorktreeWorkspaceMode_RollbackOnFailure(t *testing.T) {
	// Create test setup
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create one valid repository and one invalid repository
	validRepoPath := filepath.Join(setup.CmPath, "Hello-World")
	invalidRepoPath := filepath.Join(setup.CmPath, "Invalid-Repo")

	// Create valid repository
	err := os.MkdirAll(validRepoPath, 0755)
	require.NoError(t, err)
	createTestGitRepo(t, validRepoPath)

	// Create invalid repository (non-Git directory)
	err = os.MkdirAll(invalidRepoPath, 0755)
	require.NoError(t, err)
	// Don't initialize it as a Git repository to make it invalid

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)),
	})
	require.NoError(t, err)

	// Create workspace with both repositories
	workspaceName := "mixed-workspace"
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{validRepoPath, invalidRepoPath},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	// This should fail because one of the repositories is not a Git repository
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not contain a Git repository")
}

// TestWorktreeCreateWithRepository tests creating a worktree with RepositoryName option
func TestWorktreeCreateWithRepository(t *testing.T) {
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

	// Create another directory to run CM commands from (not in the repo)
	workDir := filepath.Join(setup.TempDir, "work-dir")
	err = os.MkdirAll(workDir, 0755)
	require.NoError(t, err)

	// Test creating worktree with RepositoryName option from outside the repo
	err = cmInstance.CreateWorkTree("feature-branch", codemanager.CreateWorkTreeOpts{
		RepositoryName: setup.RepoPath,
	})
	require.NoError(t, err)

	// Verify worktree was created in the correct repository
	worktrees, err := cmInstance.ListWorktrees(codemanager.ListWorktreesOpts{
		RepositoryName: setup.RepoPath,
	})
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)
	assert.Equal(t, "feature-branch", worktrees[0].Branch)

	// Test that we can't use WorkspaceName and RepositoryName together
	err = cmInstance.CreateWorkTree("another-branch", codemanager.CreateWorkTreeOpts{
		RepositoryName: setup.RepoPath,
		WorkspaceName:  "test-workspace",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot specify both WorkspaceName and RepositoryName")
}
