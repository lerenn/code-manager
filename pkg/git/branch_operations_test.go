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

func TestGit_CreateBranchFrom(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a base branch first
	baseBranchName := "base-branch-integration-" + strings.ReplaceAll(t.Name(), "/", "-")
	err := git.CreateBranchFrom(CreateBranchFromParams{
		RepoPath:   ".",
		NewBranch:  baseBranchName,
		FromBranch: "HEAD",
	})
	if err != nil {
		t.Fatalf("Expected no error creating base branch: %v", err)
	}

	// Test creating a new branch from the base branch
	testBranchName := "test-branch-from-" + strings.ReplaceAll(t.Name(), "/", "-")

	err = git.CreateBranchFrom(CreateBranchFromParams{
		RepoPath:   ".",
		NewBranch:  testBranchName,
		FromBranch: baseBranchName,
	})
	if err != nil {
		t.Fatalf("Expected no error creating branch from base branch: %v", err)
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
	err = git.CreateBranchFrom(CreateBranchFromParams{
		RepoPath:   ".",
		NewBranch:  testBranchName,
		FromBranch: baseBranchName,
	})
	if err == nil {
		t.Error("Expected error when creating duplicate branch")
	}

	// Test creating branch from non-existent branch
	err = git.CreateBranchFrom(CreateBranchFromParams{
		RepoPath:   ".",
		NewBranch:  "test-branch-from-non-existent",
		FromBranch: "non-existent-branch-12345",
	})
	if err == nil {
		t.Error("Expected error when creating branch from non-existent branch")
	}

	// Test in non-existent directory
	err = git.CreateBranchFrom(CreateBranchFromParams{
		RepoPath:   "/non/existent/directory",
		NewBranch:  "test-branch",
		FromBranch: "main",
	})
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
