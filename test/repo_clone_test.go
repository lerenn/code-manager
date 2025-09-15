//go:build e2e

package test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cloneRepository clones a repository using the CM instance
func cloneRepository(t *testing.T, setup *TestSetup, repoURL string, recursive bool) error {
	t.Helper()

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath),
	})

	require.NoError(t, err)
	// Create clone options
	opts := codemanager.CloneOpts{
		Recursive: recursive,
	}

	err = cmInstance.Clone(repoURL, opts)
	if err != nil {
		t.Logf("Clone failed for URL %s: %v", repoURL, err)
	}
	return err
}

// TestCloneRepositoryRepoModeSuccess tests successful cloning of a repository
func TestCloneRepositoryRepoModeSuccess(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Clone a real public repository
	repoURL := "https://github.com/octocat/Hello-World.git"

	// Clone the repository
	err := cloneRepository(t, setup, repoURL, true)
	require.NoError(t, err, "Clone should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Check that we have one repository entry
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Get the repository entry
	var normalizedURL string
	var repo Repository
	for url, r := range status.Repositories {
		normalizedURL = url
		repo = r
		break
	}

	assert.NotEmpty(t, normalizedURL, "Repository URL should be set")
	assert.NotEmpty(t, repo.Path, "Repository path should be set")

	// Check that the repository has remotes
	require.True(t, len(repo.Remotes) > 0, "Repository should have remotes")
	assert.Contains(t, repo.Remotes, "origin", "Repository should have origin remote")

	// Check that the origin remote has a default branch
	originRemote := repo.Remotes["origin"]
	assert.NotEmpty(t, originRemote.DefaultBranch, "Origin remote should have default branch")

	// Verify the cloned repository exists
	assert.DirExists(t, repo.Path, "Cloned repository should exist")

	// Verify it's a valid Git repository
	gitDir := filepath.Join(repo.Path, ".git")
	assert.DirExists(t, gitDir, "Cloned repository should be a valid Git repository")

	// Verify the default branch is checked out
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repo.Path
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, originRemote.DefaultBranch, strings.TrimSpace(string(output)), "Should be on default branch")

	// Verify the repository has content
	readmePath := filepath.Join(repo.Path, "README")
	assert.FileExists(t, readmePath, "Repository should have README file")
}

// TestCloneRepositoryRepoModeShallow tests shallow cloning of a repository
func TestCloneRepositoryRepoModeShallow(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Clone a real public repository with shallow option
	repoURL := "https://github.com/octocat/Hello-World.git"

	// Clone the repository with shallow option
	err := cloneRepository(t, setup, repoURL, false)
	require.NoError(t, err, "Shallow clone should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Check that we have one repository entry
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Get the repository entry
	var normalizedURL string
	var repo Repository
	for url, r := range status.Repositories {
		normalizedURL = url
		repo = r
		break
	}

	assert.NotEmpty(t, normalizedURL, "Repository URL should be set")
	assert.NotEmpty(t, repo.Path, "Repository path should be set")

	// Verify the cloned repository exists
	assert.DirExists(t, repo.Path, "Cloned repository should exist")

	// Verify it's a valid Git repository
	gitDir := filepath.Join(repo.Path, ".git")
	assert.DirExists(t, gitDir, "Cloned repository should be a valid Git repository")

	// For shallow clones, we can't easily verify the depth, but we can verify it's a valid repo
	cmd := exec.Command("git", "status")
	cmd.Dir = repo.Path
	_, err = cmd.Output()
	require.NoError(t, err, "Shallow clone should be a valid Git repository")
}

// TestCloneRepositoryRepoModeAlreadyExists tests cloning a repository that already exists
func TestCloneRepositoryRepoModeAlreadyExists(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Use a real public repository
	repoURL := "https://github.com/octocat/Hello-World.git"

	// Clone the repository first time
	err := cloneRepository(t, setup, repoURL, true)
	require.NoError(t, err, "First clone should succeed")

	// Try to clone the same repository again
	err = cloneRepository(t, setup, repoURL, true)
	require.Error(t, err, "Second clone should fail")
	assert.ErrorIs(t, err, codemanager.ErrRepositoryExists, "Error should indicate repository already exists")

	// Verify only one repository entry exists
	status := readStatusFile(t, setup.StatusPath)
	require.Len(t, status.Repositories, 1, "Should still have only one repository entry")
}

// TestCloneRepositoryRepoModeInvalidURL tests cloning with an invalid URL
func TestCloneRepositoryRepoModeInvalidURL(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Try to clone with an invalid URL
	err := cloneRepository(t, setup, "not-a-valid-url", true)
	require.Error(t, err, "Clone should fail with invalid URL")
	assert.ErrorIs(t, err, codemanager.ErrUnsupportedRepositoryURLFormat, "Error should indicate invalid URL format")
}

// TestCloneRepositoryRepoModeEmptyURL tests cloning with an empty URL
func TestCloneRepositoryRepoModeEmptyURL(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Try to clone with an empty URL
	err := cloneRepository(t, setup, "", true)
	require.Error(t, err, "Clone should fail with empty URL")
	assert.ErrorIs(t, err, codemanager.ErrRepositoryURLEmpty, "Error should indicate empty URL")
}

// TestCloneRepositoryRepoModeHTTPSURL tests cloning with HTTPS URL format
func TestCloneRepositoryRepoModeHTTPSURL(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Clone using HTTPS URL
	repoURL := "https://github.com/octocat/Hello-World.git"

	err := cloneRepository(t, setup, repoURL, true)
	require.NoError(t, err, "Clone with HTTPS URL should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Check that we have one repository entry
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Get the repository entry
	var normalizedURL string
	var repo Repository
	for url, r := range status.Repositories {
		normalizedURL = url
		repo = r
		break
	}

	assert.NotEmpty(t, normalizedURL, "Repository URL should be set")
	assert.NotEmpty(t, repo.Path, "Repository path should be set")

	// Verify the cloned repository exists
	assert.DirExists(t, repo.Path, "Cloned repository should exist")
}

// TestCloneRepositoryRepoModeWithDotGitSuffix tests cloning with .git suffix
func TestCloneRepositoryRepoModeWithDotGitSuffix(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Clone with .git suffix
	repoURL := "https://github.com/octocat/Hello-World.git"

	err := cloneRepository(t, setup, repoURL, true)
	require.NoError(t, err, "Clone with .git suffix should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Check that we have one repository entry
	require.Len(t, status.Repositories, 1, "Should have one repository entry")

	// Get the repository entry
	var normalizedURL string
	var repo Repository
	for url, r := range status.Repositories {
		normalizedURL = url
		repo = r
		break
	}

	assert.NotEmpty(t, normalizedURL, "Repository URL should be set")
	assert.NotEmpty(t, repo.Path, "Repository path should be set")

	// Verify the cloned repository exists
	assert.DirExists(t, repo.Path, "Cloned repository should exist")
}

// TestCloneRepositoryRepoModeDefaultBranchDetection tests that default branch is correctly detected
func TestCloneRepositoryRepoModeDefaultBranchDetection(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Use a real public repository
	repoURL := "https://github.com/octocat/Hello-World.git"

	// Clone the repository
	err := cloneRepository(t, setup, repoURL, true)
	require.NoError(t, err, "Clone should succeed")

	// Verify the status.yaml file was created and updated
	status := readStatusFile(t, setup.StatusPath)
	require.NotNil(t, status.Repositories, "Status file should have repositories section")

	// Get the repository entry
	var repo Repository
	for _, r := range status.Repositories {
		repo = r
		break
	}

	// Check that the origin remote has a default branch
	require.Contains(t, repo.Remotes, "origin", "Repository should have origin remote")
	originRemote := repo.Remotes["origin"]
	assert.NotEmpty(t, originRemote.DefaultBranch, "Origin remote should have default branch")

	// Verify the default branch is checked out in the cloned repository
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repo.Path
	output, err := cmd.Output()
	require.NoError(t, err)
	actualBranch := strings.TrimSpace(string(output))
	assert.Equal(t, originRemote.DefaultBranch, actualBranch, "Should be on the detected default branch")

	// Verify the default branch is either 'main' or 'master' (common default branches)
	assert.True(t, actualBranch == "main" || actualBranch == "master", "Default branch should be main or master")
}

// TestCloneRepositoryRepoModeSSHURL tests cloning with SSH URL format (if SSH is available)
func TestCloneRepositoryRepoModeSSHURL(t *testing.T) {
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Test SSH URL format (this will likely fail in CI environments without SSH keys)
	// but it's good to test the URL parsing logic
	repoURL := "git@github.com:octocat/Hello-World.git"

	err := cloneRepository(t, setup, repoURL, true)
	// This might fail due to SSH authentication, but the URL parsing should work
	if err != nil {
		// If it fails due to SSH auth, that's expected in test environments
		assert.Contains(t, err.Error(), "ssh", "Should fail due to SSH authentication")
	} else {
		// If it succeeds, verify the repository was cloned correctly
		status := readStatusFile(t, setup.StatusPath)
		require.NotNil(t, status.Repositories, "Status file should have repositories section")
		require.Len(t, status.Repositories, 1, "Should have one repository entry")
	}
}
