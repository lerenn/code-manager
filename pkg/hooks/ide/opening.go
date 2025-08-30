package ide

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/hooks"
)

// OpeningHook provides IDE opening functionality as a post-hook.
type OpeningHook struct{}

// NewOpeningHook creates a new OpeningHook instance and registers it for appropriate operations.
func NewOpeningHook() *OpeningHook {
	return &OpeningHook{}
}

// RegisterForOperations registers this hook for the operations that create worktrees.
func (h *OpeningHook) RegisterForOperations(cmInstance interface {
	RegisterHook(operation string, hook hooks.Hook) error
}) error {
	// Register as post-hook for operations that create worktrees
	if err := cmInstance.RegisterHook(consts.CreateWorkTree, h); err != nil {
		return err
	}

	if err := cmInstance.RegisterHook(consts.LoadWorktree, h); err != nil {
		return err
	}

	// Register as post-hook for operations that open worktrees
	if err := cmInstance.RegisterHook(consts.OpenWorktree, h); err != nil {
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

	// Get worktree path from parameters
	worktreePath := h.extractWorktreePath(ctx.Parameters)
	if worktreePath == "" {
		return fmt.Errorf("cannot open IDE: worktree path is empty")
	}

	// Store the IDE opening information in the context for the CM to handle
	ctx.Results["ideName"] = ideNameStr
	ctx.Results["worktreePath"] = worktreePath
	ctx.Results["shouldOpenIDE"] = true

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

// OnError is a no-op for OpeningHook.
func (h *OpeningHook) OnError(_ *hooks.HookContext) error {
	return nil
}
