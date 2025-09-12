package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetCurrentBranch gets the current branch name.
func (g *realGit) GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w (command: git branch --show-current, output: %s)",
			err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}
