package ide

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/logger"
)

// OpeningHook provides IDE opening functionality as a post-hook.
type OpeningHook struct {
	IDEManager ManagerInterface
}

// NewOpeningHook creates a new OpeningHook instance and registers it for appropriate operations.
func NewOpeningHook() *OpeningHook {
	// Create IDE manager internally since it should be private to this package
	fsInstance := fs.NewFS()
	loggerInstance := logger.NewNoopLogger()
	return &OpeningHook{
		IDEManager: NewManager(fsInstance, loggerInstance),
	}
}

// RegisterForOperations registers this hook for the operations that create worktrees.
func (h *OpeningHook) RegisterForOperations(registerHook func(operation string, hook hooks.Hook) error) error {
	// Register as post-hook for operations that create worktrees
	if err := registerHook(consts.CreateWorkTree, h); err != nil {
		return err
	}

	if err := registerHook(consts.LoadWorktree, h); err != nil {
		return err
	}

	// Register for OpenWorktree as well to ensure uniform IDE opening
	if err := registerHook(consts.OpenWorktree, h); err != nil {
		return err
	}

	return nil
}

// Name returns the hook name.
func (h *OpeningHook) Name() string {
	return "ide-opening"
}

// Priority returns the hook priority (lower numbers execute first).
func (h *OpeningHook) Priority() int {
	return 150
}

// Execute is a no-op for OpeningHook as it implements specific methods.
func (h *OpeningHook) Execute(_ *hooks.HookContext) error {
	return nil
}

// PreExecute is a no-op for OpeningHook.
func (h *OpeningHook) PreExecute(_ *hooks.HookContext) error {
	return nil
}

// PostExecute validates IDE opening parameters and opens the IDE after successful operations.
func (h *OpeningHook) PostExecute(ctx *hooks.HookContext) error {
	// Only proceed if operation was successful
	if ctx.Error != nil {
		// Operation failed, don't process IDE opening - this is intentional
		return nil //nolint:nilerr
	}

	// Get worktree path from parameters
	worktreePath, hasWorktreePath := ctx.Parameters["worktreePath"]
	if !hasWorktreePath {
		return fmt.Errorf("cannot open IDE: worktreePath parameter is required")
	}

	worktreePathStr, ok := worktreePath.(string)
	if !ok || worktreePathStr == "" {
		return fmt.Errorf("cannot open IDE: worktreePath must be a non-empty string")
	}

	// Get IDE name from parameters
	ideName, hasIDEName := ctx.Parameters["ideName"]
	if !hasIDEName {
		// No IDE specified, nothing to do
		return nil
	}

	ideNameStr, ok := ideName.(string)
	if !ok || ideNameStr == "" {
		// Invalid IDE name, nothing to do
		return nil
	}

	// Open the IDE
	if err := h.IDEManager.OpenIDE(ideNameStr, worktreePathStr); err != nil {
		// For OpenWorktree operations, IDE opening failure should fail the operation
		// For CreateWorkTree and LoadWorktree operations, IDE opening failure should not prevent success
		if ctx.OperationName == consts.OpenWorktree {
			return fmt.Errorf("failed to open IDE %s: %w", ideNameStr, err)
		}

		// IDE opening failed, but this should not prevent the worktree creation from succeeding
		// Log the error but don't fail the operation
		// TODO: Consider adding proper logging here when logger is available
		return nil
	}

	return nil
}

// OnError is a no-op for OpeningHook.
func (h *OpeningHook) OnError(_ *hooks.HookContext) error {
	return nil
}
