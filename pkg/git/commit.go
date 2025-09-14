package git

import (
	"fmt"
	"os/exec"
)

// Commit creates a new commit with the specified message.
func (g *realGit) Commit(repoPath, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w (command: git commit -m %s, output: %s)",
			err, message, string(output))
	}
	return nil
}
