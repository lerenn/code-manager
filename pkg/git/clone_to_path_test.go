//go:build integration

package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_CloneToPath(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Test cloning to a new path
	targetPath := filepath.Join(".", "test-clone-"+strings.ReplaceAll(t.Name(), "/", "-"))
	defer os.RemoveAll(targetPath)

	// Create a test branch first
	testBranchName := "test-clone-branch-" + strings.ReplaceAll(t.Name(), "/", "-")
	err := git.CreateBranch(".", testBranchName)
	if err != nil {
		t.Fatalf("Expected no error creating branch: %v", err)
	}

	// Clone the repository to target path
	err = git.CloneToPath(".", targetPath, testBranchName)
	if err != nil {
		t.Fatalf("Expected no error cloning repository: %v", err)
	}

	// Verify the cloned directory exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Expected cloned directory %s to exist", targetPath)
	}

	// Verify it's a standalone repository (has .git directory, not file)
	gitDir := filepath.Join(targetPath, ".git")
	gitDirInfo, err := os.Stat(gitDir)
	if err != nil {
		t.Errorf("Expected .git directory to exist in cloned repo: %v", err)
	}
	if !gitDirInfo.IsDir() {
		t.Error("Expected .git to be a directory in cloned repo, not a file")
	}

	// Verify the branch is checked out
	currentBranch, err := git.GetCurrentBranch(targetPath)
	if err != nil {
		t.Errorf("Expected no error getting current branch: %v", err)
	}
	if currentBranch != testBranchName {
		t.Errorf("Expected current branch to be %s, got %s", testBranchName, currentBranch)
	}

	// Test cloning to non-existent source
	err = git.CloneToPath("/non/existent/repo", "/tmp/test-clone", "main")
	if err == nil {
		t.Error("Expected error for non-existent source repository")
	}
}
