//go:build integration

package git

import (
	"testing"
)

func TestGit_GetRemoteURL(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test getting origin remote URL
	url, err := git.GetRemoteURL(".", "origin")
	if err != nil {
		t.Fatalf("Expected no error getting origin remote URL: %v", err)
	}
	if url == "" {
		t.Error("Expected origin remote URL to not be empty")
	}

	// Test getting non-existent remote URL
	_, err = git.GetRemoteURL(".", "non-existent-remote")
	if err == nil {
		t.Error("Expected error when getting non-existent remote URL")
	}

	// Test in non-existent directory
	_, err = git.GetRemoteURL("/non/existent/directory", "origin")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
