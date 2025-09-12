package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// WorktreeExists checks if a worktree exists for the specified branch.
func (g *realGit) WorktreeExists(repoPath, branch string) (bool, error) {
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git worktree list failed: %w (command: git worktree list, output: %s)", err, string(output))
	}

	// Check if the branch is mentioned in the worktree list
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, branch) {
			return true, nil
		}
	}

	return false, nil
}
