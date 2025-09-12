package git

import (
	"fmt"
	"os/exec"
)

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
