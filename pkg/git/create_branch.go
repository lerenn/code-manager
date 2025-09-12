package git

import (
	"fmt"
	"os/exec"
)

// CreateBranch creates a new branch from the current branch.
func (g *realGit) CreateBranch(repoPath, branch string) error {
	cmd := exec.Command("git", "branch", branch)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch failed: %w (command: git branch %s, output: %s)", err, branch, string(output))
	}

	return nil
}
