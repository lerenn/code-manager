package git

import (
	"errors"
	"fmt"
	"os/exec"
)

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
