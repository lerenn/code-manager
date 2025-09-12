//go:build integration

package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_GetWorktreePath(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test creating a worktree and then getting its path
	testBranchName := "test-get-path-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
	testWorktreePath := filepath.Join(".", "test-get-path-"+strings.ReplaceAll(t.Name(), "/", "-"))

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

	// Get the worktree path
	retrievedPath, err := git.GetWorktreePath(".", testBranchName)
	if err != nil {
		t.Fatalf("Expected no error getting worktree path: %v", err)
	}

	// Verify the path matches (use absolute paths for comparison)
	absTestPath, err := filepath.Abs(testWorktreePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	absRetrievedPath, err := filepath.Abs(retrievedPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if absRetrievedPath != absTestPath {
		t.Errorf("Expected worktree path %s, got %s", absTestPath, absRetrievedPath)
	}

	// Test getting path for non-existent worktree
	_, err = git.GetWorktreePath(".", "non-existent-branch")
	if err == nil {
		t.Error("Expected error when getting path for non-existent worktree")
	}

	// Test in non-existent directory
	_, err = git.GetWorktreePath("/non/existent/directory", testBranchName)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	// Clean up worktree directory
	os.RemoveAll(testWorktreePath)
}
