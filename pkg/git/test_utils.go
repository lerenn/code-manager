package git

import (
	"os"
	"os/exec"
	"testing"
)

// SetupTestRepo creates a temporary git repository for testing.
func SetupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, originalDir := setupTempDir(t)
	setupGitRepo(t, tmpDir, originalDir)
	
	cleanup := func() {
		_ = os.Chdir(originalDir)
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func setupTempDir(t *testing.T) (string, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	return tmpDir, originalDir
}

func setupGitRepo(t *testing.T, tmpDir, originalDir string) {
	t.Helper()
	initGitRepo(t, originalDir, tmpDir)
	configureGitUser(t, originalDir, tmpDir)
	addRemoteOrigin(t, originalDir, tmpDir)
	createInitialCommit(t, originalDir, tmpDir)
}

func initGitRepo(t *testing.T, originalDir, tmpDir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	if err := cmd.Run(); err != nil {
		cleanupOnEf(t, originalDir, tmpDir, "Failed to initialize git repository: %v", err)
	}
}

func configureGitUser(t *testing.T, originalDir, tmpDir string) {
	t.Helper()
	commands := []*exec.Cmd{
		exec.Command("git", "config", "user.name", "Test User"),
		exec.Command("git", "config", "user.email", "test@example.com"),
	}
	
	for _, cmd := range commands {
		if err := cmd.Run(); err != nil {
			cleanupOnEf(t, originalDir, tmpDir, "Failed to configure git user: %v", err)
		}
	}
}

func addRemoteOrigin(t *testing.T, originalDir, tmpDir string) {
	t.Helper()
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/octocat/Hello-World.git")
	if err := cmd.Run(); err != nil {
		cleanupOnEf(t, originalDir, tmpDir, "Failed to add remote origin: %v", err)
	}
}

func createInitialCommit(t *testing.T, originalDir, tmpDir string) {
	t.Helper()
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		createCommitWithFile(t, originalDir, tmpDir)
	}
}

func createCommitWithFile(t *testing.T, originalDir, tmpDir string) {
	t.Helper()
	if err := os.WriteFile("README.md", []byte("# Test Repository"), 0644); err != nil {
		cleanupOnEf(t, originalDir, tmpDir, "Failed to create README file: %v", err)
	}

	cmd := exec.Command("git", "add", "README.md")
	if err := cmd.Run(); err != nil {
		cleanupOnEf(t, originalDir, tmpDir, "Failed to add README to git: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		cleanupOnEf(t, originalDir, tmpDir, "Failed to create initial commit: %v", err)
	}
}

func cleanupOnEf(t *testing.T, originalDir, tmpDir, format string, args ...interface{}) {
	t.Helper()
	_ = os.Chdir(originalDir)
	_ = os.RemoveAll(tmpDir)
	t.Fatalf(format, args...)
}
