// Package base provides common functionality for CM components.
package base

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
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
		return fmt.Errorf("%w: %w", ErrGitConfiguration, err)
	}

	return nil
}

// CleanupWorktreeDirectory removes the worktree directory.
func (b *Base) CleanupWorktreeDirectory(worktreePath string) error {
	b.VerbosePrint("Cleaning up worktree directory: %s", worktreePath)

	exists, err := b.FS.Exists(worktreePath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToCheckWorktreeDirectoryExists, err)
	}

	if exists {
		if err := b.FS.RemoveAll(worktreePath); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToRemoveWorktreeDirectory, err)
		}
	}

	return nil
}

// BuildWorktreePath constructs a worktree path from base path, repository URL, remote name, and branch.
func (b *Base) BuildWorktreePath(repoURL, remoteName, branch string) string {
	// Use base path directly with structure: $base_path/<repo_url>/<remote_name>/<branch>
	return filepath.Join(b.Config.BasePath, repoURL, remoteName, branch)
}
