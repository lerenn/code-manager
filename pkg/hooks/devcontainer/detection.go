// Package devcontainer provides devcontainer detection functionality for worktree operations.
package devcontainer

import (
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/fs"
)

// Detector provides devcontainer detection functionality.
type Detector struct {
	fs fs.FS
}

// NewDetector creates a new Detector instance.
func NewDetector(fs fs.FS) *Detector {
	return &Detector{
		fs: fs,
	}
}

// DetectDevcontainer checks if a repository has a devcontainer configuration.
// It checks for both .devcontainer/devcontainer.json and .devcontainer.json files.
func (d *Detector) DetectDevcontainer(repoPath string) (bool, error) {
	// Check for .devcontainer/devcontainer.json
	devcontainerPath := filepath.Join(repoPath, ".devcontainer", "devcontainer.json")
	exists, err := d.fs.Exists(devcontainerPath)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// Check for .devcontainer.json in root
	rootDevcontainerPath := filepath.Join(repoPath, ".devcontainer.json")
	exists, err = d.fs.Exists(rootDevcontainerPath)
	if err != nil {
		return false, err
	}

	return exists, nil
}
