//go:build integration

package git

import (
	"strings"
	"testing"
)

func TestGit_CreateBranch(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
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
