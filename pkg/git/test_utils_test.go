//go:build integration

package git

import (
	"os"
	"os/exec"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	if err := cmd.Run(); err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to configure git user: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to configure git email: %v", err)
	}

	// Add remote origin (using a public repository)
	cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/octocat/Hello-World.git")
	if err := cmd.Run(); err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to add remote origin: %v", err)
	}

	// Create initial commit
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		// If commit fails, create a file first
		if err := os.WriteFile("README.md", []byte("# Test Repository"), 0644); err != nil {
			os.Chdir(originalDir)
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create README file: %v", err)
		}

		cmd = exec.Command("git", "add", "README.md")
		if err := cmd.Run(); err != nil {
			os.Chdir(originalDir)
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to add README to git: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		if err := cmd.Run(); err != nil {
			os.Chdir(originalDir)
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create initial commit: %v", err)
		}
	}

	// Return cleanup function
	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}
