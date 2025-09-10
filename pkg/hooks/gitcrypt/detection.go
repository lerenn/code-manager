// Package gitcrypt provides git-crypt functionality as a hook for worktree operations.
package gitcrypt

import (
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/fs"
)

// Detector handles detection of git-crypt usage in repositories.
type Detector struct {
	fs fs.FS
}

// NewDetector creates a new GitCryptDetector instance.
func NewDetector(fs fs.FS) *Detector {
	return &Detector{
		fs: fs,
	}
}

// DetectGitCryptUsage checks if the repository uses git-crypt by examining .gitattributes.
func (d *Detector) DetectGitCryptUsage(repoPath string) (bool, error) {
	gitattributesPath := filepath.Join(repoPath, ".gitattributes")

	// Check if .gitattributes exists
	exists, err := d.fs.Exists(gitattributesPath)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	// Read .gitattributes and check for git-crypt filter
	content, err := d.fs.ReadFile(gitattributesPath)
	if err != nil {
		return false, err
	}

	// Check for git-crypt filter patterns
	contentStr := string(content)
	return strings.Contains(contentStr, "filter=git-crypt"), nil
}
