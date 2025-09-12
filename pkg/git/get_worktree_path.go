package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetWorktreePath gets the path of a worktree for a branch.
func (g *realGit) GetWorktreePath(repoPath, branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git worktree list failed: %w (command: git worktree list, output: %s)", err, string(output))
	}

	// Parse the worktree list output to find the path for the specified branch
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse worktree list format: "worktree-path [branch-name]"
		// Example: "/path/to/worktree [feature/branch]"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			worktreePath := parts[0]
			// Check if the branch name matches (it's in brackets)
			branchPart := strings.Join(parts[1:], " ")
			if strings.Contains(branchPart, branch) {
				return worktreePath, nil
			}
		}
	}

	return "", fmt.Errorf("worktree path not found for branch %s", branch)
}
