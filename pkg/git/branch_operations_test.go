//go:build integration

package git

import (
	"strings"
	"testing"
)

func TestGit_BranchExists(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test with current branch (should exist)
	currentBranch, err := git.GetCurrentBranch(".")
	if err != nil {
		t.Fatalf("Expected no error getting current branch: %v", err)
	}

	exists, err := git.BranchExists(".", currentBranch)
	if err != nil {
		t.Errorf("Expected no error checking current branch existence: %v", err)
	}
	if !exists {
		t.Errorf("Expected current branch %s to exist", currentBranch)
	}

	// Test with non-existent branch
	exists, err = git.BranchExists(".", "non-existent-branch-12345")
	if err != nil {
		t.Errorf("Expected no error checking non-existent branch: %v", err)
	}
	if exists {
		t.Error("Expected non-existent branch to not exist")
	}

	// Test in non-existent directory
	_, err = git.BranchExists("/non/existent/directory", "main")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGit_CreateBranch(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test creating a new branch
	testBranchName := "test-branch-integration-" + strings.ReplaceAll(t.Name(), "/", "-")

	// Create a new branch
	err := git.CreateBranch(".", testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating branch: %v", err)
	}

	// Verify the branch was created
	exists, err := git.BranchExists(".", testBranchName)
	if err != nil {
		t.Errorf("Expected no error checking created branch existence: %v", err)
	}
	if !exists {
		t.Errorf("Expected created branch %s to exist", testBranchName)
	}

	// Test creating the same branch again (should fail)
	err = git.CreateBranch(".", testBranchName)
	if err == nil {
		t.Error("Expected error when creating duplicate branch")
	}

	// Test in non-existent directory
	err = git.CreateBranch("/non/existent/directory", "test-branch")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
