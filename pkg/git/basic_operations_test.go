//go:build integration

package git

import (
	"strings"
	"testing"
)

func TestGit_Status(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test in the temporary git repository
	output, err := git.Status(".")
	if err != nil {
		t.Fatalf("Expected no error in git repository: %v", err)
	}

	if !strings.Contains(output, "On branch") && !strings.Contains(output, "HEAD detached") {
		t.Errorf("Expected git status output to contain branch information, got: %s", output)
	}
}

func TestGit_ConfigGet(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test getting user.name (should exist in our test repo)
	output, err := git.ConfigGet(".", "user.name")
	if err != nil {
		t.Fatalf("Expected no error getting user.name: %v", err)
	}

	// Should return the configured name (trim newline)
	expectedName := "Test User"
	if strings.TrimSpace(output) != expectedName {
		t.Errorf("Expected %q, got: %q", expectedName, strings.TrimSpace(output))
	}

	// Test getting non-existent config
	output, err = git.ConfigGet(".", "nonexistent.key")
	if err != nil {
		t.Errorf("Expected no error for non-existent config, got: %v", err)
	}

	if output != "" {
		t.Errorf("Expected empty string for non-existent config, got: %q", output)
	}
}

func TestGit_Status_NonExistentDir(t *testing.T) {
	git := NewGit()

	// Test in non-existent directory
	_, err := git.Status("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGit_GetCurrentBranch(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test in the temporary git repository
	branch, err := git.GetCurrentBranch(".")
	if err != nil {
		t.Fatalf("Expected no error in git repository: %v", err)
	}

	// Should be on main or master branch
	if branch != "main" && branch != "master" {
		t.Errorf("Expected 'main' or 'master' branch, got: %q", branch)
	}

	// Test in non-existent directory
	_, err = git.GetCurrentBranch("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGit_GetRepositoryName(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test in the temporary git repository
	repoName, err := git.GetRepositoryName(".")
	if err != nil {
		t.Fatalf("Expected no error in git repository: %v", err)
	}

	// Should return the repository name from remote origin (trim .git suffix and newline)
	expectedRepoName := "github.com/octocat/Hello-World"
	actualRepoName := strings.TrimSuffix(strings.TrimSpace(repoName), ".git")
	if actualRepoName != expectedRepoName {
		t.Errorf("Expected %q, got: %q", expectedRepoName, actualRepoName)
	}

	// Test in non-existent directory
	_, err = git.GetRepositoryName("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGit_IsClean(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test in the temporary git repository
	isClean, err := git.IsClean(".")
	if err != nil {
		t.Fatalf("Expected no error in git repository: %v", err)
	}

	// Currently always returns true as it's a placeholder
	if !isClean {
		t.Error("Expected clean state (placeholder implementation)")
	}

	// Test in non-existent directory
	_, err = git.IsClean("/non/existent/directory")
	if err != nil {
		t.Errorf("Expected no error for non-existent directory (placeholder implementation), got: %v", err)
	}
}
