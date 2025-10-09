package git

import (
	"fmt"
	"os/exec"
)

// CloneToPath clones a local repository to a target path with a specific branch.
// This creates a standalone repository (not a worktree reference).
func (g *realGit) CloneToPath(sourceRepoPath, targetPath, branch string) error {
	// Use git clone with --branch to clone and checkout the specific branch
	cmd := exec.Command("git", "clone", "--branch", branch, sourceRepoPath, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w (command: git clone --branch %s %s %s, output: %s)",
			err, branch, sourceRepoPath, targetPath, string(output))
	}
	return nil
}
