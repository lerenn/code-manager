//go:build integration

package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_WorktreeExists(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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

func TestGit_CreateWorktree(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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

func TestGit_RemoveWorktree(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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
	err = git.RemoveWorktree(".", testWorktreePath)
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
	err = git.RemoveWorktree(".", "/non/existent/worktree")
	if err == nil {
		t.Error("Expected error when removing non-existent worktree")
	}

	// Test in non-existent directory
	err = git.RemoveWorktree("/non/existent/directory", "/tmp/test-worktree")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGit_GetWorktreePath(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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
