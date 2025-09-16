//go:build integration

package git

import (
	"testing"
)

func TestGit_GetCurrentBranch(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test in the temporary git repository
	branch, err := git.GetCurrentBranch(".")
	if err != nil {
		t.Fatalf("Expected no error in git repository: %v", err)
	}

	// Should be on main or master branch
	if branch != "main" && branch != "master" {
		t.Errorf("Expected 'main' or 'master' branch, got: %q", branch)
	}

	// Test in non-existent directory
	_, err = git.GetCurrentBranch("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
