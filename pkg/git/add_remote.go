package git

import (
	"fmt"
	"os/exec"
)

// AddRemote adds a new remote to the repository.
func (g *realGit) AddRemote(repoPath, remoteName, remoteURL string) error {
	cmd := exec.Command("git", "remote", "add", remoteName, remoteURL)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote add failed: %w (command: git remote add %s %s, output: %s)",
			err, remoteName, remoteURL, string(output))
	}
	return nil
}
