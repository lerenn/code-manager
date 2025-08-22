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
