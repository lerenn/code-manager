package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// BranchExistsOnRemote checks if a branch exists on a specific remote.
func (g *realGit) BranchExistsOnRemote(params BranchExistsOnRemoteParams) (bool, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", params.RemoteName, params.Branch)
	cmd.Dir = params.RepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git ls-remote failed: %w (command: git ls-remote --heads %s %s, output: %s)",
			err, params.RemoteName, params.Branch, string(output))
	}
	return strings.TrimSpace(string(output)) != "", nil
}
