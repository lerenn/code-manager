//go:build integration

package git

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGit_Add(t *testing.T) {
	git := NewGit()
	_, cleanup := SetupTestRepo(t)
	defer cleanup()

	// Create a test file
	testFile := "test-add-file.txt"
	testContent := "This is a test file for git add"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test adding a single file
	err = git.Add(".", testFile)
	if err != nil {
		t.Fatalf("Expected no error adding file: %v", err)
	}

	// Verify the file was added to staging area
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = "."
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check git status: %v", err)
	}

	if !strings.Contains(string(output), "A  "+testFile) {
		t.Errorf("Expected file %s to be staged, got status: %s", testFile, string(output))
	}

	// Test adding multiple files
	testFile2 := "test-add-file2.txt"
	testContent2 := "This is another test file"
	err = os.WriteFile(testFile2, []byte(testContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}

	err = git.Add(".", testFile2, "non-existent-file.txt")
	if err == nil {
		t.Error("Expected error when adding non-existent file")
	}

	// Test adding all files with .
	err = git.Add(".", ".")
	if err != nil {
		t.Fatalf("Expected no error adding all files: %v", err)
	}

	// Test in non-existent directory
	err = git.Add("/non/existent/directory", testFile)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
