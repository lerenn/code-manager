//go:build integration

package git

import (
	"strings"
	"testing"
)

func TestGit_GetRepositoryName(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
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
