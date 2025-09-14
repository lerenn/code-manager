package git

import (
	"fmt"
	"os/exec"
)

// CreateWorktreeWithNoCheckout creates a new worktree without checking out files.
func (g *realGit) CreateWorktreeWithNoCheckout(repoPath, worktreePath, branch string) error {
	cmd := exec.Command("git", "worktree", "add", "--no-checkout", worktreePath, branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add --no-checkout failed: %w "+
			"(command: git worktree add --no-checkout %s %s, output: %s)",
			err, worktreePath, branch, string(output))
	}
	return nil
}
