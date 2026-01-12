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
	output, err := g.tryCheckoutBranch(worktreePath, branch)
	if err == nil {
		return nil
	}

	// If that fails, check if the branch exists locally
	if branchExists, branchErr := g.BranchExists(worktreePath, branch); branchErr == nil && branchExists {
		// Branch exists locally, try checkout again (might need ref refresh)
		if _, err2 := g.tryCheckoutBranch(worktreePath, branch); err2 == nil {
			return nil
		}
	}

	// Try to checkout from origin/branch
	originOutput, err := g.tryCheckoutFromOrigin(worktreePath, branch)
	if err == nil {
		return nil
	}

	// Fetch and try again
	if g.FetchRemote(worktreePath, "origin") == nil {
		if originOutput, err = g.tryCheckoutFromOrigin(worktreePath, branch); err == nil {
			return nil
		}
	}

	// Final fallback: create branch from HEAD
	headOutput, err3 := g.createBranchFromHead(worktreePath, branch)
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

// tryCheckoutBranch attempts to checkout a branch directly.
func (g *realGit) tryCheckoutBranch(worktreePath, branch string) ([]byte, error) {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, err
	}
	return output, nil
}

// tryCheckoutFromOrigin attempts to checkout a branch from origin/branch if it exists.
func (g *realGit) tryCheckoutFromOrigin(worktreePath, branch string) ([]byte, error) {
	remoteBranchExists, remoteErr := g.BranchExistsOnRemote(BranchExistsOnRemoteParams{
		RepoPath:   worktreePath,
		RemoteName: "origin",
		Branch:     branch,
	})
	if remoteErr != nil || !remoteBranchExists {
		return nil, fmt.Errorf("remote branch origin/%s does not exist", branch)
	}

	// origin/branch exists, create local branch tracking it
	cmd := exec.Command("git", "checkout", "-b", branch, "origin/"+branch)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, err
	}
	return output, nil
}

// createBranchFromHead creates a new branch from HEAD.
func (g *realGit) createBranchFromHead(worktreePath, branch string) ([]byte, error) {
	cmd := exec.Command("git", "checkout", "-b", branch, "HEAD")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, err
	}
	return output, nil
}
