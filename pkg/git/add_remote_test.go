//go:build integration

package git

import (
	"testing"
)

func TestGit_AddRemote(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test adding a new remote
	remoteName := "test-remote"
	remoteURL := "https://github.com/octocat/Hello-World.git"

	err := git.AddRemote(".", remoteName, remoteURL)
	if err != nil {
		t.Fatalf("Expected no error adding remote: %v", err)
	}

	// Verify the remote was added
	exists, err := git.RemoteExists(".", remoteName)
	if err != nil {
		t.Fatalf("Expected no error checking remote existence: %v", err)
	}
	if !exists {
		t.Error("Expected remote to exist after adding")
	}

	// Verify the remote URL
	retrievedURL, err := git.GetRemoteURL(".", remoteName)
	if err != nil {
		t.Fatalf("Expected no error getting remote URL: %v", err)
	}
	if retrievedURL != remoteURL {
		t.Errorf("Expected remote URL %s, got %s", remoteURL, retrievedURL)
	}

	// Test adding duplicate remote (should fail)
	err = git.AddRemote(".", remoteName, "https://github.com/otheruser/otherrepo.git")
	if err == nil {
		t.Error("Expected error when adding duplicate remote")
	}

	// Test in non-existent directory
	err = git.AddRemote("/non/existent/directory", remoteName, remoteURL)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
