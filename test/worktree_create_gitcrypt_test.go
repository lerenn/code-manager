//go:build e2e

// Package test provides end-to-end tests for git-crypt hook functionality.
// These tests require git-crypt to be installed and will be skipped in environments
// where git-crypt is not available (e.g., CI/Dagger environments).
package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/hooks/gitcrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateWorktreeWithGitCryptRepoModeIntegration tests the git-crypt hook functionality in isolation
func TestCreateWorktreeWithGitCryptRepoModeIntegration(t *testing.T) {
	// Fail test if git-crypt is not available
	if !isGitCryptAvailable() {
		t.Fatal("git-crypt is not available on this system - this is a required dependency")
	}

	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository with git-crypt
	createGitCryptRepo(t, setup.RepoPath)

	// Test the git-crypt hook directly
	testGitCryptHook(t, setup.RepoPath)
}

// TestCreateWorktreeWithGitCryptRepoModeWithMissingKey tests the git-crypt hook when key is missing
func TestCreateWorktreeWithGitCryptRepoModeWithMissingKey(t *testing.T) {
	// Fail test if git-crypt is not available
	if !isGitCryptAvailable() {
		t.Fatal("git-crypt is not available on this system - this is a required dependency")
	}

	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create a test Git repository with git-crypt but without key
	createGitCryptRepoWithoutKey(t, setup.RepoPath)

	// Test the git-crypt hook with missing key
	testGitCryptHookWithMissingKey(t, setup.RepoPath)
}

// isGitCryptAvailable checks if git-crypt is available on the system
func isGitCryptAvailable() bool {
	_, err := exec.LookPath("git-crypt")
	return err == nil
}

// createGitCryptRepo creates a Git repository with git-crypt initialized and an encrypted file
func createGitCryptRepo(t *testing.T, repoPath string) {
	t.Helper()

	// Safely change to the repository directory for setup
	restore := safeChdir(t, repoPath)
	defer restore()

	// Set up Git environment variables for all commands
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	// Initialize Git repository
	cmd := exec.Command("git", "init")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Set the default branch to master
	cmd = exec.Command("git", "branch", "-M", "master")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Add a remote origin
	cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/octocat/Hello-World.git")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Initialize git-crypt
	cmd = exec.Command("git-crypt", "init")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create .gitattributes file to specify which files should be encrypted
	gitattributesContent := `# Encrypt sensitive files
*.secret filter=git-crypt diff=git-crypt
secrets/ filter=git-crypt diff=git-crypt
`
	gitattributesPath := filepath.Join(repoPath, ".gitattributes")
	require.NoError(t, os.WriteFile(gitattributesPath, []byte(gitattributesContent), 0644))

	// Create a secret file that will be encrypted
	secretContent := "This is a secret file that should be encrypted by git-crypt"
	secretPath := filepath.Join(repoPath, "secrets", "database.secret")
	require.NoError(t, os.MkdirAll(filepath.Dir(secretPath), 0755))
	require.NoError(t, os.WriteFile(secretPath, []byte(secretContent), 0644))

	// Create a regular file that should not be encrypted
	readmePath := filepath.Join(repoPath, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test Repository\n\nThis is a test repository with git-crypt.\n"), 0644))

	// Add all files
	cmd = exec.Command("git", "add", ".")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Commit the files
	cmd = exec.Command("git", "commit", "-m", "Initial commit with git-crypt")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create a test branch
	cmd = exec.Command("git", "checkout", "-b", "feature/test-branch")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Create another secret file in the feature branch
	featureSecretContent := "This is a feature branch secret file"
	featureSecretPath := filepath.Join(repoPath, "secrets", "feature.secret")
	require.NoError(t, os.WriteFile(featureSecretPath, []byte(featureSecretContent), 0644))

	// Add and commit the feature secret
	cmd = exec.Command("git", "add", "secrets/feature.secret")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Add feature secret")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Switch back to master branch
	cmd = exec.Command("git", "checkout", "master")
	cmd.Env = gitEnv
	require.NoError(t, cmd.Run())

	// Verify that the secret file is encrypted in the repository
	verifySecretFileEncrypted(t, repoPath, "secrets/database.secret")
}

// createGitCryptRepoWithoutKey creates a Git repository with git-crypt but removes the key
func createGitCryptRepoWithoutKey(t *testing.T, repoPath string) {
	t.Helper()

	// First create a normal git-crypt repo
	createGitCryptRepo(t, repoPath)

	// Remove the git-crypt key to simulate a scenario where the key is not available
	keyPath := filepath.Join(repoPath, ".git", "git-crypt", "keys", "default")
	if _, err := os.Stat(keyPath); err == nil {
		require.NoError(t, os.Remove(keyPath))
	}
}

// verifySecretFileEncrypted verifies that a file is encrypted in the repository
func verifySecretFileEncrypted(t *testing.T, repoPath, filePath string) {
	t.Helper()

	// Read the file from the git index to see if it's encrypted
	cmd := exec.Command("git", "show", "HEAD:"+filePath)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	require.NoError(t, err)

	// The file should be encrypted (not contain the original content)
	content := string(output)
	assert.NotContains(t, content, "This is a secret file", "File should be encrypted in repository")
	assert.NotContains(t, content, "This is a feature branch secret file", "File should be encrypted in repository")
}

// testGitCryptHook tests the git-crypt hook functionality directly
func testGitCryptHook(t *testing.T, repoPath string) {
	t.Helper()

	// Create a temporary worktree directory
	worktreeDir, err := os.MkdirTemp("", "gitcrypt-test-worktree-*")
	require.NoError(t, err)
	defer os.RemoveAll(worktreeDir)

	// First, create the worktree with --no-checkout (this is what CM would do)
	cmd := exec.Command("git", "worktree", "add", "--no-checkout", worktreeDir, "feature/test-branch")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Now test the git-crypt hook
	hook := gitcrypt.NewWorktreeCheckoutHook()

	// Create a hook context
	ctx := &hooks.HookContext{
		OperationName: "CreateWorkTree",
		Parameters: map[string]interface{}{
			"worktreePath": worktreeDir,
			"repoPath":     repoPath,
			"branch":       "feature/test-branch",
		},
		Results:  make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	// Test the hook
	err = hook.OnWorktreeCheckout(ctx)
	require.NoError(t, err, "Git-crypt hook should succeed")

	// Now checkout the branch (this is where git-crypt would normally fail without the hook)
	cmd = exec.Command("git", "checkout", "feature/test-branch")
	cmd.Dir = worktreeDir
	require.NoError(t, cmd.Run())

	// Verify that the worktree directory has git-crypt setup (after checkout)
	assertGitCryptSetupInWorktree(t, worktreeDir)

	// Verify that encrypted files are properly decrypted
	assertDecryptedFilesInWorktree(t, worktreeDir)
}

// testGitCryptHookWithMissingKey tests the git-crypt hook when the key is missing
func testGitCryptHookWithMissingKey(t *testing.T, repoPath string) {
	t.Helper()

	// Create a temporary worktree directory
	worktreeDir, err := os.MkdirTemp("", "gitcrypt-test-worktree-*")
	require.NoError(t, err)
	defer os.RemoveAll(worktreeDir)

	// Create the git-crypt hook
	hook := gitcrypt.NewWorktreeCheckoutHook()

	// Create a hook context
	ctx := &hooks.HookContext{
		OperationName: "CreateWorkTree",
		Parameters: map[string]interface{}{
			"worktreePath": worktreeDir,
			"repoPath":     repoPath,
			"branch":       "feature/test-branch",
		},
		Results:  make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	// Test the hook - this should fail because the key is missing
	err = hook.OnWorktreeCheckout(ctx)
	require.Error(t, err, "Git-crypt hook should fail when key is missing")
	assert.Contains(t, err.Error(), "git-crypt key not found", "Error should mention git-crypt key not found")
}

// assertGitCryptSetupInWorktree verifies that git-crypt is properly set up in the worktree
func assertGitCryptSetupInWorktree(t *testing.T, worktreePath string) {
	t.Helper()

	// For worktrees, the .git directory is a file that points to the main repository
	// We need to check the git-crypt setup in the main repository's worktree directory
	// First, let's find the actual git directory
	gitFile := filepath.Join(worktreePath, ".git")
	if _, err := os.Stat(gitFile); err == nil {
		// Read the .git file to get the actual git directory
		content, err := os.ReadFile(gitFile)
		require.NoError(t, err, "Should be able to read .git file")

		// The content should be something like "gitdir: /path/to/main/repo/.git/worktrees/worktree-name"
		gitDirLine := strings.TrimSpace(string(content))
		if strings.HasPrefix(gitDirLine, "gitdir: ") {
			actualGitDir := strings.TrimPrefix(gitDirLine, "gitdir: ")

			// Check that the worktree's git directory has git-crypt key
			keyPath := filepath.Join(actualGitDir, "git-crypt", "keys", "default")
			assert.FileExists(t, keyPath, "Git-crypt key should exist in worktree git directory")
		}
	}

	// Check that .gitattributes exists in the worktree
	gitattributesPath := filepath.Join(worktreePath, ".gitattributes")
	assert.FileExists(t, gitattributesPath, ".gitattributes should exist in worktree")

	// Verify git-crypt status
	cmd := exec.Command("git-crypt", "status")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	require.NoError(t, err, "git-crypt status should work in worktree")

	// The output should indicate that git-crypt is unlocked
	statusOutput := string(output)
	assert.Contains(t, statusOutput, "not encrypted", "Git-crypt should be unlocked in worktree")
}

// assertDecryptedFilesInWorktree verifies that encrypted files are properly decrypted in the worktree
func assertDecryptedFilesInWorktree(t *testing.T, worktreePath string) {
	t.Helper()

	// Check that the secret file is decrypted and contains the expected content
	secretPath := filepath.Join(worktreePath, "secrets", "database.secret")
	assert.FileExists(t, secretPath, "Secret file should exist in worktree")

	// Read the file content
	content, err := os.ReadFile(secretPath)
	require.NoError(t, err, "Should be able to read secret file")

	// The content should be decrypted (contain the original text)
	contentStr := string(content)
	assert.Contains(t, contentStr, "This is a secret file", "Secret file should be decrypted in worktree")

	// Check that the feature secret file is also decrypted
	featureSecretPath := filepath.Join(worktreePath, "secrets", "feature.secret")
	assert.FileExists(t, featureSecretPath, "Feature secret file should exist in worktree")

	featureContent, err := os.ReadFile(featureSecretPath)
	require.NoError(t, err, "Should be able to read feature secret file")

	featureContentStr := string(featureContent)
	assert.Contains(t, featureContentStr, "This is a feature branch secret file", "Feature secret file should be decrypted in worktree")

	// Check that regular files are not affected
	readmePath := filepath.Join(worktreePath, "README.md")
	assert.FileExists(t, readmePath, "README should exist in worktree")

	readmeContent, err := os.ReadFile(readmePath)
	require.NoError(t, err, "Should be able to read README file")

	readmeContentStr := string(readmeContent)
	assert.Contains(t, readmeContentStr, "This is a test repository with git-crypt", "README should contain expected content")
}
