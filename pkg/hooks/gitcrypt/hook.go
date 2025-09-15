// Package gitcrypt provides git-crypt functionality as a hook for worktree operations.
package gitcrypt

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
)

// WorktreeCheckoutHook provides git-crypt functionality as a worktree checkout hook.
type WorktreeCheckoutHook struct {
	fs            fs.FS
	git           git.Git
	prompt        prompt.Prompter
	logger        logger.Logger
	detector      *Detector
	keyManager    *KeyManager
	worktreeSetup *WorktreeSetup
}

// NewWorktreeCheckoutHook creates a new GitCryptWorktreeCheckoutHook instance.
func NewWorktreeCheckoutHook() *WorktreeCheckoutHook {
	fsInstance := fs.NewFS()
	gitInstance := git.NewGit()
	promptInstance := prompt.NewPrompt()
	loggerInstance := logger.NewNoopLogger()

	return &WorktreeCheckoutHook{
		fs:            fsInstance,
		git:           gitInstance,
		prompt:        promptInstance,
		logger:        loggerInstance,
		detector:      NewDetector(fsInstance),
		keyManager:    NewKeyManager(fsInstance, gitInstance, promptInstance),
		worktreeSetup: NewWorktreeSetup(fsInstance),
	}
}

// RegisterForOperations registers this hook for worktree operations.
func (h *WorktreeCheckoutHook) RegisterForOperations(
	registerHook func(operation string, hook hooks.WorktreeCheckoutHook) error,
) error {
	// Register for operations that create worktrees
	if err := registerHook(consts.CreateWorkTree, h); err != nil {
		return err
	}

	if err := registerHook(consts.LoadWorktree, h); err != nil {
		return err
	}

	return nil
}

// Name returns the hook name.
func (h *WorktreeCheckoutHook) Name() string {
	return "git-crypt-worktree-checkout"
}

// Priority returns the hook priority.
func (h *WorktreeCheckoutHook) Priority() int {
	return 50
}

// Execute is a no-op for GitCryptWorktreeCheckoutHook.
func (h *WorktreeCheckoutHook) Execute(_ *hooks.HookContext) error {
	return nil
}

// OnWorktreeCheckout handles git-crypt setup before worktree checkout.
func (h *WorktreeCheckoutHook) OnWorktreeCheckout(ctx *hooks.HookContext) error {
	// Get worktree path from context
	worktreePath, ok := ctx.Parameters["worktreePath"].(string)
	if !ok || worktreePath == "" {
		return ErrWorktreePathNotFound
	}

	// Get repository path
	repoPath, err := h.getRepositoryPath(ctx)
	if err != nil {
		return err
	}

	// Check if repository uses git-crypt
	usesGitCrypt, err := h.detector.DetectGitCryptUsage(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect git-crypt usage: %w", err)
	}

	if !usesGitCrypt {
		// No git-crypt usage detected, nothing to do
		return nil
	}

	// Get branch information
	branch, ok := ctx.Parameters["branch"].(string)
	if !ok || branch == "" {
		return ErrBranchNotFound
	}

	// Setup git-crypt for the worktree
	return h.setupGitCryptForWorktree(repoPath, worktreePath, branch)
}

// setupGitCryptForWorktree sets up git-crypt in the worktree.
func (h *WorktreeCheckoutHook) setupGitCryptForWorktree(repoPath, worktreePath, _ string) error {
	// Try to find key in repository
	keyPath, err := h.keyManager.FindGitCryptKey(repoPath)
	if err != nil {
		return fmt.Errorf("failed to find git-crypt key: %w", err)
	}

	if keyPath == "" {
		// Key not found, prompt user
		keyPath, err = h.keyManager.PromptUserForKeyPath()
		if err != nil {
			return fmt.Errorf("failed to get key path from user: %w", err)
		}
	}

	// Validate the key file
	if err := h.keyManager.ValidateKeyFile(keyPath); err != nil {
		return fmt.Errorf("failed to validate key file: %w", err)
	}

	// Setup git-crypt in the worktree
	if err := h.worktreeSetup.SetupGitCryptForWorktree(repoPath, worktreePath, keyPath); err != nil {
		return fmt.Errorf("failed to setup git-crypt in worktree: %w", err)
	}

	return nil
}

// getRepositoryPath extracts the repository path from the hook context.
func (h *WorktreeCheckoutHook) getRepositoryPath(ctx *hooks.HookContext) (string, error) {
	// Try to get repository path from parameters
	if repoPath, ok := ctx.Parameters["repoPath"].(string); ok && repoPath != "" {
		return repoPath, nil
	}

	// Try to get repository path from metadata
	if repoPath, ok := ctx.Metadata["repoPath"].(string); ok && repoPath != "" {
		return repoPath, nil
	}

	return "", ErrRepositoryPathNotFound
}
