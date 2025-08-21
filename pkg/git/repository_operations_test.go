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
	_, cleanup := setupTestRepo(t)
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

func TestGit_Commit(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a test file and add it to staging
	testFile := "test-commit-file.txt"
	testContent := "This is a test file for git commit"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add the file to staging
	err = git.Add(".", testFile)
	if err != nil {
		t.Fatalf("Expected no error adding file: %v", err)
	}

	// Test creating a commit
	commitMessage := "Test commit message"
	err = git.Commit(".", commitMessage)
	if err != nil {
		t.Fatalf("Expected no error creating commit: %v", err)
	}

	// Verify the commit was created
	cmd := exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = "."
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check git log: %v", err)
	}

	if !strings.Contains(string(output), commitMessage) {
		t.Errorf("Expected commit message '%s' in log, got: %s", commitMessage, string(output))
	}

	// Test creating a commit with empty staging area
	err = git.Commit(".", "Empty commit")
	if err == nil {
		t.Error("Expected error when committing with empty staging area")
	}

	// Test creating a commit with empty message
	err = git.Commit(".", "")
	if err == nil {
		t.Error("Expected error when committing with empty message")
	}

	// Test in non-existent directory
	err = git.Commit("/non/existent/directory", "test message")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGit_AddAndCommit_Workflow(t *testing.T) {
	git := NewGit()
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Test the complete workflow: create file, add, commit
	testFile := "workflow-test-file.txt"
	testContent := "This is a test file for the add and commit workflow"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add the file
	err = git.Add(".", testFile)
	if err != nil {
		t.Fatalf("Expected no error adding file: %v", err)
	}

	// Commit the file
	commitMessage := "Add workflow test file"
	err = git.Commit(".", commitMessage)
	if err != nil {
		t.Fatalf("Expected no error creating commit: %v", err)
	}

	// Verify the file is in the repository
	cmd := exec.Command("git", "ls-files", testFile)
	cmd.Dir = "."
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check if file is tracked: %v", err)
	}

	if strings.TrimSpace(string(output)) != testFile {
		t.Errorf("Expected file %s to be tracked, got: %s", testFile, string(output))
	}

	// Verify the commit message
	cmd = exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = "."
	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check git log: %v", err)
	}

	if !strings.Contains(string(output), commitMessage) {
		t.Errorf("Expected commit message '%s' in log, got: %s", commitMessage, string(output))
	}
}
