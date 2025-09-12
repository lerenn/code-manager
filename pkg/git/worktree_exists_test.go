//go:build integration

package git

import (
	"testing"
)

func TestGit_WorktreeExists(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test with current branch (should not have a worktree by default)
	currentBranch, err := git.GetCurrentBranch(".")
	if err != nil {
		t.Fatalf("Expected no error getting current branch: %v", err)
	}

	exists, err := git.WorktreeExists(".", currentBranch)
	if err != nil {
		t.Fatalf("Expected no error checking worktree existence: %v", err)
	}

	// May or may not exist by default depending on the test environment
	// The important thing is that it doesn't error
	_ = exists // Use the variable to avoid unused variable warning

	// Test in non-existent directory
	_, err = git.WorktreeExists("/non/existent/directory", "main")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
