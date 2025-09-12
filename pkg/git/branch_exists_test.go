//go:build integration

package git

import (
	"testing"
)

func TestGit_BranchExists(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
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
