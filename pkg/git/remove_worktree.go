package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RemoveWorktree removes a worktree from Git's tracking.
func (g *realGit) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %w (command: git %s, output: %s)",
			err, strings.Join(args, " "), string(output))
	}
	return nil
}
