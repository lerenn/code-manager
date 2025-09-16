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

	// Test checking out non-existent branch
	err = git.CheckoutBranch(testWorktreePath, "non-existent-branch")
	if err == nil {
		t.Error("Expected error when checking out non-existent branch")
	}

	// Test in non-existent directory
	err = git.CheckoutBranch("/non/existent/directory", testBranchName)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
