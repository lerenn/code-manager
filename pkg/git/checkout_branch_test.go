//go:build integration

package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_CheckoutBranch(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Create a test branch
	testBranchName := "test-checkout-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
	err := git.CreateBranch(".", testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating branch: %v", err)
	}

	// Create a worktree without checkout
	testWorktreePath := filepath.Join(".", "test-checkout-worktree-"+strings.ReplaceAll(t.Name(), "/", "-"))
	err = git.CreateWorktreeWithNoCheckout(".", testWorktreePath, testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating worktree without checkout: %v", err)
	}
	defer os.RemoveAll(testWorktreePath)

	// Checkout the branch in the worktree
	err = git.CheckoutBranch(testWorktreePath, testBranchName)
	if err != nil {
		t.Fatalf("Expected no error checking out branch: %v", err)
	}

	// Test checking out non-existent branch (should create from HEAD)
	nonExistentBranch := "non-existent-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
	err = git.CheckoutBranch(testWorktreePath, nonExistentBranch)
	if err != nil {
		t.Errorf("Expected no error when checking out non-existent branch (should create from HEAD): %v", err)
	}

	// Verify the branch was created
	exists, err := git.BranchExists(testWorktreePath, nonExistentBranch)
	if err != nil {
		t.Errorf("Expected no error checking created branch existence: %v", err)
	}
	if !exists {
		t.Errorf("Expected branch %s to exist after checkout", nonExistentBranch)
	}

	// Test in non-existent directory
	err = git.CheckoutBranch("/non/existent/directory", testBranchName)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

// TestGit_CheckoutBranch_FromHead tests checkout when branch doesn't exist locally or on remote
func TestGit_CheckoutBranch_FromHead(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Create a unique branch for worktree creation to avoid conflicts
	worktreeBaseBranch := "worktree-base-" + strings.ReplaceAll(t.Name(), "/", "-")
	err := git.CreateBranch(".", worktreeBaseBranch)
	if err != nil {
		t.Fatalf("Expected no error creating base branch: %v", err)
	}

	// Create a worktree without checkout
	testWorktreePath := "test-checkout-head-worktree-" + strings.ReplaceAll(t.Name(), "/", "-")
	testBranchName := "test-branch-from-head-" + strings.ReplaceAll(t.Name(), "/", "-")

	err = git.CreateWorktreeWithNoCheckout(".", testWorktreePath, worktreeBaseBranch)
	if err != nil {
		t.Fatalf("Expected no error creating worktree without checkout: %v", err)
	}
	defer os.RemoveAll(testWorktreePath)

	// Checkout a branch that doesn't exist (should create from HEAD)
	err = git.CheckoutBranch(testWorktreePath, testBranchName)
	if err != nil {
		t.Fatalf("Expected no error checking out non-existent branch (should create from HEAD): %v", err)
	}

	// Verify the branch was created
	exists, err := git.BranchExists(testWorktreePath, testBranchName)
	if err != nil {
		t.Errorf("Expected no error checking created branch existence: %v", err)
	}
	if !exists {
		t.Errorf("Expected branch %s to exist after checkout", testBranchName)
	}
}

// TestGit_CheckoutBranch_LocalBranchExists tests checkout when branch exists locally
func TestGit_CheckoutBranch_LocalBranchExists(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Create a unique branch for worktree creation to avoid conflicts
	worktreeBaseBranch := "worktree-base-local-" + strings.ReplaceAll(t.Name(), "/", "-")
	err := git.CreateBranch(".", worktreeBaseBranch)
	if err != nil {
		t.Fatalf("Expected no error creating base branch: %v", err)
	}

	// Create a test branch in main repo
	testBranchName := "test-local-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
	err = git.CreateBranch(".", testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating branch: %v", err)
	}

	// Create a worktree without checkout
	testWorktreePath := "test-checkout-local-worktree-" + strings.ReplaceAll(t.Name(), "/", "-")

	err = git.CreateWorktreeWithNoCheckout(".", testWorktreePath, worktreeBaseBranch)
	if err != nil {
		t.Fatalf("Expected no error creating worktree without checkout: %v", err)
	}
	defer os.RemoveAll(testWorktreePath)

	// Checkout the branch that exists locally
	err = git.CheckoutBranch(testWorktreePath, testBranchName)
	if err != nil {
		t.Fatalf("Expected no error checking out existing local branch: %v", err)
	}

	// Verify we're on the correct branch
	currentWorktreeBranch, err := git.GetCurrentBranch(testWorktreePath)
	if err != nil {
		t.Errorf("Expected no error getting current branch in worktree: %v", err)
	}
	if currentWorktreeBranch != testBranchName {
		t.Errorf("Expected to be on branch %s, got %s", testBranchName, currentWorktreeBranch)
	}
}
