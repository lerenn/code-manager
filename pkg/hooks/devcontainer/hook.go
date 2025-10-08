package devcontainer

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
)

// PreWorktreeCreationHook provides devcontainer detection as a pre-worktree creation hook.
type PreWorktreeCreationHook struct {
	fs       fs.FS
	logger   logger.Logger
	detector *Detector
}

// NewPreWorktreeCreationHook creates a new devcontainer PreWorktreeCreationHook instance.
func NewPreWorktreeCreationHook() *PreWorktreeCreationHook {
	fsInstance := fs.NewFS()
	loggerInstance := logger.NewNoopLogger()

	return &PreWorktreeCreationHook{
		fs:       fsInstance,
		logger:   loggerInstance,
		detector: NewDetector(fsInstance),
	}
}

// RegisterForOperations registers this hook for worktree operations.
func (h *PreWorktreeCreationHook) RegisterForOperations(
	registerHook func(operation string, hook hooks.PreWorktreeCreationHook) error,
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
func (h *PreWorktreeCreationHook) Name() string {
	return "devcontainer-detached-worktree"
}

// Priority returns the hook priority.
func (h *PreWorktreeCreationHook) Priority() int {
	return 10 // Low number = high priority, runs before other hooks
}

// Execute is a no-op for PreWorktreeCreationHook.
func (h *PreWorktreeCreationHook) Execute(_ *hooks.HookContext) error {
	return nil
}

// OnPreWorktreeCreation handles devcontainer detection before worktree creation.
func (h *PreWorktreeCreationHook) OnPreWorktreeCreation(ctx *hooks.HookContext) error {
	// Get repository path from context
	repoPath, err := h.getRepositoryPath(ctx)
	if err != nil {
		return err
	}

	// Check if repository has devcontainer configuration
	hasDevcontainer, err := h.detector.DetectDevcontainer(repoPath)
	if err != nil {
		return fmt.Errorf("failed to detect devcontainer: %w", err)
	}

	if hasDevcontainer {
		// Set metadata to indicate detached mode should be used
		ctx.Metadata["detached"] = true
		h.logger.Logf("Devcontainer detected, enabling detached mode for container compatibility")
	}

	return nil
}

// getRepositoryPath extracts the repository path from the hook context.
func (h *PreWorktreeCreationHook) getRepositoryPath(ctx *hooks.HookContext) (string, error) {
	// Try to get repository path from parameters
	if repoPath, ok := ctx.Parameters["repoPath"].(string); ok && repoPath != "" {
		return repoPath, nil
	}

	// Try to get repository path from metadata
	if repoPath, ok := ctx.Metadata["repoPath"].(string); ok && repoPath != "" {
		return repoPath, nil
	}

	return "", fmt.Errorf("repository path not found in hook context")
}
