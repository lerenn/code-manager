package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Add adds files to the Git staging area.
func (g *realGit) Add(repoPath string, files ...string) error {
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w (command: git add %s, output: %s)",
			err, strings.Join(files, " "), string(output))
	}
	return nil
}
