//go:build integration

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_GetDefaultBranch(t *testing.T) {
	git := NewGit()

	// Test getting default branch from a real public repository
	repoURL := "https://github.com/octocat/Hello-World.git"
	defaultBranch, err := git.GetDefaultBranch(repoURL)
	if err != nil {
		t.Fatalf("Expected no error getting default branch from %s: %v", repoURL, err)
	}

	// The default branch should be either 'main' or 'master'
	if defaultBranch != "main" && defaultBranch != "master" {
		t.Errorf("Expected default branch to be 'main' or 'master', got: %s", defaultBranch)
	}

	// Test getting default branch from invalid URL
	_, err = git.GetDefaultBranch("not-a-valid-url")
	if err == nil {
		t.Error("Expected error when getting default branch from invalid URL")
	}

	// Test getting default branch from empty URL
	_, err = git.GetDefaultBranch("")
	if err == nil {
		t.Error("Expected error when getting default branch from empty URL")
	}
}

func TestGit_CloneAndGetDefaultBranch_Integration(t *testing.T) {
	git := NewGit()

	// Create a temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "git-clone-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test the integration: get default branch, then clone
	repoURL := "https://github.com/octocat/Hello-World.git"

	// First, get the default branch
	defaultBranch, err := git.GetDefaultBranch(repoURL)
	if err != nil {
		t.Fatalf("Expected no error getting default branch: %v", err)
	}

	// Then clone the repository
	clonePath := filepath.Join(tmpDir, "hello-world-integration")
	err = git.Clone(CloneParams{
		RepoURL:    repoURL,
		TargetPath: clonePath,
		Recursive:  true,
	})
	if err != nil {
		t.Fatalf("Expected no error cloning repository: %v", err)
	}

	// Verify the cloned repository is on the correct default branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = clonePath
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	currentBranch := strings.TrimSpace(string(output))
	if currentBranch != defaultBranch {
		t.Errorf("Expected current branch to be %s (detected default), got: %s", defaultBranch, currentBranch)
	}

	// Verify the repository has the expected structure
	expectedFiles := []string{"README", ".git"}
	for _, file := range expectedFiles {
		filePath := filepath.Join(clonePath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file/directory %s to exist", filePath)
		}
	}
}
