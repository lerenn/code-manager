//go:build integration

package git

import (
	"os/exec"
	"strings"
	"testing"
)

func TestGit_SetUpstreamBranch(t *testing.T) {
	git := NewGit()
	repoPath, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test setting upstream for a branch that exists on remote
	t.Run("existing_remote_branch", func(t *testing.T) {
		// Create a local branch
		branchName := "test-upstream-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
		err := git.CreateBranch(repoPath, branchName)
		if err != nil {
			t.Fatalf("Expected no error creating branch: %v", err)
		}

		// Checkout the branch
		err = git.CheckoutBranch(repoPath, branchName)
		if err != nil {
			t.Fatalf("Expected no error checking out branch: %v", err)
		}

		// Create a remote reference to simulate the branch exists on remote
		cmd := exec.Command("git", "update-ref", "refs/remotes/origin/"+branchName, "HEAD")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create remote reference: %v", err)
		}

		// Set upstream branch - this should succeed and configure push settings
		err = git.SetUpstreamBranch(repoPath, "origin", branchName)
		if err != nil {
			t.Errorf("Expected no error setting upstream for existing remote branch: %v", err)
		}

		// Verify push configuration is set correctly
		cmd = exec.Command("git", "config", "--get", "branch."+branchName+".pushRemote")
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Expected no error getting pushRemote config: %v", err)
		}
		expectedPushRemote := "origin"
		actualPushRemote := strings.TrimSpace(string(output))
		if actualPushRemote != expectedPushRemote {
			t.Errorf("Expected pushRemote to be %s, got %s", expectedPushRemote, actualPushRemote)
		}
	})

	// Test setting upstream for a branch that doesn't exist on remote
	t.Run("non_existing_remote_branch", func(t *testing.T) {
		// Create a local branch
		branchName := "test-nonexistent-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
		err := git.CreateBranch(repoPath, branchName)
		if err != nil {
			t.Fatalf("Expected no error creating branch: %v", err)
		}

		// Checkout the branch
		err = git.CheckoutBranch(repoPath, branchName)
		if err != nil {
			t.Fatalf("Expected no error checking out branch: %v", err)
		}

		// Set upstream branch - this should succeed and configure push settings
		err = git.SetUpstreamBranch(repoPath, "origin", branchName)
		if err != nil {
			t.Errorf("Expected no error when setting upstream for non-existing remote branch: %v", err)
		}

		// Verify push configuration is set correctly
		cmd := exec.Command("git", "config", "--get", "branch."+branchName+".pushRemote")
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Expected no error getting pushRemote config: %v", err)
		}
		expectedPushRemote := "origin"
		actualPushRemote := strings.TrimSpace(string(output))
		if actualPushRemote != expectedPushRemote {
			t.Errorf("Expected pushRemote to be %s, got %s", expectedPushRemote, actualPushRemote)
		}
	})

	// Test setting upstream for a branch that doesn't exist locally
	t.Run("non_existing_local_branch", func(t *testing.T) {
		// Try to set upstream for a branch that doesn't exist
		branchName := "nonexistent-branch-" + strings.ReplaceAll(t.Name(), "/", "-")

		// Set upstream branch - this should fail because the local branch doesn't exist
		err := git.SetUpstreamBranch(repoPath, "origin", branchName)
		if err == nil {
			t.Error("Expected error when setting upstream for non-existing local branch")
		}

		// Verify the error message is helpful
		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "local branch") && !strings.Contains(errorMsg, "does not exist") {
			t.Errorf("Expected error message to contain 'local branch' and 'does not exist', got: %s", errorMsg)
		}
	})
}
