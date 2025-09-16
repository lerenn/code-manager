package git

import (
	"fmt"
	"os/exec"
)

// CheckoutBranch checks out a branch in the specified worktree.
func (g *realGit) CheckoutBranch(worktreePath, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w (command: git checkout %s, output: %s)",
			err, branch, string(output))
	}
	return nil
}
