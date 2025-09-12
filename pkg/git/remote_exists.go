package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RemoteExists checks if a remote exists.
func (g *realGit) RemoteExists(repoPath, remoteName string) (bool, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git remote failed: %w (command: git remote, output: %s)",
			err, string(output))
	}

	// Check if the remote name exists in the list
	remotes := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, remote := range remotes {
		if strings.TrimSpace(remote) == remoteName {
			return true, nil
		}
	}
	return false, nil
}
