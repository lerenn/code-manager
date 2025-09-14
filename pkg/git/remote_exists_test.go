//go:build integration

package git

import (
	"testing"
)

func TestGit_RemoteExists(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test checking if origin exists (should exist)
	exists, err := git.RemoteExists(".", "origin")
	if err != nil {
		t.Fatalf("Expected no error checking origin existence: %v", err)
	}
	if !exists {
		t.Error("Expected origin remote to exist")
	}

	// Test checking non-existent remote
	exists, err = git.RemoteExists(".", "non-existent-remote")
	if err != nil {
		t.Fatalf("Expected no error checking non-existent remote: %v", err)
	}
	if exists {
		t.Error("Expected non-existent remote to not exist")
	}

	// Test in non-existent directory
	_, err = git.RemoteExists("/non/existent/directory", "origin")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
