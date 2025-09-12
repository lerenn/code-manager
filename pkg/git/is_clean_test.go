//go:build integration

package git

import (
	"testing"
)

func TestGit_IsClean(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test in the temporary git repository
	isClean, err := git.IsClean(".")
	if err != nil {
		t.Fatalf("Expected no error in git repository: %v", err)
	}

	// Currently always returns true as it's a placeholder
	if !isClean {
		t.Error("Expected clean state (placeholder implementation)")
	}

	// Test in non-existent directory
	_, err = git.IsClean("/non/existent/directory")
	if err != nil {
		t.Errorf("Expected no error for non-existent directory (placeholder implementation), got: %v", err)
	}
}
