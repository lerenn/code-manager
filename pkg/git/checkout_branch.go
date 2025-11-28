package git

import (
	"fmt"
	"os/exec"
)

// CheckoutBranch checks out a branch in the specified worktree.
// If the branch doesn't exist locally, it will try to checkout from origin/branch.
func (g *realGit) CheckoutBranch(worktreePath, branch string) error {
	// First try to checkout the branch directly
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// If that fails, try to checkout with tracking from origin
	cmd = exec.Command("git", "checkout", "-b", branch, "origin/"+branch)
	cmd.Dir = worktreePath
	output2, err2 := cmd.CombinedOutput()
	if err2 != nil {
		// Return the original error if both attempts fail
		return fmt.Errorf(
			"git checkout failed: %w (command: git checkout %s, output: %s; fallback: git checkout -b %s origin/%s, output: %s)",
			err, branch, string(output), branch, branch, string(output2))
	}
	return nil
}
