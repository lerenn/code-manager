package git

import (
	"fmt"
	"os/exec"
)

// FetchRemote fetches from a specific remote.
func (g *realGit) FetchRemote(repoPath, remoteName string) error {
	cmd := exec.Command("git", "fetch", remoteName)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w (command: git fetch %s, output: %s)",
			err, remoteName, string(output))
	}
	return nil
}
