//go:build integration

package git

import (
	"strings"
	"testing"
)

func TestGit_Status(t *testing.T) {
	git := NewGit()

	// Test in current directory (should be a git repo for this test)
	output, err := git.Status(".")
	if err != nil {
		t.Skipf("Skipping test - not in a git repository: %v", err)
	}

	if !strings.Contains(output, "On branch") && !strings.Contains(output, "HEAD detached") {
		t.Errorf("Expected git status output to contain branch information, got: %s", output)
	}
}

func TestGit_ConfigGet(t *testing.T) {
	git := NewGit()

	// Test getting user.name (should exist in most git repos)
	output, err := git.ConfigGet(".", "user.name")
	if err != nil {
		t.Skipf("Skipping test - git config error: %v", err)
	}

	// Should return either a name or empty string
	if output != "" && strings.TrimSpace(output) == "" {
		t.Errorf("Expected valid user.name or empty string, got: %q", output)
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
