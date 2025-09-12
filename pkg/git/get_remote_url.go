package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetRemoteURL gets the URL of a remote.
func (g *realGit) GetRemoteURL(repoPath, remoteName string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git remote get-url failed: %w (command: git remote get-url %s, output: %s)",
			err, remoteName, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}
