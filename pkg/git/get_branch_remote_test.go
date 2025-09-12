//go:build integration

package git

import (
	"testing"
)

func TestGit_GetBranchRemote(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test getting remote for current branch
	currentBranch, err := git.GetCurrentBranch(".")
	if err != nil {
		t.Fatalf("Expected no error getting current branch: %v", err)
	}

	remote, err := git.GetBranchRemote(".", currentBranch)
	if err != nil {
		t.Errorf("Expected no error getting branch remote: %v", err)
	}
	if remote == "" {
		t.Error("Expected remote name to not be empty")
	}

	// Test getting remote for non-existent branch (should return default remote)
	remote, err = git.GetBranchRemote(".", "non-existent-branch")
	if err != nil {
		t.Errorf("Expected no error getting remote for non-existent branch: %v", err)
	}
	if remote != "origin" {
		t.Errorf("Expected default remote 'origin' for non-existent branch, got: %s", remote)
	}

	// Test in non-existent directory
	_, err = git.GetBranchRemote("/non/existent/directory", currentBranch)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
