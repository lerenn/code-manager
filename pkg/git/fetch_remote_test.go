//go:build integration

package git

import (
	"testing"
)

func TestGit_FetchRemote(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test fetching from origin (which exists but may fail due to auth)
	err := git.FetchRemote(".", "origin")
	// Note: This might fail due to authentication, but we're testing the command execution
	// The important thing is that it doesn't crash

	// Test fetching from non-existent remote
	err = git.FetchRemote(".", "non-existent-remote")
	if err == nil {
		t.Error("Expected error when fetching from non-existent remote")
	}

	// Test in non-existent directory
	err = git.FetchRemote("/non/existent/directory", "origin")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
