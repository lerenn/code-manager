package git

import (
	"fmt"
	"os/exec"
)

// SetUpstreamBranch sets the upstream branch for the current branch.
// This configures push settings so that 'git push' will work without specifying remote/branch.
// The upstream tracking will be properly set when the branch is first pushed with 'git push -u'.
func (g *realGit) SetUpstreamBranch(repoPath, remote, branch string) error {
	// First check if the local branch exists
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = repoPath
	err := cmd.Run()
	if err != nil {
		// Local branch doesn't exist, this is an error
		return fmt.Errorf("local branch %s does not exist", branch)
	}

	// Set push configuration so 'git push' works without specifying remote/branch
	cmd = exec.Command("git", "config", "branch."+branch+".pushRemote", remote)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config branch.%s.pushRemote failed: %w (output: %s)", branch, err, string(output))
	}
	return nil
}
