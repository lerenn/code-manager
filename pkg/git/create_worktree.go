package git

import (
	"fmt"
	"os/exec"
)

// CreateWorktree creates a new worktree for the specified branch.
func (g *realGit) CreateWorktree(repoPath, worktreePath, branch string) error {
	cmd := exec.Command("git", "worktree", "add", worktreePath, branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %w (command: git worktree add %s %s, output: %s)",
			err, worktreePath, branch, string(output))
	}
	return nil
}
