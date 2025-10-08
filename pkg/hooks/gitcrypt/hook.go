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

// PostWorktreeCheckoutHook provides git-crypt functionality as a post-worktree checkout hook.
type PostWorktreeCheckoutHook struct {
	fs            fs.FS
	git           git.Git
	prompt        prompt.Prompter
	logger        logger.Logger
	detector      *Detector
	keyManager    *KeyManager
	worktreeSetup *WorktreeSetup
}

// NewPostWorktreeCheckoutHook creates a new GitCryptPostWorktreeCheckoutHook instance.
func NewPostWorktreeCheckoutHook() *PostWorktreeCheckoutHook {
	fsInstance := fs.NewFS()
	gitInstance := git.NewGit()
	promptInstance := prompt.NewPrompt()
	loggerInstance := logger.NewNoopLogger()

	return &PostWorktreeCheckoutHook{
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
func (h *PostWorktreeCheckoutHook) RegisterForOperations(
	registerHook func(operation string, hook hooks.PostWorktreeCheckoutHook) error,
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
func (h *PostWorktreeCheckoutHook) Name() string {
	return "git-crypt-worktree-checkout"
}

// Priority returns the hook priority.
func (h *PostWorktreeCheckoutHook) Priority() int {
	return 50
}

// Execute is a no-op for GitCryptPostWorktreeCheckoutHook.
func (h *PostWorktreeCheckoutHook) Execute(_ *hooks.HookContext) error {
	return nil
}

// OnPostWorktreeCheckout handles git-crypt setup before worktree checkout.
func (h *PostWorktreeCheckoutHook) OnPostWorktreeCheckout(ctx *hooks.HookContext) error {
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
func (h *PostWorktreeCheckoutHook) setupGitCryptForWorktree(repoPath, worktreePath, _ string) error {
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
func (h *PostWorktreeCheckoutHook) getRepositoryPath(ctx *hooks.HookContext) (string, error) {
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
