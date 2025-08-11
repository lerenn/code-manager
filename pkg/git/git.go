package git

import (
	"errors"
	"fmt"
	"os/exec"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=git.go -destination=mockgit.gen.go -package=git

// Git interface provides Git command execution capabilities.
type Git interface {
	// Status executes `git status` in specified directory.
	Status(workDir string) (string, error)

	// ConfigGet executes `git config --get <key>` in specified directory.
	ConfigGet(workDir, key string) (string, error)
}

type realGit struct {
	// No fields needed for basic Git operations
}

// NewGit creates a new Git instance.
func NewGit() Git {
	return &realGit{}
}

// Status executes `git status` in specified directory.
func (g *realGit) Status(workDir string) (string, error) {
	cmd := exec.Command("git", "status")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w (command: git status, output: %s)", err, string(output))
	}

	return string(output), nil
}

// ConfigGet executes `git config --get <key>` in specified directory.
func (g *realGit) ConfigGet(workDir, key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		// Return empty string for missing config keys (exit code 1)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("git command failed: %w (command: git config --get %s, output: %s)", err, key, string(output))
	}

	return string(output), nil
}
