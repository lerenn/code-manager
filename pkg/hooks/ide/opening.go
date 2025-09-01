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

	// Check if IDE name is provided in parameters
	ideName, hasIDEName := ctx.Parameters["ideName"]
	if !hasIDEName {
		return nil
	}

	ideNameStr, ok := ideName.(string)
	if !ok || ideNameStr == "" {
		return nil
	}

	// Get worktree path from parameters or calculated path
	worktreePath := h.calculateWorktreePath(ctx)
	if worktreePath == "" {
		return fmt.Errorf("cannot open IDE: worktree path is empty")
	}

	// Open the IDE directly using the hook's IDE manager
	cmInstance, ok := ctx.CM.(interface {
		IsVerbose() bool
	})
	if !ok {
		return fmt.Errorf("CM instance does not support verbose mode")
	}

	// Try to open the IDE
	if err := h.IDEManager.OpenIDE(ideNameStr, worktreePath, cmInstance.IsVerbose()); err != nil {
		// For OpenWorktree operation, IDE opening is required, so fail the operation
		if ctx.OperationName == consts.OpenWorktree {
			return err
		}

		// For other operations, IDE opening is optional, so just log the error
		if cmInstance.IsVerbose() {
			fmt.Printf("Warning: Failed to open IDE %s: %v\n", ideNameStr, err)
		}
		return nil
	}

	return nil
}

// extractWorktreePath extracts the worktree path from parameters.
func (h *OpeningHook) extractWorktreePath(params map[string]interface{}) string {
	if branch, hasBranch := params["branch"]; hasBranch {
		if branchStr, ok := branch.(string); ok && branchStr != "" {
			return branchStr
		}
	}
	if worktreeName, hasWorktreeName := params["worktreeName"]; hasWorktreeName {
		if worktreeNameStr, ok := worktreeName.(string); ok && worktreeNameStr != "" {
			return worktreeNameStr
		}
	}
	return ""
}

// calculateWorktreePath calculates the worktree path for OpenWorktree operation.
func (h *OpeningHook) calculateWorktreePath(ctx *hooks.HookContext) string {
	// For OpenWorktree operation, use the worktreePath from parameters
	if ctx.OperationName == consts.OpenWorktree {
		if worktreePath, hasWorktreePath := ctx.Parameters["worktreePath"]; hasWorktreePath {
			if worktreePathStr, ok := worktreePath.(string); ok && worktreePathStr != "" {
				return worktreePathStr
			}
		}
	}

	// For other operations, use the existing logic
	return h.extractWorktreePath(ctx.Parameters)
}

// OnError is a no-op for OpeningHook.
func (h *OpeningHook) OnError(_ *hooks.HookContext) error {
	return nil
}
