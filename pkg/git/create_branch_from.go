package git

import (
	"fmt"
	"os/exec"
)

// CreateBranchFrom creates a new branch from a specific branch.
func (g *realGit) CreateBranchFrom(params CreateBranchFromParams) error {
	cmd := exec.Command("git", "branch", params.NewBranch, params.FromBranch)
	cmd.Dir = params.RepoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch failed: %w (command: git branch %s %s, output: %s)",
			err, params.NewBranch, params.FromBranch, string(output))
	}

	return nil
}
