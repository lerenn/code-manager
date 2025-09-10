//go:build integration

package git

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckReferenceConflict(t *testing.T) {
	git := NewGit()

	t.Run("no conflict for simple branch name", func(t *testing.T) {
		tempDir := setupRefConflictTestRepo(t)
		err := git.CheckReferenceConflict(tempDir, "feature-branch")
		assert.NoError(t, err)
	})

	t.Run("no conflict for nested branch name without existing parent", func(t *testing.T) {
		tempDir := setupRefConflictTestRepo(t)
		err := git.CheckReferenceConflict(tempDir, "feature/new-feature")
		assert.NoError(t, err)
	})

	t.Run("conflict when parent branch exists", func(t *testing.T) {
		tempDir := setupRefConflictTestRepo(t)

		// Create a branch called "feat"
		cmd := exec.Command("git", "branch", "feat")
		cmd.Dir = tempDir
		require.NoError(t, cmd.Run())

		// Try to create a branch called "feat/test" - should conflict
		err := git.CheckReferenceConflict(tempDir, "feat/test")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrBranchParentExists)
	})

	t.Run("conflict when parent tag exists", func(t *testing.T) {
		tempDir := setupRefConflictTestRepo(t)

		// Create a tag called "feature"
		cmd := exec.Command("git", "tag", "feature")
		cmd.Dir = tempDir
		require.NoError(t, cmd.Run())

		// Try to create a branch called "feature/branch" - should conflict
		err := git.CheckReferenceConflict(tempDir, "feature/branch")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrTagParentExists)
	})

	t.Run("no conflict for deeply nested branch without conflicts", func(t *testing.T) {
		tempDir := setupRefConflictTestRepo(t)
		err := git.CheckReferenceConflict(tempDir, "feature/subfeature/implementation")
		assert.NoError(t, err)
	})

	t.Run("conflict with deeply nested existing reference", func(t *testing.T) {
		tempDir := setupRefConflictTestRepo(t)

		// Create a branch called "feature/subfeature"
		cmd := exec.Command("git", "branch", "feature/subfeature")
		cmd.Dir = tempDir
		require.NoError(t, cmd.Run())

		// Try to create a branch called "feature/subfeature/implementation" - should conflict
		err := git.CheckReferenceConflict(tempDir, "feature/subfeature/implementation")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrBranchParentExists)
	})
}

// setupRefConflictTestRepo creates a temporary Git repository for testing
func setupRefConflictTestRepo(t *testing.T) string {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "git-ref-conflict-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Initialize a Git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())

	// Configure Git user (required for commits)
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())

	// Create an initial commit
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = tempDir
	require.NoError(t, cmd.Run())

	return tempDir
}
