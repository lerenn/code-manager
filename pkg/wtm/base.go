package wtm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
)

// base provides common functionality for WTM components.
type base struct {
	fs            fs.FS
	git           git.Git
	config        *config.Config
	statusManager status.Manager
	logger        logger.Logger
	verbose       bool
}

// newBase creates a new base instance.
func newBase(
	fs fs.FS,
	git git.Git,
	config *config.Config,
	statusManager status.Manager,
	logger logger.Logger,
	verbose bool,
) *base {
	return &base{
		fs:            fs,
		git:           git,
		config:        config,
		statusManager: statusManager,
		logger:        logger,
		verbose:       verbose,
	}
}

// verbosePrint prints a formatted message only in verbose mode.
func (b *base) verbosePrint(msg string, args ...interface{}) {
	if b.verbose {
		b.logger.Logf(fmt.Sprintf(msg, args...))
	}
}

// validateGitConfiguration validates that Git configuration is functional.
func (b *base) validateGitConfiguration(workDir string) error {
	b.verbosePrint("Validating Git configuration in: %s", workDir)

	// Execute git status to ensure Git is working
	_, err := b.git.Status(workDir)
	if err != nil {
		b.verbosePrint("Error: %v", err)
		return fmt.Errorf("git configuration error: %w", err)
	}

	return nil
}

// cleanupWorktreeDirectory removes the worktree directory.
func (b *base) cleanupWorktreeDirectory(worktreePath string) error {
	b.verbosePrint("Cleaning up worktree directory: %s", worktreePath)

	exists, err := b.fs.Exists(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}

	if exists {
		if err := b.fs.RemoveAll(worktreePath); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
	}

	return nil
}

// buildWorktreePath constructs a worktree path from base path, repository URL, and branch.
func (b *base) buildWorktreePath(repoURL, branch string) string {
	// Use configurable worktrees directory if specified, otherwise fall back to base path
	worktreesBase := b.config.WorktreesDir
	if worktreesBase == "" {
		worktreesBase = b.config.BasePath
	}
	return filepath.Join(worktreesBase, repoURL, branch)
}

// parseConfirmationInput parses confirmation input from user.
func (b *base) parseConfirmationInput(input string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "y", "yes":
		return true, nil
	case "n", "no", "":
		return false, nil
	case "q", "quit", "exit", "cancel":
		return false, fmt.Errorf("user cancelled")
	default:
		return false, fmt.Errorf("invalid input")
	}
}

// isQuitCommand checks if the input is a quit command.
func (b *base) isQuitCommand(input string) bool {
	quitCommands := []string{"q", "quit", "exit", "cancel"}
	for _, cmd := range quitCommands {
		if input == cmd {
			return true
		}
	}
	return false
}

// parseNumericInput parses numeric input from string.
func (b *base) parseNumericInput(input string) (int, error) {
	var choice int
	_, err := fmt.Sscanf(input, "%d", &choice)
	return choice, err
}

// isValidChoice checks if the choice is within valid range.
func (b *base) isValidChoice(choice, maxChoice int) bool {
	return choice >= 1 && choice <= maxChoice
}
