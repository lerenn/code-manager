//go:build integration

package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_CreateWorktree(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test creating a worktree
	testBranchName := "test-worktree-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
	testWorktreePath := filepath.Join(".", "test-worktree-"+strings.ReplaceAll(t.Name(), "/", "-"))

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

	// Verify the worktree exists in git
	exists, err := git.WorktreeExists(".", testBranchName)
	if err != nil {
		t.Errorf("Expected no error checking worktree existence: %v", err)
	}
	if !exists {
		t.Error("Expected worktree to exist in git")
	}

	// Test creating the same worktree again (should fail)
	err = git.CreateWorktree(".", testWorktreePath, testBranchName)
	if err == nil {
		t.Error("Expected error when creating duplicate worktree")
	}

	// Clean up worktree directory
	os.RemoveAll(testWorktreePath)

	// Test in non-existent directory
	err = git.CreateWorktree("/non/existent/directory", "/tmp/test-worktree", "main")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
