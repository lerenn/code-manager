package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckReferenceConflict checks if creating a branch would conflict with existing references.
func (g *realGit) CheckReferenceConflict(repoPath, branch string) error {
	// Check if any parent reference exists that would conflict
	parts := strings.Split(branch, "/")
	for i := 1; i < len(parts); i++ {
		parentRef := strings.Join(parts[:i], "/")

		// Check if parent reference exists as a branch
		cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+parentRef)
		cmd.Dir = repoPath
		if err := cmd.Run(); err == nil {
			return fmt.Errorf("%w: cannot create branch '%s': reference 'refs/heads/%s' already exists",
				ErrBranchParentExists, branch, parentRef)
		}

		// Also check tags
		cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/tags/"+parentRef)
		cmd.Dir = repoPath
		if err := cmd.Run(); err == nil {
			return fmt.Errorf("%w: cannot create branch '%s': tag 'refs/tags/%s' already exists",
				ErrTagParentExists, branch, parentRef)
		}
	}
	return nil
}
