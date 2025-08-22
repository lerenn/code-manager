//go:build integration

package git

import (
	"testing"
)

func TestGit_AddRemote(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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

func TestGit_FetchRemote(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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

func TestGit_BranchExistsOnRemote(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a test branch
	testBranchName := "test-remote-branch"
	err := git.CreateBranch(".", testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating branch: %v", err)
	}

	// Test checking if branch exists on origin (may fail due to auth, but shouldn't crash)
	_, err = git.BranchExistsOnRemote(BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "origin",
		Branch:     testBranchName,
	})
	// Note: This might fail due to authentication, but we're testing the command execution
	// The important thing is that it doesn't crash

	// Test checking non-existent branch (may fail due to auth, but shouldn't crash)
	_, err = git.BranchExistsOnRemote(BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "origin",
		Branch:     "non-existent-branch",
	})
	// Note: This might fail due to authentication, but we're testing the command execution
	// The important thing is that it doesn't crash

	// Test in non-existent directory
	_, err = git.BranchExistsOnRemote(BranchExistsOnRemoteParams{
		RepoPath:   "/non/existent/directory",
		RemoteName: "origin",
		Branch:     testBranchName,
	})
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGit_GetRemoteURL(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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

func TestGit_RemoteExists(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
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
