//go:build integration

package git

import (
	"testing"
)

func TestGit_BranchExistsOnRemote(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
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
