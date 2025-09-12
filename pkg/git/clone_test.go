//go:build integration

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_Clone(t *testing.T) {
	git := NewGit()

	// Create a temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "git-clone-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test cloning a real public repository
	repoURL := "https://github.com/octocat/Hello-World.git"
	clonePath := filepath.Join(tmpDir, "hello-world")

	err = git.Clone(CloneParams{
		RepoURL:    repoURL,
		TargetPath: clonePath,
		Recursive:  true,
	})
	if err != nil {
		t.Fatalf("Expected no error cloning %s: %v", repoURL, err)
	}

	// Verify the repository was cloned
	if _, err := os.Stat(clonePath); os.IsNotExist(err) {
		t.Errorf("Expected cloned repository directory %s to exist", clonePath)
	}

	// Verify it's a valid Git repository
	gitDir := filepath.Join(clonePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("Expected .git directory %s to exist", gitDir)
	}

	// Verify the default branch is checked out
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = clonePath
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	currentBranch := strings.TrimSpace(string(output))
	if currentBranch != "main" && currentBranch != "master" {
		t.Errorf("Expected current branch to be 'main' or 'master', got: %s", currentBranch)
	}

	// Verify the repository has content
	readmePath := filepath.Join(clonePath, "README")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Errorf("Expected README file %s to exist", readmePath)
	}

	// Test shallow cloning
	shallowClonePath := filepath.Join(tmpDir, "hello-world-shallow")

	err = git.Clone(CloneParams{
		RepoURL:    repoURL,
		TargetPath: shallowClonePath,
		Recursive:  false,
	})
	if err != nil {
		t.Fatalf("Expected no error shallow cloning %s: %v", repoURL, err)
	}

	// Verify the shallow repository was cloned
	if _, err := os.Stat(shallowClonePath); os.IsNotExist(err) {
		t.Errorf("Expected shallow cloned repository directory %s to exist", shallowClonePath)
	}

	// Verify it's a valid Git repository
	shallowGitDir := filepath.Join(shallowClonePath, ".git")
	if _, err := os.Stat(shallowGitDir); os.IsNotExist(err) {
		t.Errorf("Expected .git directory %s to exist", shallowGitDir)
	}

	// Test cloning to existing directory (should fail)
	err = git.Clone(CloneParams{
		RepoURL:    repoURL,
		TargetPath: clonePath, // Use the same path that already exists
		Recursive:  true,
	})
	if err == nil {
		t.Error("Expected error when cloning to existing directory")
	}

	// Test cloning with invalid URL
	invalidPath := filepath.Join(tmpDir, "invalid")
	err = git.Clone(CloneParams{
		RepoURL:    "not-a-valid-url",
		TargetPath: invalidPath,
		Recursive:  true,
	})
	if err == nil {
		t.Error("Expected error when cloning with invalid URL")
	}

	// Test cloning with empty URL
	emptyPath := filepath.Join(tmpDir, "empty")
	err = git.Clone(CloneParams{
		RepoURL:    "",
		TargetPath: emptyPath,
		Recursive:  true,
	})
	if err == nil {
		t.Error("Expected error when cloning with empty URL")
	}
}
