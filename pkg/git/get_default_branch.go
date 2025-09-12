package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetDefaultBranch gets the default branch name from a remote repository.
func (g *realGit) GetDefaultBranch(remoteURL string) (string, error) {
	// Use git ls-remote --symref to get the default branch
	cmd := exec.Command("git", "ls-remote", "--symref", remoteURL, "HEAD")
	// Set working directory to a temporary directory to avoid conflicts with worktrees
	cmd.Dir = "/tmp"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git ls-remote failed: %w (command: git ls-remote --symref %s HEAD, output: %s)",
			err, remoteURL, string(output))
	}

	// Parse the output to extract the default branch name
	// Output format: "ref: refs/heads/main\t<commit-hash>"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for the HEAD reference line
		if strings.HasPrefix(line, "ref:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ref := parts[1]
				// Extract branch name from refs/heads/branch-name
				if strings.HasPrefix(ref, "refs/heads/") {
					return strings.TrimPrefix(ref, "refs/heads/"), nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not determine default branch from remote URL: %s", remoteURL)
}
