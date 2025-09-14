//go:build integration

package git

import (
	"strings"
	"testing"
)

func TestGit_Status(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
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

func TestGit_Status_NonExistentDir(t *testing.T) {
	git := NewGit()

	// Test in non-existent directory
	_, err := git.Status("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
