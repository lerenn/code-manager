//go:build integration

package git

import (
	"strings"
	"testing"
)

func TestGit_ConfigGet(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
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
