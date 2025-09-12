package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// BranchExists checks if a branch exists locally or remotely.
func (g *realGit) BranchExists(repoPath, branch string) (bool, error) {
	// Check local branches
	cmd := exec.Command("git", "branch", "--list", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git branch --list failed: %w (command: git branch --list %s, output: %s)",
			err, branch, string(output))
	}
	if strings.TrimSpace(string(output)) != "" {
		return true, nil
	}

	// Check remote branches
	cmd = exec.Command("git", "branch", "-r", "--list", "origin/"+branch)
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git branch -r --list failed: %w (command: git branch -r --list origin/%s, output: %s)",
			err, branch, string(output))
	}
	return strings.TrimSpace(string(output)) != "", nil
}
