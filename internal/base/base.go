// Package base provides common functionality for CM components.
package base

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/logger"
	"github.com/lerenn/cm/pkg/prompt"
	"github.com/lerenn/cm/pkg/status"
)

// Base provides common functionality for CM components.
type Base struct {
	FS            fs.FS
	Git           git.Git
	Config        *config.Config
	StatusManager status.Manager
	Logger        logger.Logger
	Prompt        prompt.Prompt
	verbose       bool
}

// NewBaseParams contains parameters for creating a new Base instance.
type NewBaseParams struct {
	FS            fs.FS
	Git           git.Git
	Config        *config.Config
	StatusManager status.Manager
	Logger        logger.Logger
	Prompt        prompt.Prompt
	Verbose       bool
}

// NewBase creates a new Base instance.
func NewBase(params NewBaseParams) *Base {
	return &Base{
		FS:            params.FS,
		Git:           params.Git,
		Config:        params.Config,
		StatusManager: params.StatusManager,
		Logger:        params.Logger,
		Prompt:        params.Prompt,
		verbose:       params.Verbose,
	}
}

// VerbosePrint prints a formatted message only in verbose mode.
func (b *Base) VerbosePrint(msg string, args ...interface{}) {
	if b.verbose {
		b.Logger.Logf(fmt.Sprintf(msg, args...))
	}
}

// IsVerbose returns whether verbose mode is enabled.
func (b *Base) IsVerbose() bool {
	return b.verbose
}

// ValidateGitConfiguration validates that Git configuration is functional.
func (b *Base) ValidateGitConfiguration(workDir string) error {
	b.VerbosePrint("Validating Git configuration in: %s", workDir)

	// Execute git status to ensure Git is working
	_, err := b.Git.Status(workDir)
	if err != nil {
		b.VerbosePrint("Error: %v", err)
		return fmt.Errorf("git configuration error: %w", err)
	}

	return nil
}

// CleanupWorktreeDirectory removes the worktree directory.
func (b *Base) CleanupWorktreeDirectory(worktreePath string) error {
	b.VerbosePrint("Cleaning up worktree directory: %s", worktreePath)

	exists, err := b.FS.Exists(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}

	if exists {
		if err := b.FS.RemoveAll(worktreePath); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
	}

	return nil
}

// BuildWorktreePath constructs a worktree path from base path, repository URL, and branch.
func (b *Base) BuildWorktreePath(repoURL, branch string) string {
	// Use computed worktrees directory
	worktreesBase := b.Config.GetWorktreesDir()
	return filepath.Join(worktreesBase, repoURL, branch)
}
