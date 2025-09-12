package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Clone clones a repository to the specified path.
func (g *realGit) Clone(params CloneParams) error {
	args := []string{"clone"}

	// Add --no-recursive flag if not recursive
	if !params.Recursive {
		args = append(args, "--no-recursive")
	}

	// Add repository URL and target path
	args = append(args, params.RepoURL, params.TargetPath)

	cmd := exec.Command("git", args...)
	// Set working directory to /tmp to avoid working directory issues
	cmd.Dir = "/tmp"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w (command: git %s, output: %s)",
			err, strings.Join(args, " "), string(output))
	}
	return nil
}
