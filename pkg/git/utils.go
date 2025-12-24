package git

import (
	"fmt"
	"os/exec"
	"strings"
)

const defaultRemote = "origin"

// extractRepoNameFromURL extracts the repository name from a Git remote URL.
func (g *realGit) extractRepoNameFromURL(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle ssh:// format: ssh://git@host/path/to/repo
	if strings.HasPrefix(url, "ssh://") {
		return g.extractSSHProtocolRepoName(url)
	}

	// Handle SSH format: git@host:user/repo
	if strings.Contains(url, "@") && strings.Contains(url, ":") && !strings.Contains(url, "://") {
		return g.extractSSHRepoName(url)
	}

	// Handle HTTPS format: https://host/user/repo
	if strings.HasPrefix(url, "http") {
		return g.extractHTTPSRepoName(url)
	}

	return ""
}

// extractSSHProtocolRepoName extracts repository name from ssh:// URLs.
func (g *realGit) extractSSHProtocolRepoName(url string) string {
	// Remove ssh:// prefix
	url = strings.TrimPrefix(url, "ssh://")
	// Now it's in format git@host/path/to/repo
	if !strings.Contains(url, "@") {
		return ""
	}

	parts := strings.SplitN(url, "/", 2)
	if len(parts) != 2 {
		return ""
	}

	hostParts := strings.Split(parts[0], "@")
	if len(hostParts) != 2 {
		return ""
	}

	return hostParts[1] + "/" + parts[1]
}

// extractSSHRepoName extracts repository name from SSH URLs (git@host:user/repo).
func (g *realGit) extractSSHRepoName(url string) string {
	parts := strings.Split(url, ":")
	if len(parts) != 2 {
		return ""
	}

	hostParts := strings.Split(parts[0], "@")
	if len(hostParts) != 2 {
		return ""
	}

	return hostParts[1] + "/" + parts[1]
}

// extractHTTPSRepoName extracts repository name from HTTPS URLs.
func (g *realGit) extractHTTPSRepoName(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return ""
	}

	// Extract host and path: host/user/repo
	host := parts[2] // host
	if len(parts) < 4 {
		return host
	}

	user := parts[3] // user
	if len(parts) < 5 {
		return host + "/" + user
	}

	repo := parts[4] // repo
	return host + "/" + user + "/" + repo
}

// getUpstreamRemote tries to get the remote from the branch's upstream configuration.
func (g *realGit) getUpstreamRemote(repoPath, branch string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", branch+"@{upstream}")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("no upstream configured: %w", err)
	}

	// Parse the upstream branch name to extract remote
	// Format: "refs/remotes/remote/branch"
	upstream := strings.TrimSpace(string(output))
	parts := strings.Split(upstream, "/")
	if len(parts) >= 3 && parts[1] == "remotes" {
		return parts[2], nil
	}

	return defaultRemote, nil
}

// findRemoteFromBranchList searches through remote branches to find which remote has the specified branch.
func (g *realGit) findRemoteFromBranchList(repoPath, branch string) (string, error) {
	cmd := exec.Command("git", "branch", "-r")
	cmd.Dir = repoPath
	remoteOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git branch -r failed: %w (command: git branch -r, output: %s)",
			err, string(remoteOutput))
	}

	// Parse remote branches to find which remote has this branch
	lines := strings.Split(string(remoteOutput), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remote branch format: "remote/branch"
		if strings.HasSuffix(line, "/"+branch) {
			parts := strings.SplitN(line, "/", 2)
			if len(parts) == 2 {
				return parts[0], nil
			}
		}
	}

	// If we can't find the remote, return "origin" as default
	return defaultRemote, nil
}
