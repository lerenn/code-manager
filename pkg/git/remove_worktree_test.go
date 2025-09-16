//go:build integration

package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_RemoveWorktree(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test creating and then removing a worktree
	testBranchName := "test-remove-worktree-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
	testWorktreePath := filepath.Join(".", "test-remove-worktree-"+strings.ReplaceAll(t.Name(), "/", "-"))

	// Create a test branch first
	err := git.CreateBranch(".", testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating branch: %v", err)
	}

	// Create a worktree
	err = git.CreateWorktree(".", testWorktreePath, testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating worktree: %v", err)
	}

	// Verify the worktree was created
	if _, err := os.Stat(testWorktreePath); os.IsNotExist(err) {
		t.Errorf("Expected worktree directory %s to exist", testWorktreePath)
	}

	// Remove the worktree
	err = git.RemoveWorktree(".", testWorktreePath, false)
	if err != nil {
		t.Fatalf("Expected no error removing worktree: %v", err)
	}

	// Verify the worktree directory was removed
	if _, err := os.Stat(testWorktreePath); !os.IsNotExist(err) {
		t.Errorf("Expected worktree directory %s to be removed", testWorktreePath)
	}

	// Verify the worktree no longer exists in git
	exists, err := git.WorktreeExists(".", testBranchName)
	if err != nil {
		t.Errorf("Expected no error checking worktree existence: %v", err)
	}
	if exists {
		t.Error("Expected worktree to not exist in git after removal")
	}

	// Test removing non-existent worktree
	err = git.RemoveWorktree(".", "/non/existent/worktree", false)
	if err == nil {
		t.Error("Expected error when removing non-existent worktree")
	}

	// Test in non-existent directory
	err = git.RemoveWorktree("/non/existent/directory", "/tmp/test-worktree", false)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
