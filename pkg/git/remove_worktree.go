package git

import (
	"fmt"
	"os/exec"
)

// RemoveWorktree removes a worktree from Git's tracking.
func (g *realGit) RemoveWorktree(repoPath, worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %w (command: git worktree remove %s, output: %s)",
			err, worktreePath, string(output))
	}
	return nil
}
