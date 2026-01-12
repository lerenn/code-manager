package git

import (
	"fmt"
	"os/exec"
)

// CheckoutBranch checks out a branch in the specified worktree.
// If the branch doesn't exist locally, it will try to checkout from origin/branch.
// If origin/branch doesn't exist, it will create the branch from HEAD.
func (g *realGit) CheckoutBranch(worktreePath, branch string) error {
	// First try to checkout the branch directly
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// If that fails, check if the branch exists locally
	branchExists, branchErr := g.BranchExists(worktreePath, branch)
	if branchErr == nil && branchExists {
		// Branch exists locally, try checkout again (might need ref refresh)
		cmd = exec.Command("git", "checkout", branch)
		cmd.Dir = worktreePath
		_, err2 := cmd.CombinedOutput()
		if err2 == nil {
			return nil
		}
		// If it still fails, continue to fallback logic
	}

	// Check if origin/branch exists on remote
	remoteBranchExists, remoteErr := g.BranchExistsOnRemote(BranchExistsOnRemoteParams{
		RepoPath:   worktreePath,
		RemoteName: "origin",
		Branch:     branch,
	})

	var originOutput []byte
	if remoteErr == nil && remoteBranchExists {
		// origin/branch exists, create local branch tracking it
		cmd = exec.Command("git", "checkout", "-b", branch, "origin/"+branch)
		cmd.Dir = worktreePath
		var err2 error
		originOutput, err2 = cmd.CombinedOutput()
		if err2 == nil {
			return nil
		}
		// If that fails, continue to final fallback
	}

	// origin/branch doesn't exist or checking failed, fetch to ensure we have latest refs
	if fetchErr := g.FetchRemote(worktreePath, "origin"); fetchErr == nil {
		// After fetch, check again if origin/branch exists
		remoteBranchExists, remoteErr = g.BranchExistsOnRemote(BranchExistsOnRemoteParams{
			RepoPath:   worktreePath,
			RemoteName: "origin",
			Branch:     branch,
		})
		if remoteErr == nil && remoteBranchExists {
			// origin/branch exists after fetch, create local branch tracking it
			cmd = exec.Command("git", "checkout", "-b", branch, "origin/"+branch)
			cmd.Dir = worktreePath
			var err2 error
			originOutput, err2 = cmd.CombinedOutput()
			if err2 == nil {
				return nil
			}
		}
	}

	// Final fallback: create branch from HEAD
	cmd = exec.Command("git", "checkout", "-b", branch, "HEAD")
	cmd.Dir = worktreePath
	headOutput, err3 := cmd.CombinedOutput()
	if err3 != nil {
		// Return comprehensive error with all attempted operations
		errorMsg := fmt.Sprintf(
			"git checkout failed: %v (command: git checkout %s, output: %s",
			err, branch, string(output))
		if len(originOutput) > 0 {
			errorMsg += fmt.Sprintf("; fallback: git checkout -b %s origin/%s, output: %s", branch, branch, string(originOutput))
		}
		errorMsg += fmt.Sprintf("; final fallback: git checkout -b %s HEAD, output: %s)", branch, string(headOutput))
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}
